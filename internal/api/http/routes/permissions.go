package routes

import (
	"github.com/VEDA95/OpenBoard-API/internal/api/http/responses"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v2"
)

func PermissionsGET(context *fiber.Ctx) error {
	permissions, err := auth.GetPermissions()

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKCollectionResponse(fiber.StatusOK, permissions))
}

func PermissionsGETByID(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}

	if err := validators.Instance.Validate(validationParamData); err != nil {
		return util.CreateValidationError(err)
	}

	permission, err := auth.GetPermission(validationParamData.Id)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, permission))
}

func WorkspacePermissionsGET(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}

	if err := validators.Instance.Validate(validationParamData); err != nil {
		return util.CreateValidationError(err)
	}

	permissions, err := auth.GetWorkspacePermissions(validationParamData.Id)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKCollectionResponse(fiber.StatusOK, permissions))
}

func BoardPermissionsGET(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}

	if err := validators.Instance.Validate(validationParamData); err != nil {
		return util.CreateValidationError(err)
	}

	permissions, err := auth.GetBoardPermissions(validationParamData.Id)

	if err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKCollectionResponse(fiber.StatusOK, permissions))
}

func PermissionsPOST(context *fiber.Ctx) error {
	validationData := new(validators.PermissionValidator)

	if err := context.BodyParser(validationData); err != nil {
		return err
	}

	if err := validators.Instance.Validate(validationData); err != nil {
		return util.CreateValidationError(err)
	}

	createPermissionQuery := db.Instance.From("open_board_role_permission").Prepared(true).
		Insert().
		Rows(goqu.Record{"path": validationData.Path}).
		Executor()

	if _, err := createPermissionQuery.Exec(); err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusCreated, responses.OKResponse(
		fiber.StatusCreated,
		auth.Permission{Path: validationData.Path},
	))
}

func PermissionsPUT(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}
	validationData := new(validators.PermissionUpdateValidator)

	if err := context.BodyParser(&validationData); err != nil {
		return err
	}

	if err := validators.Instance.Validate(validationParamData); err != nil {
		return util.CreateValidationError(err)
	}

	if err := validators.Instance.Validate(validationData); err != nil {
		return util.CreateValidationError(err)
	}

	permission, err := auth.GetPermission(validationParamData.Id)

	if err != nil {
		return err
	}

	updatePayload := goqu.Record{}

	if validationData.Path != "" && (validationData.Path != permission.Path) {
		updatePayload["path"] = validationData.Path
	}

	if len(updatePayload) == 0 {
		return util.JSONResponse(context, fiber.StatusOK, permission)
	}

	updatePermissionQuery := db.Instance.From("open_board_role_permission").Prepared(true).
		Update().
		Where(goqu.Ex{"id": validationParamData.Id}).
		Set(updatePayload).
		Executor()

	if _, err := updatePermissionQuery.Exec(); err != nil {
		return err
	}

	permission.Path = validationData.Path

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, permission))
}

func PermissionsDELETE(context *fiber.Ctx) error {
	validationParamData := validators.RoleIDValidator{Id: context.Params("id")}

	if err := validators.Instance.Validate(validationParamData); err != nil {
		return util.CreateValidationError(err)
	}

	var permissionId string

	_, err := db.Instance.From("open_board_role_permission").Prepared(true).
		Select("id").
		Where(goqu.Ex{"id": validationParamData.Id}).
		ScanVal(&permissionId)

	if err != nil {
		return err
	}

	if permissionId == "" {
		return util.JSONResponse(context, fiber.StatusNotFound, responses.GenericMessage{Message: "permission not found"})
	}

	transaction, err := db.Instance.Begin()

	if err != nil {
		return err
	}

	deletePermissionQuery := transaction.From("open_board_role_permission").Prepared(true).
		Delete().
		Where(goqu.Ex{"id": validationParamData.Id}).
		Executor()
	deleteRolePermissionsQuery := transaction.From("open_board_role_permissions").Prepared(true).
		Delete().
		Where(goqu.Ex{"permission_id": validationParamData.Id}).
		Executor()
	deleteWorkspacePermissionsQuery := transaction.From("open_board_workspace_permissions").Prepared(true).
		Delete().
		Where(goqu.Ex{"permission_id": validationParamData.Id}).
		Executor()
	deleteBoardPermissionsQuery := transaction.From("open_board_board_permissions").Prepared(true).
		Delete().
		Where(goqu.Ex{"permission_id": validationParamData.Id}).
		Executor()

	if _, err := deleteBoardPermissionsQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if _, err := deleteWorkspacePermissionsQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if _, err := deleteRolePermissionsQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if _, err := deletePermissionQuery.Exec(); err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}

		return err
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	return util.JSONResponse(context, fiber.StatusOK, responses.OKResponse(fiber.StatusOK, fiber.Map{}))
}
