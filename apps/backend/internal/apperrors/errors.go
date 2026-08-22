package apperrors

import "errors"

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")

	// User errors
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")

	// Project errors
	ErrProjectNotFound            = errors.New("project not found")
	ErrProjectForbidden           = errors.New("forbidden")
	ErrProjectAlreadyExists       = errors.New("project already exists")
	ErrProjectTeamAlreadyAssigned = errors.New("team already assigned to project")
	ErrInvalidProjectName         = errors.New("project name must be between 3 and 100 characters")
	ErrInvalidProjectDescription  = errors.New("project description must be 300 characters or fewer")

	// Organization errors
	ErrOrganizationNotFound           = errors.New("organization not found")
	ErrOrganizationForbidden          = errors.New("forbidden")
	ErrOrganizationAlreadyExists      = errors.New("organization already exists")
	ErrInvalidOrganizationName        = errors.New("organization name must be between 3 and 100 characters")
	ErrInvalidOrganizationDescription = errors.New("organization description must be 500 characters or fewer")
	ErrInvalidOrganizationSlug        = errors.New("organization slug must be between 3 and 120 characters")

	// Team errors
	ErrTeamNotFound            = errors.New("team not found")
	ErrTeamForbidden           = errors.New("forbidden")
	ErrTeamAlreadyExists       = errors.New("team already exists")
	ErrInvalidTeamName         = errors.New("team name must be between 3 and 100 characters")
	ErrInvalidTeamDescription  = errors.New("team description must be 500 characters or fewer")
	ErrTeamMemberAlreadyExists = errors.New("team member already exists")
	ErrTeamMemberNotFound      = errors.New("team member not found")

	// Invitation errors
	ErrInvitationNotFound      = errors.New("invitation not found")
	ErrInvitationForbidden     = errors.New("forbidden")
	ErrInvitationAlreadyExists = errors.New("pending invitation already exists")
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrInvitationNotPending    = errors.New("invitation is not pending")
	ErrInvitationEmailMismatch = errors.New("invitation email does not match current user")
	ErrInvalidInvitationEmail  = errors.New("invalid invitation email")
	ErrInvalidInvitationRole   = errors.New("invalid invitation role")
	ErrInvalidInvitationToken  = errors.New("invalid invitation token")

	// Incident errors
	ErrIncidentNotFound = errors.New("incident not found")
	ErrInvalidSeverity  = errors.New("invalid severity")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrInvalidProject   = errors.New("invalid project")

	// Alert errors
	ErrAlertNotFound            = errors.New("alert not found")
	ErrInvalidAlertSeverity     = errors.New("invalid alert severity")
	ErrInvalidAlertStatus       = errors.New("invalid alert status")
	ErrInvalidAlertSource       = errors.New("invalid alert source")
	ErrInvalidAlertResourceType = errors.New("invalid alert resource type")

	// Metric errors
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrInvalidTimeRange  = errors.New("end time must be greater than or equal to start time")

	// Cluster errors
	ErrClusterNotFound              = errors.New("cluster not found")
	ErrInvalidClusterProvider       = errors.New("invalid cluster provider")
	ErrInvalidClusterStatus         = errors.New("invalid cluster status")
	ErrInvalidClusterConnectionType = errors.New("invalid cluster connection type")
	ErrClusterCredentialRequired    = errors.New("cluster credential is required")
	ErrClusterCredentialCorrupted   = errors.New("stored cluster credential could not be decrypted")

	// Application errors
	ErrApplicationNotFound            = errors.New("application not found")
	ErrApplicationForbidden           = errors.New("forbidden")
	ErrApplicationAlreadyExists       = errors.New("application already exists")
	ErrApplicationTeamAlreadyAssigned = errors.New("team already assigned to application")
	ErrInvalidApplicationName         = errors.New("application name is required")
	ErrInvalidApplication             = errors.New("invalid application")
	ErrInvalidApplicationSlug         = errors.New("invalid application slug")
	ErrInvalidApplicationRuntime      = errors.New("invalid application runtime")
	ErrInvalidApplicationStatus       = errors.New("invalid application status")
	ErrInvalidApplicationPort         = errors.New("invalid application port")
	ErrInvalidApplicationEnvironment  = errors.New("invalid application environment")

	// Deployment errors
	ErrDeploymentNotFound                  = errors.New("deployment not found")
	ErrDeploymentForbidden                 = errors.New("forbidden")
	ErrInvalidDeploymentImage              = errors.New("deployment image is required")
	ErrInvalidDeploymentReplica            = errors.New("deployment replica count must be greater than zero")
	ErrInvalidDeploymentStrategy           = errors.New("invalid deployment strategy")
	ErrInvalidDeploymentEnvironment        = errors.New("invalid deployment environment")
	ErrInvalidDeploymentNamespace          = errors.New("invalid deployment namespace")
	ErrInvalidDeploymentStatus             = errors.New("invalid deployment status")
	ErrDeploymentHistoryNotFound           = errors.New("deployment history not found")
	ErrInvalidDeploymentRevision           = errors.New("invalid deployment revision")
	ErrDeploymentRollbackLatest            = errors.New("cannot rollback to latest revision")
	ErrDeploymentClusterUnreachable        = errors.New("cluster unreachable")
	ErrDeploymentInvalidKubeconfig         = errors.New("invalid kubeconfig")
	ErrDeploymentNamespaceNotFound         = errors.New("deployment namespace missing")
	ErrDeploymentExecutionFailed           = errors.New("deployment create failed")
	ErrDeploymentExecutionTimeout          = errors.New("deployment execution timeout")
	ErrDeploymentImageInvalid              = errors.New("image invalid")
	ErrDeploymentExecutionPermissionDenied = errors.New("permission denied")

	// Resource errors
	ErrResourceNotFound      = errors.New("resource not found")
	ErrInvalidResourceKind   = errors.New("invalid resource kind")
	ErrInvalidResourceStatus = errors.New("invalid resource status")
	ErrInvalidResourceHealth = errors.New("invalid resource health")

	// Pod errors
	ErrPodNotFound           = errors.New("pod not found")
	ErrPodForbidden          = errors.New("forbidden")
	ErrPodInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrPodClusterUnreachable = errors.New("cluster unreachable")

	// Kubernetes service errors
	ErrServiceNotFound           = errors.New("service not found")
	ErrServiceNamespaceNotFound  = errors.New("namespace not found")
	ErrServiceForbidden          = errors.New("forbidden")
	ErrServiceInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrServiceClusterUnavailable = errors.New("cluster unavailable")
	ErrServiceTimeout            = errors.New("kubernetes timeout")

	// Kubernetes ingress errors
	ErrIngressNotFound           = errors.New("ingress not found")
	ErrIngressNamespaceNotFound  = errors.New("namespace not found")
	ErrIngressForbidden          = errors.New("forbidden")
	ErrIngressInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrIngressClusterUnavailable = errors.New("cluster unavailable")
	ErrIngressTimeout            = errors.New("kubernetes timeout")

	// Kubernetes event errors
	ErrEventNotFound           = errors.New("event not found")
	ErrEventNamespaceNotFound  = errors.New("namespace not found")
	ErrEventForbidden          = errors.New("forbidden")
	ErrEventInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrEventClusterUnavailable = errors.New("cluster unavailable")
	ErrEventTimeout            = errors.New("kubernetes timeout")

	// Kubernetes pod logs errors
	ErrLogPodNotFound        = errors.New("pod not found")
	ErrLogContainerNotFound  = errors.New("container not found")
	ErrLogNamespaceNotFound  = errors.New("namespace not found")
	ErrLogForbidden          = errors.New("forbidden")
	ErrLogInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrLogClusterUnavailable = errors.New("cluster unavailable")
	ErrLogTimeout            = errors.New("kubernetes timeout")
	ErrLogPreviousNotFound   = errors.New("no previous container instance found")

	// Kubernetes configmap errors
	ErrConfigMapNotFound           = errors.New("configmap not found")
	ErrConfigMapNamespaceNotFound  = errors.New("namespace not found")
	ErrConfigMapForbidden          = errors.New("forbidden")
	ErrConfigMapInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrConfigMapClusterUnavailable = errors.New("cluster unavailable")
	ErrConfigMapTimeout            = errors.New("kubernetes timeout")

	// Kubernetes secret errors
	ErrSecretNotFound           = errors.New("secret not found")
	ErrSecretNamespaceNotFound  = errors.New("namespace not found")
	ErrSecretForbidden          = errors.New("forbidden")
	ErrSecretInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrSecretClusterUnavailable = errors.New("cluster unavailable")
	ErrSecretTimeout            = errors.New("kubernetes timeout")

	// Kubernetes runtime deployment errors
	ErrRuntimeDeploymentNotFound           = errors.New("deployment not found")
	ErrRuntimeDeploymentNamespaceNotFound  = errors.New("namespace not found")
	ErrRuntimeDeploymentForbidden          = errors.New("forbidden")
	ErrRuntimeDeploymentInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrRuntimeDeploymentClusterUnavailable = errors.New("cluster unavailable")
	ErrRuntimeDeploymentTimeout            = errors.New("kubernetes timeout")

	// Kubernetes node errors
	ErrNodeForbidden          = errors.New("forbidden")
	ErrNodeInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrNodeClusterUnavailable = errors.New("cluster unavailable")
	ErrNodeTimeout            = errors.New("kubernetes timeout")

	// Kubernetes namespace errors
	ErrNamespaceForbidden          = errors.New("forbidden")
	ErrNamespaceInvalidKubeconfig  = errors.New("invalid kubeconfig")
	ErrNamespaceClusterUnavailable = errors.New("cluster unavailable")
	ErrNamespaceTimeout            = errors.New("kubernetes timeout")

	// Comment errors
	ErrCommentNotFound       = errors.New("comment not found")
	ErrCommentForbidden      = errors.New("forbidden")
	ErrInvalidCommentContent = errors.New("invalid comment content")

	// Generic errors
	ErrInternal = errors.New("internal server error")
)
