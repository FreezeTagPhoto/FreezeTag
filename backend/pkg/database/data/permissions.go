package data

import "slices"

type Permission struct {
	Slug        string `json:"permission"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Permissions []Permission

var (
	ReadFiles = register("read:files",
		"Read Files",
		"Allows downloading/viewing stored files.")
	WriteFiles = register("write:files",
		"Write Files",
		"Allows uploading files.")
	ReadTags = register("read:tags",
		"Read Tags",
		"Allows viewing metadata tags.")
	WriteTags = register("write:tags",
		"Write Tags",
		"Allows creating and editing metadata tags.")

	WriteUser = register("write:user",
		"Write User",
		"Allows updating existing user information. (e.g passwords).")
	ReadUser = register("read:user",
		"Read User",
		"Allows viewing user profiles.")

	ReadPlugins = register("read:plugins",
		"Read Plugins",
		"Allows viewing the plugins enabled on the system.")
	WritePlugins = register("write:plugins",
		"Write Plugins",
		"Allows changing plugins' enabled status, as well as downloading new plugins.")

	WriteToken = register("write:tokens",
		"Write API Tokens",
		"Allows creating and revoking API tokens.")
	ReadToken = register("read:tokens",
		"Read API Tokens",
		"Allows viewing existing API token information.")

	WriteJobs = register("write:jobs",
		"Write/Cancel Jobs",
		"Allows modifying or cancelling running jobs (uploads, plugins, etc...)")

	// More powerful permissions for admin users
	WriteAnyToken    = register("write:any_token", "Write Any API Token", "Allows revoking and deleting any API token, including those not owned by the user.")
	ReadAnyToken     = register("read:any_token", "Read Any API Token", "Allows viewing any API token, including those not owned by the user.")
	ReadPermissions  = register("read:permissions", "Read Permissions", "Administrative access to read system roles and rights.")
	WritePermissions = register("write:permissions", "Write Permissions", "Administrative access to modify system roles and rights.")
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
