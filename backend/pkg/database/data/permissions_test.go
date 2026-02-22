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


func TestPermissionsContains(t *testing.T) {
	perms1 := Permissions{ReadUser, WriteFiles}
	perms2 := Permissions{ReadUser}

	if !perms1.Contains(perms2) {
		t.Errorf("Expected permissions %v to contain %v, but it did not", perms1, perms2)
	}

	perms3 := Permissions{DeleteUser}
	if perms1.Contains(perms3) {
		t.Errorf("Expected permissions %v not to contain %v, but it did", perms1, perms3)
	}
}