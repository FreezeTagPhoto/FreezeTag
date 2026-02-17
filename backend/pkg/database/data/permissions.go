package data

import "slices"

type Permission struct {
	Slug        string `json:"permission"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Permissions []Permission

var (
	ReadUser         Permission = register("read:user", "Read User", "Allows viewing user profiles.")
	DeleteUser       Permission = register("delete:user", "Delete User", "Allows deleting user accounts.")
	ReadFiles        Permission = register("read:files", "Read Files", "Allows downloading/viewing stored files.")
	WriteFiles       Permission = register("write:files", "Write Files", "Allows uploading files.")
	WritePermissions Permission = register("write:permissions", "Write Permissions", "Administrative access to modify system roles and rights.")
	ReadTags         Permission = register("read:tags", "Read Tags", "Allows viewing metadata tags.")
	WriteTags        Permission = register("write:tags", "Write Tags", "Allows creating and editing metadata tags.")
	CreateUser       Permission = register("create:user", "Create User", "Allows registering new users.")
	WriteUser        Permission = register("write:user", "Write User", "Allows updating existing user information. (e.g passwords).")
)

var allPermissions Permissions

func register(slug string, name string, description string) Permission {
	permission := Permission{Slug: slug, Name: name, Description: description}
	allPermissions = append(allPermissions, permission)
	return permission
}

func (p Permissions) HasPermission(permission Permission) bool {
	return slices.Contains(p, permission)
}

func (p Permissions) HasPermissionString(permission string) bool {
	for _, perm := range p {
		if perm.Slug == permission {
			return true
		}
	}
	return false
}

func GetPermissionFromSlug(permission string) (Permission, bool) {
	for _, perm := range allPermissions {
		if perm.Slug == permission {
			return perm, true
		}
	}
	return Permission{}, false
}

func All() Permissions {
	return allPermissions
}
