#!/usr/bin/env bash

set -o errexit
set -o pipefail
# set -o xtrace

main() {
	scriptname=$0
	check_prerequisites

	while (( $# ))
	do
		case $1 in
			-c)
				getTERRAFORM
				echo "Using $TERRAFORM ...
				"
				oneClusterPerClone
				shift
				if [ ! -z "${1+x}" ]
				then
					echo "Reading configuration from ${1}"
					readConfig "$1"
				else
					getTritonAccountConfig
					getClusterManagerConfig
					verifySetup "cluster"
				fi
				setModuleClusterManager
				startProvisioning "create_rancher"
				showClusterDetails
				exit 0
			;;
			-e)
				getTERRAFORM
				echo "Using $TERRAFORM ...
				"
				shift
				if [ ! -z "${1+x}" ]
				then
					echo "Reading configuration from ${1}"
					readConfig "$1"
					if [ ! -z "${_etcd_triton_machine_package+x}" ]
					then
						setModuleTritonEnvironment
					elif [ ! -z "${_azure_subscription_id+x}" ]
					then
						setModuleAzureEnvironment
					elif [ ! -z "${_aws_secret_key+x}" ]
					then
						setModuleAWSEnvironment
					elif [ ! -z "${_gcp_path_to_credentials+x}" ]
					then
						setModuleGCPEnvironment
					fi
				else
					getEnvironmentConfig
				fi
				startProvisioning "${_name}"
				showEnvironmentDetails
				exit 0
			;;
			-d)
				getTERRAFORM
				shift
				if [ ! -z "${1+x}" ]
				then
					deleteEnvironment "${1}"
				fi
				exit 1
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
		$TERRAFORM get
		$TERRAFORM apply -target "module.${1}"
	)
}
deleteEnvironment() {
	(
		cd terraform
		$TERRAFORM get
		echo "You are about to delete ${1} environment and all the associated hosts..."
		if [ "$(getVerification "Do you want to destroy this environmemnt")" == "false" ]
		then
			echo ""
			exit 1
		fi
		local _date
		_date=$(date '+%Y%m%d%H%M%S')
		$TERRAFORM destroy -force -target "module.${1}"
		sed "s;module \"${1}\" {$;module \"${1}-deleted-${_date}\" {;g" create-rancher-env.tf > tmp.tf && mv tmp.tf create-rancher-env.tf
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
	elif [ "$1" == "aws" ]
	then
		echo "Environment ${_name} will be created on AWS."
		if [ "$_ha" == "true" ]
		then
			echo "${_name} will be running in HA configuration ..."
			echo "6 dedicated hosts will be created ..."
			echo "    ${_name}-etcd-[123] ${_etcd_aws_instance_type}"
			echo "    ${_name}-orchestration-[123] ${_orchestration_aws_instance_type}"
		else
			echo "${_name} will be running in non-HA configuration ..."
		fi
		echo "${_compute_node_count} compute nodes will be created for this environment ..."
		echo "    ${_name}-compute-# ${_compute_aws_instance_type}"
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
	elif [ "$1" == "gcp" ]
	then
		echo "Environment ${_name} will be created on GCP."
		if [ "$_ha" == "true" ]
		then
			echo "${_name} will be running in HA configuration ..."
			echo "6 dedicated hosts will be created ..."
			echo "    ${_name}-etcd-[123] ${_etcd_gcp_instance_type}"
			echo "    ${_name}-orchestration-[123] ${_orchestration_gcp_instance_type}"
		else
			echo "${_name} will be running in non-HA configuration ..."
		fi
		echo "${_compute_node_count} compute nodes will be created for this environment ..."
		echo "    ${_name}-compute-# ${_compute_gcp_instance_type}"
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
	echo "2. AWS"
	echo "3. Azure"
	echo "4. GCP"
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
			'2') # AWS
				getAWSAccountConfig
				getAWSEnvironmentConfig
				verifySetup "aws"
				setModuleAWSEnvironment
				__cloudOption=$__cloudChoice
			;;
			'3') # AZURE
				getAzureAccountConfig
				getAzureEnvironmentConfig
				verifySetup "azure"
				setModuleAzureEnvironment
				__cloudOption=$__cloudChoice
			;;
			'4') # GCP
				getGCPAccountConfig
				getGCPEnvironmentConfig
				verifySetup "gcp"
				setModuleGCPEnvironment
				__cloudOption=$__cloudChoice
		esac
	done
}

getAWSAccountConfig() {
	_aws_access_key="$(getArgument "AWS Access Key")"
	_aws_secret_key="$(getArgument "AWS Secret Key")"
}
getAzureAccountConfig() {
	_azure_subscription_id="$(getArgument "Azure subscription id")"
	_azure_client_id="$(getArgument "Azure client id")"
	_azure_client_secret="$(getArgument "Azure client secret")"
	_azure_tenant_id="$(getArgument "Azure tenant id")"
}
getGCPAccountConfig() {
	_gcp_path_to_credentials="$(getArgument "Path to GCP credentials file")"
	_gcp_project_id="$(getArgument "GCP Project ID")"
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
	#   _gcm_node_count
	_name="$(getArgument "Name your Global Cluster Manager" "global-cluster")"
	_ha="$(getVerification "Do you want to set up the Global Cluster Manager in HA mode")"
	if [ "$_ha" == "true" ]
	then
		_gcm_node_count="$(getArgument "Number of cluster manager nodes for $_name Global Cluster Manager" "2")"
	fi
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
	docker_engine_install_url="$(getArgument "docker-engine install script" "https://releases.rancher.com/install-docker/1.12.sh")"
}
getAWSEnvironmentConfig() {
	_name="$(getArgument "Name your environment" "aws-test")"
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
	_aws_region="$(getArgument "Where should the $_name environment be located" "us-west-2")"
	_aws_ami_id="$(getArgument "Which image should be used for the nodes" "ami-0def3275")"
	if [ "$_ha" == "true" ]
	then
		_etcd_aws_instance_type="$(getArgument "What size hosts should be used for $_name environment etcd nodes" "t2.micro")"
		_orchestration_aws_instance_type="$(getArgument "What size hosts should be used for $_name environment orchestration nodes running apiserver/scheduler/controllermanager/..." "t2.micro")"
	fi
	_compute_aws_instance_type="$(getArgument "What size hosts should be used for $_name environment compute nodes" "t2.micro")"
	_aws_public_key_path="$(getArgument "Which ssh public key should these hosts be set up with" "${HOME}/.ssh/id_rsa.pub")"
	docker_engine_install_url="$(getArgument "docker-engine install script" "https://releases.rancher.com/install-docker/1.12.sh")"
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
	docker_engine_install_url="$(getArgument "docker-engine install script" "https://releases.rancher.com/install-docker/1.12.sh")"
}
getGCPEnvironmentConfig() {
	_name="$(getArgument "Name your environment" "gcp-test")"
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
	_gcp_compute_region="$(getArgument "Compute Region" "us-west1")"
	_gcp_instance_zone="$(getArgument "Instance Zone" "us-west1-a")"
	if [ "$_ha" == "true" ]
	then
		_etcd_gcp_instance_type="$(getArgument "What size hosts should be used for $_name environment etcd nodes" "n1-standard-1")"
		_orchestration_gcp_instance_type="$(getArgument "What size hosts should be used for $_name environment orchestration nodes running apiserver/scheduler/controllermanager/..." "n1-standard-1")"
	fi
	_compute_gcp_instance_type="$(getArgument "What size hosts should be used for $_name environment compute nodes" "n1-standard-1")"
	docker_engine_install_url="$(getArgument "docker-engine install script" "https://releases.rancher.com/install-docker/1.12.sh")"
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
	docker_engine_install_url="$(getArgument "docker-engine install script" "https://releases.rancher.com/install-docker/1.12.sh")"
}
readConfig() {
	docker_engine_install_url=https://releases.rancher.com/install-docker/1.12.sh
	while read -r line || [[ -n "$line" ]]
	do
		line=$(echo "$line" | tr -d ' ')
		if echo "$line" | grep -q ^SDC_ACCOUNT=
		then
			_triton_account=${line//SDC_ACCOUNT=/}
			echo "Triton Account: $_triton_account"
		elif echo "$line" | grep -q ^SDC_URL=
		then
			_triton_url=${line//SDC_URL=/}
			echo "Triton URL: $_triton_url"
		elif echo "$line" | grep -q ^SDC_KEY_ID=
		then
			_triton_key_id=${line//SDC_KEY_ID=/}
			echo "Triton Key ID: $_triton_key_id"
		elif echo "$line" | grep -q ^SDC_KEY=
		then
			_triton_key_path=${line//SDC_KEY=/}
			echo "Triton Key Material: $_triton_key_path"
		elif echo "$line" | grep -q ^name=
		then
			_name=${line//name=/}
			echo "Name: $_name"
		elif echo "$line" | grep -q ^ha=
		then
			_ha=${line//ha=/}
			echo "HA: $_ha"
		elif echo "$line" | grep -q ^networks=
		then
			_network_choice="triton_network_names = [ \"${line//networks=/}\" ]"
			_network_choice=${_network_choice//,/\",\ \"}
			echo "Networks: ${_network_choice//triton_network_names = /}"
		elif echo "$line" | grep -q ^master_package=
		then
			_master_triton_machine_package=${line//master_package=/}
			echo "Cluster Manager Package: $_master_triton_machine_package"
		elif echo "$line" | grep -q ^mysqldb_package=
		then
			_mysqldb_triton_machine_package=${line//mysqldb_package=/}
			echo "Cluster Manager DB Package: $_mysqldb_triton_machine_package"
		elif echo "$line" | grep -q ^compute_count=
		then
			_compute_node_count=${line//compute_count=/}
			echo "Number of Compute Nodes: $_compute_node_count"
		elif echo "$line" | grep -q ^etcd_package=
		then
			_etcd_triton_machine_package=${line//etcd_package=/}
			echo "Environment ETCD Package: $_etcd_triton_machine_package"
		elif echo "$line" | grep -q ^orchestration_package=
		then
			_orchestration_triton_machine_package=${line//orchestration_package=/}
			echo "Environment Orchestration Package $_orchestration_triton_machine_package"
		elif echo "$line" | grep -q ^compute_package=
		then
			_compute_triton_machine_package=${line//compute_package=/}
			echo "Environment Compute Package $_compute_triton_machine_package"
		elif echo "$line" | grep -q ^triton_image_name=
		then
			triton_image_name=${line//triton_image_name=/}
			echo "Image Name: $triton_image_name"
		elif echo "$line" | grep -q ^triton_image_version=
		then
			triton_image_version=${line//triton_image_version=/}
			echo "Image Version: $triton_image_version"
		elif echo "$line" | grep -q ^azure_subscription_id=
		then
			_azure_subscription_id=${line//azure_subscription_id=/}
			echo "Azure Subscription ID: $_azure_subscription_id"
		elif echo "$line" | grep -q ^azure_client_id=
		then
			_azure_client_id=${line//azure_client_id=/}
			echo "Azure Client ID ..."
		elif echo "$line" | grep -q ^azure_client_secret=
		then
			_azure_client_secret=${line//azure_client_secret=/}
			echo "Azure Client Secret ..."
		elif echo "$line" | grep -q ^azure_tenant_id=
		then
			_azure_tenant_id=${line//azure_tenant_id=/}
			echo "Azure Tenant ID ..."
		elif echo "$line" | grep -q ^azure_location=
		then
			_azure_location=${line//azure_location=/}
			echo "Location: $_azure_location"
		elif echo "$line" | grep -q ^azure_public_key_path=
		then
			_azure_public_key_path=${line//azure_public_key_path=/}
			echo "Public Key Path: $_azure_public_key_path"
		elif echo "$line" | grep -q ^etcd_azure_size=
		then
			_etcd_azure_size=${line//etcd_azure_size=/}
			echo "ETCD Image Size: $_etcd_azure_size"
		elif echo "$line" | grep -q ^orchestration_azure_size=
		then
			_orchestration_azure_size=${line//orchestration_azure_size=/}
			echo "Orchestration Image Size: $_orchestration_azure_size"
		elif echo "$line" | grep -q ^compute_azure_size=
		then
			_compute_azure_size=${line//compute_azure_size=/}
			echo "Compute Image Size: $_compute_azure_size"
		elif echo "$line" | grep -q ^rancher_server_image=
		then
			rancher_server_image=${line//rancher_server_image=/}
			echo "Rancher Server Image: $rancher_server_image"
		elif echo "$line" | grep -q ^rancher_agent_image=
		then
			rancher_agent_image=${line//rancher_agent_image=/}
			echo "Rancher Agent Image: $rancher_agent_image"
		elif echo "$line" | grep -q ^rancher_registry=
		then
			rancher_registry=${line//rancher_registry=/}
			echo "Rancher Registry: $rancher_registry"
		elif echo "$line" | grep -q ^rancher_registry_username=
		then
			rancher_registry_username=${line//rancher_registry_username=/}
			echo "Rancher Registry Username: $rancher_registry_username"
		elif echo "$line" | grep -q ^rancher_registry_password=
		then
			rancher_registry_password=${line//rancher_registry_password=/}
			echo "Rancher Registry Password ..."
		elif echo "$line" | grep -q ^k8s_registry=
		then
			k8s_registry=${line//k8s_registry=/}
			echo "K8s Registry: $k8s_registry"
		elif echo "$line" | grep -q ^k8s_registry_username=
		then
			k8s_registry_username=${line//k8s_registry_username=/}
			echo "K8s Registry Username: $k8s_registry_username"
		elif echo "$line" | grep -q ^k8s_registry_password=
		then
			k8s_registry_password=${line//k8s_registry_password=/}
			echo "K8s Registry Password ..."
		elif echo "$line" | grep -q ^access_key=
		then
			_aws_access_key=${line//access_key=/}
			echo "Access Key ..."
		elif echo "$line" | grep -q ^secret_key=
		then
			_aws_secret_key=${line//secret_key=/}
			echo "Secret Key ..."
		elif echo "$line" | grep -q ^etcd_aws_instance_type=
		then
			_etcd_aws_instance_type=${line//etcd_aws_instance_type=/}
			echo "ETCD instance size ${_etcd_aws_instance_type}"
		elif echo "$line" | grep -q ^orchestration_aws_instance_type=
		then
			_orchestration_aws_instance_type=${line//orchestration_aws_instance_type=/}
			echo "Orchestration instance size ${_orchestration_aws_instance_type}"
		elif echo "$line" | grep -q ^compute_aws_instance_type=
		then
			_compute_aws_instance_type=${line//compute_aws_instance_type=/}
			echo "Compute instance size ${_compute_aws_instance_type}"
		elif echo "$line" | grep -q ^aws_region=
		then
			_aws_region=${line//aws_region=/}
			echo "Region ${_aws_region}"
		elif echo "$line" | grep -q ^aws_ami_id=
		then
			_aws_ami_id=${line//aws_ami_id=/}
			echo "AMI ID ${_aws_ami_id}"
		elif echo "$line" | grep -q ^aws_public_key_path=
		then
			_aws_public_key_path=${line//aws_public_key_path=/}
			echo "Public ssh key ${_aws_public_key_path}"
		elif echo "$line" | grep -q ^gcp_path_to_credentials=
		then
			_gcp_path_to_credentials=${line//gcp_path_to_credentials=/}
			echo "GCP credentials ${_gcp_path_to_credentials}"
		elif echo "$line" | grep -q ^gcp_project_id=
		then
			_gcp_project_id=${line//gcp_project_id=/}
			echo "Project ID ${_gcp_project_id}"
		elif echo "$line" | grep -q ^etcd_gcp_instance_type=
		then
			_etcd_gcp_instance_type=${line//etcd_gcp_instance_type=/}
			echo "ETCD instance type ${_etcd_gcp_instance_type}"
		elif echo "$line" | grep -q ^orchestration_gcp_instance_type=
		then
			_orchestration_gcp_instance_type=${line//orchestration_gcp_instance_type=/}
			echo "Orchestration instance type ${_orchestration_gcp_instance_type}"
		elif echo "$line" | grep -q ^compute_gcp_instance_type=
		then
			_compute_gcp_instance_type=${line//compute_gcp_instance_type=/}
			echo "Compute instance type ${_compute_gcp_instance_type}"
		elif echo "$line" | grep -q ^gcp_compute_region=
		then
			_gcp_compute_region=${line//gcp_compute_region=/}
			echo "Compute Region ${_gcp_compute_region}"
		elif echo "$line" | grep -q ^gcp_instance_zone=
		then
			_gcp_instance_zone=${line//gcp_instance_zone=/}
			echo "Instane Zone ${_gcp_instance_zone}"
		elif echo "$line" | grep -q ^docker_engine_install_url=
		then
			docker_engine_install_url=${line//docker_engine_install_url=/}
			echo "docker-engine Install Script: $docker_engine_install_url"
		elif echo "$line" | grep -q ^gcm_node_count=
		then
			_gcm_node_count=${line//gcm_node_count=/}
			echo "Number of cluster manager Nodes: $_gcm_node_count"
		fi

		if [ "$_ha" == "true" ]
		then
			_etcd_node_count=3
			_orchestration_node_count=3
		else
			_etcd_node_count=0
			_orchestration_node_count=0
		fi
	done < "${1}"
}

setModuleClusterManager() {
	# TERRAFORM, _triton_account, _triton_url, _triton_key_id, _triton_key_path,  _master_triton_machine_package, _name, _ha
	(
		rancher_server_image_line='# rancher_server_image      = "docker-registry.joyent.com:5000/rancher/server:v1.6.10"'
		rancher_agent_image_line='# rancher_agent_image       = "docker-registry.joyent.com:5000/rancher/agent:v1.2.6"'
		rancher_registry_line='# rancher_registry          = "docker-registry.joyent.com:5000"'
		rancher_registry_username_line='# rancher_registry_username = "username"'
		rancher_registry_password_line='# rancher_registry_password = "password"'
		if [ ! -z "${rancher_server_image}" ]
		then
			rancher_server_image_line="rancher_server_image      = \"${rancher_server_image}\""
			rancher_agent_image_line="rancher_agent_image       = \"${rancher_agent_image}\""
			rancher_registry_line="rancher_registry          = \"${rancher_registry}\""
			rancher_registry_username_line="rancher_registry_username = \"${rancher_registry_username}\""
			rancher_registry_password_line="rancher_registry_password = \"${rancher_registry_password}\""
		fi

		if [ "$_ha" == "false" ]
		then
			_gcm_node_count="1"
		fi

		cd terraform
		cat >>create-rancher.tf <<-EOF
		
		module "create_rancher" {
		  source = "./modules/triton-rancher"

		  ${_network_choice}

		  triton_image_name    = "${triton_image_name:-"ubuntu-certified-16.04"}"
		  triton_image_version = "${triton_image_version:-"20170619.1"}"
		  triton_ssh_user      = "ubuntu"

		  triton_account                 = "${_triton_account}"
		  triton_key_path                = "${_triton_key_path}"
		  triton_key_id                  = "${_triton_key_id}"
		  triton_url                     = "${_triton_url}"
		  name                           = "${_name}"
		  master_triton_machine_package  = "${_master_triton_machine_package}"
		  mysqldb_triton_machine_package = "${_mysqldb_triton_machine_package}"
		  ha                             = ${_ha}
		  gcm_node_count                 = ${_gcm_node_count}
		  
		  ${rancher_server_image_line}
		  ${rancher_agent_image_line}
		  ${rancher_registry_line}
		  ${rancher_registry_username_line}
		  ${rancher_registry_password_line}
		  docker_engine_install_url = "${docker_engine_install_url}"
		}
		
		EOF
	)
}
setModuleAWSEnvironment() {
	(
		rancher_registry_line='# rancher_registry          = "docker-registry.joyent.com:5000"'
		rancher_registry_username_line='# rancher_registry_username = "username"'
		rancher_registry_password_line='# rancher_registry_password = "password"'
		if [ ! -z "${rancher_registry}" ]
		then
			rancher_registry_line="rancher_registry          = \"${rancher_registry}\""
			rancher_registry_username_line="rancher_registry_username = \"${rancher_registry_username}\""
			rancher_registry_password_line="rancher_registry_password = \"${rancher_registry_password}\""
		fi
		k8s_registry_line='# k8s_registry              = "docker-registry.joyent.com:5000"'
		k8s_registry_username_line='# k8s_registry_username     = "username"'
		k8s_registry_password_line='# k8s_registry_password     = "password"'
		if [ ! -z "${rancher_registry}" ]
		then
			k8s_registry_line="k8s_registry          = \"${k8s_registry}\""
			k8s_registry_username_line="k8s_registry_username = \"${k8s_registry_username}\""
			k8s_registry_password_line="k8s_registry_password = \"${k8s_registry_password}\""
		fi

		cd terraform
		__k8s_plane_isolation="none"
		if [ "$_ha" == "true" ]
		then
			__k8s_plane_isolation="required"
		fi
		cat >>create-rancher-env.tf <<-EOF

		module "${_name}" {
		  source = "./modules/aws-rancher-k8s"

		  api_url    = "http://\${element(data.terraform_remote_state.rancher.masters, 0)}:8080"
		  access_key = ""
		  secret_key = ""

		  name = "${_name}"

		  etcd_node_count          = "${_etcd_node_count}"
		  orchestration_node_count = "${_orchestration_node_count}"
		  compute_node_count       = "${_compute_node_count}"

		  k8s_plane_isolation = "${__k8s_plane_isolation}"

		  aws_access_key = "${_aws_access_key}"
		  aws_secret_key = "${_aws_secret_key}"

		  aws_region = "${_aws_region}"
		  aws_ami_id = "${_aws_ami_id}"

		  aws_public_key_path = "${_aws_public_key_path}"
		  aws_key_name        = "${_name}-key"

		  etcd_aws_instance_type          = "${_etcd_aws_instance_type}"
		  orchestration_aws_instance_type = "${_orchestration_aws_instance_type}"
		  compute_aws_instance_type       = "${_compute_aws_instance_type}"

		  ${rancher_registry_line}
		  ${rancher_registry_username_line}
		  ${rancher_registry_password_line}

		  ${k8s_registry_line}
		  ${k8s_registry_username_line}
		  ${k8s_registry_password_line}
		  docker_engine_install_url = "${docker_engine_install_url}"
		}

		EOF
	)
}
setModuleAzureEnvironment() {
	(
		rancher_registry_line='# rancher_registry          = "docker-registry.joyent.com:5000"'
		rancher_registry_username_line='# rancher_registry_username = "username"'
		rancher_registry_password_line='# rancher_registry_password = "password"'
		if [ ! -z "${rancher_registry}" ]
		then
			rancher_registry_line="rancher_registry          = \"${rancher_registry}\""
			rancher_registry_username_line="rancher_registry_username = \"${rancher_registry_username}\""
			rancher_registry_password_line="rancher_registry_password = \"${rancher_registry_password}\""
		fi
		k8s_registry_line='# k8s_registry              = "docker-registry.joyent.com:5000"'
		k8s_registry_username_line='# k8s_registry_username     = "username"'
		k8s_registry_password_line='# k8s_registry_password     = "password"'
		if [ ! -z "${rancher_registry}" ]
		then
			k8s_registry_line="k8s_registry          = \"${k8s_registry}\""
			k8s_registry_username_line="k8s_registry_username = \"${k8s_registry_username}\""
			k8s_registry_password_line="k8s_registry_password = \"${k8s_registry_password}\""
		fi
	
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

		  azure_resource_group_name           = "${_name}-k8s"
		  azure_virtual_network_name          = "${_name}-k8snetwork"
		  azure_subnet_name                   = "${_name}-k8ssubnet"
		  azurerm_network_security_group_name = "${_name}-k8sfirewall"

		  etcd_azure_size          = "${_etcd_azure_size}"
		  orchestration_azure_size = "${_orchestration_azure_size}"
		  compute_azure_size       = "${_compute_azure_size}"

		  ${rancher_registry_line}
		  ${rancher_registry_username_line}
		  ${rancher_registry_password_line}

		  ${k8s_registry_line}
		  ${k8s_registry_username_line}
		  ${k8s_registry_password_line}
		  docker_engine_install_url = "${docker_engine_install_url}"
		}

		EOF
	)
}
setModuleGCPEnvironment() {
	(
		rancher_registry_line='# rancher_registry          = "docker-registry.joyent.com:5000"'
		rancher_registry_username_line='# rancher_registry_username = "username"'
		rancher_registry_password_line='# rancher_registry_password = "password"'
		if [ ! -z "${rancher_registry}" ]
		then
			rancher_registry_line="rancher_registry          = \"${rancher_registry}\""
			rancher_registry_username_line="rancher_registry_username = \"${rancher_registry_username}\""
			rancher_registry_password_line="rancher_registry_password = \"${rancher_registry_password}\""
		fi
		k8s_registry_line='# k8s_registry              = "docker-registry.joyent.com:5000"'
		k8s_registry_username_line='# k8s_registry_username     = "username"'
		k8s_registry_password_line='# k8s_registry_password     = "password"'
		if [ ! -z "${rancher_registry}" ]
		then
			k8s_registry_line="k8s_registry          = \"${k8s_registry}\""
			k8s_registry_username_line="k8s_registry_username = \"${k8s_registry_username}\""
			k8s_registry_password_line="k8s_registry_password = \"${k8s_registry_password}\""
		fi

		cd terraform
		__k8s_plane_isolation="none"
		if [ "$_ha" == "true" ]
		then
			__k8s_plane_isolation="required"
		fi
		cat >>create-rancher-env.tf <<-EOF

		module "${_name}" {
		  source = "./modules/gcp-rancher-k8s"

		  api_url    = "http://\${element(data.terraform_remote_state.rancher.masters, 0)}:8080"
		  access_key = ""
		  secret_key = ""

		  name = "${_name}"

		  etcd_node_count          = "${_etcd_node_count}"
		  orchestration_node_count = "${_orchestration_node_count}"
		  compute_node_count       = "${_compute_node_count}"

		  k8s_plane_isolation = "${__k8s_plane_isolation}"

		  gcp_path_to_credentials = "${_gcp_path_to_credentials}"
		  gcp_project_id = "${_gcp_project_id}"

		  gcp_compute_region = "${_gcp_compute_region}"
		  gcp_instance_zone  = "${_gcp_instance_zone}"
		  compute_firewall   = "${_name}-firewall"

		  etcd_gcp_instance_type          = "${_etcd_gcp_instance_type}"
		  orchestration_gcp_instance_type = "${_orchestration_gcp_instance_type}"
		  compute_gcp_instance_type       = "${_compute_gcp_instance_type}"

		  ${rancher_registry_line}
		  ${rancher_registry_username_line}
		  ${rancher_registry_password_line}

		  ${k8s_registry_line}
		  ${k8s_registry_username_line}
		  ${k8s_registry_password_line}
		  docker_engine_install_url = "${docker_engine_install_url}"
		}

		EOF
	)
}
setModuleTritonEnvironment() {
	(
		rancher_registry_line='# rancher_registry          = "docker-registry.joyent.com:5000"'
		rancher_registry_username_line='# rancher_registry_username = "username"'
		rancher_registry_password_line='# rancher_registry_password = "password"'
		if [ ! -z "${rancher_registry}" ]
		then
			rancher_registry_line="rancher_registry          = \"${rancher_registry}\""
			rancher_registry_username_line="rancher_registry_username = \"${rancher_registry_username}\""
			rancher_registry_password_line="rancher_registry_password = \"${rancher_registry_password}\""
		fi
		k8s_registry_line='# k8s_registry              = "docker-registry.joyent.com:5000"'
		k8s_registry_username_line='# k8s_registry_username     = "username"'
		k8s_registry_password_line='# k8s_registry_password     = "password"'
		if [ ! -z "${rancher_registry}" ]
		then
			k8s_registry_line="k8s_registry          = \"${k8s_registry}\""
			k8s_registry_username_line="k8s_registry_username = \"${k8s_registry_username}\""
			k8s_registry_password_line="k8s_registry_password = \"${k8s_registry_password}\""
		fi

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

		  triton_image_name    = "${triton_image_name:-"ubuntu-certified-16.04"}"
		  triton_image_version = "${triton_image_version:-"20170619.1"}"

		  ${_network_choice}

		  etcd_triton_machine_package          = "${_etcd_triton_machine_package}"
		  orchestration_triton_machine_package = "${_orchestration_triton_machine_package}"
		  compute_triton_machine_package       = "${_compute_triton_machine_package}"

		  ${rancher_registry_line}
		  ${rancher_registry_username_line}
		  ${rancher_registry_password_line}

		  ${k8s_registry_line}
		  ${k8s_registry_username_line}
		  ${k8s_registry_password_line}
		  docker_engine_install_url = "${docker_engine_install_url}"
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

	local __MasterCNS __name
	__name=$_name
	__MasterCNS=$(triton instance get "${__name}-master-1" 2>&1 | jq '.dns_names' 2>&1 | grep "${__name}.svc.*.triton.zone" | tr -d '"' | tr -d ' ' | tr -d ',' || true)
	if [ "$__MasterCNS" ]
	then
		echo "    http://${__MasterCNS}:8080/"
	else
		echo "    http://${__MasterIP}:8080/"
	fi
	cat <<-EOM

	Next step is adding Kubernetes environments to be managed here.
	To start your first environment, run:
	    $scriptname -e

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

	local __MasterCNS __name
	__name=$(triton ls -l 2>&1 | grep "$__MasterIP" | awk '{print $2}' 2>&1 | sed 's/-master-1//g' || true)
	__MasterCNS=$(triton instance get "${__name}-master-1" 2>&1 | jq '.dns_names' 2>&1 | grep "${__name}.svc.*.triton.zone" | tr -d '"' | tr -d ' ' | tr -d ',' || true)
	if [ "$__MasterCNS" ]
	then
		cat <<-EOM
		Cluster Manager URL:
		    http://${__MasterCNS}:8080/settings/env
		Kubernetes Hosts URL:
		    http://${__MasterCNS}:8080/env/$($TERRAFORM output -state=terraform/terraform.tfstate -module="${_name}" -json | jq '.environment_id.value' | tr -d '"')/infra/hosts?mode=dot
		Kubernetes Health:
		    http://${__MasterCNS}:8080/env/$($TERRAFORM output -state=terraform/terraform.tfstate -module="${_name}" -json | jq '.environment_id.value' | tr -d '"')/apps/stacks?which=cattle
		EOM
	else
		cat <<-EOM
		Cluster Manager URL:
		    http://${__MasterIP}:8080/settings/env
		Kubernetes Hosts URL:
		    http://${__MasterIP}:8080/env/$($TERRAFORM output -state=terraform/terraform.tfstate -module="${_name}" -json | jq '.environment_id.value' | tr -d '"')/infra/hosts?mode=dot
		Kubernetes Health:
		    http://${__MasterIP}:8080/env/$($TERRAFORM output -state=terraform/terraform.tfstate -module="${_name}" -json | jq '.environment_id.value' | tr -d '"')/apps/stacks?which=cattle
		EOM
	fi

	cat <<-EOM
	
	NOTE: Nodes might take a few minutes to connect and come up.

	To start another environment, run:
	    $scriptname -e

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
			__TERRAFORM_URL="https://releases.hashicorp.com/terraform/0.10.8/terraform_0.10.8_$(uname | tr '[:upper:]' '[:lower:]')_amd64.zip"
			mkdir bin >/dev/null 2>&1 || true
			cd bin

			if [ ! -e "$__TERRAFORM_ZIP_FILE" ]
			then
				echo ""
				echo "Downloading Terraform v0.10.8 ..."
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
		$scriptname (-c|-e) [conf file]
		$scriptname -d <environment name>
	Options:
		-c  create a cluster manager on Triton
		-e  create and add a Kubernetes environment to an existing cluster manager
		-d  delete an environment
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
