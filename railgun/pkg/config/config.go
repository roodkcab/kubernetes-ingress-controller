package config

import (
	"context"
	"fmt"

	"github.com/kong/go-kong/kong"
	"github.com/kong/kubernetes-ingress-controller/pkg/adminapi"
	"github.com/kong/kubernetes-ingress-controller/pkg/admission"
	"github.com/kong/kubernetes-ingress-controller/pkg/annotations"
	"github.com/kong/kubernetes-ingress-controller/pkg/util"
	"github.com/kong/kubernetes-ingress-controller/railgun/internal/proxy"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// -----------------------------------------------------------------------------
// Controller Manager - Config
// -----------------------------------------------------------------------------

// Config collects all configuration that the controller manager takes from the environment.
type Config struct {
	// See flag definitions in RegisterFlags(...) for documentation of the fields defined here.

	// Logging configurations
	LogLevel  string
	LogFormat string

	// Kong high-level controller manager configurations
	KongAdminAPIConfig adminapi.HTTPClientOpts
	KongAdminToken     string
	KongStateEnabled   util.EnablementStatus
	KongWorkspace      string
	AnonymousReports   bool
	EnableReverseSync  bool

	// Kong Proxy configurations
	APIServerHost            string
	MetricsAddr              string
	ProbeAddr                string
	KongAdminURL             string
	ProxySyncSeconds         float32
	KongCustomEntitiesSecret string

	// Kubernetes configurations
	KubeconfigPath       string
	IngressClassName     string
	EnableLeaderElection bool
	LeaderElectionID     string
	Concurrency          int
	FilterTags           []string
	WatchNamespace       string

	// Kubernetes API toggling
	IngressExtV1beta1Enabled util.EnablementStatus
	IngressNetV1beta1Enabled util.EnablementStatus
	IngressNetV1Enabled      util.EnablementStatus
	UDPIngressEnabled        util.EnablementStatus
	TCPIngressEnabled        util.EnablementStatus
	KongIngressEnabled       util.EnablementStatus
	KongClusterPluginEnabled util.EnablementStatus
	KongPluginEnabled        util.EnablementStatus
	KongConsumerEnabled      util.EnablementStatus
	ServiceEnabled           util.EnablementStatus

	// Admission Webhook server config
	AdmissionServer admission.ServerConfig
}

// -----------------------------------------------------------------------------
// Controller Manager - Config - Constants & Vars
// -----------------------------------------------------------------------------

// onOffUsage is used to indicate what textual options are used to enable or disable a feature.
const onOffUsage = "Can be one of [enabled, disabled]."

// -----------------------------------------------------------------------------
// Controller Manager - Config - Methods
// -----------------------------------------------------------------------------

// FlagSet binds the provided Config to commandline flags.
func (c *Config) FlagSet() *pflag.FlagSet {

	flagSet := flagSet{*pflag.NewFlagSet("", pflag.ExitOnError)}

	// Logging configurations
	flagSet.StringVar(&c.LogLevel, "log-level", "info", `Level of logging for the controller. Allowed values are trace, debug, info, warn, error, fatal and panic.`)
	flagSet.StringVar(&c.LogFormat, "log-format", "text", `Format of logs of the controller. Allowed values are text and json.`)

	// Kong high-level controller manager configurations
	flagSet.BoolVar(&c.KongAdminAPIConfig.TLSSkipVerify, "kong-admin-tls-skip-verify", false, "Disable verification of TLS certificate of Kong's Admin endpoint.")
	flagSet.StringVar(&c.KongAdminAPIConfig.TLSServerName, "kong-admin-tls-server-name", "", "SNI name to use to verify the certificate presented by Kong in TLS.")
	flagSet.StringVar(&c.KongAdminAPIConfig.CACertPath, "kong-admin-ca-cert-file", "", `Path to PEM-encoded CA certificate file to verify Kong's Admin SSL certificate.`)
	flagSet.StringVar(&c.KongAdminAPIConfig.CACert, "kong-admin-ca-cert", "", `PEM-encoded CA certificate to verify Kong's Admin SSL certificate.`)
	flagSet.StringSliceVar(&c.KongAdminAPIConfig.Headers, "kong-admin-header", nil, `add a header (key:value) to every Admin API call, this flag can be used multiple times to specify multiple headers`)
	flagSet.StringVar(&c.KongAdminToken, "kong-admin-token", "", `The Kong Enterprise RBAC token used by the controller.`)
	flagSet.enablementStatusVar(&c.KongStateEnabled, "controller-kongstate", util.EnablementStatusEnabled, "Enable or disable the KongState controller. "+onOffUsage)
	flagSet.StringVar(&c.KongWorkspace, "kong-workspace", "", "Kong Enterprise workspace to configure. Leave this empty if not using Kong workspaces.")
	flagSet.BoolVar(&c.AnonymousReports, "anonymous-reports", true, `Send anonymized usage data to help improve Kong`)
	flagSet.BoolVar(&c.EnableReverseSync, "enable-reverse-sync", false, `Send configuration to Kong even if the configuration checksum has not changed since previous update.`)

	// Kong Proxy and Proxy Cache configurations
	flagSet.StringVar(&c.APIServerHost, "apiserver-host", "", `The Kubernetes API server URL. If not set, the controller will use cluster config discovery.`)
	flagSet.StringVar(&c.MetricsAddr, "metrics-bind-address", fmt.Sprintf(":%v", MetricsPort), "The address the metric endpoint binds to.")
	flagSet.StringVar(&c.ProbeAddr, "health-probe-bind-address", fmt.Sprintf(":%v", HealthzPort), "The address the probe endpoint binds to.")
	flagSet.StringVar(&c.KongAdminURL, "kong-admin-url", "http://localhost:8001", `The Kong Admin URL to connect to in the format "protocol://address:port".`)
	flagSet.Float32Var(&c.ProxySyncSeconds, "sync-rate-limit", proxy.DefaultSyncSeconds,
		fmt.Sprintf(
			"Define the rate (in seconds) in which configuration updates will be applied to the Kong Admin API. (default: %g seconds) (DEPRECATED, use --proxy-sync-seconds instead)",
			proxy.DefaultSyncSeconds,
		))
	flagSet.Float32Var(&c.ProxySyncSeconds, "proxy-sync-seconds", proxy.DefaultSyncSeconds,
		fmt.Sprintf(
			"Define the rate (in seconds) in which configuration updates will be applied to the Kong Admin API. (default: %g seconds)",
			proxy.DefaultSyncSeconds,
		))
	flagSet.StringVar(&c.KongCustomEntitiesSecret, "kong-custom-entities-secret", "", `A Secret containing custom entities for DB-less mode, in "namespace/name" format`)

	// Kubernetes configurations
	flagSet.StringVar(&c.KubeconfigPath, "kubeconfig", "", "Path to the kubeconfig file.")
	flagSet.StringVar(&c.IngressClassName, "ingress-class", annotations.DefaultIngressClass, `Name of the ingress class to route through this controller.`)
	flagSet.BoolVar(&c.EnableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flagSet.StringVar(&c.LeaderElectionID, "election-id", "5b374a9e.konghq.com", `Election id to use for status update.`)
	flagSet.StringSliceVar(&c.FilterTags, "kong-admin-filter-tag", []string{"managed-by-ingress-controller"}, "The tag used to manage and filter entities in Kong. This flag can be specified multiple times to specify multiple tags")
	flagSet.IntVar(&c.Concurrency, "kong-admin-concurrency", 10, "Max number of concurrent requests sent to Kong's Admin API.")
	flagSet.StringVar(&c.WatchNamespace, "watch-namespace", corev1.NamespaceAll,
		`Namespace(s) to watch for Kubernetes resources. Defaults to all namespaces. To watch multiple namespaces, use
		a comma-separated list of namespaces.`)

	// Kubernetes API toggling
	flagSet.enablementStatusVar(&c.IngressNetV1Enabled, "controller-ingress-networkingv1", util.EnablementStatusEnabled, "Enable or disable the Ingress controller (using API version networking.k8s.io/v1)."+onOffUsage)
	// TODO the other Ingress versions remain disabled for now. 2.x does not yet support version negotiation
	flagSet.enablementStatusVar(&c.IngressNetV1beta1Enabled, "controller-ingress-networkingv1beta1", util.EnablementStatusDisabled, "Enable or disable the Ingress controller (using API version networking.k8s.io/v1beta1)."+onOffUsage)
	flagSet.enablementStatusVar(&c.IngressExtV1beta1Enabled, "controller-ingress-extensionsv1beta1", util.EnablementStatusDisabled, "Enable or disable the Ingress controller (using API version extensions/v1beta1)."+onOffUsage)
	flagSet.enablementStatusVar(&c.UDPIngressEnabled, "controller-udpingress", util.EnablementStatusDisabled, "Enable or disable the UDPIngress controller. "+onOffUsage)
	flagSet.enablementStatusVar(&c.TCPIngressEnabled, "controller-tcpingress", util.EnablementStatusEnabled, "Enable or disable the TCPIngress controller. "+onOffUsage)
	flagSet.enablementStatusVar(&c.KongIngressEnabled, "controller-kongingress", util.EnablementStatusEnabled, "Enable or disable the KongIngress controller. "+onOffUsage)
	flagSet.enablementStatusVar(&c.KongClusterPluginEnabled, "controller-kongclusterplugin", util.EnablementStatusEnabled, "Enable or disable the KongClusterPlugin controller. "+onOffUsage)
	flagSet.enablementStatusVar(&c.KongPluginEnabled, "controller-kongplugin", util.EnablementStatusEnabled, "Enable or disable the KongPlugin controller. "+onOffUsage)
	flagSet.enablementStatusVar(&c.KongConsumerEnabled, "controller-kongconsumer", util.EnablementStatusEnabled, "Enable or disable the KongConsumer controller. "+onOffUsage)
	flagSet.enablementStatusVar(&c.ServiceEnabled, "controller-service", util.EnablementStatusEnabled, "Enable or disable the Service controller. "+onOffUsage)

	// Admission Webhook server config
	flagSet.StringVar(&c.AdmissionServer.ListenAddr, "admission-webhook-listen", "off",
		`The address to start admission controller on (ip:port).  Setting it to 'off' disables the admission controller.`)
	flagSet.StringVar(&c.AdmissionServer.CertPath, "admission-webhook-cert-file", "",
		`admission server PEM certificate file path; `+
			`if both this and the cert value is unset, defaults to `+admission.DefaultAdmissionWebhookCertPath)
	flagSet.StringVar(&c.AdmissionServer.KeyPath, "admission-webhook-key-file", "",
		`admission server PEM private key file path; `+
			`if both this and the key value is unset, defaults to `+admission.DefaultAdmissionWebhookKeyPath)
	flagSet.StringVar(&c.AdmissionServer.Cert, "admission-webhook-cert", "",
		`admission server PEM certificate value`)
	flagSet.StringVar(&c.AdmissionServer.Key, "admission-webhook-key", "",
		`admission server PEM private key value`)

	return &flagSet.FlagSet
}

func (c *Config) GetKongClient(ctx context.Context) (*kong.Client, error) {
	if c.KongAdminToken != "" {
		c.KongAdminAPIConfig.Headers = append(c.KongAdminAPIConfig.Headers, "kong-admin-token:"+c.KongAdminToken)
	}
	httpclient, err := adminapi.MakeHTTPClient(&c.KongAdminAPIConfig)
	if err != nil {
		return nil, err
	}

	return adminapi.GetKongClientForWorkspace(ctx, c.KongAdminURL, c.KongWorkspace, httpclient)
}

func (c *Config) GetKubeconfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags(c.APIServerHost, c.KubeconfigPath)
}

func (c *Config) GetKubeClient() (client.Client, error) {
	conf, err := c.GetKubeconfig()
	if err != nil {
		return nil, err
	}
	return client.New(conf, client.Options{})
}
