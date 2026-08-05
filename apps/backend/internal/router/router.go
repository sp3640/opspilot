package router

import (
	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/handlers"
	"github.com/sp3640/opspilot/backend/internal/metrics"
	"github.com/sp3640/opspilot/backend/internal/middleware"
)

func RegisterRoutes(
	r *gin.Engine,
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	organizationHandler *handlers.OrganizationHandler,
	invitationHandler *handlers.InvitationHandler,
	projectHandler *handlers.ProjectHandler,
	applicationHandler *handlers.ApplicationHandler,
	podHandler *handlers.PodHandler,
	kubernetesLogHandler *handlers.KubernetesLogHandler,
	kubernetesConfigMapHandler *handlers.KubernetesConfigMapHandler,
	kubernetesSecretHandler *handlers.KubernetesSecretHandler,
	kubernetesRuntimeDeploymentHandler *handlers.KubernetesRuntimeDeploymentHandler,
	kubernetesServiceHandler *handlers.KubernetesServiceHandler,
	kubernetesIngressHandler *handlers.KubernetesIngressHandler,
	kubernetesEventHandler *handlers.KubernetesEventHandler,
	deploymentHandler *handlers.DeploymentHandler,
	deploymentHistoryHandler *handlers.DeploymentHistoryHandler,
	teamHandler *handlers.TeamHandler,
	projectTeamHandler *handlers.ProjectTeamHandler,
	incidentHandler *handlers.IncidentHandler,
	alertHandler *handlers.AlertHandler,
	metricHandler *handlers.MetricHandler,
	clusterHandler *handlers.ClusterHandler,
	resourceHandler *handlers.ResourceHandler,
	commentHandler *handlers.CommentHandler,
	auditHandler *handlers.AuditHandler,
	dashboardHandler *handlers.DashboardHandler,
	healthHandler *handlers.HealthHandler,
	collector *metrics.Collector,
) {
	if healthHandler != nil {
		r.GET("/health", healthHandler.HealthCheck)
		r.GET("/ready", healthHandler.Ready)
		r.GET("/live", healthHandler.Live)
	}
	if collector != nil {
		r.GET("/metrics", collector.Handler)
	}

	api := r.Group("/api/v1")
	{
		// =========================
		// Public Routes
		// =========================
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// =========================
		// Protected User Routes
		// =========================
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg))
		{
			users.GET("/me", userHandler.Me)
		}

		// =========================
		// Protected Organization Routes
		// =========================
		organizations := api.Group("/organizations")
		organizations.Use(middleware.AuthMiddleware(cfg))
		{
			organizations.POST("", organizationHandler.Create)
			organizations.GET("", organizationHandler.List)
			organizations.GET("/:id", organizationHandler.GetByID)
			organizations.PUT("/:id", organizationHandler.Update)
			organizations.DELETE("/:id", organizationHandler.Delete)
		}

		invitations := api.Group("/invitations")
		invitations.Use(middleware.AuthMiddleware(cfg))
		{
			invitations.POST("", invitationHandler.Invite)
			invitations.GET("", invitationHandler.List)
			invitations.POST("/accept", invitationHandler.Accept)
			invitations.DELETE("/:id", invitationHandler.Revoke)
		}

		// =========================
		// Protected Project Routes
		// =========================
		projects := api.Group("/projects")
		projects.Use(middleware.AuthMiddleware(cfg))
		{
			projects.POST("", projectHandler.Create)
			projects.GET("", projectHandler.List)
			projects.GET("/:id", projectHandler.GetByID)
			projects.PUT("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)
			projects.POST("/:id/applications", applicationHandler.Create)
			projects.GET("/:id/applications", applicationHandler.ListByProject)
			projects.GET("/:id/deployments", deploymentHandler.ListByProject)
			projects.POST("/:id/teams", projectTeamHandler.AssignTeam)
			projects.GET("/:id/teams", projectTeamHandler.ListProjectTeams)
			projects.DELETE("/:id/teams/:teamId", projectTeamHandler.RemoveTeam)
			projects.GET("/:id/audit-logs", auditHandler.GetProjectAuditLogs)
		}

		applications := api.Group("/applications")
		applications.Use(middleware.AuthMiddleware(cfg))
		{
			applications.GET("/:id", applicationHandler.GetByID)
			applications.PUT("/:id", applicationHandler.Update)
			applications.DELETE("/:id", applicationHandler.Delete)
			applications.GET("/:id/deployments", deploymentHandler.ListByApplication)
			if podHandler != nil {
				applications.GET("/:id/pods", podHandler.ListByApplication)
			}
			if kubernetesServiceHandler != nil {
				applications.GET("/:id/services", kubernetesServiceHandler.ListByApplication)
			}
			if kubernetesIngressHandler != nil {
				applications.GET("/:id/ingresses", kubernetesIngressHandler.ListByApplication)
			}
			if kubernetesEventHandler != nil {
				applications.GET("/:id/events", kubernetesEventHandler.ListByApplication)
			}
			if kubernetesConfigMapHandler != nil {
				applications.GET("/:id/configmaps", kubernetesConfigMapHandler.ListByApplication)
			}
			if kubernetesSecretHandler != nil {
				applications.GET("/:id/secrets", kubernetesSecretHandler.ListByApplication)
			}
			if kubernetesRuntimeDeploymentHandler != nil {
				applications.GET("/:id/runtime/deployments", kubernetesRuntimeDeploymentHandler.ListByApplication)
			}
		}

		pods := api.Group("/pods")
		pods.Use(middleware.AuthMiddleware(cfg))
		{
			if podHandler != nil {
				pods.GET("/:namespace/:name", podHandler.GetByNamespaceAndName)
			}
			if kubernetesLogHandler != nil {
				pods.GET("/:namespace/:name/logs", kubernetesLogHandler.GetPodLogs)
			}
		}

		services := api.Group("/services")
		services.Use(middleware.AuthMiddleware(cfg))
		{
			if kubernetesServiceHandler != nil {
				services.GET("/:namespace/:name", kubernetesServiceHandler.GetByNamespaceAndName)
			}
		}

		ingresses := api.Group("/ingresses")
		ingresses.Use(middleware.AuthMiddleware(cfg))
		{
			if kubernetesIngressHandler != nil {
				ingresses.GET("/:namespace/:name", kubernetesIngressHandler.GetByNamespaceAndName)
			}
		}

		events := api.Group("/events")
		events.Use(middleware.AuthMiddleware(cfg))
		{
			if kubernetesEventHandler != nil {
				events.GET("/:namespace/:name", kubernetesEventHandler.GetByNamespaceAndName)
			}
		}

		configMaps := api.Group("/configmaps")
		configMaps.Use(middleware.AuthMiddleware(cfg))
		{
			if kubernetesConfigMapHandler != nil {
				configMaps.GET("/:namespace/:name", kubernetesConfigMapHandler.GetByNamespaceAndName)
			}
		}

		secrets := api.Group("/secrets")
		secrets.Use(middleware.AuthMiddleware(cfg))
		{
			if kubernetesSecretHandler != nil {
				secrets.GET("/:namespace/:name", kubernetesSecretHandler.GetByNamespaceAndName)
			}
		}

		runtime := api.Group("/runtime")
		runtime.Use(middleware.AuthMiddleware(cfg))
		{
			runtimeDeployments := runtime.Group("/deployments")
			if kubernetesRuntimeDeploymentHandler != nil {
				runtimeDeployments.GET("/:namespace/:name", kubernetesRuntimeDeploymentHandler.GetByNamespaceAndName)
			}
		}

		deployments := api.Group("/deployments")
		deployments.Use(middleware.AuthMiddleware(cfg))
		{
			deployments.POST("", deploymentHandler.Create)
			deployments.GET("", deploymentHandler.List)
			deployments.GET("/:id/history", deploymentHistoryHandler.ListByDeployment)
			deployments.GET("/:id/history/:revision", deploymentHistoryHandler.GetRevision)
			deployments.GET("/:id", deploymentHandler.GetByID)
			deployments.PATCH("/:id", deploymentHandler.Update)
			deployments.DELETE("/:id", deploymentHandler.Delete)
			deployments.PATCH("/:id/cancel", deploymentHandler.Cancel)
			deployments.POST("/:id/rollback", deploymentHandler.Rollback)
		}

		teams := api.Group("/teams")
		teams.Use(middleware.AuthMiddleware(cfg))
		{
			teams.POST("", teamHandler.Create)
			teams.GET("", teamHandler.List)
			teams.GET("/:id", teamHandler.GetByID)
			teams.GET("/:id/projects", projectTeamHandler.ListTeamProjects)
			teams.PUT("/:id", teamHandler.Update)
			teams.DELETE("/:id", teamHandler.Delete)
			teams.POST("/:id/members", teamHandler.AddMember)
			teams.DELETE("/:id/members/:userId", teamHandler.RemoveMember)
			teams.GET("/:id/members", teamHandler.ListMembers)
		}

		// =========================
		// Protected Incident Routes
		// =========================
		incidents := api.Group("/incidents")
		incidents.Use(middleware.AuthMiddleware(cfg))
		{
			incidents.POST("", incidentHandler.Create)
			incidents.GET("", incidentHandler.List)
			incidents.GET("/:id", incidentHandler.GetByID)
			incidents.PUT("/:id", incidentHandler.Update)
			incidents.DELETE("/:id", incidentHandler.Delete)
			incidents.POST("/:id/comments", commentHandler.Create)
			incidents.GET("/:id/comments", commentHandler.List)
			incidents.GET("/:id/audit-logs", auditHandler.GetIncidentAuditLogs)
		}

		alerts := api.Group("/alerts")
		alerts.Use(middleware.AuthMiddleware(cfg))
		{
			alerts.POST("", alertHandler.Create)
			alerts.GET("", alertHandler.List)
			alerts.GET("/:id", alertHandler.GetByID)
			alerts.PUT("/:id", alertHandler.Update)
			alerts.DELETE("/:id", alertHandler.Delete)
			alerts.POST("/:id/acknowledge", alertHandler.Acknowledge)
			alerts.POST("/:id/resolve", alertHandler.Resolve)
			alerts.POST("/:id/reopen", alertHandler.Reopen)
			alerts.POST("/:id/incident", alertHandler.AttachIncident)
		}

		metricsRoutes := api.Group("/metrics")
		metricsRoutes.Use(middleware.AuthMiddleware(cfg))
		{
			metricsRoutes.GET("", metricHandler.List)
			metricsRoutes.GET("/latest", metricHandler.GetLatest)
			metricsRoutes.GET("/history", metricHandler.GetHistory)
			metricsRoutes.GET("/aggregate", metricHandler.Aggregate)
		}

		clusters := api.Group("/clusters")
		clusters.Use(middleware.AuthMiddleware(cfg))
		{
			clusters.POST("", clusterHandler.Create)
			clusters.GET("", clusterHandler.List)
			clusters.GET("/:id/metrics", metricHandler.GetClusterMetrics)
			clusters.GET("/:id", clusterHandler.GetByID)
			clusters.PUT("/:id", clusterHandler.Update)
			clusters.DELETE("/:id", clusterHandler.Delete)
			clusters.POST("/:id/validate", clusterHandler.Validate)
			clusters.POST("/:id/default", clusterHandler.SetDefault)
		}

		resources := api.Group("/resources")
		resources.Use(middleware.AuthMiddleware(cfg))
		{
			resources.POST("/sync", resourceHandler.Sync)
			resources.POST("", resourceHandler.Create)
			resources.GET("", resourceHandler.List)
			resources.GET("/:id/metrics", metricHandler.GetResourceMetrics)
			resources.GET("/:id", resourceHandler.GetByID)
			resources.PUT("/:id", resourceHandler.Update)
			resources.DELETE("/:id", resourceHandler.Delete)
		}

		comments := api.Group("/comments")
		comments.Use(middleware.AuthMiddleware(cfg))
		{
			comments.PUT("/:id", commentHandler.Update)
			comments.DELETE("/:id", commentHandler.Delete)
		}

		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware(cfg))
		{
			dashboard.GET("/overview", dashboardHandler.Overview)
			dashboard.GET("/resources", dashboardHandler.Resources)
			dashboard.GET("/alerts", dashboardHandler.Alerts)
			dashboard.GET("/clusters", dashboardHandler.Clusters)
			dashboard.GET("/metrics", dashboardHandler.Metrics)
		}
	}
}
