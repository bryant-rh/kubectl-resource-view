package cmd

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bryant-rh/kubectl-resource-view/pkg/kube"
	"github.com/bryant-rh/kubectl-resource-view/pkg/writer"

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
	ResourceTypeslice  []string
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
		Display resource (cpu/memory/gpu/podcount) usage of nodes.

		The resource node command allows you to see the resource consumption of nodes.`))

	ResourceNodeExample = templates.Examples(i18n.T(`
		  # Show metrics for all nodes
		  kubectl resource-view node

		  # Show metrics for a given node
		  kubectl resource-view node NODE_NAME

		  # Show metrics for the node defined by type name=cpu,memory,gpu,pod
		  kubectl resource-view node -t cpu,memory,gpu,pod

		  `))
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
		Short:                 i18n.T("Display resource (cpu/memory/gpu/podcount) usage of nodes"),
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

	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().StringVarP(&o.ResourceType, "type", "t", o.ResourceType, "Type information hierarchically (default: All Type)[possible values: cpu,memory,pod,gpu], Multiple can be specified, separated by commas")
	cmd.Flags().BoolVar(&o.NoFormat, "no-format", o.NoFormat, "If present, print output without format table")
	cmd.Flags().StringVar(&o.SortBy, "sort-by", o.SortBy, "If non-empty, sort nodes list using specified field. The field can be either 'cpu' or 'memory' ")

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
	o.Client, err = kube.NewClient(config)
	if err != nil {
		return err
	}
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

	o.ResourceTypeslice = strings.Split(o.ResourceType, ",")
	if len(o.ResourceType) > 0 {
		for _, str := range o.ResourceTypeslice {
			if !MapKeyInIntSlice(nodeResourceType, str) {
				return errors.New("--type accepts only cpu,memory,pod,gpu")
			}
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
	//ResourceType := strings.Split(o.ResourceType, ",")

	// 添加context用于超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 修改GetNodeResources调用，传入context
	data, err := o.Client.GetNodeResources(ctx, o.ResourceName, o.ResourceTypeslice, o.SortBy, selector)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.New("operation timed out - too many nodes or slow API response")
		}
		return err
	}

	writer.NodeWrite(data, o.ResourceTypeslice, o.NoFormat)
	return nil
}
