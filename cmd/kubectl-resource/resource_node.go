package cmd

import (
	//"context"
	"errors"

	"github.com/bryant-rh/kubectl-resource/pkg/writer"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"k8s.io/kubectl/pkg/metricsutil"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"


	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type ResourceNodeOptions struct {
	ResourceName       string
	ResourceType       string
	Selector           string
	SortBy             string
	NoFormat           bool
	UseProtocolBuffers bool

	NodeClient      corev1client.CoreV1Interface
	Printer         *metricsutil.TopCmdPrinter
	DiscoveryClient discovery.DiscoveryInterface
	MetricsClient   metricsclientset.Interface
	Client          *kube.KubeClient

	genericclioptions.IOStreams
}

var (
	ResourceNodeLong = templates.LongDesc(i18n.T(`
		Display resource (CPU/Memory/PodCount) usage of nodes.

		The resource-node command allows you to see the resource consumption of nodes.`))

	ResourceNodeExample = templates.Examples(i18n.T(`
		  # Show metrics for all nodes
		  kubectl resource node

		  # Show metrics for a given node
		  kubectl resource node NODE_NAME`))
)

func NewCmdResouceNode(f cmdutil.Factory, o *ResourceNodeOptions, streams genericclioptions.IOStreams) *cobra.Command {
	if o == nil {
		o = &ResourceNodeOptions{
			IOStreams:          streams,
			UseProtocolBuffers: true,
		}
	}

	cmd := &cobra.Command{
		Use:                   "node [NAME | -l label]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Display resource (CPU/Memory/PodCount) usage of nodes"),
		Long:                  ResourceNodeLong,
		Example:               ResourceNodeExample,
		ValidArgsFunction:     util.ResourceNameCompletionFunc(f, "node"),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.RunResourceNode())
		},
		Aliases: []string{"nodes", "no"},
	}
	// fsets := cmd.PersistentFlags()
	// cfgFlags := genericclioptions.NewConfigFlags(true)
	// //cfgFlags := defaultConfigFlags
	// cfgFlags.AddFlags(fsets)
	// matchVersionFlags := cmdutil.NewMatchVersionFlags(cfgFlags)
	// matchVersionFlags.AddFlags(fsets)

	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().StringVarP(&o.ResourceType, "type", "t", o.ResourceType, "Type information hierarchically (default: All Type)[possible values: cpu, memory, pod]")
	cmd.Flags().BoolVar(&o.NoFormat, "no-format", o.NoFormat, "If present, print output without format table")
	cmd.Flags().StringVar(&o.SortBy, "sort-by", o.SortBy, "If non-empty, sort nodes list using specified field. The field can be either 'cpu' or 'memory' or ''.")
	// cmd.Flags().BoolVar(&o.NoHeaders, "no-headers", o.NoHeaders, "If present, print output without headers")
	// cmd.Flags().BoolVar(&o.UseProtocolBuffers, "use-protocol-buffers", o.UseProtocolBuffers, "Enables using protocol-buffers to access Metrics API.")
	// cmd.Flags().BoolVar(&o.ShowCapacity, "show-capacity", o.ShowCapacity, "Print node resources based on Capacity instead of Allocatable(default) of the nodes.")

	return cmd
}

func (o *ResourceNodeOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		o.ResourceName = args[0]
	} else if len(args) > 1 {
		return cmdutil.UsageErrorf(cmd, "%s", cmd.Use)
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
	// if o.UseProtocolBuffers {
	// 	config.ContentType = "application/vnd.kubernetes.protobuf"
	// }
	//o.MetricsClient, err = metricsclientset.NewForConfig(config)
	o.Client, err = kube.NewClient(config)
	if err != nil {
		return err
	}
	//o.MetricsClient	= client

	//o.NodeClient = clientset.CoreV1()

	//o.Printer = metricsutil.NewTopCmdPrinter(o.Out)
	return nil
}

func (o *ResourceNodeOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(o.SortBy) > 0 {
		if o.SortBy != sortByCPU && o.SortBy != sortByMemory {
			return errors.New("--sort-by accepts only cpu or memory")
		}
	}
	if len(o.ResourceName) > 0 && len(o.Selector) > 0 {
		return errors.New("only one of NAME or --selector can be provided")
	}
	if len(o.ResourceType) > 0 {
		if o.ResourceType != sortByCPU && o.ResourceType != sortByMemory && o.ResourceType != sortByPod {
			return errors.New("--type accepts only cpu or memory or pod")
		}
	}
	return nil
}

func (o ResourceNodeOptions) RunResourceNode() error {
	var err error
	selector := labels.Everything()
	if len(o.Selector) > 0 {
		selector, err = labels.Parse(o.Selector)
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

	data, err := o.Client.GetNodeResources(o.ResourceName, o.ResourceType, o.SortBy, selector)
	if err != nil {
		return err
	}
	writer.NodeWrite(data, o.ResourceType, o.NoFormat)
	return nil
}
