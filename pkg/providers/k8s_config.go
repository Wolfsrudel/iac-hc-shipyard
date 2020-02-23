package providers

import (
	hclog "github.com/hashicorp/go-hclog"
	"github.com/shipyard-run/shipyard/pkg/clients"
	"github.com/shipyard-run/shipyard/pkg/config"
	"github.com/shipyard-run/shipyard/pkg/utils"
	"golang.org/x/xerrors"
)

type K8sConfig struct {
	config *config.K8sConfig
	client clients.Kubernetes
	log    hclog.Logger
}

// NewK8sConfig creates a provider which can create and destroy kubernetes configuration
func NewK8sConfig(c *config.K8sConfig, kc clients.Kubernetes, l hclog.Logger) *K8sConfig {
	return &K8sConfig{c, kc, l}
}

// Create the Kubernetes resources defined by the config
func (c *K8sConfig) Create() error {
	c.log.Info("Applying Kubernetes configuration", "ref", c.config.Name, "config", c.config.Paths)

	err := c.setup()
	if err != nil {
		return err
	}

	err = c.client.Apply(c.config.Paths, c.config.WaitUntilReady)
	if err != nil {
		return err
	}

	// set the status
	c.config.Status = config.Applied

	return nil
}

// Destroy the Kubernetes resources defined by the config
func (c *K8sConfig) Destroy() error {
	c.log.Info("Destroy Kubernetes configuration", "ref", c.config.Name, "config", c.config.Paths)

	err := c.setup()
	if err != nil {
		return err
	}

	return c.client.Delete(c.config.Paths)
}

// Lookup the Kubernetes resources defined by the config
func (c *K8sConfig) Lookup() ([]string, error) {
	return []string{}, nil
}

func (c *K8sConfig) setup() error {
	cluster, err := c.config.FindDependentResource(c.config.Cluster)
	if err != nil {
		return xerrors.Errorf("Unable to find associated cluster: %w", cluster)
	}

	_, destPath, _ := utils.CreateKubeConfigPath(cluster.Info().Name)
	err = c.client.SetConfig(destPath)
	if err != nil {
		return xerrors.Errorf("unable to create Kubernetes client: %w", err)
	}

	return nil
}
