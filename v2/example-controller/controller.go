package examplecontroller

import (
	"context"

	chv1 "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse.altinity.com/v1"
	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"github.com/mayankshah1607/everest-runtime/v2/sdk"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func main() {
	// Each controller can be broken down into 5 parts:
	// 1. Watchers
	// 2. Validators
	// 3. Reconciliation
	// 4. Status
	// 5. Cleanup

	b := sdk.NewControllerBuilder()

	// Controller watches and owns ClickHouseInstallation resources
	b.Owns(&chv1.ClickHouseInstallation{})

	// Add watch on custom objects.
	b.WithWatch(&chv1.ClickHouseInstallation{}, func(obj client.Object) []reconcile.Request {
		return []reconcile.Request{} //todo
	}, sdk.WithWatchOptions{})
	b.WithWatch(&chv1.ClickHouseInstallation{}, func(obj client.Object) []reconcile.Request {
		return []reconcile.Request{} //todo
	}, sdk.WithWatchOptions{})

	// Add validators
	b.WithValidator(func(obj v2alpha1.DatabaseCluster) error {
		return nil // todo
	})
	b.WithValidator(func(obj v2alpha1.DatabaseCluster) error {
		return nil // todo
	})

	// Add reconcile steps
	b.WithSyncStep("Ensure ClickHouseInstallation", func(ctx context.Context, c client.Client, dc v2alpha1.DatabaseCluster) (sdk.StepResult, error) {
		return sdk.StepResult{}, nil
	}, true)
	b.WithSyncStep("Ensure ClickHouseKeeper", func(ctx context.Context, c client.Client, dc v2alpha1.DatabaseCluster) (sdk.StepResult, error) {
		return sdk.StepResult{}, nil
	}, true)
	b.WithSyncStep("Sync clickhouse users", func(ctx context.Context, c client.Client, dc v2alpha1.DatabaseCluster) (sdk.StepResult, error) {
		return sdk.StepResult{}, nil
	}, true)

	// Set status
	b.WithStatusGetter(func() (v2alpha1.DatabaseClusterStatus, error) {
		return v2alpha1.DatabaseClusterStatus{}, nil // implement your logic here
	})

	b.WithCleanupStep("Cleanup ClickHouseInstallation finalizers", func(ctx context.Context, c client.Client, dc v2alpha1.DatabaseCluster) (sdk.StepResult, error) {
		return sdk.StepResult{}, nil // implement your logic here
	}, true)
	b.WithCleanupStep("Cleanup ClickhHouseKeeper finalizers", func(ctx context.Context, c client.Client, dc v2alpha1.DatabaseCluster) (sdk.StepResult, error) {
		return sdk.StepResult{}, nil // implement your logic here
	}, true)
}
