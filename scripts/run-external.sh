#!/usr/bin/env bash
# NOTE:
#  -e exit immediately when a command fails
#  -u error on unset variables
#  -o pipefail  exit with zero only if all commands of the pipeline exit successfully
set -eu -o pipefail

trap cleanup EXIT INT

# globals
declare CONTEXT=""
declare SHOW_USAGE=false
declare SKIP_OPERATOR_RUN_CHECK=false
declare USE_DEFAULT_CONTEXT=false
declare API_SERVER=""
declare IMPERSONATE_USER="${IMPERSONATE_USER:-}"
declare FEATURE_GATES="${FEATURE_GATES:-}"
declare LOG_LEVEL="${LOG_LEVEL:-info}"

# tmp operator files that needs to be cleaned up
declare -r CA_FILE="tmp/CA_FILE"
declare -r KEY_FILE="tmp/KEY_FILE"
declare -r CERT_FILE="tmp/CERT_FILE"

cleanup() {
	rm -f "$CA_FILE" "$KEY_FILE" "$CERT_FILE"
}

run() {
	echo " ❯ $* "
	"$@"
}

header() {
	local title="🔆🔆🔆  $*  🔆🔆🔆 "

	local len=40
	if [[ ${#title} -gt $len ]]; then
		len=${#title}
	fi

	echo -e "\n\n  \033[1m${title}\033[0m"
	echo -n "━━━━━"
	printf '━%.0s' $(seq "$len")
	echo "━━━━━"
}

info() {
	echo " 🔔 $*"
}

ok() {
	echo " ✅ $*"
}

warn() {
	echo " ⚠️  $*"
}

err() {
	echo " 🛑 ERROR: $*"
}

update_crds() {
	header "Update Operator CRDs"

	run kubectl apply --server-side --force-conflicts \
		-f example/prometheus-operator-crd-full

	info "Wait for CRDs to be registered"
	run kubectl wait --for=condition=Established crds --all --timeout=120s
}

init_context() {
	header "Set Context"
	$USE_DEFAULT_CONTEXT && {
		CONTEXT=$(kubectl config current-context)
	}

	info "Using context - '$CONTEXT'"
	return 0
}

extract_certs() {
        local ca key cert context_cluster context_user
  
        context_cluster=$(kubectl config view --raw -o jsonpath="{.contexts[?(@.name == \"$CONTEXT\")].context.cluster}")
        context_user=$(kubectl config view --raw -o jsonpath="{.contexts[?(@.name == \"$CONTEXT\")].context.user}")
        
        local cluster=".clusters[?(@.name == \"$context_cluster\")].cluster"
        local user=".users[?(@.name == \"$context_user\")].user"

	API_SERVER=$(kubectl config view -o jsonpath="{$cluster.server}")
	ca=$(kubectl config view --raw -o jsonpath="{$cluster.certificate-authority-data}" | base64 -d)
	key=$(kubectl config view --raw -o jsonpath="{$user.client-key-data}" | base64 -d)
	cert=$(kubectl config view --raw -o jsonpath="{$user.client-certificate-data}" | base64 -d)

	local fail=0
	[[ -z "$API_SERVER" ]] && {
		err "failed to get api server details; did you specify the right context?"
		fail=1
	}

	[[ -z "$ca" ]] && {
		err "CA is empty; did you specify the right context?"
		fail=1
	}

	[[ -z "$cert" ]] && {
		err "cert is empty; did you specify the right context?"
		fail=1
	}

	[[ -z "$key" ]] && {
		err "key is empty; did you specify the right context?"
		fail=1
	}

	[[ "$fail" -ne 0 ]] && return 1

	echo "$ca" >"$CA_FILE"
	echo "$key" >"$KEY_FILE"
	echo "$cert" >"$CERT_FILE"
	return 0
}

run_operator() {
	header "Run Operator"

	info "Running operator against context: $CONTEXT - $API_SERVER"
	echo "──────────────────────────────────────────────────────────────────"

	run ./operator \
		--as="$IMPERSONATE_USER" \
		--apiserver="$API_SERVER" \
		--ca-file="$CA_FILE" \
		--cert-file="$CERT_FILE" \
		--feature-gates="$FEATURE_GATES" \
		--log-level="$LOG_LEVEL" \
		--key-file="$KEY_FILE" 2>&1 | tee tmp/operator.log
}

validate_args() {
	if [[ -z "$CONTEXT" ]] && ! $USE_DEFAULT_CONTEXT; then
		err "missing context name or --use-default-context"
		return 1
	fi

	# ensure both context and use-default-context isn't set
	if [[ -n "$CONTEXT" ]] && $USE_DEFAULT_CONTEXT; then
		err "Cannot provide both context name '$CONTEXT' and --use-default-context"
		return 1
	fi

	return 0
}

ensure_operator_not_running() {
	header "Ensure no other prometheus-operator is running"

	$SKIP_OPERATOR_RUN_CHECK && {
		info "skipping operator run check"
		return 0
	}

	local po_label='app.kubernetes.io/name=prometheus-operator'

	local po_pods
	po_pods=$(kubectl get pods -A -l "$po_label" -o name | wc -l)

	[[ "$po_pods" -gt 0 ]] && {
		warn "Found $po_pods pods matching $po_label in cluster!"
		echo "If it is safe to continue, rerun the script with --no-operator-run-check option"
		return 1
	}
	ok "No pods with matching label $po_label running in cluster"

	return 0
}

show_usage() {

	local scr
	scr="$(basename "$0")"

	read -r -d '' help <<-EOF_HELP || true
		Usage:
		────────────────────────────────────────────
		  ❯ $scr <context-name>
		  ❯ $scr  -h|--help

		Options:
		⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻
		  --help | -h                   show this help
		  --no-operator-run-check | -f: skips the check that ensures no prometheus-operator
		                                is running in the cluster.
		  --use-default-context | -c:   Runs operator using current context
																		NOTE: cannot use both -c and <context-name>

		⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻
		💡 Tips
		⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻
		To get all context names in current config, run:
		   ❯ kubectl config get-contexts -o name

		To get current/default context name, run:
		   ❯ kubectl config current-context

	EOF_HELP

	echo -e "$help"
	return 0
}

parse_args() {
	### while there are args parse them
	while [[ -n "${1+xxx}" ]]; do
		case $1 in
		-h | --help)
			SHOW_USAGE=true
			# exit the loop
			break
			;;
		--no-operator-run-check | -f)
			shift
			SKIP_OPERATOR_RUN_CHECK=true
			shift
			;;
		--use-default-context | -c)
			shift
			USE_DEFAULT_CONTEXT=true
			;;
		*)
			CONTEXT="$1"
			shift
			;;
		esac
	done
	return 0
}

build_operator() {
	header "Building Operator"
	run make operator
}

main() {
	parse_args "$@" || {
		show_usage
		exit 1
	}

	$SHOW_USAGE && {
		show_usage
		exit 0
	}

	validate_args || {
		show_usage
		exit 1
	}

	# all files are relative to the root of the project
	cd "$(git rev-parse --show-toplevel)"
	mkdir -p tmp

	init_context
	extract_certs

	ensure_operator_not_running
	build_operator
	update_crds
	run_operator
}

main "$@"
