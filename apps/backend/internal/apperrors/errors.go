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

	// Application errors
	ErrApplicationNotFound       = errors.New("application not found")
	ErrApplicationForbidden      = errors.New("forbidden")
	ErrApplicationAlreadyExists  = errors.New("application already exists")
	ErrInvalidApplicationName    = errors.New("application name is required")
	ErrInvalidApplicationSlug    = errors.New("invalid application slug")
	ErrInvalidApplicationRuntime = errors.New("invalid application runtime")
	ErrInvalidApplicationStatus  = errors.New("invalid application status")
	ErrInvalidApplicationPort    = errors.New("invalid application port")

	// Resource errors
	ErrResourceNotFound      = errors.New("resource not found")
	ErrInvalidResourceKind   = errors.New("invalid resource kind")
	ErrInvalidResourceStatus = errors.New("invalid resource status")
	ErrInvalidResourceHealth = errors.New("invalid resource health")

	// Comment errors
	ErrCommentNotFound       = errors.New("comment not found")
	ErrCommentForbidden      = errors.New("forbidden")
	ErrInvalidCommentContent = errors.New("invalid comment content")

	// Generic errors
	ErrInternal = errors.New("internal server error")
)
