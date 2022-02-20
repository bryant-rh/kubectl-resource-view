# kubectl resource-view
A plugin to access Kubernetes resource requests, limits, and usage.

Display Resource (CPU/Memory/Gpu/PodCount) Usage and Request and Limit.

The resource command allows you to see the resource consumption for nodes or pods.

```diff
- This command requires Metrics Server to be correctly configured and working on the server
```


## Installation

    go install github.com/bryant-rh/kubectl-resource-view@VERSION


krew is submitting PR, and it is still being approved. After approval, it can be installed in the following way

Use [krew](https://krew.sigs.k8s.io/) plugin manager to instal:

    kubectl krew install resource-view
    kubectl resource-view --help


## Usage
```bash
$ kubectl resource-view -h  # or kubectl-resource-view -h
Display Resource (cpu/memory/gpu/podcount) Usage and Request and Limit.

 The resource command allows you to see the resource consumption for nodes or pods.

 This command requires Metrics Server to be correctly configured and working on the server.

Usage:
  kubectl-resource-view [flags] [options]
  kubectl-resource-view [command]

Examples:
  node        Display Resource (cpu/memory/gpu/podcount) usage of nodes
  pod         Display Resource (cpu/memory/gpu)          usage of pods

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  node        Display resource (cpu/memory/gpu/podcount) usage of nodes
  pod         Display resource (cpu/memory/gpu) usage of pods

```
### node
```bash
$ kubectl resource-view node -h  # or kubectl-resource-view node -h
Display resource (cpu/memory/gpu/podcount) usage of nodes.

 The resource node command allows you to see the resource consumption of nodes.

Usage:
  kubectl-resource-view node [NAME | -l label]

Aliases:
  node, nodes, no

Examples:
  # Show metrics for all nodes
  kubectl resource-view node

  # Show metrics for a given node
  kubectl resource-view node NODE_NAME

  # Show metrics for the node defined by type name=cpu,memory,gpu,pod
  kubectl resource-view node -t cpu,memory,gpu,pod

Flags:
  -h, --help              help for node
      --no-format         If present, print output without format table
  -l, --selector string   Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
      --sort-by string    If non-empty, sort nodes list using specified field. The field can be either 'cpu' or 'memory'
  -t, --type string       Type information hierarchically (default: All Type)[possible values: cpu,memory,pod,gpu], Multiple can be specified, separated by commas

```

### pod
``` bash
$ kubectl resource-view pod -h  # or kubectl-resource-view  pod -h
Display resource (cpu/memory/gpu) usage of pods.

 The 'resource-view pod' command allows you to see the resource consumption of pods.

 Due to the metrics pipeline delay, they may be unavailable for a few minutes since pod creation.

Usage:
  kubectl-resource-view pod [NAME | -l label]

Aliases:
  pod, pods, po

Examples:
  # Show metrics for all pods in the default namespace
  kubectl resource-view pod

  # Show metrics for all pods in the given namespace
  kubectl resource-view pod --namespace=NAMESPACE

  # Show metrics for a given pod
  kubectl resource-view pod POD_NAME

  # Show metrics for the pods defined by label name=myLabel
  kubectl resource-view pod -l name=myLabel

  # Show metrics for the pods defined by type name=cpu,memory,gpu
  kubectl resource-view pod -t cpu,memory,gpu

Flags:
  -A, --all-namespaces          If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
      --field-selector string   Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.
  -h, --help                    help for pod
      --no-format               If present, print output without format table
  -l, --selector string         Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
      --sort-by string          If non-empty, sort pods list using specified field. The field can be either 'cpu' or 'memory'.
  -t, --type string             Type information hierarchically (default: All Type)[possible values: cpu,memory,gpu],Multiple can be specified, separated by commas

```

## Demo

### node

Show aggregate resource requests and limits  and pods Capacity. This is the same information displayed by kubectl describe nodes but in a easier to view format.

```diff
+ When the percentage of cpu request and memory request is greater than 90, it will be marked in yellow, and if it is greater than 95, it will be marked in red
```


Example (node):

![example Kubernetes node](assets/demo-node-1.png)

Example (Only display one of the data of the cpu or memroy or pod or gpu):
```bash
-t  [cpu,memory,gpu,pod] Multiple can be specified, separated by commas
```

![example Kubernetes node cpu or memroy or pod](assets/demo-node-2.png)

Example (According to cpu use or memory use sort):
```bash
--sort-by  [cpu/memory]
```

![example Kubernetes According to cpu use or memory use sort](assets/demo-node-3.png)

Example (Support filtering by lable or by node name):
```bash
-l   key1=value1,key2=value2
```

![example Kubernetes filtering by lable](assets/demo-node-4.png)
![example Kubernetes filtering by node name](assets/demo-node-5.png)


Example (Support similar kubectl format display):
```bash
--no-format
```

![example Kubernetes similar kubectl format display](assets/demo-node-6.png)

### pod

Show aggregate resource requests and limits  and pods Capacity. This is the same information displayed by kubectl describe nodes but in a easier to view format.

```diff
+ When the percentage of cpu use percent and memory use percent is greater than 90, it will be marked in yellow, and if it is greater than 95, it will be marked in red
```

Example (pod):

![example Kubernetes pod](assets/demo-pod-1.png)

Example (Only display one of the data of the cpu or memroy):
```bash
-t  [cpu/memory/gpu] Multiple can be specified, separated by commas
```

![example Kubernetes node cpu or memroy or pod](assets/demo-pod-7.png)

Example (According to cpu use or memory use sort):
```bash
--sort-by  [cpu/memory]
```

![example Kubernetes According to cpu use or memory use sort](assets/demo-pod-3.png)

Example (Support filtering by lable or by pod name or namespace):
```bash
-l   key1=value1,key2=value2
-n   namespace name
```

![example Kubernetes filtering by lable](assets/demo-pod-2.png)
![example Kubernetes filtering by lable](assets/demo-pod-4.png)
![example Kubernetes filtering by node name](assets/demo-pod-5.png)


Example (Support similar kubectl format display):
```bash
--no-format
```

![example Kubernetes similar kubectl format display](assets/demo-pod-6.png)

## License

Apache 2.0. See [LICENSE](./LICENSE).

---

This is not an official Google project.


