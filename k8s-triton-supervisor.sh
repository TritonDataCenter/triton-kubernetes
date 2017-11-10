#!/usr/bin/env bash

set -o errexit
set -o pipefail
# set -o xtrace

main() {
	check_prerequisites

	while (( $# ))
	do
		case $1 in
			-c)
				getTERRAFORM
				echo "Using $TERRAFORM ...
				"
				oneClusterPerClone
				getTritonAccountConfig
				getClusterManagerConfig
				verifySetup "cluster"
				setModuleClusterManager
				startProvisioning "create_rancher"
				showClusterDetails
				exit 0
			;;
			-e)
				getTERRAFORM
				echo "Using $TERRAFORM ...
				"
				getEnvironmentConfig
				startProvisioning "${_name}"
				showEnvironmentDetails
				exit 0
			;;
			--cleanAll)
				shift
				if [ "$1" == "true" ]; then
					echo "WARN: files could be deleted..."
					rm -rvf bin || true
					rm -rvf terraform/.terraform || true
					rm -rvf terraform/terraform.tfstate* || true
					curl -s -o terraform/create-rancher.tf  https://raw.githubusercontent.com/joyent/k8s-triton-supervisor/multicloud-cli/terraform/create-rancher.tf
					curl -s -o terraform/create-rancher-env.tf https://raw.githubusercontent.com/joyent/k8s-triton-supervisor/multicloud-cli/terraform/create-rancher-env.tf
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
		$TERRAFORM apply -target "module.${1}"
	)
}

verifySetup() {
	echo "############################################################"
	echo ""
	if [ "$1" == "cluster" ]
	then
		echo "Cluster Manager ${_name} will be created on Triton."
		if [ "$_ha" == "true" ]
		then
			echo "${_name} will be running in HA configuration and provision three Triton machines ..."
			echo "    ${_name}-master-1 ${_master_triton_machine_package}"
			echo "    ${_name}-master-2 ${_master_triton_machine_package}"
			echo "    ${_name}-mysqldb ${_mysqldb_triton_machine_package}"
		else
			echo "${_name} will be running in non-HA configuration ..."
			echo "    ${_name}-master-1 ${_master_triton_machine_package}"
		fi
	elif [ "$1" == "triton" ]
	then
		echo "Environment ${_name} will be created on Triton."
		if [ "$_ha" == "true" ]
		then
			echo "${_name} will be running in HA configuration ..."
			echo "6 dedicated hosts will be created ..."
			echo "    ${_name}-etcd-[123] ${_etcd_triton_machine_package}"
			echo "    ${_name}-orchestration-[123] ${_orchestration_triton_machine_package}"
		else
			echo "${_name} will be running in non-HA configuration ..."
		fi
		echo "${_compute_node_count} compute nodes will be created for this environment ..."
		echo "    ${_name}-compute-# ${_compute_triton_machine_package}"
	elif [ "$1" == "azure" ]
	then
		echo "Environment ${_name} will be created on Azure."
		if [ "$_ha" == "true" ]
		then
			echo "${_name} will be running in HA configuration ..."
			echo "6 dedicated hosts will be created ..."
			echo "    ${_name}-etcd-[123] ${_etcd_azure_size}"
			echo "    ${_name}-orchestration-[123] ${_orchestration_azure_size}"
		else
			echo "${_name} will be running in non-HA configuration ..."
		fi
		echo "${_compute_node_count} compute nodes will be created for this environment ..."
		echo "    ${_name}-compute-# ${_compute_azure_size}"
	fi
	echo ""
	if [ "$(getVerification "Do you want to start the setup")" == "false" ]
	then
		echo "Canceling the configuration ..."
		echo ""
		exit 1
	fi
}

getEnvironmentConfig() {
	local __cloudOption
	echo "From clouds below:"
	echo "1. Triton"
	echo "2. Azure"
	while [ ! "$__cloudOption" ]
	do
		__cloudChoice=$(getArgument "Which cloud do you want to run your environment on" "1")
		case $__cloudChoice in
			'1') # TRITON
				getTritonAccountConfig
				getTritonEnvironmentConfig
				verifySetup "triton"
				setModuleTritonEnvironment
				__cloudOption=$__cloudChoice
			;;
			'2') # AZURE
				getAzureAccountConfig
				getAzureEnvironmentConfig
				verifySetup "azure"
				setModuleAzureEnvironment
				__cloudOption=$__cloudChoice
			;;
		esac
	done
}

getAzureAccountConfig() {
	_azure_subscription_id="$(getArgument "Azure subscription id")"
	_azure_client_id="$(getArgument "Azure client id")"
	_azure_client_secret="$(getArgument "Azure client secret")"
	_azure_tenant_id="$(getArgument "Azure tenant id")"
}
getTritonAccountConfig() {
	_triton_account="${SDC_ACCOUNT:-"$(getArgument "Your Triton account login name")"}"
	_triton_url="${SDC_URL:-"$(getArgument "The Triton CloudAPI endpoint URL" "https://us-east-1.api.joyent.com")"}"
	_triton_key_id="${SDC_KEY_ID:-"$(getArgument "Your Triton account key id")"}"

	for f in ~/.ssh/*
	do
		if [[ "$(ssh-keygen -E md5 -lf "${f//.pub$/}" 2> /dev/null | awk '{print $2}' | sed 's/^MD5://')" == "$_triton_key_id" ]]
		then
			export SDC_KEY=${f//.pub$/}
			break
		fi
		# old version of ssh-keygen, defaults to md5
		if [ "$(ssh-keygen -l -f "${f//.pub$/}" 2> /dev/null | awk '{print $2}' | grep ^SHA256)" == "" ]
		then
			if [[ "$(ssh-keygen -l -f "${f//.pub$/}" 2> /dev/null | awk '{print $2}')" == "$_triton_key_id" ]]
			then
				export SDC_KEY=${f//.pub$/}
				break
			fi
		fi
	done
	_triton_key_path="${SDC_KEY:-"$(getArgument "Your Triton account private key")"}"
}

getClusterManagerConfig() {
    # SET the following variables:
    #	_master_triton_machine_package
    #	_mysqldb_triton_machine_package
	#   _name
	#   _ha
	_name="$(getArgument "Name your Global Cluster Manager" "global-cluster")"
	_ha="$(getVerification "Do you want to set up the Global Cluster Manager in HA mode")"
	echo "From below options:"
	echo "Joyent-SDC-Public"
	echo "Joyent-SDC-Private"
	echo "Both"
	_network_choice="$(getArgument "Which Triton networks should be used for this environment" "Joyent-SDC-Public")"
	if [ "${_network_choice}" == "Joyent-SDC-Public" ]
	then
		_network_choice="triton_network_names = [ \"Joyent-SDC-Public\" ]"
	elif [ "${_network_choice}" == "Joyent-SDC-Private" ]
	then
		_network_choice="triton_network_names = [ \"Joyent-SDC-Private\" ]"
	elif [ "${_network_choice}" == "Both" ]
	then
		_network_choice="triton_network_names = [ \"Joyent-SDC-Public\", \"Joyent-SDC-Private\" ]"
	else
		echo "error: no networks selected"
	fi
	if triton profile get >/dev/null 2>&1
	then
		echo "From below packages:"
		triton packages -o name | grep kvm
	fi
	_master_triton_machine_package="$(getArgument "Which Triton package should be used for Global Cluster Manager server(s)" "k4-highcpu-kvm-1.75G")"
	if [ "$_ha" == "true" ]
	then
		_mysqldb_triton_machine_package="$(getArgument "Which Triton package should be used for Global Cluster Manager database server" "k4-highcpu-kvm-1.75G")"
	fi
}
getAzureEnvironmentConfig() {
	_name="$(getArgument "Name your environment" "azure-test")"
	_ha="$(getVerification "Do you want this environment to run in HA mode")"
	if [ "$_ha" == "true" ]
	then
		_etcd_node_count=3
		_orchestration_node_count=3
	else
		_etcd_node_count=0
		_orchestration_node_count=0
	fi
	_compute_node_count="$(getArgument "Number of compute nodes for $_name environment" "3")"
	# TODO: networks for environment
	_azure_location="$(getArgument "Where should the $_name environment be located" "West US 2")"
	if [ "$_ha" == "true" ]
	then
		_etcd_azure_size="$(getArgument "What size hosts should be used for $_name environment etcd nodes" "Standard_A1")"
		_orchestration_azure_size="$(getArgument "What size hosts should be used for $_name environment orchestration nodes running apiserver/scheduler/controllermanager/..." "Standard_A1")"
	fi
	_compute_azure_size="$(getArgument "What size hosts should be used for $_name environment compute nodes" "Standard_A1")"
	_azure_public_key_path="$(getArgument "Which ssh public key should these hosts be set up with" "${HOME}/.ssh/id_rsa.pub")"
}
getTritonEnvironmentConfig() {
	_name="$(getArgument "Name your environment" "triton-test")"
	_ha="$(getVerification "Do you want this environment to run in HA mode")"
	if [ "$_ha" == "true" ]
	then
		_etcd_node_count=3
		_orchestration_node_count=3
	else
		_etcd_node_count=0
		_orchestration_node_count=0
	fi
	_compute_node_count="$(getArgument "Number of compute nodes for $_name environment" "3")"
	echo "From below options:"
	echo "Joyent-SDC-Public"
	echo "Joyent-SDC-Private"
	echo "Both"
	_network_choice="$(getArgument "Which Triton networks should be used for this environment" "Joyent-SDC-Public")"
	if [ "${_network_choice}" == "Joyent-SDC-Public" ]
	then
		_network_choice="triton_network_names = [ \"Joyent-SDC-Public\" ]"
	elif [ "${_network_choice}" == "Joyent-SDC-Private" ]
	then
		_network_choice="triton_network_names = [ \"Joyent-SDC-Private\" ]"
	elif [ "${_network_choice}" == "Both" ]
	then
		_network_choice="triton_network_names = [ \"Joyent-SDC-Public\", \"Joyent-SDC-Private\" ]"
	else
		echo "error: no networks selected"
	fi
	if triton profile get >/dev/null 2>&1
	then
		echo "From below packages:"
		triton packages -o name | grep kvm
	fi
	if [ "$_ha" == "true" ]
	then
		_etcd_triton_machine_package="$(getArgument "Which Triton package should be used for $_name environment etcd nodes" "k4-highcpu-kvm-1.75G")"
		_orchestration_triton_machine_package="$(getArgument "Which Triton package should be used for $_name environment orchestration nodes running apiserver/scheduler/controllermanager/..." "k4-highcpu-kvm-1.75G")"
	fi
	_compute_triton_machine_package="$(getArgument "Which Triton package should be used for $_name environment compute nodes" "k4-highcpu-kvm-1.75G")"
}

setModuleClusterManager() {
	# TERRAFORM, _triton_account, _triton_url, _triton_key_id, _triton_key_path,  _master_triton_machine_package, _name, _ha
	(
		cd terraform
		sed "s; triton_account.*=.*; triton_account                 = \"${_triton_account}\";g" create-rancher.tf \
		| sed "s; triton_key_path.*=.*; triton_key_path                = \"${_triton_key_path}\";g" \
		| sed "s; triton_key_id.*=.*; triton_key_id                  = \"${_triton_key_id}\";g" \
		| sed "s; triton_url.*=.*; triton_url                     = \"${_triton_url}\";g" \
		| sed "s; name.*=.*; name                           = \"${_name}\";g" \
		| sed "s; master_triton_machine_package.*=.*; master_triton_machine_package  = \"${_master_triton_machine_package}\";g" \
		| sed "s; triton_network_names.*=.*; ${_network_choice};g" \
		| sed "s; mysqldb_triton_machine_package.*=.*; mysqldb_triton_machine_package = \"${_mysqldb_triton_machine_package}\";g" \
		| sed "s; ha.*=.*; ha                             = ${_ha};g" > tmp.cfg && mv tmp.cfg create-rancher.tf
	)
}
setModuleAzureEnvironment() {
	(
		cd terraform
		__k8s_plane_isolation="none"
		if [ "$_ha" == "true" ]
		then
			__k8s_plane_isolation="required"
		fi
		cat >>create-rancher-env.tf <<-EOF

		module "${_name}" {
		  source = "./modules/azure-rancher-k8s"

		  api_url    = "http://\${element(data.terraform_remote_state.rancher.masters, 0)}:8080"
		  access_key = ""
		  secret_key = ""

		  name = "${_name}"

		  etcd_node_count          = "${_etcd_node_count}"
		  orchestration_node_count = "${_orchestration_node_count}"
		  compute_node_count       = "${_compute_node_count}"

		  k8s_plane_isolation = "${__k8s_plane_isolation}"

		  azure_subscription_id = "${_azure_subscription_id}"
		  azure_client_id       = "${_azure_client_id}"
		  azure_client_secret   = "${_azure_client_secret}"
		  azure_tenant_id       = "${_azure_tenant_id}"

		  azure_location = "${_azure_location}"

		  azure_ssh_user        = "ubuntu"
		  azure_public_key_path = "${_azure_public_key_path}"

		  etcd_azure_size          = "${_etcd_azure_size}"
		  orchestration_azure_size = "${_orchestration_azure_size}"
		  compute_azure_size       = "${_compute_azure_size}"
		}

		EOF
	)
}
setModuleTritonEnvironment() {
	(
		cd terraform
		__k8s_plane_isolation="none"
		if [ "$_ha" == "true" ]
		then
			__k8s_plane_isolation="required"
		fi
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

		  k8s_plane_isolation = "${__k8s_plane_isolation}"

		  triton_account       = "${_triton_account}"
		  triton_key_path      = "${_triton_key_path}"
		  triton_key_id        = "${_triton_key_id}"
		  triton_url           = "${_triton_url}"

		  ${_network_choice}

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
	while true
	do
		if [ -z "${2+x}" ]
		then
			read -r -p "$1: " theargument
			if [ ! -z "$theargument" ]
			then
				echo "$theargument"
				break
			fi
		else
			read -r -p "$1: ($2) " theargument
			if [ -z "$theargument" ]
			then
				echo "$2"
			else
				echo "$theargument"
			fi
			break
		fi
	done
}
getVerification() {
	while true
	do
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
			* )
			;;
		esac
	done
}

showClusterDetails() {
	local __MasterIP
	__MasterIP=$($TERRAFORM output -state=terraform/terraform.tfstate -json | jq ".masters.value[0]" | tr -d '"')
	echo ""
	echo "Cluster Manager ${_name} has been started."
	if [ "$_ha" == "true" ]
	then
		echo "This is an HA Active/Active setup so you can use either of the IP addresses."
	else
		echo "This is a non-HA setup so there is only one cluster manager node."
	fi
	cat <<-EOM
	    http://${__MasterIP}:8080/settings/env

	Next step is adding Kubernetes environments to be managed here.
	To start your first environment, run:
	    $0 -e

	EOM
}
showEnvironmentDetails() {
	local __MasterIP
	__MasterIP=$($TERRAFORM output -state=terraform/terraform.tfstate -json | jq ".masters.value[0]" | tr -d '"')
	echo ""
	echo "Environment ${_name} has been started."
	if [ "$_ha" == "true" ]
	then
		echo "This is an HA setup of Kubernetes cluster so there are 3 dedicated etcd and 3 orchestration nodes."
	else
		echo "This is a non-HA setup so Kubernetes services could run on any of the compute nodes."
	fi
	cat <<-EOM
	Cluster Manager URL:
	    http://${__MasterIP}:8080/settings/env
	Kubernetes Hosts URL:
	    http://${__MasterIP}:8080/env/$($TERRAFORM output -state=terraform/terraform.tfstate -module="${_name}" -json | jq '.environment_id.value' | tr -d '"')/infra/hosts?mode=dot
	Kubernetes Health:
	    http://${__MasterIP}:8080/env/$($TERRAFORM output -state=terraform/terraform.tfstate -module="${_name}" -json | jq '.environment_id.value' | tr -d '"')/apps/stacks?which=cattle
	
	NOTE: Nodes might take a few minutes to connect and come up.

	To start another environment, run:
	    $0 -e

	EOM
}

getTERRAFORM() {
	local __TERRAFORM
	local __TERRAFORM_VERSION
	local __TERRAFORM_URL
	local __TERRAFORM_ZIP_FILE

	if [ -n "$(command -v terraform)" ]
	then
		__TERRAFORM="$(command -v terraform)"
		__TERRAFORM_VERSION="$(terraform version | grep 'Terraform v' | sed 's/Terraform v//')"
	fi
	if [ "$__TERRAFORM_VERSION" != "$(curl -s https://checkpoint-api.hashicorp.com/v1/check/terraform | jq -r -M '.current_version')" ]
	then
		(
			__TERRAFORM_ZIP_FILE="terraform_""$(uname | tr '[:upper:]' '[:lower:]')""_amd64.zip"
			__TERRAFORM_URL="https://releases.hashicorp.com/terraform/$(curl -s https://checkpoint-api.hashicorp.com/v1/check/terraform | jq -r -M '.current_version')/terraform_$(curl -s https://checkpoint-api.hashicorp.com/v1/check/terraform | jq -r -M '.current_version')_$(uname | tr '[:upper:]' '[:lower:]')_amd64.zip"
			mkdir bin >/dev/null 2>&1 || true
			cd bin

			if [ ! -e "$__TERRAFORM_ZIP_FILE" ]
			then
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
		-c  create a cluster manager on Triton
		-e  create and add a Kubernetes environment to an existing cluster manager
	"
}
check_prerequisites() {
	if ! which jq > /dev/null
	then
		echo "error: jq is not in your PATH ..."
		echo "       Make sure it is in your PATH and run this command again."
		echo ""
		exit 1
	fi
}
oneClusterPerClone() {
	(
		cd terraform
		if $TERRAFORM output > /dev/null 2>&1
		then
			$TERRAFORM output
			echo "error: there seems to be a Cluster Manager running already ..."
			echo ""
			exit 1
		fi
	)
}

main "$@"
