#!/bin/bash

# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

# Author: eedugon

# Description: Tool to maintain grafana dashboards configmap for a grafana deployed
#   with kube-prometheus (a tool inside prometheus-operator)
# The tool reads the content of a directory with grafana .json resources
#   that need to be moved into a configmap.
# Based on a configurable size limit, the tool will create 1 or N configmaps
#   to allocate the .json resources (bin packing)

# Update: 20170914
# The tool also generates a grafana deployment manifest (-g option)

# parameters
# -o, --output-file
# -g, --grafana-manifest-file
# -i, --input-dir
# -s, --size-limit
# -x, --apply-configmap : true or false (default = false)
# --apply-type : create, replace, apply (default = apply)

#
# Basic Functions
#
echoSyntax() {
  echo "Usage: ${0} [options]"
  echo "Options:"
  echo -e "\t-i dir, --input-dir dir"
  echo -e "\t\tDirectory with grafana dashboards to process."
  echo -e "\t\tImportant notes:"
  echo -e "\t\t\tFiles should be suffixed with -dashboard.json or -datasource.json."
  echo -e "\t\t\tWe don't recommend file names with spaces."
  echo
  echo -e "\t-o file, --output-file file"
  echo -e "\t\tOutput file for config maps."
  echo
  echo -e "\t-s NUM, --size-limit NUM"
  echo -e "\t\tSize limit in bytes for each dashboard (default: 240000)"
  echo
  echo -e "\t-n namespace, --namespace namespace"
  echo -e "\t\tNamespace for the configmap (default: monitoring)."
  echo
  echo -e "\t-x, --apply-configmap"
  echo -e "\t\tApplies the generated configmap with kubectl."
  echo
  echo -e "\t--apply-type"
  echo -e "\t\tType of kubectl command. Accepted values: apply, replace, create (default: apply)."
}


# # Apply changes --> environment allowed
# test -z "$APPLY_CONFIGMAP" && APPLY_CONFIGMAP="false"
# # Size limit --> environment set allowed
# test -z "$DATA_SIZE_LIMIT" && DATA_SIZE_LIMIT="240000" # in bytes
# # Changes type: in case of problems with k8s configmaps, try replace. Should be apply
# test -z "$APPLY_TYPE" && APPLY_TYPE="apply"
# # Input values verification
# echo "$DATA_SIZE_LIMIT" | grep -q "^[0-9]\+$" || { echo "ERROR: Incorrect value for DATA_SIZE_LIMIT: $DATA_SIZE_LIMIT. Number expected"; exit 1; }

# Base variables (do not change them)
DATE_EXEC="$(date "+%Y-%m-%d-%H%M%S")"
BIN_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TOOL_HOME="$(dirname $BIN_DIR)"
SCRIPT_BASE=`basename $0 | sed "s/\.[Ss][Hh]//"`
CONFIGMAP_DASHBOARD_PREFIX="grafana-dashboards"

TEMPLATES_DIR="$TOOL_HOME/templates"
DASHBOARD_HEADER_FILE="$TEMPLATES_DIR/dashboard.header"
DASHBOARD_FOOT_FILE="$TEMPLATES_DIR/dashboard.foot"
CONFIGMAP_HEADER="$TEMPLATES_DIR/ConfigMap.header"
GRAFANA_DEPLOYMENT_TEMPLATE="$TEMPLATES_DIR/grafana-deployment-template.yaml"
OUTPUT_BASE_DIR="$TOOL_HOME/output"

# Some default values
OUTPUT_FILE="$OUTPUT_BASE_DIR/grafana-dashboards-configMap-$DATE_EXEC.yaml"
GRAFANA_OUTPUT_FILE="$OUTPUT_BASE_DIR/grafana-deployment-$DATE_EXEC.yaml"
DASHBOARDS_DIR="$TEMPLATES_DIR/grafana-dashboards"

APPLY_CONFIGMAP="false"
APPLY_TYPE="apply"
DATA_SIZE_LIMIT="240000"
NAMESPACE="monitoring"

# Input parameters
while (( "$#" )); do
    case "$1" in
        "-o" | "--output-file")
            OUTPUT_FILE="$2"
            shift
            ;;
        "-g" | "--grafana-output-file")
            GRAFANA_OUTPUT_FILE="$2"
            shift
            ;;
        "-i" | "--input-dir")
            DASHBOARDS_DIR="$2"
            shift
            ;;
        "-n" | "--namespace")
            NAMESPACE="$2"
            shift
            ;;
        "-x" | "--apply-configmap")
            APPLY_CONFIGMAP="true"
            ;;
        "--apply-type")
            APPLY_TYPE="$2"
            test "$APPLY_TYPE" != "create" && test "$APPLY_TYPE" != "apply" && test "$APPLY_TYPE" != "replace" && { echo "Unexpected APPLY_TYPE: $APPLY_TYPE"; exit 1; }
            shift
            ;;
        "-s"|"--size-limit")
            if ! ( echo $2 | grep -q '^[0-9]\+$') || [ $2 -eq 0 ]; then
                echo "Invalid value for size limit '$2'"
                exit 1
            fi
            DATA_SIZE_LIMIT=$2
            shift
            ;;
        "-h"|"--help")
            echoSyntax
            exit 0
            ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
    shift
done

#
# Auxiliary Functions
#
indentMultiLineString() {
    # Indent a given string (in one line including multiple \n)
    test "$#" -eq 2 || { echo "INTERNAL ERROR: wrong call to function indentMultiLineString"; exit 1; }
    local indent_number="$1"
    local string="$2"

    test "$indent_number" -ge 0 || { echo "INTERNAL ERROR: wrong indent number parameter: $indent_number"; exit 1; }

    # prepare indentation text
    local indent_string=""
    for (( c=0; c<$indent_number; c++ )); do
      indent_string="$indent_string "
    done

    echo "$string" | sed -e "s#^#$indent_string#" -e "s#\\\n#\\\n$indent_string#g"
}

#
# Main Functions
#
addConfigMapHeader() {
  # If a parameter is provided it will be used as the configmap index.
  # If no parameter is provided, the name will be kept
  test "$#" -le 1 || { echo "# INTERNAL ERROR: Wrong call to function addConfigMapHeader"; return 1; }
  test "$#" -eq 1 && local id="$1" || local id=""

  if [ "$id" ]; then
    cat "$CONFIGMAP_HEADER" | sed "s/name: $CONFIGMAP_DASHBOARD_PREFIX/name: $CONFIGMAP_DASHBOARD_PREFIX-$id/"
  else
    cat "$CONFIGMAP_HEADER"
  fi
}

addArrayToConfigMap() {
  # This function process the array to_process into a configmap
  local file=""
  local OLDIFS=$IFS
  local IFS=$'\n'
  for file in ${to_process[@]}; do
    # check that file exists
    test -f "$file" || { echo "# INTERNAL ERROR IN ARRAY: File not found: $file"; continue; }

    # detection of type (dashboard or datasource)
    type=""
    basename "$file" | grep -q "\-datasource" && type="datasource"
    basename "$file" | grep -q "\-dashboard" && type="dashboard"
    test "$type" || { echo "# ERROR: Unrecognized file type: $(basename $file)"; return 1; }

    #echo "# Processing $type $file"
    # Indent 2
    echo "  $(basename $file): |+"

    # Dashboard header: No indent needed
    test "$type" = "dashboard" && cat $DASHBOARD_HEADER_FILE

    # File content: Indent 4
    cat $file | sed "s/^/    /"

    # Dashboard foot
    test "$type" = "dashboard" && cat $DASHBOARD_FOOT_FILE
    [ "$(tail -c 1 "$file")" ] && echo
  done
  echo "---"

  IFS=$OLDIFS
  return 0
}

initialize-bin-pack() {
  # We separate initialization to reuse the bin-pack for different sets of files.
  n="0"
  to_process=()
  bytes_to_process="0"
  total_files_processed="0"
  total_configmaps_created="0"
}

bin-pack-files() {
  # Algorithm:
  # We process the files with no special order consideration
  # We create an array/queue of "files to add to configmap" called "to_process"
  # Size of the file is analyzed to determine if it can be added to the queue or not.
  # the max size of the queue is limited by DATA_SIZE_LIMIT
  # while there's room available in the queue we add files.
  # when there's no room we create a configmap with the members of the queue
  #  before adding the file to a cleaned queue

  # Counters initialization is not in the scope of this function
  local file=""
  OLDIFS=$IFS
  IFS=$'\n'
#  echo "DEBUG bin-pack:"
#  echo "$@"

  for file in $@; do
    test -f "$file" || { echo "# INTERNAL ERROR: File not found: $file"; continue; }
#    echo "debug: Processing file $(basename $file)"

    file_size_bytes="$(stat -c%s "$file")" || true

    # If the file is bigger than the configured limit we skip it file
    if [ "$file_size_bytes" -gt "$DATA_SIZE_LIMIT" ]; then
      echo "ERROR: File $(basename $file) bigger than size limit: $DATA_SIZE_LIMIT ($file_size_bytes). Skipping"
      continue
    fi
    (( total_files_processed++ )) || true

    if test "$(expr "$bytes_to_process" + "$file_size_bytes")" -le "$DATA_SIZE_LIMIT"; then
      # We have room to include the file in the configmap
      # test "$to_process" && to_process="$to_process $file" || to_process="$file"
      to_process+=("$file")
      (( bytes_to_process = bytes_to_process + file_size_bytes )) || true
      echo "# File $(basename $file) : added to queue"
    else
      # There's no room to add this file to the queue. so we process what we have and add the file to the queue
      if [ "$to_process" ]; then
        echo
        echo "# Size limit ($DATA_SIZE_LIMIT) reached. Processing queue with $bytes_to_process bytes. Creating configmap with id $n"
        echo
        # Create a new configmap
        addConfigMapHeader $n >> $OUTPUT_FILE || { echo "ERROR in call to addConfigMapHeader function"; exit 1; }
        addArrayToConfigMap >> $OUTPUT_FILE || { echo "ERROR in call to addArrayToConfigMap function"; exit 1; }
        # Initialize variables with info about file not processed
        (( total_configmaps_created++ )) || true
        (( n++ )) || true
        # to_process="$file"
        to_process=()
        to_process+=("$file")
        bytes_to_process="$file_size_bytes"
        echo "# File $(basename $file) : added to queue"
      else
        # based on the algorithm the queue should never be empty if we reach this part of the code
        # if this happens maybe bytes_to_process was not aligned with the queue (to_process)
        echo "ERROR (unexpected)"
      fi
    fi
  done
  IFS=$OLDIFS
}

# prepareGrafanaDeploymentManifest() {
#   local num_configmaps="$1"
#
#   for (( i=0; i<$total_configmaps_created; i++ )); do
#     echo "Creating deployment for $CONFIGMAP_DASHBOARD_PREFIX-$i"
#
#   done
# }


# Some variables checks...
test ! -d "$TEMPLATES_DIR" && { echo "ERROR: missing templates directory $TEMPLATES_DIR"; exit 1; }

test -f "$DASHBOARD_FOOT_FILE" || { echo "Template $DASHBOARD_FOOT_FILE not found"; exit 1; }
test -f "$DASHBOARD_HEADER_FILE" || { echo "Template $DASHBOARD_HEADER_FILE not found"; exit 1; }
test -f "$CONFIGMAP_HEADER" || { echo "Template $CONFIGMAP_HEADER not found"; exit 1; }
test -f "$GRAFANA_DEPLOYMENT_TEMPLATE" || { echo "Template $GRAFANA_DEPLOYMENT_TEMPLATE not found"; exit 1; }

test ! -d "$OUTPUT_BASE_DIR" && { echo "ERROR: missing directory $OUTPUT_BASE_DIR"; exit 1; }

# Initial checks
test -d "$DASHBOARDS_DIR" || { echo "ERROR: Dashboards directory not found: $DASHBOARDS_DIR"; echoSyntax; exit 1; }

test -f "$OUTPUT_FILE" && { echo "ERROR: Output file already exists: $OUTPUT_FILE"; exit 1; }
test -f "$GRAFANA_OUTPUT_FILE" && { echo "ERROR: Output file already exists: $GRAFANA_OUTPUT_FILE"; exit 1; }
touch $OUTPUT_FILE || { echo "ERROR: Unable to create or modify $OUTPUT_FILE"; exit 1; }
touch $GRAFANA_OUTPUT_FILE || { echo "ERROR: Unable to create or modify $GRAFANA_OUTPUT_FILE"; exit 1; }

# Main code start

echo "# Starting execution of $SCRIPT_BASE on $DATE_EXEC"
echo "# Configured size limit: $DATA_SIZE_LIMIT bytes"
echo "# Grafna input dashboards and datasources will be read from: $DASHBOARDS_DIR"
echo "# Grafana Dashboards ConfigMap will be created into file:"
echo "$OUTPUT_FILE"
echo "# Grafana Deployment manifest will be created into file:"
echo "$GRAFANA_OUTPUT_FILE"
echo

# Loop variables initialization
initialize-bin-pack

# Process dashboards
bin-pack-files "$(find $DASHBOARDS_DIR -maxdepth 1 -type f -name "*-dashboard.json" | sort)"

# Continue processing datasources (maintaining the same queue)
bin-pack-files "$(find $DASHBOARDS_DIR -maxdepth 1 -type f -name "*-datasource.json" | sort )"

# Processing remaining data in the queue (or unique)
if [ "$to_process" ]; then
  if [ "$n" -eq 0 ]; then
    echo
    echo "# Size limit not reached ($bytes_to_process). Adding all files into basic configmap"
    echo
    addConfigMapHeader $n >> $OUTPUT_FILE || { echo "ERROR in call to addConfigMapHeader function"; exit 1; }
  else
    echo
    echo "# Size limit not reached ($bytes_to_process). Adding remaining files into configmap with id $n"
    echo
    addConfigMapHeader $n >> $OUTPUT_FILE || { echo "ERROR in call to addConfigMapHeader function"; exit 1; }
  fi
  addArrayToConfigMap >> $OUTPUT_FILE || { echo "ERROR in call to addArrayToConfigMap function"; exit 1; }
  (( total_configmaps_created++ )) || true
  to_process=()
fi

echo "# Process completed, configmap created: $(basename $OUTPUT_FILE)"
echo "# Summary"
echo "# Total files processed: $total_files_processed"
echo "# Total amount of ConfigMaps inside the manifest: $total_configmaps_created"
echo
# Grafana deployment Processing (for every configmap)
#prepareGrafanaDeploymentManifest "$total_configmaps_created"
VOLUMES=""
VOLUME_MOUNTS=""
WATCH_DIR=""
for (( i=0; i<$total_configmaps_created; i++ )); do
  configmap="$CONFIGMAP_DASHBOARD_PREFIX-$i"
  echo "# Preparing grafana deployment to support configmap: $configmap"

  test "$VOLUME_MOUNTS" && VOLUME_MOUNTS="$VOLUME_MOUNTS\n- name: $configmap\n  mountPath: /var/$configmap" || VOLUME_MOUNTS="- name: $configmap\n  mountPath: /var/$configmap"
  test "$VOLUMES" && VOLUMES="$VOLUMES\n- name: $configmap\n  configMap:\n    name: $configmap" || VOLUMES="- name: $configmap\n  configMap:\n    name: $configmap"
  test "$WATCH_DIR" && WATCH_DIR="$WATCH_DIR\n- '--watch-dir=/var/$configmap'" || WATCH_DIR="- '--watch-dir=/var/$configmap'"
  # echo "DEBUG:"
  # echo "VOLUMES: $VOLUMES"
  # echo "VOLUME_MOUNTS: $VOLUME_MOUNTS"
  # echo "WATCH_DIR: $WATCH_DIR"
  echo
done

echo "# Processing grafana deployment template into $GRAFANA_OUTPUT_FILE"
sed -e "s#XXX_VOLUMES_XXX#$(indentMultiLineString 6 "$VOLUMES")#" \
   -e "s#XXX_VOLUME_MOUNTS_XXX#$(indentMultiLineString 8 "$VOLUME_MOUNTS")#" \
   -e "s#XXX_WATCH_DIR_XXX#$(indentMultiLineString 10 "$WATCH_DIR")#" \
   $GRAFANA_DEPLOYMENT_TEMPLATE > $GRAFANA_OUTPUT_FILE

# If output file is empty we can delete it and exit
test ! -s "$OUTPUT_FILE" && { echo "# Configmap empty, deleting file"; rm $OUTPUT_FILE; exit 0; }
test ! -s "$GRAFANA_OUTPUT_FILE" && { echo "# Configmap empty, deleting file"; rm $GRAFANA_OUTPUT_FILE; exit 0; }

if [ "$APPLY_CONFIGMAP" = "true" ]; then
  test -x "$(which kubectl)" || { echo "ERROR: kubectl command not available. Apply configmap not possible"; exit 1; }
  echo "# Applying configuration with $APPLY_TYPE method on namespace $NAMESPACE"
  if kubectl -n $NAMESPACE $APPLY_TYPE -f "$OUTPUT_FILE"; then
    echo
    echo "# ConfigMap updated. Updating grafana deployment"
    kubectl -n $NAMESPACE $APPLY_TYPE -f "$GRAFANA_OUTPUT_FILE" || { echo "Error applying Grafana deployment. Check yaml file: $GRAFANA_OUTPUT_FILE"; exit 1; }
  else
    echo "Error applying Configmap. Check yaml file: $OUTPUT_FILE"
  fi
else
  echo
  echo "# To apply the new configMap to your k8s system do something like:"
  echo "kubectl -n monitoring $APPLY_TYPE -f $OUTPUT_FILE"
  echo "kubectl -n monitoring $APPLY_TYPE -f $GRAFANA_OUTPUT_FILE"
  echo
fi
