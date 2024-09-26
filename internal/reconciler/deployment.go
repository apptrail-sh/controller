package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	appVersionMetricName = "apptrail_app_version"
)

var (
	appVersionGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: appVersionMetricName,
		Help: "App version for a given deployment",
	}, []string{
		"namespace",
		"app",
		"previous_version",
		"current_version",
		"last_updated",
	})
)

type AppVersion struct {
	PreviousVersion string
	CurrentVersion  string
	LastUpdated     time.Time
}

type DeploymentReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Recorder           record.EventRecorder
	deploymentVersions map[string]AppVersion
}

func NewDeploymentReconciler(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) *DeploymentReconciler {
	metrics.Registry.MustRegister(appVersionGauge)
	return &DeploymentReconciler{
		Client:             client,
		Scheme:             scheme,
		Recorder:           recorder,
		deploymentVersions: make(map[string]AppVersion),
	}
}

func (dr *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling Deployment")

	resource := &v1.Deployment{}
	if err := dr.Get(ctx, req.NamespacedName, resource); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Deployment found", "Deployment", resource)

	appkey := req.Namespace + "/" + req.Name
	stored := dr.deploymentVersions[appkey]

	versionLabel := resource.Labels["app.kubernetes.io/version"]
	if versionLabel == "" {
		log.Info("Deployment version label not found",
			"Deployment", fmt.Sprintf("%s/%s", req.Namespace, req.Name))
		return ctrl.Result{}, nil
	}

	if stored.CurrentVersion != versionLabel {
		newAppVer := AppVersion{
			PreviousVersion: stored.CurrentVersion,
			CurrentVersion:  versionLabel,
			LastUpdated:     time.Now(),
		}
		dr.deploymentVersions[appkey] = newAppVer

		timeFormatted := newAppVer.LastUpdated.Format(time.RFC3339)

		appVersionGauge.WithLabelValues(
			resource.Namespace,
			resource.Name,
			newAppVer.PreviousVersion,
			newAppVer.CurrentVersion,
			timeFormatted).Set(1)
		log.Info("Deployment version updated", "Deployment", resource)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (dr *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Deployment{}).
		Complete(dr)
}
