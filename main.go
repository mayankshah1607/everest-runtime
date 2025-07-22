package main

import (
	"context"

	"github.com/mayankshah1607/everest-runtime/internal/clickhouse"
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	"github.com/mayankshah1607/everest-runtime/pkg/reconciler"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	chkv1 "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse-keeper.altinity.com/v1"
	chv1 "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse.altinity.com/v1"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	// contains the logic for clickhouse
	chctrl := clickhouse.New()

	bldr := controller.NewControllerBuilder()

	// Set the name of the DatabaseEngine we will be managing.
	bldr.WithEngineName("clickhouse")

	// We own the following objects.
	bldr.Owns(&chv1.ClickHouseInstallation{})
	bldr.Owns(&chkv1.ClickHouseKeeperInstallation{})

	// Configure additional watchers.
	bldr.WithWatch(&corev1.Pod{}, func(ctx context.Context, obj client.Object) []reconcile.Request {
		return nil
	}, controller.WithWatchOptions{})

	// Prepare reconciliation steps
	bldr.WithSyncStep("Prepare default users", chctrl.EnsureDefaultSecret, true)
	bldr.WithSyncStep("Ensure ClickHouse Keeper", chctrl.ReconcileClickhouseKeeper, true)
	bldr.WithSyncStep("Ensure ClickHouse Installation", chctrl.ReconcileClickhouseInstallation, true)

	// Status reconciliation
	bldr.WithStatusGetter(chctrl.GetStatus)

	// Add validation
	bldr.WithValidator(chctrl.Validate)

	// Prepare cleanup steps
	bldr.WithCleanupStep("Cleanup ClickHouse Installation", chctrl.CleanupClickhouseInstallation, true)

	// Start reconciler
	reconciler := reconciler.New(bldr.Build())
	if err := reconciler.Start(); err != nil {
		panic(err)
	}
}
