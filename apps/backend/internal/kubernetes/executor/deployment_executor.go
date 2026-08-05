package executor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
)

type clientsetFactory interface {
	Clientset(kubeconfig []byte) (kubernetes.Interface, error)
}

type ClientsetFactory struct{}

func NewClientsetFactory() *ClientsetFactory {
	return &ClientsetFactory{}
}

func (f *ClientsetFactory) Clientset(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}

type DeploymentExecutor struct {
	deploymentRepo  *repository.DeploymentRepository
	applicationRepo *repository.ApplicationRepository
	clusterRepo     *repository.ClusterRepository
	cipher          security.ClusterCredentialCipher
	manifestBuilder *DeploymentManifestBuilder
	statusUpdater   *DeploymentStatusUpdater
	factory         clientsetFactory
	timeout         time.Duration
}

func NewDeploymentExecutor(
	deploymentRepo *repository.DeploymentRepository,
	applicationRepo *repository.ApplicationRepository,
	clusterRepo *repository.ClusterRepository,
	cipher security.ClusterCredentialCipher,
	manifestBuilder *DeploymentManifestBuilder,
	statusUpdater *DeploymentStatusUpdater,
	factory clientsetFactory,
	timeout time.Duration,
) *DeploymentExecutor {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	if factory == nil {
		factory = NewClientsetFactory()
	}

	return &DeploymentExecutor{
		deploymentRepo:  deploymentRepo,
		applicationRepo: applicationRepo,
		clusterRepo:     clusterRepo,
		cipher:          cipher,
		manifestBuilder: manifestBuilder,
		statusUpdater:   statusUpdater,
		factory:         factory,
		timeout:         timeout,
	}
}

func (e *DeploymentExecutor) ExecuteDeployment(ctx context.Context, deploymentID, organizationID uuid.UUID, userID uint) (*models.Deployment, error) {
	executionCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	deployment, application, cluster, err := e.loadExecutionContext(executionCtx, deploymentID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := e.newClientset(cluster)
	if err != nil {
		return nil, mapExecutionError(err)
	}

	if err := e.ensureNamespace(executionCtx, clientset, deployment.Namespace); err != nil {
		return nil, mapExecutionError(err)
	}

	if err := e.statusUpdater.MarkRunning(executionCtx, deployment, userID); err != nil {
		return nil, err
	}

	manifest := e.manifestBuilder.Build(deployment, application)
	if err := e.applyDeployment(executionCtx, clientset, manifest); err != nil {
		execErr := mapExecutionError(err)
		failureReason := execErr.Error()
		if markErr := e.statusUpdater.MarkFailed(executionCtx, deployment, userID, failureReason); markErr != nil {
			return nil, fmt.Errorf("%w (status update failed: %v)", execErr, markErr)
		}
		return nil, execErr
	}

	if err := e.statusUpdater.MarkSucceeded(executionCtx, deployment, userID); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (e *DeploymentExecutor) loadExecutionContext(ctx context.Context, deploymentID, organizationID uuid.UUID) (*models.Deployment, *models.Application, *models.Cluster, error) {
	_ = ctx

	deployment, err := e.deploymentRepo.GetByID(deploymentID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, apperrors.ErrDeploymentNotFound
		}
		return nil, nil, nil, err
	}

	application, err := e.applicationRepo.GetApplication(deployment.ApplicationID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, apperrors.ErrApplicationNotFound
		}
		return nil, nil, nil, err
	}

	cluster, err := e.clusterRepo.FindByID(deployment.TargetClusterID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, apperrors.ErrClusterNotFound
		}
		return nil, nil, nil, err
	}

	return deployment, application, cluster, nil
}

func (e *DeploymentExecutor) newClientset(cluster *models.Cluster) (kubernetes.Interface, error) {
	encryptedKubeconfig := strings.TrimSpace(cluster.KubeconfigEncrypted)
	if encryptedKubeconfig == "" {
		encryptedKubeconfig = strings.TrimSpace(cluster.EncryptedCredential)
	}
	if encryptedKubeconfig == "" {
		return nil, apperrors.ErrDeploymentInvalidKubeconfig
	}

	decryptedKubeconfig, err := e.cipher.Decrypt(encryptedKubeconfig)
	if err != nil {
		return nil, apperrors.ErrDeploymentInvalidKubeconfig
	}

	clientset, err := e.factory.Clientset([]byte(decryptedKubeconfig))
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func (e *DeploymentExecutor) ensureNamespace(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return errNamespaceMissing
		}
		return err
	}

	return nil
}

func (e *DeploymentExecutor) applyDeployment(ctx context.Context, clientset kubernetes.Interface, manifest *appsv1.Deployment) error {
	resource := clientset.AppsV1().Deployments(manifest.Namespace)

	existing, err := resource.Get(ctx, manifest.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, createErr := resource.Create(ctx, manifest, metav1.CreateOptions{})
			return createErr
		}
		return err
	}

	manifest.ResourceVersion = existing.ResourceVersion
	_, err = resource.Update(ctx, manifest, metav1.UpdateOptions{})
	return err
}
