#!/usr/bin/env bash

set -o errexit
set -o pipefail
# set -o xtrace

main() {
	while (( $# )); do
		case $1 in
			-c)
				getTERRAFORM
				echo "Using $TERRAFORM ...
				"
				declare -x cluster_module=module.create_rancher
				declare -a environment_modules
				getTritonAccountConfig
				addCluster
				addEnvironment
				startProvisioning cluster_module
				exit 0
			;;
			-e)
				getTERRAFORM
				echo "Using $TERRAFORM ...
				"
				declare -a environment_modules
				getTritonAccountConfig
				addEnvironment
				startProvisioning
				exit 0
			;;
			--cleanAll)
				shift
				if [ "$1" == "true" ]; then
					echo "WARN: files could be deleted..."
					rm -rvf bin terraform/.terraform terraform/terraform.tfstate* || true
					exit 1
				fi
				USAGE
				exit 1
			;;
		esac
		shift
	done
	USAGE
	exit 1
}

startProvisioning() {
	(
	cd terraform
	$TERRAFORM init
	if [ ! -z "${1+x}" ]; then
		$TERRAFORM apply -target module.${cluster_module}
	fi
	for __modules in "${environment_modules[@]}"; do
		$TERRAFORM apply -target "module.${__modules}"
	done
	)
}

addCluster() {
    # SET the following variables:
    #	_master_triton_machine_package
	#   _name
	#   _ha
	_name="$(getArgument "Name your Global Cluster Manager" "global-cluster")"
	_ha="$(getVerification "Do you want to set up the Global Cluster Manager in HA mode")"
	# TODO: networks for global cluster manager
	if triton profile get >/dev/null 2>&1; then
		echo "From below packages:"
		triton packages -o name | grep kvm
	fi
	_master_triton_machine_package="$(getArgument "Which Triton package should be used for Global Cluster Manager" "k4-highcpu-kvm-1.75G")"
	if [ "$(getVerification "Do you want to create this cluster manager")" == "false" ] ; then
		echo "Canceling the configuration ..."
		exit 1
	fi
	cluster_module=$(grep "module \".*{$" terraform/create-rancher.tf | awk '{print $2}' | tr -d '"')
	createClusterManager
}
addEnvironment() {
	local __addEnvironment
	local __cloudOption
	local __ha
	__addEnvironment="$(getVerification "Do you want to add an environment to this cluster")"
	while [ "$__addEnvironment" == "true" ]; do
		echo "From clouds below:"
		echo "1. Triton"
		echo "2. Azure"
		echo "3. GCP"
		__cloudOption=0
		while [ $__cloudOption != "1" ] && [ $__cloudOption != "2" ] && [ $__cloudOption != "3" ]; do
			__cloudOption=$(getArgument "Which cloud do you want to run your environment on" "1")
			case $__cloudOption in
				'1') # _name, _etcd_node_count, _orchestration_node_count, _compute_node_count, _etcd_triton_machine_package, _orchestration_triton_machine_package, _compute_triton_machine_package
					_name="$(getArgument "Name your environment" "dev-test")"
					__ha="$(getVerification "Do you want this environment to run in HA mode")"
					if [ "$__ha" == "true" ]; then
						_etcd_node_count=3
						_orchestration_node_count=3
					else
						_etcd_node_count=0
						_orchestration_node_count=0
					fi
					_compute_node_count="$(getArgument "Number of compute nodes for $_name environment" "3")"
					# TODO: networks for environment
					if triton profile get >/dev/null 2>&1; then
						echo "From below packages:"
						triton packages -o name | grep kvm
					fi
					if [ "$__ha" == "true" ]; then
						_etcd_triton_machine_package="$(getArgument "Which Triton package should be used for $_name environment etcd nodes" "k4-highcpu-kvm-1.75G")"
						_orchestration_triton_machine_package="$(getArgument "Which Triton package should be used for $_name environment orchestration nodes running apiserver/scheduler/controllermanager/..." "k4-highcpu-kvm-1.75G")"
					fi
					_compute_triton_machine_package="$(getArgument "Which Triton package should be used for $_name environment compute nodes" "k4-highcpu-kvm-1.75G")"
					environment_modules+=(${_name})
					createEnvironmentOnTriton
				;;
			esac
		done
		__addEnvironment="$(getVerification "Do you want to add another environment to this cluster")"
	done
}
getTritonAccountConfig() {
	_triton_account="${SDC_ACCOUNT:-"$(getArgument "Your Triton account login name")"}"
	_triton_url="${SDC_URL:-"$(getArgument "The Triton CloudAPI endpoint URL" "https://us-east-1.api.joyent.com")"}"
	_triton_key_id="${SDC_KEY_ID:-"$(getArgument "Your Triton account key id")"}"
	
	for f in ~/.ssh/*; do
		if [[ "$(ssh-keygen -E md5 -lf "${f//.pub$/}" 2> /dev/null | awk '{print $2}' | sed 's/^MD5://')" == "$_triton_key_id" ]]; then
			export SDC_KEY=${f//.pub$/}
			break
		fi
		# old version of ssh-keygen, defaults to md5
		if [ "$(ssh-keygen -l -f "${f//.pub$/}" 2> /dev/null | awk '{print $2}' | grep ^SHA256)" == "" ]; then
			if [[ "$(ssh-keygen -l -f "${f//.pub$/}" 2> /dev/null | awk '{print $2}')" == "$_triton_key_id" ]]; then
				export SDC_KEY=${f//.pub$/}
				break
			fi
		fi
	done
	_triton_key_path="${SDC_KEY:-"$(getArgument "Your Triton account private key")"}"
}

createClusterManager() {
	# TERRAFORM, _triton_account, _triton_url, _triton_key_id, _triton_key_path,  _master_triton_machine_package, _name, _ha
	(
	cd terraform
	sed "s; triton_account.*=.*; triton_account                = \"${_triton_account}\";g" create-rancher.tf \
	| sed "s; triton_url.*=.*; triton_url                    = \"${_triton_url}\";g" \
	| sed "s; triton_key_id.*=.*; triton_key_id                 = \"${_triton_key_id}\";g" \
	| sed "s; triton_key_path.*=.*; triton_key_path               = \"${_triton_key_path}\";g" \
	| sed "s; master_triton_machine_package.*=.*; master_triton_machine_package = \"${_master_triton_machine_package}\";g" \
	| sed "s; name.*=.*; name                          = \"${_name}\";g" \
	| sed "s; ha.*=.*; ha                            = ${_ha};g" > tmp.cfg && mv tmp.cfg create-rancher.tf
	)
}
createEnvironmentOnTriton() {
	(
	cd terraform
	cat >>create-rancher-env.tf <<-EOF

	module "${_name}" {
	  source = "./modules/triton-rancher-k8s"

	  api_url    = "http://\${element(data.terraform_remote_state.rancher.masters, 0)}:8080"
	  access_key = ""
	  secret_key = ""

	  name = "${_name}"

	  etcd_node_count          = "${_etcd_node_count}"
	  orchestration_node_count = "${_orchestration_node_count}"
	  compute_node_count       = "${_compute_node_count}"

	  triton_account       = "${_triton_account}"
	  triton_key_path      = "${_triton_key_path}"
	  triton_key_id        = "${_triton_key_id}"
	  triton_url           = "${_triton_url}"

	  triton_network_names = [
	    "Joyent-SDC-Public",
	    "Joyent-SDC-Private",
	  ]

	  etcd_triton_machine_package          = "${_etcd_triton_machine_package}"
	  orchestration_triton_machine_package = "${_orchestration_triton_machine_package}"
	  compute_triton_machine_package       = "${_compute_triton_machine_package}"
	}

	EOF
	)
}

getArgument() {
	# $1 message
	# $2 default
	while true; do
		if [ -z "${2+x}" ]; then
			read -r -p "$1: " theargument
			if [ ! -z "$theargument" ]; then
				echo "$theargument"
				break
			fi
		else
			read -r -p "$1: ($2) " theargument
			if [ -z "$theargument" ]; then
				echo "$2"
			else
				echo "$theargument"
			fi
			break
		fi
	done
}
getVerification() {
	while true; do
		read -r -p "$1? (yes | no) " yn
		case $yn in
			yes )
			echo "true"
			break
			;;
			no )
			echo "false"
			break
			;;
			* ) ;;
		esac
	done
}

getTERRAFORM() {
	local __TERRAFORM
	local __TERRAFORM_VERSION
	local __TERRAFORM_URL
	local __TERRAFORM_ZIP_FILE
	
	if [ -n "$(command -v terraform)" ]; then
		__TERRAFORM="$(command -v terraform)"
		__TERRAFORM_VERSION="$(terraform version | grep 'Terraform v' | sed 's/Terraform v//')"
	fi
	if [ "$__TERRAFORM_VERSION" != "$(curl -s https://checkpoint-api.hashicorp.com/v1/check/terraform | jq -r -M '.current_version')" ]; then
		(
		__TERRAFORM_ZIP_FILE="terraform_""$(uname | tr '[:upper:]' '[:lower:]')""_amd64.zip"
		__TERRAFORM_URL="https://releases.hashicorp.com/terraform/$(curl -s https://checkpoint-api.hashicorp.com/v1/check/terraform | jq -r -M '.current_version')/terraform_$(curl -s https://checkpoint-api.hashicorp.com/v1/check/terraform | jq -r -M '.current_version')_$(uname | tr '[:upper:]' '[:lower:]')_amd64.zip"
		mkdir bin >/dev/null 2>&1 || true
		cd bin
		
		if [ ! -e "$__TERRAFORM_ZIP_FILE" ]; then
			echo ""
			echo "Downloading latest Terraform zip'd binary"
			curl -s -o "$__TERRAFORM_ZIP_FILE" "$__TERRAFORM_URL"
			
			echo ""
			echo "Extracting Terraform executable"
			unzip -q -o "$__TERRAFORM_ZIP_FILE"
		fi
		)
		__TERRAFORM="$(pwd)/bin/terraform"
		__TERRAFORM_VERSION="$("$__TERRAFORM" version | grep 'Terraform v' | sed 's/Terraform v//')"
	fi
	TERRAFORM=$__TERRAFORM
}
USAGE() {
	echo "
	Usage:
		$0 [-c|-e]
	Options:
		-c  create cluster manager
		-e  create environment
	"
}

main "$@"
