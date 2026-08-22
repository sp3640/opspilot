package rbac

import "testing"

func TestPlatformAdminHasEveryPermission(t *testing.T) {
	for _, permission := range allPermissions {
		if !HasPermission(string(RolePlatformAdmin), permission) {
			t.Fatalf("expected Platform Admin to hold permission %q", permission)
		}
	}
}

func TestDevOpsEngineerPermissions(t *testing.T) {
	role := string(RoleDevOpsEngineer)

	for _, permission := range []Permission{
		PermissionProjectRead, PermissionTeamRead, PermissionClusterRead, PermissionApplicationRead,
		PermissionDeploymentRead, PermissionIncidentRead, PermissionAlertRead, PermissionAuditRead, PermissionOrganizationRead,
		PermissionClusterManage, PermissionClusterValidate, PermissionClusterSetDefault,
		PermissionApplicationManage,
		PermissionApplicationTeamRead,
		PermissionDeploymentCreate, PermissionDeploymentUpdate, PermissionDeploymentCancel, PermissionDeploymentRollback, PermissionDeploymentDelete,
		PermissionIncidentManage, PermissionIncidentComment,
		PermissionAlertManage, PermissionAlertAcknowledge, PermissionAlertResolve, PermissionAlertReopen,
	} {
		if !HasPermission(role, permission) {
			t.Fatalf("expected DevOps Engineer to hold permission %q", permission)
		}
	}

	for _, permission := range []Permission{
		PermissionOrganizationManage,
		PermissionInvitationCreate, PermissionInvitationRead, PermissionInvitationRevoke,
		PermissionTeamManage,
		PermissionProjectCreate, PermissionProjectUpdate, PermissionProjectDelete,
		PermissionProjectTeamManage,
		PermissionApplicationTeamManage,
		PermissionUserManage,
	} {
		if HasPermission(role, permission) {
			t.Fatalf("expected DevOps Engineer NOT to hold permission %q", permission)
		}
	}
}

func TestDeveloperPermissions(t *testing.T) {
	role := string(RoleDeveloper)

	for _, permission := range []Permission{
		PermissionProjectRead, PermissionTeamRead, PermissionClusterRead, PermissionApplicationRead,
		PermissionDeploymentRead, PermissionAuditRead, PermissionOrganizationRead,
		PermissionApplicationTeamRead,
		PermissionIncidentRead, PermissionIncidentManage, PermissionIncidentComment,
		PermissionAlertRead, PermissionAlertManage, PermissionAlertAcknowledge, PermissionAlertResolve, PermissionAlertReopen,
	} {
		if !HasPermission(role, permission) {
			t.Fatalf("expected Developer to hold permission %q", permission)
		}
	}

	for _, permission := range []Permission{
		PermissionOrganizationManage,
		PermissionUserManage,
		PermissionTeamManage,
		PermissionProjectCreate, PermissionProjectUpdate, PermissionProjectDelete,
		PermissionClusterManage, PermissionClusterValidate, PermissionClusterSetDefault,
		PermissionApplicationManage,
		PermissionApplicationTeamManage,
		PermissionDeploymentCreate, PermissionDeploymentUpdate, PermissionDeploymentCancel, PermissionDeploymentRollback, PermissionDeploymentDelete,
		PermissionKubernetesOperate,
		PermissionInvitationCreate, PermissionInvitationRead, PermissionInvitationRevoke,
		PermissionProjectTeamManage,
	} {
		if HasPermission(role, permission) {
			t.Fatalf("expected Developer NOT to hold permission %q", permission)
		}
	}
}

func TestViewerPermissions(t *testing.T) {
	role := string(RoleViewer)

	for _, permission := range []Permission{
		PermissionProjectRead, PermissionTeamRead, PermissionClusterRead, PermissionApplicationRead,
		PermissionDeploymentRead, PermissionIncidentRead, PermissionAlertRead, PermissionAuditRead, PermissionOrganizationRead,
	} {
		if !HasPermission(role, permission) {
			t.Fatalf("expected Viewer to hold read permission %q", permission)
		}
	}

	for _, permission := range allPermissions {
		if permission == PermissionProjectRead || permission == PermissionTeamRead || permission == PermissionClusterRead ||
			permission == PermissionApplicationRead || permission == PermissionDeploymentRead || permission == PermissionIncidentRead ||
			permission == PermissionAlertRead || permission == PermissionAuditRead || permission == PermissionOrganizationRead ||
			permission == PermissionKubernetesRead || permission == PermissionProjectTeamRead || permission == PermissionApplicationTeamRead {
			continue
		}
		if HasPermission(role, permission) {
			t.Fatalf("expected Viewer NOT to hold permission %q", permission)
		}
	}
}

func TestHasPermissionNormalizesRole(t *testing.T) {
	if !HasPermission("  platform admin  ", PermissionOrganizationManage) {
		t.Fatalf("expected case/whitespace-insensitive role matching to succeed")
	}
}

func TestHasPermissionRejectsUnknownRole(t *testing.T) {
	if HasPermission("User", PermissionProjectRead) {
		t.Fatalf("expected legacy 'User' role to hold no permissions under the new matrix")
	}
	if HasPermission("", PermissionProjectRead) {
		t.Fatalf("expected empty role to hold no permissions")
	}
}

func TestIsValidRole(t *testing.T) {
	for _, role := range AllRoles {
		if !IsValidRole(string(role)) {
			t.Fatalf("expected %q to be a valid role", role)
		}
	}
	if IsValidRole("User") {
		t.Fatalf("expected legacy 'User' role to be invalid under the new model")
	}
	if IsValidRole("") {
		t.Fatalf("expected empty string to be invalid")
	}
}
