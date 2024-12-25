package routes

import (
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exec"
	"github.com/gofiber/fiber/v2"
	"slices"
)

func RolesGET(context *fiber.Ctx) error {
	roles, err := auth.GetRoles()

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, roles)
}

func RolesGETByID(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}

	if errs := validators.Instance.Validate(validationParamData); errs != nil {
		return util.CreateValidationError(errs)
	}

	role, err := auth.GetRole(validationParamData.Id)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, role)
}

func MeRolesGET(context *fiber.Ctx) error {
	session := context.Locals("session").(*auth.AuthSession)
	roles, err := auth.GetUserRoles(session.UserId)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, roles)
}

func MeRolesGETByID(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}

	if errs := validators.Instance.Validate(validationParamData); errs != nil {
		return util.CreateValidationError(errs)
	}

	session := context.Locals("session").(*auth.AuthSession)
	roles, err := auth.GetUserRoles(session.UserId)

	if err != nil {
		return err
	}

	role := new(auth.Role)

	for _, authRole := range roles {
		if authRole.Id == validationParamData.Id {
			role = &authRole
			break
		}
	}

	if role == nil {
		return util.JSONResponse(context, fiber.StatusNotFound, responses.GenericMessage{Message: "role not found"})
	}

	return util.JSONResponse(context, fiber.StatusOK, role)
}
func UserRolesGET(context *fiber.Ctx) error {
	validationParamData := validators.UserIdParamValidator{Id: context.Params("id")}

	if errs := validators.Instance.Validate(validationParamData); errs != nil {
		return util.CreateValidationError(errs)
	}

	roles, err := auth.GetUserRoles(validationParamData.Id)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, roles)
}

func UserRolesGETByID(context *fiber.Ctx) error {
	validationParamData := validators.UserIdParamValidator{
		Id:     context.Params("id"),
		RoleId: context.Params("role_id"),
	}

	if errs := validators.Instance.Validate(&validationParamData); errs != nil {
		return util.CreateValidationError(errs)
	}

	roles, err := auth.GetUserRoles(validationParamData.Id)

	if err != nil {
		return err
	}

	role := new(auth.Role)

	for _, authRole := range roles {
		if authRole.Id == validationParamData.RoleId {
			role = &authRole
			break
		}
	}

	if role == nil {
		return util.JSONResponse(context, fiber.StatusNotFound, responses.GenericMessage{Message: "role not found"})
	}

	return util.JSONResponse(context, fiber.StatusOK, role)
}

func RolesPOST(context *fiber.Ctx) error {
	validationData := new(validators.RoleValidator)

	if err := context.BodyParser(&validationData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(&validationData); errs != nil {
		return util.CreateValidationError(errs)
	}

	transaction, err := db.Instance.Begin()

	if err != nil {
		return err
	}

	role := new(auth.Role)
	roleQuery := transaction.From("open_board_role").Prepared(true).
		Insert().
		Returning("*").
		Rows(goqu.Record{"name": validationData.Name}).
		Executor()

	_, dbErr := roleQuery.ScanStruct(&role)

	if dbErr != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return dbErr
	}

	rolePermissionRecords := make([]goqu.Record, len(validationData.Permissions))

	for _, permission := range validationData.Permissions {
		rolePermissionRecords = append(rolePermissionRecords, goqu.Record{
			"permission_id": permission,
			"role_id":       role.Id,
		})
	}

	rolePermissionsQuery := transaction.From("open_board_role_permissions").Prepared(true).
		Insert().
		Rows(rolePermissionRecords).
		Executor()

	if _, err := roleQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if _, err := rolePermissionsQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	dbErr2 := db.Instance.From("open_board_role_permission").Prepared(true).
		Select("*").
		Where(goqu.Ex{"permission_id": validationData.Permissions}).
		ScanStructs(&role.Permissions)

	if dbErr2 != nil {
		return dbErr2
	}

	return util.JSONResponse(context, fiber.StatusOK, role)
}

func RolesPUT(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}
	validationData := new(validators.RoleUpdateValidator)

	if err := context.BodyParser(&validationData); err != nil {
		return err
	}

	if errs := validators.Instance.Validate(&validationParamData); errs != nil {
		return util.CreateValidationError(errs)
	}

	if errs := validators.Instance.Validate(&validationData); errs != nil {
		return util.CreateValidationError(errs)
	}

	role, err := auth.GetRole(validationParamData.Id)

	if err != nil {
		return err
	}

	roleUpdatePayload := goqu.Record{}
	permissionIds := make([]string, len(role.Permissions))
	permissionsToAdd := make([]string, 0)
	permissionsToRemove := make([]string, 0)

	for _, permission := range role.Permissions {
		permissionIds = append(permissionIds, permission.Id)
	}

	if validationData.Name != role.Name {
		roleUpdatePayload["name"] = validationData.Name
	}

	if slices.Compare(validationData.Permissions, permissionIds) != 0 {
		for _, permission := range validationData.Permissions {
			if !slices.Contains(permissionIds, permission) {
				permissionsToAdd = append(permissionsToAdd, permission)
			}
		}

		for _, permissionId := range permissionIds {
			if !slices.Contains(validationData.Permissions, permissionId) {
				permissionsToRemove = append(permissionsToRemove, permissionId)
			}
		}
	}

	if len(roleUpdatePayload) == 0 && len(permissionsToAdd) == 0 && len(permissionsToRemove) == 0 {
		return util.JSONResponse(context, fiber.StatusOK, role)
	}

	transaction, err := db.Instance.Begin()

	if err != nil {
		return err
	}

	var permissionsToAddQuery exec.QueryExecutor
	var permissionsToRemoveQuery exec.QueryExecutor

	roleUpdateQuery := transaction.From("open_board_role").Prepared(true).
		Update().
		Where(goqu.Ex{"id": validationParamData.Id}).
		Set(goqu.Record{"name": validationData.Name}).
		Executor()

	if len(permissionsToAdd) > 0 {
		permissionsToAddPayload := make([]goqu.Record, len(permissionsToAdd))

		for _, permission := range permissionsToAdd {
			permissionsToAddPayload = append(permissionsToAddPayload, goqu.Record{"role_id": role.Id, "permission_id": permission})
		}

		permissionsToAddQuery = transaction.From("open_board_role_permissions").Prepared(true).
			Insert().
			Rows(permissionsToAddPayload).
			Executor()
	}

	if len(permissionsToRemove) > 0 {
		permissionsToRemoveQuery = transaction.From("open_board_role_permissions").Prepared(true).
			Delete().
			Where(goqu.Ex{"id": validationParamData.Id, "role_id": role.Id}).
			Executor()
	}

	if _, err := roleUpdateQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if &permissionsToAddQuery != nil {
		if _, err := permissionsToAddQuery.Exec(); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}

			return err
		}
	}

	if &permissionsToRemoveQuery != nil {
		if _, err := permissionsToRemoveQuery.Exec(); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}

			return err
		}
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	updatedRole, err := auth.GetRole(validationParamData.Id)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, updatedRole)
}

func RolesDELETE(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}

	if err := validators.Instance.Validate(&validationParamData); err != nil {
		return util.CreateValidationError(err)
	}

	_, err := auth.GetRole(validationParamData.Id)

	if err != nil {
		return err
	}

	transaction, err := db.Instance.Begin()

	if err != nil {
		return err
	}

	deleteRolePermissionsQuery := transaction.From("open_board_role_permissions").Prepared(true).
		Delete().
		Where(goqu.Ex{"role_id": validationParamData.Id}).
		Executor()
	deleteUserRolesQuery := transaction.From("open_board_user_roles").Prepared(true).
		Delete().
		Where(goqu.Ex{"role_id": validationParamData.Id}).
		Executor()
	deleteRoleQuery := transaction.From("open_board_role").Prepared(true).
		Delete().
		Where(goqu.Ex{"id": validationParamData.Id}).
		Executor()

	if _, err := deleteRolePermissionsQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if _, err := deleteUserRolesQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if _, err := deleteRoleQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.GenericMessage{Message: fmt.Sprintf("Role: %s has deleted successfully", validationParamData.Id)})
}

func UserRolesPUT(context *fiber.Ctx) error {
	validationParamData := validators.UserIdParamValidator{Id: context.Params("id")}
	validationData := new(validators.UserRolesUpdateValidator)

	if err := validators.Instance.Validate(&validationParamData); err != nil {
		return util.CreateValidationError(err)
	}

	if err := validators.Instance.Validate(&validationData); err != nil {
		return util.CreateValidationError(err)
	}

	roles, err := auth.GetUserRoles(validationParamData.Id)

	if err != nil {
		return err
	}

	roleIds := make([]string, len(roles))

	for _, role := range roles {
		roleIds = append(roleIds, role.Id)
	}

	if slices.Compare(validationData.RoleIds, roleIds) == 0 {
		return util.JSONResponse(context, fiber.StatusOK, responses.GenericMessage{Message: "Roles were not updated"})
	}

	rolesToAdd := make([]string, 0)
	rolesToRemove := make([]string, 0)

	for _, roleId := range validationData.RoleIds {
		if !slices.Contains(roleIds, roleId) {
			rolesToAdd = append(rolesToAdd, roleId)
		}
	}

	for _, roleId := range roleIds {
		if !slices.Contains(validationData.RoleIds, roleId) {
			rolesToRemove = append(rolesToRemove, roleId)
		}
	}

	var permissionsToAddQuery exec.QueryExecutor
	var permissionsToRemoveQuery exec.QueryExecutor

	transaction, err := db.Instance.Begin()

	if err != nil {
		return err
	}

	if len(rolesToAdd) > 0 {
		rolesAddPayload := make([]goqu.Record, len(rolesToAdd))

		for _, role := range rolesToAdd {
			rolesAddPayload = append(rolesAddPayload, goqu.Record{"user_id": validationParamData.Id, "role_id": role})
		}

		permissionsToAddQuery = transaction.From("open_board_user_roles").Prepared(true).
			Insert().
			Rows(rolesAddPayload).
			Executor()
	}

	if len(rolesToRemove) > 0 {
		permissionsToRemoveQuery = transaction.From("open_board_user_roles").Prepared(true).
			Delete().
			Where(goqu.Ex{"user_id": validationParamData.Id, "role_id": rolesToRemove}).
			Executor()
	}

	if &permissionsToAddQuery != nil {
		if _, err := permissionsToAddQuery.Exec(); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}

			return err
		}
	}

	if &permissionsToRemoveQuery != nil {
		if _, err := permissionsToRemoveQuery.Exec(); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}

			return err
		}
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	updatedRoles, err := auth.GetUserRoles(validationParamData.Id)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, updatedRoles)
}
