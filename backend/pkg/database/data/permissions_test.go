package data

import "testing"

func TestPermissionsHasPermission(t *testing.T) {
	perms := Permissions{ReadUser, WriteFiles}

	if !perms.HasPermission(ReadUser) {
		t.Errorf("Expected to have permission %s, but it was not found", ReadUser)
	}

	if !perms.HasPermission(WriteFiles) {
		t.Errorf("Expected to have permission %s, but it was not found", WriteFiles)
	}

	if perms.HasPermission(DeleteUser) {
		t.Errorf("Expected not to have permission %s, but it was found", DeleteUser)
	}
}

func TestGetPermissionFromSlug(t *testing.T) {
	perm, err := GetPermissionFromSlug("read:user")
	if !err {
		t.Errorf("Expected to find permission with slug 'read:user', but it was not found")
	}
	if perm != ReadUser {
		t.Errorf("Expected to get permission %s, but got %s", ReadUser, perm)
	}

	_, err = GetPermissionFromSlug("nonexistent:permission")
	if err {
		t.Errorf("Expected not to find permission with slug 'nonexistent:permission', but it was found")
	}
}
func TestPermissionsHasPermissionSlug(t *testing.T) {
	perms := Permissions{ReadUser, WriteFiles}

	if !perms.HasPermissionSlug("read:user") {
		t.Errorf("Expected to have permission with slug 'read:user', but it was not found")
	}

	if !perms.HasPermissionSlug("write:files") {
		t.Errorf("Expected to have permission with slug 'write:files', but it was not found")
	}

	if perms.HasPermissionSlug("delete:user") {
		t.Errorf("Expected not to have permission with slug 'delete:user', but it was found")
	}
}
