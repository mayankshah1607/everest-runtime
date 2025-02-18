package clickhouse

import (
	"context"
	"errors"
	"fmt"

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
	return srcs
}

func (p *databaseClusterImpl) Reconcile(ctx context.Context, c client.Client, db *v2alpha1.DatabaseCluster) (reconcile.Result, error) {
	desired, err := p.getDesiredCHI(db)
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := createDefaultUserSecret(ctx, c, db); err != nil {
		return reconcile.Result{}, err
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
			return reconcile.Result{}, c.Create(ctx, desired)
		}
		return reconcile.Result{}, err
	}

	existing.Spec = desired.Spec
	existing.ObjectMeta.SetLabels(desired.ObjectMeta.GetLabels())
	existing.ObjectMeta.SetAnnotations(desired.ObjectMeta.GetAnnotations())
	return reconcile.Result{}, c.Update(ctx, existing)
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
		return v2alpha1.DatabaseClusterStatus{}, err
	}

	// TODO
	sts := v2alpha1.DatabaseClusterStatus{
		Phase: v2alpha1.DatabaseClusterPhase(chi.Status.Status),
	}

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
