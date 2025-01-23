package pxc

import (
	"github.com/mayankshah1607/everest-runtime/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type engine struct {
	client.Client
	name string
}

func (e *engine) GetName() string {
	return e.name
}

func (e *engine) GetInstalledOperatorVersion() (string, error) {
	// TODO: read the deployment image tag
	return "1.0.0", nil
}

func (e *engine) GetVersionInfo() (*runtime.VersionInfo, error) {
	// TODO: read some local configmap.
	return &runtime.VersionInfo{}, nil
}
