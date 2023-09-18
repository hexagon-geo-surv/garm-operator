// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"flag"
	"os"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	garmoperatorv1alpha1 "github.com/mercedes-benz/garm-operator/api/v1alpha1"
	"github.com/mercedes-benz/garm-operator/internal/controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(garmoperatorv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
		syncPeriod           time.Duration

		watchNamespace string

		garmServer   string
		garmUsername string
		garmPassword string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.DurationVar(&syncPeriod, "sync-period", 5*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 15m)")

	flag.StringVar(&watchNamespace, "namespace", "",
		"Namespace that the controller watches to reconcile garm objects. If unspecified, the controller watches for garm objects across all namespaces.")

	flag.StringVar(&garmServer, "garm-server", "", "The address of the GARM server")
	flag.StringVar(&garmUsername, "garm-username", "", "The username for the GARM server")
	flag.StringVar(&garmPassword, "garm-password", "", "The password for the GARM server")

	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	ctrl.SetLogger(klogr.New())

	// configure garm client from environment variables
	if len(os.Getenv("GARM_SERVER")) > 0 {
		setupLog.Info("Using garm-server from environment variable")
		garmServer = os.Getenv("GARM_SERVER")
	}
	if len(os.Getenv("GARM_USERNAME")) > 0 {
		setupLog.Info("Using garm-username from environment variable")
		garmUsername = os.Getenv("GARM_USERNAME")
	}
	if len(os.Getenv("GARM_PASSWORD")) > 0 {
		setupLog.Info("Using garm-password from environment variable")
		garmPassword = os.Getenv("GARM_PASSWORD")
	}
	if len(os.Getenv("WATCH_NAMESPACE")) > 0 {
		setupLog.Info("using watch-namespace from environment variable")
		watchNamespace = os.Getenv("WATCH_NAMESPACE")
	}

	var watchNamespaces []string
	if watchNamespace != "" {
		watchNamespaces = []string{watchNamespace}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "b608d8b3.mercedes-benz.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
		//
		// Default Sync Period = 10 hours.
		// Set default via flag to 5 minutes
		SyncPeriod: &syncPeriod,
		Cache: cache.Options{
			Namespaces: watchNamespaces,
			SyncPeriod: &syncPeriod,
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if garmServer == "" {
		setupLog.Error(errors.New("unable to fetch garm server from either flag or os_env"), "unable to start manager")
		os.Exit(1)
	}
	if garmUsername == "" {
		setupLog.Error(errors.New("unable to fetch garm username from either flag or os_env"), "unable to start manager")
		os.Exit(1)
	}
	if garmPassword == "" {
		setupLog.Error(errors.New("unable to fetch garm password from either flag or os_env"), "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.EnterpriseReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("enterprise-controller"),

		BaseURL:  garmServer,
		Username: garmUsername,
		Password: garmPassword,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Enterprise")
		os.Exit(1)
	}
	if err = (&controller.PoolReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("pool-controller"),

		BaseURL:  garmServer,
		Username: garmUsername,
		Password: garmPassword,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pool")
		os.Exit(1)
	}

	if os.Getenv("CREATE_WEBHOOK") == "true" {
		if err = (&garmoperatorv1alpha1.Pool{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Pool")
			os.Exit(1)
		}
		if err = (&garmoperatorv1alpha1.Image{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Image")
			os.Exit(1)
		}
		if err = (&garmoperatorv1alpha1.Repository{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Repository")
			os.Exit(1)
		}
	}

	if err = (&controller.OrganizationReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("organization-controller"),

		BaseURL:  garmServer,
		Username: garmUsername,
		Password: garmPassword,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Organization")
		os.Exit(1)
	}

	if err = (&controller.RepositoryReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("repository-controller"),

		BaseURL:  garmServer,
		Username: garmUsername,
		Password: garmPassword,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Repository")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}