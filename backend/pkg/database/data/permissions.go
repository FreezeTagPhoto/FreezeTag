package data

import "slices"

type Permission struct {
	Slug        string `json:"permission"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Permissions []Permission

var (
	ReadFiles  Permission = register("read:files", "Read Files", "Allows downloading/viewing stored files.")
	WriteFiles Permission = register("write:files", "Write Files", "Allows uploading files.")
	ReadTags   Permission = register("read:tags", "Read Tags", "Allows viewing metadata tags.")
	WriteTags  Permission = register("write:tags", "Write Tags", "Allows creating and editing metadata tags.")

	WriteUser Permission = register("write:user", "Write User", "Allows updating existing user information. (e.g passwords).")
	ReadUser  Permission = register("read:user", "Read User", "Allows viewing user profiles.")

	ReadPlugins  Permission = register("read:plugins", "Read Plugins", "Allows viewing the plugins enabled on the system.")
	WritePlugins Permission = register("write:plugins", "Write Plugins", "Allows changing plugins' enabled status, as well as downloading new plugins.")

	WriteToken Permission = register("write:tokens", "Write API Tokens", "Allows creating and revoking API tokens.")
	ReadToken  Permission = register("read:tokens", "Read API Tokens", "Allows viewing existing API token information.")

	// More powerful permissions for admin users
	WriteAnyToken    Permission = register("write:any_token", "Write Any API Token", "Allows revoking and deleting any API token, including those not owned by the user.")
	ReadAnyToken     Permission = register("read:any_token", "Read Any API Token", "Allows viewing any API token, including those not owned by the user.")
	ReadPermissions  Permission = register("read:permissions", "Read Permissions", "Administrative access to read system roles and rights.")
	WritePermissions Permission = register("write:permissions", "Write Permissions", "Administrative access to modify system roles and rights.")
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

func (p Permissions) HasPermissionSlug(slug string) bool {
	for _, perm := range p {
		if perm.Slug == slug {
			return true
		}
	}
	return false
}

func GetPermissionFromSlug(slug string) (Permission, bool) {
	for _, perm := range allPermissions {
		if perm.Slug == slug {
			return perm, true
		}
	}
	return Permission{}, false
}

func (p Permissions) Contains(permissions Permissions) bool {
	for _, perm := range permissions {
		if !p.HasPermission(perm) {
			return false
		}
	}
	return true
}

func AllPermissions() Permissions {
	return allPermissions
}
