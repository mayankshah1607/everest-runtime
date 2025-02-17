package main

import (
	chv1 "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse.altinity.com/v1"
	"github.com/mayankshah1607/everest-runtime/internal/providers/clickhouse"
	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"github.com/mayankshah1607/everest-runtime/pkg/plugin"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var scheme = runtime.NewScheme()

func main() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}

	chProv := clickhouse.New(scheme)

	plugin := &plugin.Plugin{
		Manager: mgr,
		Name:    "clickhouse",
		Controllers: plugin.Controllers{
			DatabaseController: chProv.DatabaseCluster,
		},
	}

	if err := plugin.Run(ctrl.SetupSignalHandler()); err != nil {
		panic(err)
	}
}

func init() {
	v2alpha1.AddToScheme(scheme)
	chv1.AddToScheme(scheme)
}
