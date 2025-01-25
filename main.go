package main

import (
	"github.com/mayankshah1607/everest-runtime/internal/providers/pxc"
	"github.com/mayankshah1607/everest-runtime/pkg/reconcilers"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	pgv2 "github.com/percona/percona-postgresql-operator/pkg/apis/pgv2.percona.com/v2"
	psmdbv1 "github.com/percona/percona-server-mongodb-operator/pkg/apis/psmdb/v1"
	pxcv1 "github.com/percona/percona-xtradb-cluster-operator/pkg/apis/pxc/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme = runtime.NewScheme()
)

func main() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}
	addToSchemes()

	pxcProv := pxc.New(mgr.GetClient())

	c := reconcilers.Reconciler{
		Manager:                          mgr,
		DatabaseClusterController:        pxcProv.DatabaseClusterController,
		DatabaseClusterBackupController:  pxcProv.DatabaseClusterBackupController,
		DatabaseClusterRestoreController: pxcProv.DatabaseClusterRestoreController,
	}

	if err := c.Start(ctrl.SetupSignalHandler()); err != nil {
		panic(err)
	}
}

func addToSchemes() {
	utilruntime.Must(everestv1alpha1.AddToScheme(scheme))
	utilruntime.Must(pgv2.AddToScheme(scheme))
	utilruntime.Must(psmdbv1.SchemeBuilder.AddToScheme(scheme))
	utilruntime.Must(pxcv1.SchemeBuilder.AddToScheme(scheme))
}
