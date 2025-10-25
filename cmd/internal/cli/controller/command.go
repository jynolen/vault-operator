/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package controller

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	vaultoperatoriov1alpha1 "github.com/jynolen/vault-operator/api/v1alpha1"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/jynolen/vault-operator/internal/controller"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type ControllerCliConfig struct {
	MetricsAddr          string
	MetricsCertPath      string
	MetricsCertName      string
	MetricsCertKey       string
	WebhookCertPath      string
	WebhookCertName      string
	WebhookCertKey       string
	ProbeAddr            string
	EnableHTTP2          bool
	SecureMetrics        bool
	EnableLeaderElection bool
	LeaderElectionID     string
	zapOption            zap.Options
	tlsOptions           []func(*tls.Config)
}

// controllerCmd represents the controller command
var (
	config = ControllerCliConfig{
		zapOption: zap.Options{
			Development: true,
		},
		LeaderElectionID: "09ba50ab.vault-operator.io",
	}
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func disableHTTP2(c *tls.Config) {
	setupLog.Info("disabling http/2")
	c.NextProtos = []string{"http/1.1"}
}

func initCmd(controllerCmd *cobra.Command) {
	flags := controllerCmd.Flags()
	flags.BoolVarP(&config.EnableLeaderElection, "leader-elect", "", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flags.BoolVarP(&config.SecureMetrics, "metrics-secure", "", true, "If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flags.BoolVarP(&config.EnableHTTP2, "enable-http2", "", false, "If set, HTTP/2 will be enabled for the metrics and webhook servers")

	flags.StringVarP(&config.MetricsAddr, "metrics-bind-address", "", "0", "The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flags.StringVarP(&config.ProbeAddr, "health-probe-bind-address", "", ":8081", "The address the probe endpoint binds to.")
	flags.StringVarP(&config.WebhookCertPath, "webhook-cert-path", "", "", "The directory that contains the webhook certificate.")
	flags.StringVarP(&config.WebhookCertName, "webhook-cert-name", "", "tls.crt", "The name of the webhook certificate file.")
	flags.StringVarP(&config.WebhookCertKey, "webhook-cert-key", "", "tls.key", "The name of the webhook key file.")
	flags.StringVarP(&config.MetricsCertPath, "metrics-cert-path", "", "", "The directory that contains the metrics server certificate.")
	flags.StringVarP(&config.MetricsCertName, "metrics-cert-name", "", "", "The name of the metrics server certificate file.")
	flags.StringVarP(&config.MetricsCertKey, "metrics-cert-key", "", "", "The name of the metrics server key file.")

	flagSet := &flag.FlagSet{}
	config.zapOption.BindFlags(flagSet)
	flags.AddGoFlagSet(flagSet)

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(vaultoperatoriov1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func initWebHookServer() (webhook.Server, *certwatcher.CertWatcher) {
	if len(config.WebhookCertPath) == 0 {
		return webhook.NewServer(webhook.Options{
			TLSOpts: config.tlsOptions,
		}), nil
	}
	setupLog.Info("Initializing webhook certificate watcher using provided certificates",
		"webhook-cert-path", config.WebhookCertPath, "webhook-cert-name", config.WebhookCertName, "webhook-cert-key", config.WebhookCertKey)

	var err error
	var webhookCertWatcher *certwatcher.CertWatcher

	webhookCertWatcher, err = certwatcher.New(
		filepath.Join(config.WebhookCertPath, config.WebhookCertName),
		filepath.Join(config.WebhookCertPath, config.WebhookCertKey),
	)
	if err != nil {
		setupLog.Error(err, "Failed to initialize webhook certificate watcher")
		os.Exit(1)
	}

	return webhook.NewServer(webhook.Options{
		TLSOpts: append(config.tlsOptions, func(config *tls.Config) {
			config.GetCertificate = webhookCertWatcher.GetCertificate
		}),
	}), webhookCertWatcher
}

func initMetricsServerOptions() (metricsserver.Options, *certwatcher.CertWatcher) {
	metricsServerOptions := metricsserver.Options{
		BindAddress:   config.MetricsAddr,
		SecureServing: config.SecureMetrics,
		TLSOpts:       config.tlsOptions,
	}
	if config.SecureMetrics {
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	if len(config.MetricsCertPath) == 0 {
		return metricsServerOptions, nil
	}

	setupLog.Info("Initializing metrics certificate watcher using provided certificates",
		"metrics-cert-path", config.MetricsCertPath, "metrics-cert-name", config.MetricsCertName, "metrics-cert-key", config.MetricsCertKey)

	metricsCertWatcher, err := certwatcher.New(
		filepath.Join(config.MetricsCertPath, config.MetricsCertName),
		filepath.Join(config.MetricsCertPath, config.MetricsCertKey),
	)
	if err != nil {
		setupLog.Error(err, "to initialize metrics certificate watcher", "error", err)
		os.Exit(1)
	}

	metricsServerOptions.TLSOpts = append(metricsServerOptions.TLSOpts, func(config *tls.Config) {
		config.GetCertificate = metricsCertWatcher.GetCertificate
	})

	return metricsServerOptions, metricsCertWatcher
}

func registerWatcher(mgr *manager.Manager, watcher *certwatcher.CertWatcher, name string) {
	if watcher != nil {
		setupLog.Info(fmt.Sprintf("Adding %s certificate watcher to manager", name))
		if err := (*mgr).Add(watcher); err != nil {
			setupLog.Error(err, fmt.Sprintf("unable to add %s certificate watcher to manager", name))
			os.Exit(1)
		}
	}
}

func controllerMain(cmd *cobra.Command, args []string) {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&config.zapOption)))
	if !config.EnableHTTP2 {
		config.tlsOptions = append(config.tlsOptions, disableHTTP2)
	}

	webhookServer, webhookCertWatcher := initWebHookServer()
	metricsServerOptions, metricsCertWatcher := initMetricsServerOptions()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: config.ProbeAddr,
		LeaderElection:         config.EnableLeaderElection,
		LeaderElectionID:       config.LeaderElectionID,
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
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	if err := (&controller.VaultServerReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("vault-operator"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VaultServer")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	registerWatcher(&mgr, metricsCertWatcher, "metrics")
	registerWatcher(&mgr, webhookCertWatcher, "webhook")

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

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Execute the controller part of vault-operator",
		Run:   controllerMain,
	}
	initCmd(cmd)
	return cmd
}
