package cmd

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/shipyard-run/shipyard/pkg/clients/mocks"
)

const (
	logStdOut = 1
	logStdErr = 0
)

type testWriter struct {
	Buffer *bytes.Buffer
	mutex  sync.Mutex
}

func newTestWriter() *testWriter {
	return &testWriter{
		Buffer: bytes.NewBuffer([]byte("")),
		mutex:  sync.Mutex{},
	}
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	return tw.Buffer.Write(p)
}

func setupLog(t *testing.T, logStream int) (*cobra.Command, *mocks.MockDocker, *bytes.Buffer, *bytes.Buffer) {
	// setup the statefile
	t.Cleanup(setupState(logState))

	// hijack stdout and stderr
	stdout := newTestWriter()
	stderr := newTestWriter()

	log := createLogOutput(logStream)

	md := &mocks.MockDocker{}
	md.On("ServerVersion", mock.Anything).Return(types.Version{}, nil)

	md.On("ContainerLogs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
		io.NopCloser(bytes.NewBuffer(log)),
		nil,
	)

	md.On("ContainerLogs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
		io.NopCloser(bytes.NewBuffer(log)),
		nil,
	)

	md.On("ContainerLogs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
		io.NopCloser(bytes.NewBuffer(log)),
		nil,
	)

	md.On("ContainerLogs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
		io.NopCloser(bytes.NewBuffer(log)),
		nil,
	)

	md.On("ContainerLogs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
		io.NopCloser(bytes.NewBuffer(log)),
		nil,
	)

	md.On("ContainerLogs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
		io.NopCloser(bytes.NewBuffer(log)),
		nil,
	)

	lc := newLogCmd(nil, md, stdout, stderr)

	return lc, md, stdout.Buffer, stderr.Buffer
}

// createLogOutput creates a byte array that is formatted as a docker log
func createLogOutput(logStream int) []byte {
	out := []byte{}
	for _, line := range logLines {
		hdr := make([]byte, 8)

		// stdout
		hdr[0] = byte(logStream)

		// line length
		l := uint32(len(line))
		binary.BigEndian.PutUint32(hdr[4:], l)
		out = append(out, hdr...)

		out = append(out, []byte(line)...)
	}

	return out
}

func TestLogWithAllCallsDockerLog(t *testing.T) {
	lc, md, _, _ := setupLog(t, logStdOut)

	// call the command
	err := lc.Execute()
	require.NoError(t, err)

	// check that the docker client was called
	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "40",
	}

	md.AssertNumberOfCalls(t, "ContainerLogs", 6)
	md.AssertCalled(t, "ContainerLogs", mock.Anything, "consul.container.shipyard.run", logOptions)
	md.AssertCalled(t, "ContainerLogs", mock.Anything, "docker-cache.image-cache.shipyard.run", logOptions)
	md.AssertCalled(t, "ContainerLogs", mock.Anything, "server.dev.k8s-cluster.shipyard.run", logOptions)
	md.AssertCalled(t, "ContainerLogs", mock.Anything, "server.dev.nomad-cluster.shipyard.run", logOptions)
	md.AssertCalled(t, "ContainerLogs", mock.Anything, "1.client.dev.nomad-cluster.shipyard.run", logOptions)
	md.AssertCalled(t, "ContainerLogs", mock.Anything, "2.client.dev.nomad-cluster.shipyard.run", logOptions)
}

func TestLogWithSpecificResourceCallsDockerLog(t *testing.T) {
	lc, md, _, _ := setupLog(t, logStdErr)

	// call the command
	lc.SetArgs([]string{"consul.container.shipyard.run"})
	err := lc.Execute()
	require.NoError(t, err)

	// check that the logs were written to stdout
	md.AssertNumberOfCalls(t, "ContainerLogs", 1)
	md.AssertCalled(t, "ContainerLogs", mock.Anything, "consul.container.shipyard.run", mock.Anything)
}

//func TestLogWithInvalidSpecificResourceReturnsError(t *testing.T) {
//	lc, md, _, _ := setupLog(t, logStdErr)
//
//	// call the command
//	lc.SetArgs([]string{"container.consul2"})
//	err := lc.Execute()
//	require.Error(t, err)
//
//	md.AssertNumberOfCalls(t, "ContainerLogs", 0)
//}

func TestLogWritesDockerLogToStdOut(t *testing.T) {
	lc, _, stdout, _ := setupLog(t, logStdOut)

	// call the command
	err := lc.Execute()
	require.NoError(t, err)

	// check that the logs were written to stdout
	require.Contains(t, stdout.String(), "[docker-cache.image-cache]   [16:10:20] [main/INFO]: Applying mixin: R1_17.MixinBlockEntity...")
	require.Contains(t, stdout.String(), "[consul.container]   [16:10:20] [main/INFO]: Applying mixin: R1_17.MixinBlockEntity...")
}

func TestLogWritesDockerLogToStdErr(t *testing.T) {
	lc, _, _, stderr := setupLog(t, logStdErr)

	// call the command
	err := lc.Execute()
	require.NoError(t, err)

	// check that the logs were written to stdout
	require.Contains(t, stderr.String(), "[docker-cache.image-cache]   [16:10:20] [main/INFO]: Applying mixin: R1_17.MixinBlockEntity...")
	require.Contains(t, stderr.String(), "[consul.container]   [16:10:20] [main/INFO]: Applying mixin: R1_17.MixinBlockEntity...")
}

var logState = `
{
 "resources": [
    {
      "name": "docker-cache",
      "type": "image_cache",
      "status": "applied",
      "depends_on": [
        "network.onprem"
      ],
      "networks": [
        "network.onprem"
      ]
    },
    {
      "name": "dev",
      "type": "nomad_cluster",
      "status": "applied",
			"client_nodes": 2
    },
    {
      "name": "dev",
      "type": "k8s_cluster",
      "status": "applied"
    },
    {
      "name": "consul_disabled",
      "type": "container",
      "status": "disabled",
      "disabled": true,
      "image": {
        "name": "consul:1.8.1"
      },
      "build": null
    },
    {
      "name": "consul",
      "type": "container",
      "status": "applied",
      "depends_on": [
        "network.onprem",
        "template.consul_config"
      ],
      "depends": [
        "template.consul_config"
      ],
      "networks": [
        {
          "name": "network.onprem",
          "ip_address": "10.6.0.200",
          "aliases": [
            "myalias"
          ]
        }
      ],
      "image": {
        "name": "consul:1.8.1"
      },
      "build": null,
      "command": [
        "consul",
        "agent",
        "-config-file=/config/consul.hcl"
      ],
      "env": [
        {
          "key": "something",
          "value": "blah blah"
        },
        {
          "key": "foo",
          "value": ""
        },
        {
          "key": "file",
          "value": "this is the contents of a file"
        },
        {
          "key": "abc",
          "value": "123"
        },
        {
          "key": "SHIPYARD_FOLDER",
          "value": "/home/nicj/.shipyard"
        },
        {
          "key": "HOME_FOLDER",
          "value": "/home/nicj"
        }
      ],
      "volumes": [
        {
          "source": "/home/nicj/go/src/github.com/shipyard-run/shipyard/examples/container/consul_config",
          "destination": "/config"
        }
      ],
      "port_ranges": [
        {
          "local": "8500-8502",
          "enable_host": true
        }
      ],
      "resources": {
        "cpu": 2000,
        "cpu_pin": [
          0,
          1
        ],
        "memory": 1024
      }
    }
	]
}`

var logLines = []string{
	"[16:10:20] [main/INFO]: Applying mixin: R1_17.MixinNbtTag...",
	"[16:10:20] [main/INFO]: Applying mixin: R1_17.MixinBlockEntity...",
	"[16:10:20] [main/INFO]: Applying mixin: R1_17.MixinChestBlockEntity...",
	"[16:10:20] [main/INFO]: Applying mixin: R1_17.MixinScreenHandler...",
	"[16:10:20] [main/INFO]: Applying mixin: R1_17.MixinChunkGenerator...",
	"[16:10:20] [main/INFO]: Applying mixin: R1_17.MixinPersistentStateManager...",
}
