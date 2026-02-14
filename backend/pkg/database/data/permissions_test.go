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