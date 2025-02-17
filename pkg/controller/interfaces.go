package controller

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DatabaseClusterController interface {
	GetSources(manager.Manager) []source.Source
	Reconcile(context.Context, client.Client, *v2alpha1.DatabaseCluster) (reconcile.Result, error)
	Delete(context.Context, client.Client, *v2alpha1.DatabaseCluster) (bool, error)
	GetStatus(context.Context, client.Client, *v2alpha1.DatabaseCluster) (v2alpha1.DatabaseClusterStatus, error)
	GetDefaultCredentials(context.Context, client.Client, *v2alpha1.DatabaseCluster) (*Credentials, error)
}
