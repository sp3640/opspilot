package kubernetes

import (
	"context"
	"errors"
	"time"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Validator interface {
	ValidateConnection(ctx context.Context) (*ValidationResult, error)
}

type ValidationResult struct {
	Connected      bool
	ClusterVersion *ServerVersion
	APIServerURL   string
	Latency        time.Duration
	ValidatedAt    time.Time
}

type validator struct {
	client Client
}

func NewValidator(client Client) Validator {
	return &validator{client: client}
}

func (v *validator) ValidateConnection(ctx context.Context) (*ValidationResult, error) {
	startedAt := time.Now()

	cfg, err := v.client.RESTConfig()
	if err != nil {
		return nil, err
	}

	set, err := v.client.Clientset()
	if err != nil {
		return nil, err
	}

	if _, err := set.Discovery().RESTClient().Get().AbsPath("/readyz").DoRaw(ctx); err != nil {
		return nil, wrapConnectionError(err)
	}

	review, err := set.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:     "list",
				Group:    "",
				Resource: "namespaces",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	if review == nil {
		return nil, &ErrAuthorizationFailed{Err: errors.New("permission review response was nil")}
	}
	if !review.Status.Allowed {
		reason := review.Status.Reason
		if reason == "" {
			reason = review.Status.EvaluationError
		}
		if reason == "" {
			reason = "required Kubernetes API permissions were not granted"
		}

		return nil, &ErrAuthorizationFailed{Err: errors.New(reason)}
	}

	version, err := GetServerVersion(ctx, v.client)
	if err != nil {
		return nil, err
	}

	return &ValidationResult{
		Connected:      true,
		ClusterVersion: version,
		APIServerURL:   cfg.Host,
		Latency:        time.Since(startedAt),
		ValidatedAt:    time.Now().UTC(),
	}, nil
}
