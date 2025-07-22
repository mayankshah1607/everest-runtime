package controller

import (
	"context"
	"time"

	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// ControllerBuilder implements the builder pattern for DatabaseClusterController
type ControllerBuilder struct {
	controller *DatabaseClusterController
}

// NewControllerBuilder creates a new builder for DatabaseClusterController
func NewControllerBuilder() *ControllerBuilder {
	return &ControllerBuilder{
		controller: &DatabaseClusterController{
			watches: &watches{
				owned:         []client.Object{},
				customWatches: []customWatch{},
			},
			plan: &plan{
				sync:         &syncPlan{},
				cleanup:      &cleanupPlan{},
				statusGetter: nil,
			},
			validators: []validator{},
		},
	}
}

// WithEngineName sets the engine name for the controller
func (b *ControllerBuilder) WithEngineName(name string) *ControllerBuilder {
	b.controller.engineName = name
	return b
}

// Owns registers objects to be watched that are owned by the controller
func (b *ControllerBuilder) Owns(obj client.Object) *ControllerBuilder {
	b.controller.watches.owned = append(b.controller.watches.owned, obj)
	return b
}

type WithWatchOptions struct {
	LabelSelector map[string]string
}

// WithWatch adds a custom watch with a request mapper function
func (b *ControllerBuilder) WithWatch(obj client.Object, fn handler.MapFunc, opts WithWatchOptions) *ControllerBuilder {
	b.controller.watches.customWatches = append(b.controller.watches.customWatches, customWatch{
		obj:           obj,
		labelSelector: opts.LabelSelector,
		requestMapper: fn,
	})
	return b
}

// WithValidator adds a validation function to the controller
func (b *ControllerBuilder) WithValidator(fn validator) *ControllerBuilder {
	b.controller.validators = append(b.controller.validators, fn)
	return b
}

// WithSyncStep adds a synchronization step to the controller
func (b *ControllerBuilder) WithSyncStep(description string, fn stepFn, wait bool) *ControllerBuilder {
	b.controller.plan.sync.steps = append(b.controller.plan.sync.steps, step{
		Description: description,
		Wait:        wait,
		Fn:          fn,
	})
	return b
}

// WithCleanupStep adds a cleanup step to the controller
func (b *ControllerBuilder) WithCleanupStep(description string, fn stepFn, wait bool) *ControllerBuilder {
	b.controller.plan.cleanup.steps = append(b.controller.plan.cleanup.steps, step{
		Description: description,
		Wait:        wait,
		Fn:          fn,
	})
	return b
}

// WithStatusGetter sets the status getter function for the controller
func (b *ControllerBuilder) WithStatusGetter(fn statusGetter) *ControllerBuilder {
	b.controller.plan.statusGetter = fn
	return b
}

// Build finalizes the controller configuration and returns the controller
func (b *ControllerBuilder) Build() *DatabaseClusterController {
	return b.controller
}

// Type definitions

type validator func(obj *v2alpha1.DatabaseCluster) error

type DatabaseClusterController struct {
	engineName string
	watches    *watches
	plan       *plan
	validators []validator
}

// GetEngineName returns the engine name
func (c *DatabaseClusterController) GetEngineName() string {
	return c.engineName
}

// GetWatches returns the watches configuration
func (c *DatabaseClusterController) GetWatches() *watches {
	return c.watches
}

// GetPlan returns the reconciliation plan
func (c *DatabaseClusterController) GetPlan() *plan {
	return c.plan
}

// GetValidators returns the validators
func (c *DatabaseClusterController) GetValidators() []validator {
	return c.validators
}

type watches struct {
	owned         []client.Object
	customWatches []customWatch
}

// GetOwned returns the owned objects
func (w *watches) GetOwned() []client.Object {
	return w.owned
}

// GetCustomWatches returns the custom watches
func (w *watches) GetCustomWatches() []customWatch {
	return w.customWatches
}

type customWatch struct {
	obj           client.Object
	labelSelector map[string]string
	requestMapper handler.MapFunc
}

// GetObject returns the watched object
func (w *customWatch) GetObject() client.Object {
	return w.obj
}

// GetLabelSelector returns the label selector
func (w *customWatch) GetLabelSelector() map[string]string {
	return w.labelSelector
}

// GetRequestMapper returns the request mapper function
func (w *customWatch) GetRequestMapper() handler.MapFunc {
	return w.requestMapper
}

type plan struct {
	sync         *syncPlan
	cleanup      *cleanupPlan
	statusGetter statusGetter
}

// GetSync returns the sync plan
func (p *plan) GetSync() *syncPlan {
	return p.sync
}

// GetCleanup returns the cleanup plan
func (p *plan) GetCleanup() *cleanupPlan {
	return p.cleanup
}

// GetStatusGetter returns the status getter function
func (p *plan) GetStatusGetter() statusGetter {
	return p.statusGetter
}

type statusGetter func(context.Context, client.Client, *v2alpha1.DatabaseCluster) (v2alpha1.DatabaseClusterStatus, error)

type syncPlan struct {
	steps []step
}

// GetSteps returns the sync steps
func (s *syncPlan) GetSteps() []step {
	return s.steps
}

type cleanupPlan struct {
	steps []step
}

// GetSteps returns the cleanup steps
func (c *cleanupPlan) GetSteps() []step {
	return c.steps
}

type StepResult struct {
	Done         bool
	RequeueAfter time.Duration
	Message      string
}

type stepFn func(context.Context, client.Client, *v2alpha1.DatabaseCluster) (StepResult, error)

type step struct {
	Description string
	Wait        bool
	Fn          stepFn
}
