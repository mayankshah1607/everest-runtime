package plugin

import (
	"context"
	"log"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"github.com/mayankshah1607/everest-runtime/pkg/reconcilers/databaseclusters"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Controllers struct {
	DatabaseController controller.DatabaseClusterController
}

type Plugin struct {
	Manager      manager.Manager
	Name         string
	Controllers  Controllers
	Capabilities []string
}

func (p *Plugin) Run(ctx context.Context) error {
	err := (&databaseclusters.Reconciler{
		Client:     p.Manager.GetClient(),
		Controller: p.Controllers.DatabaseController,
	}).Setup(p.Manager)
	if err != nil {
		return err
	}

	log.Println("Starting manager")
	utilruntime.Must(clientgoscheme.AddToScheme(p.Manager.GetScheme()))
	return p.Manager.Start(ctx)
}
