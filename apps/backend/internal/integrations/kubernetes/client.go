package kubernetes

import (
	"bytes"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

		return cfg, nil
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig(c.kubeconfig)
	if err != nil {
		return nil, &ErrInvalidKubeconfig{Err: err}
	}

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
