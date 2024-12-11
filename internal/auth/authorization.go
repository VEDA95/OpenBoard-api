package auth

import (
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/doug-martin/goqu/v9"
	"slices"
)

type Role struct {
	Id          string       `json:"id" db:"id,omitempty"`
	Name        string       `json:"name" db:"name,omitempty"`
	UserId      string       `json:"-" db:"user_id"`
	Permissions []Permission `json:"permissions" db:"-"`
}

type Permission struct {
	Id     string `json:"id" db:"id,omitempty"`
	Path   string `json:"path" db:"path,omitempty"`
	RoleId string `json:"-" db:"role_id,omitempty"`
}

func IsUserAuthorized(userId string, permissions ...string) (bool, error) {
	if len(permissions) == 0 {
		return true, nil
	}

	roles := make([]string, 0)
	err := db.Instance.From("open_board_user_roles").Prepared(true).
		Select("role_id").
		Where(goqu.Ex{"user_id": userId}).
		ScanVals(&roles)

	if err != nil {
		return false, err
	}

	if len(roles) == 0 && len(permissions) > 0 {
		return false, nil
	}

	rolePermissions := make([]string, 0)
	err2 := db.Instance.From("open_board_role_permissions").Prepared(true).
		Select("permission.path").
		Where(goqu.Ex{"role_id": roles}).
		Join(goqu.T("open_board_role_permission"), goqu.On(goqu.Ex{
			"open_board_role_permissions.permission_id": "open_board_role_permission.id",
		})).
		As("permission").
		ScanVals(&rolePermissions)

	if err2 != nil {
		return false, err2
	}

	if len(rolePermissions) == 0 && len(permissions) > 0 {
		return false, nil
	}

	for _, permission := range permissions {
		if !slices.Contains(rolePermissions, permission) {
			return false, nil
		}
	}

	return true, nil
}

func GetRoles() ([]Role, error) {
	roles := make([]Role, 0)
	err := db.Instance.From("open_board_roles").Prepared(true).
		Select("*").
		ScanStructs(&roles)

	if err != nil {
		return nil, err
	}

	roleIds := make([]string, len(roles))

	for _, role := range roles {
		if !slices.Contains(roleIds, role.Id) {
			roleIds = append(roleIds, role.Id)
		}
	}

	permissions := make([]Permission, 0)
	err2 := db.Instance.From("open_board_role_permissions").Prepared(true).
		Select("role_id", goqu.C("permission.id").As("id"), goqu.C("permission.path").As("path")).
		Where(goqu.Ex{"role_id": roleIds}).
		Join(goqu.T("open_board_role_permission"), goqu.On(goqu.Ex{
			"open_board_role_permissions.permission_id": "open_board_role_permission.id",
		})).
		As("permission").
		ScanStructs(&permissions)

	if err2 != nil {
		return nil, err2
	}

	for _, permission := range permissions {
		for index := range roles {
			if roles[index].Id == permission.RoleId {
				roles[index].Permissions = append(roles[index].Permissions, permission)
				break
			}
		}
	}

	return roles, nil
}

func GetRole(roleId string) (*Role, error) {
	var role Role
	exists, err := db.Instance.From("open_board_roles").Prepared(true).
		Select("*").
		Where(goqu.Ex{"id": roleId}).
		ScanStruct(&role)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	permissions := make([]Permission, 0)
	err2 := db.Instance.From("open_board_role_permissions").Prepared(true).
		Select(goqu.C("permission.id").As("id"), goqu.C("permission.path").As("path")).
		Where(goqu.Ex{"role_id": roleId}).
		Join(goqu.T("open_board_role_permission"), goqu.On(goqu.Ex{
			"open_board_role_permissions.permission_id": "open_board_role_permission.id",
		})).
		As("permission").
		ScanStructs(&permissions)

	if err2 != nil {
		return nil, err2
	}

	for _, permission := range permissions {
		role.Permissions = append(role.Permissions, permission)
	}

	return &role, nil
}

func GetUserRoles(userId string) ([]Role, error) {
	roles := make([]Role, 0)
	err := db.Instance.From("open_board_user_roles").Prepared(true).
		Select(goqu.C("role.id").As("id"), goqu.C("role.name").As("name")).
		Where(goqu.Ex{"user_id": userId}).
		Join(goqu.T("open_board_role"), goqu.On(goqu.Ex{
			"open_board_user_roles.role_id": "open_board_role.id",
		})).
		As("role").
		ScanStructs(&roles)

	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return roles, nil
	}

	roleIds := make([]string, len(roles))

	for _, role := range roles {
		if !slices.Contains(roleIds, role.Id) {
			roleIds = append(roleIds, role.Id)
		}
	}

	rolePermissions := make([]Permission, 0)
	err2 := db.Instance.From("open_board_role_permissions").Prepared(true).
		Select("role_id", goqu.C("permission.id").As("id"), goqu.C("permission.path").As("path")).
		Where(goqu.Ex{"role_id": roleIds}).
		Join(goqu.T("open_board_role_permission"), goqu.On(goqu.Ex{
			"open_board_role_permissions.permission_id": "open_board_role_permission.id",
		})).
		As("permission").
		ScanStructs(&rolePermissions)

	if err2 != nil {
		return nil, err2
	}

	for _, permission := range rolePermissions {
		for index := range roles {
			if permission.RoleId == roles[index].Id {
				roles[index].Permissions = append(roles[index].Permissions, permission)
			}
		}
	}

	return roles, nil
}

func GetAllUserRoles() ([]Role, error) {
	roles := make([]Role, 0)
	err := db.Instance.From("open_board_user_roles").Prepared(true).
		Select("user_id", goqu.C("role.id").As("id"), goqu.C("role.name").As("name")).
		Join(goqu.T("open_board_role"), goqu.On(goqu.Ex{
			"open_board_user_roles.role_id": "open_board_role.id",
		})).
		As("role").
		ScanStructs(&roles)

	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return roles, nil
	}

	roleIds := make([]string, len(roles))

	for _, role := range roles {
		if !slices.Contains(roleIds, role.Id) {
			roleIds = append(roleIds, role.Id)
		}
	}

	rolePermissions := make([]Permission, 0)
	err2 := db.Instance.From("open_board_role_permissions").Prepared(true).
		Select("role_id", goqu.C("permission.id").As("id"), goqu.C("permission.path").As("path")).
		Where(goqu.Ex{"role_id": roleIds}).
		Join(goqu.T("open_board_role_permission"), goqu.On(goqu.Ex{
			"open_board_role_permissions.permission_id": "open_board_role_permission.id",
		})).
		As("permission").
		ScanStructs(&rolePermissions)

	if err2 != nil {
		return nil, err2
	}

	for _, permission := range rolePermissions {
		for index := range roles {
			if permission.RoleId == roles[index].Id {
				roles[index].Permissions = append(roles[index].Permissions, permission)
			}
		}
	}

	return roles, nil
}

func GetBoardPermissions(boardId string) ([]Permission, error) {
	permissions := make([]Permission, 0)
	err := db.Instance.From("open_board_board_permissions").Prepared(true).
		Select(goqu.C("permission.id").As("id"), goqu.C("permission.path").As("path")).
		Where(goqu.Ex{"board_id": boardId}).
		ScanStructs(&permissions)

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func GetWorkspacePermissions(workspaceId string) ([]Permission, error) {
	permissions := make([]Permission, 0)
	err := db.Instance.From("open_board_workspace_permissions").Prepared(true).
		Select(goqu.C("permission.id").As("id"), goqu.C("permission.path").As("path")).
		Where(goqu.Ex{"workspace_id": workspaceId}).
		ScanStructs(&permissions)

	if err != nil {
		return nil, err
	}

	return permissions, nil
}
