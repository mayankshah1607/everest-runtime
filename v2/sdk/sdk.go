package sdk

import (
	"context"
	"time"

	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ControllerBuilder implements the builder pattern for DatabaseClusterController
type ControllerBuilder struct {
	controller *DatabaseClusterController
}

// NewControllerBuilder creates a new builder for DatabaseClusterController
func NewControllerBuilder() *ControllerBuilder {
	return &ControllerBuilder{
		controller: &DatabaseClusterController{
			Watches: watches{
				owned:         []client.Object{},
				customWatches: []customWatch{},
			},
			Plan: Plan{
				Sync:         SyncPlan{},
				Cleanup:      CleanupPlan{},
				StatusGetter: nil,
			},
			Validators: []Validator{},
		},
	}
}

// Owns registers objects to be watched that are owned by the controller
func (b *ControllerBuilder) Owns(obj client.Object) *ControllerBuilder {
	b.controller.Watches.owned = append(b.controller.Watches.owned, obj)
	return b
}

type WithWatchOptions struct {
	LabelSelector map[string]string
}

// WithWatch adds a custom watch with a request mapper function
func (b *ControllerBuilder) WithWatch(obj client.Object, fn requestMapper, opts WithWatchOptions) *ControllerBuilder {
	b.controller.Watches.customWatches = append(b.controller.Watches.customWatches, customWatch{
		obj:           obj,
		labelSelector: opts.LabelSelector,
		requestMapper: fn,
	})
	return b
}

// WithValidator adds a validation function to the controller
func (b *ControllerBuilder) WithValidator(fn Validator) *ControllerBuilder {
	b.controller.Validators = append(b.controller.Validators, fn)
	return b
}

// WithSyncStep adds a synchronization step to the controller
func (b *ControllerBuilder) WithSyncStep(description string, fn StepFn, wait bool) *ControllerBuilder {
	b.controller.Plan.Sync.Steps = append(b.controller.Plan.Sync.Steps, Step{
		Description: description,
		Wait:        wait,
		Fn:          fn,
	})
	return b
}

// WithCleanupStep adds a cleanup step to the controller
func (b *ControllerBuilder) WithCleanupStep(description string, fn StepFn, wait bool) *ControllerBuilder {
	b.controller.Plan.Cleanup.Steps = append(b.controller.Plan.Cleanup.Steps, Step{
		Description: description,
		Wait:        wait,
		Fn:          fn,
	})
	return b
}

// WithStatusGetter sets the status getter function for the controller
func (b *ControllerBuilder) WithStatusGetter(fn StatusGetter) *ControllerBuilder {
	b.controller.Plan.StatusGetter = fn
	return b
}

// Build finalizes the controller configuration and returns the controller
func (b *ControllerBuilder) Build() *DatabaseClusterController {
	return b.controller
}

// Type definitions

type Validator func(obj v2alpha1.DatabaseCluster) error

type DatabaseClusterController struct {
	Watches    watches
	Plan       Plan
	Validators []Validator
}

type watches struct {
	owned         []client.Object
	customWatches []customWatch
}

type requestMapper func(client.Object) []reconcile.Request
type customWatch struct {
	obj           client.Object
	labelSelector map[string]string
	requestMapper requestMapper
}

type Plan struct {
	Sync         SyncPlan
	Cleanup      CleanupPlan
	StatusGetter StatusGetter
}

type StatusGetter func() (v2alpha1.DatabaseClusterStatus, error)

type SyncPlan struct {
	Steps []Step
}

type CleanupPlan struct {
	Steps []Step
}

type StepResult struct {
	Done         bool
	RequeueAfter time.Duration
	Message      string
}

type StepFn func(context.Context, client.Client, v2alpha1.DatabaseCluster) (StepResult, error)

type Step struct {
	Description string
	Wait        bool
	Fn          StepFn
}
