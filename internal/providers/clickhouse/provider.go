package clickhouse

import (
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
)

type Provider struct {
	DatabaseCluster controller.DatabaseClusterController
}

func New(scheme *runtime.Scheme) *Provider {
	return &Provider{
		DatabaseCluster: &databaseClusterImpl{
			schema: scheme,
		},
	}
}
