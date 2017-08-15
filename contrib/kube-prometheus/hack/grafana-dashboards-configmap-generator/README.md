# Grafana Dashboards Configmap Generator

## Description:
Tool to maintain grafana dashboards' configmap for a grafana deployed with kube-prometheus (a tool inside prometheus-operator).

The tool reads the content of a directory with grafana .json resources (dashboards and datasources) and creates a manifest file under output/ directory with all the content from the files in a Kubernetes ConfigMap format.

Based on a configurable size limit, the tool will create 1 or N configmaps to allocate the .json resources (bin packing). If the limit is reached then the configmaps generated will have names like grafana-dashboards-0, grafana-dashboards-1, etc, and if the limit is not reached the configmap generated will be called "grafana-dashboards".

Input Parameters Allowed:
```bash
-i dir, --input-dir dir
  Directory with grafana dashboards to process.
  Important notes:
    Files should be suffixed with -dashboard.json or -datasource.json.
    We don't recommend file names with spaces.

-o file, --output-file file
  Output file for config maps.

-s NUM, --size-limit NUM
  Size limit in bytes for each dashboard (default: 240000)

-n namespace, --namespace namespace
  Namespace for the configmap (default: monitoring).

-x, --apply-configmap
  Applies the generated configmap with kubectl.

--apply-type
  Type of kubectl command. Accepted values: apply, replace, create (default: apply).
```

## Usage

Just execute the .sh under bin/ directory. The output will be placed in the output/ directory.

Examples:
```bash
$ ./grafana_dashboards_generate.sh
$ bin/grafana_dashboards_generate.sh -o manifests/grafana/grafana-dashboards.yaml -i assets/grafana-dashboards
$ bin/grafana_dashboards_generate.sh -s 1000000 --apply-configmap --apply-type replace

# Note: the output file, if provided with -o, shouldn't exist.
```

## Configuration and options

* Put the json files you want to pack in the templates/grafana-dashboards/ directory
* Size limit default is 240000 bytes due to the annotations size limit in kubernetes of 256KB.

