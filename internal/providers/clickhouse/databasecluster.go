package clickhouse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	chkv1 "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse-keeper.altinity.com/v1"
	chv1 "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse.altinity.com/v1"
	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type CustomCHConfig struct {
	Zookeeper *chv1.ZookeeperConfig `json:"zookeeper,omitempty" yaml:"zookeeper,omitempty"`
}

type databaseClusterImpl struct {
	schema *runtime.Scheme
}

func (p *databaseClusterImpl) GetSources(m manager.Manager) []source.Source {
	srcs := []source.Source{}

	// Everest will watch the ClickHouseInstallation CR and enqueue DatabaseClusters with the same name.
	srcs = append(srcs, source.Kind(
		m.GetCache(),
		&chv1.ClickHouseInstallation{},
		&handler.TypedEnqueueRequestForObject[*chv1.ClickHouseInstallation]{}))

	srcs = append(srcs, source.Kind(
		m.GetCache(),
		&chkv1.ClickHouseKeeperInstallation{},
		&handler.TypedEnqueueRequestForObject[*chkv1.ClickHouseKeeperInstallation]{}))
	return srcs
}

func (p *databaseClusterImpl) Reconcile(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (reconcile.Result, error) {
	// in this PoC, we are providing the user info thorugh a Secret.
	// But some operators support fetching users from external sources like Vault.
	if err := createDefaultUserSecret(ctx, c, db); err != nil {
		return reconcile.Result{}, err
	}

	if done, err := p.reconcileClickhouseKeeper(ctx, c, db); err != nil {
		return reconcile.Result{}, err
	} else if !done {
		return reconcile.Result{Requeue: true}, nil
	}

	if err := p.reconcileClickhouse(ctx, c, db); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (p *databaseClusterImpl) reconcileClickhouseKeeper(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (bool, error) {
	components := db.GetComponentsOfType("clickhouse-keeper")
	if len(components) == 0 {
		return true, nil
	}

	desired := p.getDesiredCHK(db.GetName(), db.GetNamespace(), &components[0])
	if err := controllerutil.SetControllerReference(db, desired, p.schema); err != nil {
		return false, err
	}

	existing := &chkv1.ClickHouseKeeperInstallation{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      db.GetName(),
		Namespace: db.GetNamespace(),
	}, existing); err != nil {
		if k8serrors.IsNotFound(err) {
			return false, c.Create(ctx, desired)
		}
		return false, err
	}

	existing.Spec = desired.Spec
	existing.ObjectMeta.SetLabels(desired.ObjectMeta.GetLabels())
	existing.ObjectMeta.SetAnnotations(desired.ObjectMeta.GetAnnotations())
	return existing.Status.Status == chkv1.StatusCompleted, c.Update(ctx, existing)
}

func (p *databaseClusterImpl) getDesiredCHK(name, namespace string, cmp *v2alpha1.ComponentSpec) *chkv1.ClickHouseKeeperInstallation {
	chk := &chkv1.ClickHouseKeeperInstallation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	chk.Spec.Templates = chv1.NewTemplates()
	chk.Spec.Configuration = chkv1.NewConfiguration()

	// configure VCTs
	vcts := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: dataVolumeName,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: cmp.Storage.Size,
					},
				},
				StorageClassName: cmp.Storage.StorageClass,
			},
		},
	}
	vcts = append(vcts, cmp.PodSpec.AdditionalVolumeClaimTemplates...)
	chk.Spec.Templates.VolumeClaimTemplates = intoCHVolumeClaim(vcts)

	// configure pod template
	mainContainer := cmp.PodSpec.Container
	containers := []corev1.Container{*mainContainer}
	containers = append(containers, cmp.PodSpec.Sidecars...)
	chk.Spec.Templates.PodTemplates = []chv1.PodTemplate{
		{
			Name: defaultPodTemplateName,
			Spec: corev1.PodSpec{
				Containers: containers,
			},
		},
	}

	// configure cluster
	cluster := &chkv1.Cluster{
		Name: cmp.Name,
		Templates: &chv1.TemplatesList{
			PodTemplate:         defaultPodTemplateName,
			VolumeClaimTemplate: dataVolumeName,
		},
	}
	if cmp.Shards != nil || cmp.Replicas != nil {
		cluster.Layout = &chkv1.ChkClusterLayout{}
		if cmp.Shards != nil {
			cluster.Layout.ShardsCount = int(*cmp.Shards)
		}
		if cmp.Replicas != nil {
			cluster.Layout.ReplicasCount = int(*cmp.Replicas)
		}
	}
	chk.Spec.Configuration.Clusters = []*chkv1.Cluster{cluster}
	return chk
}

func (p *databaseClusterImpl) reconcileClickhouse(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) error {
	desired, err := p.getDesiredCHI(db)
	if err != nil {
		return err
	}

	// We should NOT use controllerutil.CreateOrUpdate here
	// because chv1.ClickHouseInstallation contains private fields.
	// controllerutil.CreateOrUpdate uses the DeepCopy() method which panics
	// when it encounters private fields.

	existing := &chv1.ClickHouseInstallation{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      db.GetName(),
		Namespace: db.GetNamespace(),
	}, existing); err != nil {
		if k8serrors.IsNotFound(err) {
			return c.Create(ctx, desired)
		}
		return err
	}

	existing.Spec = desired.Spec
	existing.ObjectMeta.SetLabels(desired.ObjectMeta.GetLabels())
	existing.ObjectMeta.SetAnnotations(desired.ObjectMeta.GetAnnotations())
	return c.Update(ctx, existing)
}

func (p *databaseClusterImpl) Delete(context.Context, client.Client, *v2alpha1.DatabaseCluster) (bool, error) {
	return false, nil
}

func (p databaseClusterImpl) GetStatus(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (v2alpha1.DatabaseClusterStatus, error) {
	chi := &chv1.ClickHouseInstallation{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      db.GetName(),
		Namespace: db.GetNamespace(),
	}, chi,
	); err != nil {
		return v2alpha1.DatabaseClusterStatus{}, client.IgnoreNotFound(err)
	}

	sts := v2alpha1.DatabaseClusterStatus{
		Components: []v2alpha1.ComponentStatus{},
	}
	switch chi.Status.Status {
	case chv1.StatusCompleted:
		sts.Phase = v2alpha1.DatabaseClusterPhaseRunning
	case chv1.StatusAborted:
		sts.Phase = v2alpha1.DatabaseClusterPhaseFailed
	case chv1.StatusInProgress:
		sts.Phase = v2alpha1.DatabaseClusterPhaseCreating
	case chv1.StatusTerminating:
		sts.Phase = v2alpha1.DatabaseClusterPhaseDeleting
	}

	chiPods := []corev1.LocalObjectReference{}
	for _, pod := range chi.Status.Pods {
		chiPods = append(chiPods, corev1.LocalObjectReference{
			Name: pod,
		})
	}
	sts.Components = append(sts.Components, v2alpha1.ComponentStatus{
		Pods: chiPods,
	})

	return sts, nil
}

func (p databaseClusterImpl) GetDefaultCredentials(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (*controller.Credentials, error) {
	secret := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      db.GetName() + "-admin-password",
		Namespace: db.GetNamespace(),
	}, secret); err != nil {
		return nil, err
	}
	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	return &controller.Credentials{
		Username: username,
		Password: password,
	}, nil
}

const (
	dataVolumeName         = "data"
	defaultPodTemplateName = "clickhouse-default"
)

func (p *databaseClusterImpl) getDesiredCHI(db *v2alpha1.DatabaseCluster) (*chv1.ClickHouseInstallation, error) {
	chi := &chv1.ClickHouseInstallation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.GetName(),
			Namespace: db.GetNamespace(),
		},
		Spec: chv1.ChiSpec{
			Configuration: chv1.NewConfiguration(),
			Templates:     chv1.NewTemplates(),
		},
	}

	clusterCmp, err := p.getCHCmp(db)
	if err != nil {
		return nil, err
	}

	parsedCustomSpec := &CustomCHConfig{}
	if customSpec := clusterCmp.CustomSpec; customSpec != nil {
		if err := json.Unmarshal(customSpec.Raw, parsedCustomSpec); err != nil {
			return nil, err
		}
	}

	p.configureZookeeperNodes(chi, parsedCustomSpec)
	p.configureUsers(chi, db)
	cluster := p.configureCluster(clusterCmp)
	p.configureVolumeClaims(chi, clusterCmp)
	p.configurePodTemplate(chi, clusterCmp)

	cluster.Templates = chv1.NewTemplatesList()
	cluster.Templates.PodTemplate = defaultPodTemplateName
	chi.Spec.Configuration.Clusters = []*chv1.Cluster{cluster}

	if err := controllerutil.SetControllerReference(db, chi, p.schema); err != nil {
		return nil, err
	}

	return chi, nil
}

func (p *databaseClusterImpl) initializeCHI(db *v2alpha1.DatabaseCluster) *chv1.ClickHouseInstallation {
	return &chv1.ClickHouseInstallation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.GetName(),
			Namespace: db.GetNamespace(),
		},
		Spec: chv1.ChiSpec{
			Configuration: chv1.NewConfiguration(),
			Templates:     chv1.NewTemplates(),
		},
	}
}

func (p *databaseClusterImpl) getCHCmp(db *v2alpha1.DatabaseCluster) (*v2alpha1.ComponentSpec, error) {
	cmps := db.GetComponentsOfType("clickhouse")
	if len(cmps) != 1 {
		return nil, errors.New("invalid number of clickhouse components")
	}
	return &cmps[0], nil
}

func (p *databaseClusterImpl) configureUsers(chi *chv1.ClickHouseInstallation, db *v2alpha1.DatabaseCluster) {
	userSetting := chv1.NewSettingSource(&chv1.SettingSource{
		ValueFrom: &chv1.DataSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: db.GetName() + "-admin-password",
				},
				Key: "password",
			},
		},
	})
	chi.Spec.Configuration.Users = chv1.NewSettings()
	chi.Spec.Configuration.Users.Set(fmt.Sprintf("%s/password", defaultUser), userSetting)
}

func (p *databaseClusterImpl) configureCluster(clusterCmp *v2alpha1.ComponentSpec) *chv1.Cluster {
	cluster := &chv1.Cluster{
		Name: clusterCmp.Name,
	}

	if clusterCmp.Shards != nil || clusterCmp.Replicas != nil {
		cluster.Layout = &chv1.ChiClusterLayout{}
		if clusterCmp.Shards != nil {
			cluster.Layout.ShardsCount = int(*clusterCmp.Shards)
		}
		if clusterCmp.Replicas != nil {
			cluster.Layout.ReplicasCount = int(*clusterCmp.Replicas)
		}
	}

	return cluster
}

func (p *databaseClusterImpl) configureVolumeClaims(chi *chv1.ClickHouseInstallation, clusterCmp *v2alpha1.ComponentSpec) {
	vcts := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: dataVolumeName,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: clusterCmp.Storage.Size,
					},
				},
				StorageClassName: clusterCmp.Storage.StorageClass,
			},
		},
	}
	vcts = append(vcts, clusterCmp.PodSpec.AdditionalVolumeClaimTemplates...)
	chi.Spec.Templates.VolumeClaimTemplates = intoCHVolumeClaim(vcts)
}

func (p *databaseClusterImpl) configurePodTemplate(chi *chv1.ClickHouseInstallation, clusterCmp *v2alpha1.ComponentSpec) {
	container := p.configureContainer(clusterCmp)
	containers := []corev1.Container{container}
	containers = append(containers, clusterCmp.PodSpec.Sidecars...)

	chi.Spec.Templates.PodTemplates = []chv1.PodTemplate{
		{
			Name: defaultPodTemplateName,
			ObjectMeta: metav1.ObjectMeta{
				Labels:      clusterCmp.PodSpec.Labels,
				Annotations: clusterCmp.PodSpec.Annotations,
			},
			Spec: corev1.PodSpec{
				Containers: containers,
			},
		},
	}
}

func (p *databaseClusterImpl) configureContainer(clusterCmp *v2alpha1.ComponentSpec) corev1.Container {
	var container corev1.Container
	if clusterCmp.PodSpec.Container != nil {
		container = *clusterCmp.PodSpec.Container
	}
	if container.Name == "" {
		container.Name = "clickhouse"
	}
	if clusterCmp.Image != "" {
		container.Image = clusterCmp.Image
	}
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      dataVolumeName,
		MountPath: "/var/lib/clickhouse",
	})
	return container
}

func intoCHVolumeClaim(in []corev1.PersistentVolumeClaim) []chv1.VolumeClaimTemplate {
	result := make([]chv1.VolumeClaimTemplate, 0, len(in))
	for _, pvc := range in {
		result = append(result, chv1.VolumeClaimTemplate{
			Name: pvc.Name,
			Spec: pvc.Spec,
		})
	}
	return result
}

const defaultUser = "admin"

// TODO: should reconcile
func createDefaultUserSecret(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.GetName() + "-admin-password",
			Namespace: db.GetNamespace(),
		},
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("admin"), // TODO: randomize
		},
	}
	if err := c.Create(ctx, secret); err != nil {
		return client.IgnoreAlreadyExists(err)
	}
	return nil
}

func (p *databaseClusterImpl) configureZookeeperNodes(chi *chv1.ClickHouseInstallation, parsedCustomSpec *CustomCHConfig) {
	if parsedCustomSpec.Zookeeper == nil {
		return
	}

	zkc := parsedCustomSpec.Zookeeper
	if zkc.IsEmpty() {
		return
	}
	chi.Spec.Configuration.Zookeeper = zkc
}
