package clickhouse

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clickhouseController struct {
}

func New() *clickhouseController {
	return &clickhouseController{}
}

func (chc *clickhouseController) EnsureDefaultSecret(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (controller.StepResult, error) {
	return controller.StepResult{}, nil //todo
}

func (chc *clickhouseController) ReconcileClickhouseInstallation(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (controller.StepResult, error) {
	return controller.StepResult{}, nil //todo
}

func (chc *clickhouseController) ReconcileClickhouseKeeper(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (controller.StepResult, error) {
	return controller.StepResult{}, nil //todo
}

func (chc *clickhouseController) CleanupClickhouseInstallation(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (controller.StepResult, error) {
	return controller.StepResult{}, nil //todo
}

func (chc *clickhouseController) Validate(db *v2alpha1.DatabaseCluster) error {
	return nil //todo
}

func (chc *clickhouseController) GetStatus(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (v2alpha1.DatabaseClusterStatus, error) {
	return v2alpha1.DatabaseClusterStatus{}, nil //todo
}
