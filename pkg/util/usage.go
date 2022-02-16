package util

import (
	"fmt"

	aurora "github.com/logrusorgru/aurora/v3"
)

// var UsageTemplate = fmt.Sprintf(`%v:{{if .Runnable}}
//   kubectl {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
//   {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}
// %v:
//   {{.NameAndAliases}}{{end}}{{if .HasExample}}
// %v:
// {{.Example}}{{end}}{{if .HasAvailableSubCommands}}
// Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
//   {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}
// %v:
//   -h, --help                   Display this help message
//   -n, --namespace string       Change the namespace scope for this CLI request
//   -l, --selector='': Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
//       --sort-by='': If non-empty, sort nodes list using specified field. The field can be either 'cpu' or 'memory'.
//   -o, --options                List of all options for this command
//       --version                Show version for this command
// Use "kubectl resource --options" for a list of all options (applies to this command).
// `,
// 	aurora.Cyan("Usage"),
// 	aurora.Cyan("Aliases"),
// 	aurora.Cyan("Available Commands"),
// 	aurora.Cyan("Options"))

var UsageTemplate = fmt.Sprintf(`%v:{{if .Runnable}}
  kubectl {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}
%v:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}
%v:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{end}}
%v:
  -h, --help                   Display this help message
  -n, --namespace string       Change the namespace scope for this CLI request
  -l, --selector=''            Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
      --sort-by=''             If non-empty, sort nodes list using specified field. The field can be either 'cpu' or 'memory'.
  -t, --type string            Type information hierarchically (default: All Type) [possible values: cpu, memory, pod]
  -o, --options                List of all options for this command
      --version                Show version for this command
Use "kubectl resource --options" for a list of all options (applies to this command).
`,
	aurora.Cyan("Usage"),
	aurora.Cyan("Aliases"),
	aurora.Cyan("Available Commands"),
	aurora.Cyan("Options"))

var OptionTemplate = `The following options can be passed to this command:
{{if .HasAvailableLocalFlags}}{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}`

var VersionTemplate = fmt.Sprintf(` %v {{.Name}} version {{.Version}}
`, aurora.Yellow(">"))
