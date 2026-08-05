package executor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
	"github.com/sp3640/opspilot/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

type staticClientsetFactory struct {
	clientset kubernetes.Interface
	err       error
}

func (f *staticClientsetFactory) Clientset(_ []byte) (kubernetes.Interface, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.clientset, nil
}

type executorFixture struct {
	db              *gorm.DB
	deploymentRepo  *repository.DeploymentRepository
	applicationRepo *repository.ApplicationRepository
	clusterRepo     *repository.ClusterRepository
	historyService  *services.DeploymentHistoryService
	auditService    *services.AuditService
	cipher          security.ClusterCredentialCipher
	organizationID  uuid.UUID
	projectID       uuid.UUID
	applicationID   uuid.UUID
	clusterID       uuid.UUID
	deployment      *models.Deployment
	userID          uint
}

func TestDeploymentExecutorIntegration_Success(t *testing.T) {
	t.Parallel()

	fixture := setupExecutorFixture(t)
	clientset := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: fixture.deployment.Namespace}})

	exec := NewDeploymentExecutor(
		fixture.deploymentRepo,
		fixture.applicationRepo,
		fixture.clusterRepo,
		fixture.cipher,
		NewDeploymentManifestBuilder(),
		NewDeploymentStatusUpdater(fixture.deploymentRepo, fixture.historyService, fixture.auditService),
		&staticClientsetFactory{clientset: clientset},
		15*time.Second,
	)

	updated, err := exec.ExecuteDeployment(context.Background(), fixture.deployment.ID, fixture.organizationID, fixture.userID)
	if err != nil {
		t.Fatalf("execute deployment: %v", err)
	}
	if updated.Status != constants.DeploymentStatusSucceeded {
		t.Fatalf("expected succeeded status, got %s", updated.Status)
	}

	list, err := clientset.AppsV1().Deployments(fixture.deployment.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list deployments from fake cluster: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected one kubernetes deployment, got %d", len(list.Items))
	}
	manifest := list.Items[0]
	if manifest.Spec.Replicas == nil || *manifest.Spec.Replicas != int32(fixture.deployment.ReplicaCount) {
		t.Fatalf("expected replicas %d", fixture.deployment.ReplicaCount)
	}
	if len(manifest.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("expected one container in manifest")
	}
	if manifest.Spec.Template.Spec.Containers[0].Image != fixture.deployment.Image+":"+fixture.deployment.ImageTag {
		t.Fatalf("expected container image %s:%s", fixture.deployment.Image, fixture.deployment.ImageTag)
	}

	var histories []models.DeploymentHistory
	if err := fixture.db.Where("deployment_id = ?", fixture.deployment.ID).Order("revision asc").Find(&histories).Error; err != nil {
		t.Fatalf("query deployment histories: %v", err)
	}
	if len(histories) != 3 {
		t.Fatalf("expected 3 history rows (created, running, succeeded), got %d", len(histories))
	}
	if histories[1].Status != constants.DeploymentStatusRunning {
		t.Fatalf("expected running history status at revision 2")
	}
	if histories[2].Status != constants.DeploymentStatusSucceeded {
		t.Fatalf("expected succeeded history status at revision 3")
	}

	var auditCount int64
	if err := fixture.db.Model(&models.AuditLog{}).Where("entity_type = ? AND entity_id = ? AND field_name = ?", "deployment", fixture.deployment.ID.String(), "status").Count(&auditCount).Error; err != nil {
		t.Fatalf("count deployment audit logs: %v", err)
	}
	if auditCount < 2 {
		t.Fatalf("expected at least 2 status audit logs, got %d", auditCount)
	}
}

func TestDeploymentExecutorIntegration_NamespaceMissing(t *testing.T) {
	t.Parallel()

	fixture := setupExecutorFixture(t)
	clientset := fake.NewSimpleClientset()

	exec := NewDeploymentExecutor(
		fixture.deploymentRepo,
		fixture.applicationRepo,
		fixture.clusterRepo,
		fixture.cipher,
		NewDeploymentManifestBuilder(),
		NewDeploymentStatusUpdater(fixture.deploymentRepo, fixture.historyService, fixture.auditService),
		&staticClientsetFactory{clientset: clientset},
		15*time.Second,
	)

	_, err := exec.ExecuteDeployment(context.Background(), fixture.deployment.ID, fixture.organizationID, fixture.userID)
	if !errors.Is(err, apperrors.ErrDeploymentNamespaceNotFound) {
		t.Fatalf("expected namespace missing error, got %v", err)
	}

	fresh, getErr := fixture.deploymentRepo.GetByID(fixture.deployment.ID, fixture.organizationID)
	if getErr != nil {
		t.Fatalf("load deployment after failed execution: %v", getErr)
	}
	if fresh.Status != constants.DeploymentStatusPending {
		t.Fatalf("expected pending status when namespace lookup fails, got %s", fresh.Status)
	}
}

func TestDeploymentExecutorIntegration_PermissionDenied(t *testing.T) {
	t.Parallel()

	fixture := setupExecutorFixture(t)
	clientset := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: fixture.deployment.Namespace}})
	clientset.Fake.PrependReactor("create", "deployments", func(_ ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Group: "apps", Resource: "deployments"}, "x", errors.New("forbidden"))
	})

	exec := NewDeploymentExecutor(
		fixture.deploymentRepo,
		fixture.applicationRepo,
		fixture.clusterRepo,
		fixture.cipher,
		NewDeploymentManifestBuilder(),
		NewDeploymentStatusUpdater(fixture.deploymentRepo, fixture.historyService, fixture.auditService),
		&staticClientsetFactory{clientset: clientset},
		15*time.Second,
	)

	_, err := exec.ExecuteDeployment(context.Background(), fixture.deployment.ID, fixture.organizationID, fixture.userID)
	if !errors.Is(err, apperrors.ErrDeploymentExecutionPermissionDenied) {
		t.Fatalf("expected permission denied error, got %v", err)
	}

	fresh, getErr := fixture.deploymentRepo.GetByID(fixture.deployment.ID, fixture.organizationID)
	if getErr != nil {
		t.Fatalf("load deployment after failed execution: %v", getErr)
	}
	if fresh.Status != constants.DeploymentStatusFailed {
		t.Fatalf("expected failed status on apply permission failure, got %s", fresh.Status)
	}

	var histories []models.DeploymentHistory
	if err := fixture.db.Where("deployment_id = ?", fixture.deployment.ID).Order("revision asc").Find(&histories).Error; err != nil {
		t.Fatalf("query deployment histories: %v", err)
	}
	if len(histories) != 3 {
		t.Fatalf("expected 3 histories after failed apply (created, running, failed), got %d", len(histories))
	}
	if histories[2].Status != constants.DeploymentStatusFailed {
		t.Fatalf("expected final failed history status")
	}
}

func TestDeploymentExecutorIntegration_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	fixture := setupExecutorFixture(t)

	cluster, err := fixture.clusterRepo.FindByID(fixture.clusterID, fixture.organizationID)
	if err != nil {
		t.Fatalf("load cluster: %v", err)
	}
	cluster.KubeconfigEncrypted = "not-base64"
	cluster.EncryptedCredential = "not-base64"
	if err := fixture.clusterRepo.Update(cluster); err != nil {
		t.Fatalf("update cluster credential: %v", err)
	}

	exec := NewDeploymentExecutor(
		fixture.deploymentRepo,
		fixture.applicationRepo,
		fixture.clusterRepo,
		fixture.cipher,
		NewDeploymentManifestBuilder(),
		NewDeploymentStatusUpdater(fixture.deploymentRepo, fixture.historyService, fixture.auditService),
		&staticClientsetFactory{err: &intkube.ErrInvalidKubeconfig{Err: errors.New("bad")}},
		15*time.Second,
	)

	_, err = exec.ExecuteDeployment(context.Background(), fixture.deployment.ID, fixture.organizationID, fixture.userID)
	if !errors.Is(err, apperrors.ErrDeploymentInvalidKubeconfig) {
		t.Fatalf("expected invalid kubeconfig error, got %v", err)
	}
}

func setupExecutorFixture(t *testing.T) *executorFixture {
	t.Helper()

	dsn := "file:deployment_executor_integration?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.Project{},
		&models.Application{},
		&models.Cluster{},
		&models.Deployment{},
		&models.DeploymentHistory{},
		&models.AuditLog{},
	); err != nil {
		t.Fatalf("auto migrate schema: %v", err)
	}

	user := &models.User{Name: "Executor User", Email: uuid.NewString() + "@opspilot.dev", PasswordHash: "hash", Role: models.RolePlatformAdmin}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	organization := &models.Organization{Name: "Executor Org", Slug: "executor-org-" + uuid.NewString()[:8], Description: "org", OwnerID: user.ID}
	if err := db.Create(organization).Error; err != nil {
		t.Fatalf("create organization: %v", err)
	}
	user.OrganizationID = &organization.ID
	if err := db.Model(user).Update("organization_id", organization.ID).Error; err != nil {
		t.Fatalf("assign user org: %v", err)
	}

	project := &models.Project{OrganizationID: organization.ID, Name: "Executor Project", Slug: "executor-project-" + uuid.NewString()[:8], Description: "project", Environment: "development", Health: "healthy", OwnerID: user.ID, Members: 1, Services: 1}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}

	application := &models.Application{OrganizationID: organization.ID, ProjectID: project.ID, Name: "Executor App", Slug: "executor-app-" + uuid.NewString()[:8], Runtime: constants.ApplicationRuntimeGo, Port: 8080, Status: constants.ApplicationStatusReady}
	if err := db.Create(application).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	cipherKey := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	cipher, err := security.NewClusterCredentialCipher(cipherKey)
	if err != nil {
		t.Fatalf("create cipher: %v", err)
	}
	encryptedKubeconfig, err := cipher.Encrypt("apiVersion: v1\nclusters: []\ncontexts: []\ncurrent-context: ''\nkind: Config\npreferences: {}\nusers: []")
	if err != nil {
		t.Fatalf("encrypt kubeconfig: %v", err)
	}

	cluster := &models.Cluster{
		OrganizationID:      organization.ID,
		ProjectID:           project.ID,
		Name:                "Executor Cluster",
		Provider:            constants.ClusterProviderKubernetes,
		ConnectionType:      constants.ClusterConnectionTypeKubeconfig,
		CredentialType:      constants.ClusterConnectionTypeKubeconfig,
		EncryptedCredential: encryptedKubeconfig,
		KubeconfigEncrypted: encryptedKubeconfig,
		Status:              constants.ClusterStatusConnected,
		CreatedBy:           user.ID,
		Metadata:            json.RawMessage(`{}`),
	}
	if err := db.Create(cluster).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	deployment := &models.Deployment{
		ApplicationID:      application.ID,
		ProjectID:          project.ID,
		OrganizationID:     organization.ID,
		Image:              "ghcr.io/opspilot/executor",
		ImageTag:           "v1.2.3",
		Environment:        constants.DeploymentEnvironmentProduction,
		Namespace:          "executor-app",
		ReplicaCount:       2,
		Status:             constants.DeploymentStatusPending,
		DeploymentStrategy: constants.DeploymentStrategyRollingUpdate,
		TargetClusterID:    cluster.ID,
		CreatedBy:          user.ID,
		UpdatedBy:          user.ID,
	}
	if err := db.Create(deployment).Error; err != nil {
		t.Fatalf("create deployment: %v", err)
	}

	deploymentRepo := repository.NewDeploymentRepository(db)
	historyRepo := repository.NewDeploymentHistoryRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	historyService := services.NewDeploymentHistoryService(historyRepo, deploymentRepo)
	auditService := services.NewAuditService(auditRepo)
	if err := historyService.CreateHistoryFromDeployment(deployment, "Deployment created", user.ID); err != nil {
		t.Fatalf("seed deployment history: %v", err)
	}

	return &executorFixture{
		db:              db,
		deploymentRepo:  deploymentRepo,
		applicationRepo: repository.NewApplicationRepository(db),
		clusterRepo:     repository.NewClusterRepository(db),
		historyService:  historyService,
		auditService:    auditService,
		cipher:          cipher,
		organizationID:  organization.ID,
		projectID:       project.ID,
		applicationID:   application.ID,
		clusterID:       cluster.ID,
		deployment:      deployment,
		userID:          user.ID,
	}
}

var _ = appsv1.Deployment{}
