package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCreatesK8sCluster(t *testing.T) {
	c := NewK8sCluster("abc")

	assert.Equal(t, "abc", c.Name)
	assert.Equal(t, TypeK8sCluster, c.Type)
}

func TestK8sClusterCreatesCorrectly(t *testing.T) {
	c, _ := CreateConfigFromStrings(t, clusterDefault)

	cl, err := c.FindResource("k8s_cluster.testing")
	assert.NoError(t, err)

	assert.Equal(t, "testing", cl.Info().Name)
	assert.Equal(t, TypeK8sCluster, cl.Info().Type)
	assert.Equal(t, PendingCreation, cl.Info().Status)
}

func TestK8sClusterSetsDisabled(t *testing.T) {
	c, _ := CreateConfigFromStrings(t, clusterDisabled)

	cl, err := c.FindResource("k8s_cluster.testing")
	assert.NoError(t, err)

	assert.Equal(t, Disabled, cl.Info().Status)
}

const clusterDefault = `
k8s_cluster "testing" {
	network {
		name = "network.test"
	}
	driver = "k3s"
}
`
const clusterDisabled = `
k8s_cluster "testing" {
	disabled = true

	network {
		name = "network.test"
	}
	driver = "k3s"
}
`
