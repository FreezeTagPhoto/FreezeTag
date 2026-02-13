package data

import "slices"

type Permission string

type Permissions []Permission

var (
	ReadUser   Permission = registerPermission("read:user")
	DeleteUser Permission = registerPermission("delete:user")
	ReadFiles  Permission = registerPermission("read:files")
	WriteFiles Permission = registerPermission("write:files")

	WritePermissions Permission = registerPermission("write:permissions")
	ReadTags          Permission = registerPermission("read:tags")
	WriteTags        Permission = registerPermission("write:tags")
	CreateUser       Permission = registerPermission("create:user")
)

var allPermissions Permissions

func registerPermission(permission Permission) Permission {
	allPermissions = append(allPermissions, permission)
	return permission
}

func (p Permissions) HasPermission(permission Permission) bool {
	return slices.Contains(p, permission)
}

func All() Permissions {
	return allPermissions
}
