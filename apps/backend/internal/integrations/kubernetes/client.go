package kubernetes

import (
	"bytes"
	"time"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// requestTimeout bounds every individual HTTP request client-go makes
// through a *rest.Config built here (list/get calls across every
// internal/kubernetes/* service - none of them use Watch or log-follow, so
// nothing here is meant to be long-running). Without this, a managed
// cluster whose kubeconfig points somewhere that TCP-hangs rather than
// fails fast leaves the calling request goroutine blocked indefinitely -
// well past this server's own HTTP write timeout, so the client just sees a
// truncated/reset response while the goroutine (and cluster connection)
// keeps running server-side.
const requestTimeout = 30 * time.Second

type Client interface {
	RESTConfig() (*rest.Config, error)
	Clientset() (clientset.Interface, error)
}

type client struct {
	kubeconfig []byte
}

func NewClient(kubeconfig []byte) Client {
	return &client{
		kubeconfig: bytes.Clone(kubeconfig),
	}
}

func (c *client) RESTConfig() (*rest.Config, error) {
	if len(bytes.TrimSpace(c.kubeconfig)) == 0 {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, &ErrConnectionFailed{Err: err}
		}

		cfg.Timeout = requestTimeout
		return cfg, nil
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig(c.kubeconfig)
	if err != nil {
		return nil, &ErrInvalidKubeconfig{Err: err}
	}

	cfg.Timeout = requestTimeout
	return cfg, nil
}

func (c *client) Clientset() (clientset.Interface, error) {
	cfg, err := c.RESTConfig()
	if err != nil {
		return nil, err
	}

	set, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, wrapConnectionError(err)
	}

	return set, nil
}
