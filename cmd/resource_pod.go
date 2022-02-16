package cmd

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/discovery"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/metricsutil"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/bryant-rh/kubectl-resource/pkg/kube"
	"github.com/bryant-rh/kubectl-resource/pkg/writer"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ResourcePodOptions struct {
	ResourceName       string
	Namespace          string
	ResourceType       string
	LabelSelector      string
	FieldSelector      string
	SortBy             string
	NoFormat           bool
	AllNamespaces      bool
	PrintContainers    bool
	NoHeaders          bool
	UseProtocolBuffers bool

	PodClient       corev1client.PodsGetter
	Printer         *metricsutil.TopCmdPrinter
	DiscoveryClient discovery.DiscoveryInterface
	MetricsClient   metricsclientset.Interface
	Client          *kube.KubeClient

	genericclioptions.IOStreams
}

//const metricsCreationDelay = 2 * time.Minute

var (
	resourcePodLong = templates.LongDesc(i18n.T(`
		Display resource (CPU/Memory) usage of pods.

		The 'resource pod' command allows you to see the resource consumption of pods.

		Due to the metrics pipeline delay, they may be unavailable for a few minutes
		since pod creation.`))

	resourcePodExample = templates.Examples(i18n.T(`
		# Show metrics for all pods in the default namespace
		kubectl resource pod

		# Show metrics for all pods in the given namespace
		kubectl resource pod --namespace=NAMESPACE

		# Show metrics for a given pod and its containers
		kubectl resource pod POD_NAME --containers

		# Show metrics for the pods defined by label name=myLabel
		kubectl resource pod -l name=myLabel`))
)

func NewCmdResoucePod(f cmdutil.Factory, o *ResourcePodOptions, streams genericclioptions.IOStreams) *cobra.Command {
	if o == nil {
		o = &ResourcePodOptions{
			IOStreams:          streams,
			UseProtocolBuffers: true,
		}
	}

	cmd := &cobra.Command{
		Use:                   "pod [NAME | -l label]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Display resource (CPU/memory) usage of pods"),
		Long:                  resourcePodLong,
		Example:               resourcePodExample,
		ValidArgsFunction:     util.ResourceNameCompletionFunc(f, "pod"),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunResourcePod())
		},
		Aliases: []string{"pods", "po"},
	}
	cmd.Flags().StringVarP(&o.LabelSelector, "selector", "l", o.LabelSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().StringVarP(&o.ResourceType, "type", "t", o.ResourceType, "Type information hierarchically (default: All Type)[possible values: cpu, memory, pod]")
	cmd.Flags().StringVar(&o.FieldSelector, "field-selector", o.FieldSelector, "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.")
	cmd.Flags().StringVar(&o.SortBy, "sort-by", o.SortBy, "If non-empty, sort pods list using specified field. The field can be either 'cpu' or 'memory'.")
	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().BoolVar(&o.NoFormat, "no-format", o.NoFormat, "If present, print output without format table")
	return cmd
}

func (o *ResourcePodOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error
	if len(args) == 1 {
		o.ResourceName = args[0]
	} else if len(args) > 1 {
		return cmdutil.UsageErrorf(cmd, "%s", cmd.Use)
	}

	o.Namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	clientset, err := f.KubernetesClientSet()
	if err != nil {
		return err
	}

	o.DiscoveryClient = clientset.DiscoveryClient
	config, err := f.ToRESTConfig()
	if err != nil {
		return err
	}

	o.Client, err = kube.NewClient(config)
	if err != nil {
		return err
	}
	return nil
}

func (o *ResourcePodOptions) Validate() error {
	if len(o.SortBy) > 0 {
		if o.SortBy != sortByCPU && o.SortBy != sortByMemory {
			return errors.New("--sort-by accepts only cpu or memory")
		}
	}
	if len(o.ResourceName) > 0 && len(o.LabelSelector) > 0 {
		return errors.New("only one of NAME or --selector can be provided")
	}
	if len(o.ResourceType) > 0 {
		if o.ResourceType != sortByCPU && o.ResourceType != sortByMemory {
			return errors.New("--type accepts only cpu or memory")
		}
	}
	return nil
}

func (o ResourcePodOptions) RunResourcePod() error {
	var err error
	labelSelector := labels.Everything()
	if len(o.LabelSelector) > 0 {
		labelSelector, err = labels.Parse(o.LabelSelector)
		if err != nil {
			return err
		}
	}
	fieldSelector := fields.Everything()
	if len(o.FieldSelector) > 0 {
		fieldSelector, err = fields.ParseSelector(o.FieldSelector)
		if err != nil {
			return err
		}
	}

	apiGroups, err := o.DiscoveryClient.ServerGroups()
	if err != nil {
		return err
	}

	metricsAPIAvailable := SupportedMetricsAPIVersionAvailable(apiGroups)

	if !metricsAPIAvailable {
		return errors.New("metrics API not available")
	}
	metrics, err := o.Client.GetPodMetricsFromMetricsAPI(o.Namespace, o.ResourceName, o.AllNamespaces, labelSelector, fieldSelector)
	if err != nil {
		return err
	}

	// First we check why no metrics have been received.
	if len(metrics.Items) == 0 {
		// If the API server query is successful but all the pods are newly created,
		// the metrics are probably not ready yet, so we return the error here in the first place.
		// err := verifyEmptyMetrics(o, labelSelector, fieldSelector)
		// if err != nil {
		// 	return err
		// }

		// if we had no errors, be sure we output something.
		if o.AllNamespaces {
			fmt.Fprintln(o.ErrOut, "No resources found")
		} else {
			fmt.Fprintf(o.ErrOut, "No resources found in %s namespace.\n", o.Namespace)
		}
	}

	data, err := o.Client.GetPodResources(metrics.Items, o.Namespace, o.ResourceName, o.AllNamespaces, o.ResourceType, o.SortBy, labelSelector, fieldSelector)
	if err != nil {

		return err
	}
	writer.PodWrite(data, o.ResourceType, o.NoFormat)
	return nil
}
