package cmd

import (
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
)

const (
	sortByCPU    = "cpu"
	sortByMemory = "memory"
)

var (
	nodeResourceType = []string{"cpu", "memory", "pod", "gpu"}
	podResourceType  = []string{"cpu", "memory", "gpu"}
)

var (
	// set values via build flags
	version string
	//commit                      string
	supportedMetricsAPIVersions = []string{
		"v1beta1",
	}
	topLong = templates.LongDesc(i18n.T(`
		Display Resource (cpu/memory/gpu/podcount) Usage and Request and Limit.

		The resource command allows you to see the resource consumption for nodes or pods.

		This command requires Metrics Server to be correctly configured and working on the server. `))
	rolesumExample = templates.Examples(i18n.T(`
	   node        Display Resource (cpu/memory/gpu/podcount) usage of nodes
	   pod         Display Resource (cpu/memory/gpu)          usage of pods`))
)

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

// versionString returns the version prefixed by 'v'
// or an empty string if no version has been populated by goreleaser.
// In this case, the --version flag will not be added by cobra.
func versionString() string {
	if len(version) == 0 {
		return ""
	}
	return "v" + version
}

//NewCmdResource
func NewCmdResource() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "kubectl-resource-view [flags] [options]",
		Version:               versionString(),
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Short:                 i18n.T("Display resource (cpu/memory/gpu/podcount) usage"),
		Long:                  topLong,
		Example:               templates.Examples(rolesumExample),
		Run: func(cmd *cobra.Command, args []string) {
			runHelp(cmd, args)
			//brutil.CheckErr(Validate(cmd, args))
		},
	}
	//cmd.SetVersionTemplate(brutil.VersionTemplate)
	//cmd.SetUsageTemplate(brutil.UsageTemplate)

	fsets := cmd.PersistentFlags()
	cfgFlags := genericclioptions.NewConfigFlags(true)
	cfgFlags.AddFlags(fsets)
	matchVersionFlags := cmdutil.NewMatchVersionFlags(cfgFlags)
	matchVersionFlags.AddFlags(fsets)

	f := cmdutil.NewFactory(matchVersionFlags)
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	//create subcommands
	cmd.AddCommand(NewCmdResouceNode(f, nil, streams))
	cmd.AddCommand(NewCmdResoucePod(f, nil, streams))

	return cmd
}

//SupportedMetricsAPIVersionAvailable
func SupportedMetricsAPIVersionAvailable(discoveredAPIGroups *metav1.APIGroupList) bool {
	for _, discoveredAPIGroup := range discoveredAPIGroups.Groups {
		if discoveredAPIGroup.Name != metricsapi.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true
				}
			}
		}
	}
	return false
}

//MapKeyInIntSlice
func MapKeyInIntSlice(haystack []string, needle string) bool {
	set := make(map[string]struct{})

	for _, e := range haystack {
		set[e] = struct{}{}
	}

	_, ok := set[needle]
	return ok
}
