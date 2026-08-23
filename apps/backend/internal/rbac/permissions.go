package rbac

// Permission identifies one discrete authorization capability. Handlers gate
// actions on permissions rather than roles directly, so the role→permission
// matrix in matrix.go is the only place capabilities are assigned to roles.
type Permission string

const (
	PermissionProjectRead   Permission = "project:read"
	PermissionProjectCreate Permission = "project:create"
	PermissionProjectUpdate Permission = "project:update"
	PermissionProjectDelete Permission = "project:delete"

	PermissionTeamRead   Permission = "team:read"
	PermissionTeamManage Permission = "team:manage"

	PermissionInvitationRead   Permission = "invitation:read"
	PermissionInvitationCreate Permission = "invitation:create"
	PermissionInvitationRevoke Permission = "invitation:revoke"

	PermissionProjectTeamRead   Permission = "project-team:read"
	PermissionProjectTeamManage Permission = "project-team:manage"

	PermissionClusterRead       Permission = "cluster:read"
	PermissionClusterManage     Permission = "cluster:manage"
	PermissionClusterValidate   Permission = "cluster:validate"
	PermissionClusterSetDefault Permission = "cluster:set-default"

	PermissionApplicationRead   Permission = "application:read"
	PermissionApplicationManage Permission = "application:manage"

	PermissionApplicationTeamRead   Permission = "application-team:read"
	PermissionApplicationTeamManage Permission = "application-team:manage"

	PermissionDeploymentRead     Permission = "deployment:read"
	PermissionDeploymentCreate   Permission = "deployment:create"
	PermissionDeploymentUpdate   Permission = "deployment:update"
	PermissionDeploymentCancel   Permission = "deployment:cancel"
	PermissionDeploymentRollback Permission = "deployment:rollback"
	PermissionDeploymentDelete   Permission = "deployment:delete"

	PermissionIncidentRead    Permission = "incident:read"
	PermissionIncidentManage  Permission = "incident:manage"
	PermissionIncidentComment Permission = "incident:comment"

	PermissionAlertRead        Permission = "alert:read"
	PermissionAlertManage      Permission = "alert:manage"
	PermissionAlertAcknowledge Permission = "alert:acknowledge"
	PermissionAlertResolve     Permission = "alert:resolve"
	PermissionAlertReopen      Permission = "alert:reopen"

	// PermissionKubernetesRead gates the existing (read-only) Kubernetes
	// resource endpoints. PermissionKubernetesOperate is defined for
	// forward-compatibility but is not currently wired to any endpoint: the
	// backend exposes no Kubernetes mutation capability today.
	PermissionKubernetesRead    Permission = "kubernetes:read"
	PermissionKubernetesOperate Permission = "kubernetes:operate"

	PermissionAuditRead Permission = "audit:read"

	PermissionOrganizationRead   Permission = "organization:read"
	PermissionOrganizationManage Permission = "organization:manage"

	PermissionNotificationRead   Permission = "notification:read"
	PermissionNotificationManage Permission = "notification:manage"

	PermissionSLORead   Permission = "slo:read"
	PermissionSLOManage Permission = "slo:manage"

	// PermissionUserManage is defined for forward-compatibility but is not
	// currently wired to any endpoint: the backend exposes no user
	// management capability beyond GET /users/me today.
	PermissionUserManage Permission = "user:manage"
)
