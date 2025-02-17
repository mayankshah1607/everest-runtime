package clickhouse

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlekSi/pointer"
	chv1 "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse.altinity.com/v1"
	"github.com/mayankshah1607/everest-runtime/pkg/apis/v2alpha1"
	"github.com/mayankshah1607/everest-runtime/pkg/controller"
	corev1 "k8s.io/api/core/v1"
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
	chi := &chv1.ClickHouseInstallation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db.GetName(),
			Namespace: db.GetNamespace(),
		},
	}
	if err := createDefaultUserSecret(ctx, c, db); err != nil {
		return reconcile.Result{}, err
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, c, chi, func() error {
		return p.reconcileCHI(ctx, c, db, chi)
	}); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
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

func (p *databaseClusterImpl) reconcileCHI(
	ctx context.Context,
	c client.Client,
	db *v2alpha1.DatabaseCluster,
	chi *chv1.ClickHouseInstallation,
) error {
	var clusterCmp *v2alpha1.ComponentSpec
	// TODO: we need to validate somehow before even creating this object.
	if cmps := db.GetComponentsOfType("clickhouse"); len(cmps) != 1 {
		return errors.New("invalid number of clickhouse components")
	} else {
		clusterCmp = &cmps[0]
	}

	// configure users.
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

	cluster := &chv1.Cluster{
		Name: clusterCmp.Name,
	}

	if clusterCmp.Shards != nil {
		cluster.Layout.ShardsCount = int(*clusterCmp.Shards)
	}
	if clusterCmp.Replicas != nil {
		cluster.Layout.ReplicasCount = int(*clusterCmp.Replicas)
	}

	// configure volume claims
	vcts := []corev1.PersistentVolumeClaim{}
	vcts = append(vcts, corev1.PersistentVolumeClaim{
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
			StorageClassName: pointer.To(clusterCmp.Storage.StorageClass),
		},
	})
	vcts = append(vcts, clusterCmp.PodSpec.AdditionalVolumeClaimTemplates...)
	chi.Spec.Templates.VolumeClaimTemplates = intoCHVolumeClaim(vcts)

	// configure the default container.
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

	if err := controllerutil.SetControllerReference(db, chi, p.schema); err != nil {
		return err
	}

	return nil
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
