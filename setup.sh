#!/usr/bin/env bash

set -o errexit
set -o pipefail
# set -o xtrace

db_name=cattle
db_user=cattle
db_pass=cattle

TERRAFORM=$(pwd)/bin/terraform

MYNAME=$0

main() {
    if [ ! -e $TERRAFORM ]; then
        echo "Getting the correct version of terraform ..."
        # Detect the platform
        rm -rf bin || true
        mkdir bin
        cd bin
        OS="`uname`"
        case $OS in
            'Linux')
                wget https://releases.hashicorp.com/terraform/0.9.11/terraform_0.9.11_linux_amd64.zip
                unzip -o terraform_0.9.11_linux_amd64.zip
                rm terraform_0.9.11_linux_amd64.zip
                cd ..
                ;;
            'FreeBSD')
                wget https://releases.hashicorp.com/terraform/0.9.11/terraform_0.9.11_freebsd_amd64.zip
                unzip terraform_0.9.11_freebsd_amd64.zip
                rm terraform_0.9.11_freebsd_amd64.zip
                cd ..
                ;;
            'Darwin')
                wget https://releases.hashicorp.com/terraform/0.9.11/terraform_0.9.11_darwin_amd64.zip
                unzip terraform_0.9.11_darwin_amd64.zip
                rm terraform_0.9.11_darwin_amd64.zip
                cd ..
                ;;
            'SunOS')
                wget https://releases.hashicorp.com/terraform/0.9.11/terraform_0.9.11_solaris_amd64.zip
                unzip terraform_0.9.11_solaris_amd64.zip
                rm terraform_0.9.11_solaris_amd64.zip
                cd ..
                ;;
            *)
                cd ..
                echo "Couldn't determine os type."
                exit 1
                ;;
        esac
        echo ""
        echo "Terraform for $OS added to $(pwd)/bin/ directory."
        echo ""
    fi
    
    if [[ ! -z "$1" && "$1" == "-c" ]]; then
        cleanRunner
        exit 0
    fi

    if [ -e terraform/rancher.tf ]; then
        echo "error: configuration for a previous run has been found"
        echo "    clean the configuration ($MYNAME -c)"
        exit
    fi

    mkdir ansible/tmp > /dev/null 2>&1 || true
    # SET default variables
    setVarDefaults
    # SET configuration from current triton profile (triton account information)
    setConfigFromTritonENV
    # GET updated configuration from user
    getConfigFromUser
    # VERIFY with user that parameters are correct
    verifyConfig
    # UPDATE config file with parameters
    setConfigToFile

    exportVars

    echo "################################################################################"
    echo "### Starting terraform tasks..."
    echo "################################################################################"
    sleep 2
    runTerraformTasks
    echo "################################################################################"
    echo "### Creating ansible configs..."
    echo "################################################################################"
    sleep 2
    createAnsibleConfigs
    echo "################################################################################"
    echo "### Running ansible tasks..."
    echo "################################################################################"
    sleep 2
    runAnsible

    echo ""
    echo "Congratulations, your Kubernetes cluster setup has been complete."
    echo ""
    echo "It will take a few minutes for all the Kubernetes process to start up before you can access Kubernetes Dashboard"
    echo "----> To check what processes/containers are coming up, go to http://$(tail -1 terraform/masters.ip):8080/env/$(cat ansible/tmp/kubernetes_environment.id)/infra/containers"
    echo "    once all these containers are up, you should be able to access Kubernetes by its dashboard or using CLI"

    echo "Waiting on Kubernetes dashboard to come up."
    echo ""

    KUBERNETES_DASHBOARD_UP=false
    DASHBOARD_CONTAINER_COUNT=0
    while ! "$KUBERNETES_DASHBOARD_UP"; do
        echo -ne "."
        sleep 1
        if ! ((`date +%s` % 15)); then
            DASHBOARD_CONTAINER_COUNT=0
            cRetVal=$(curl --connect-timeout 5 --max-time 5 -s http://$(tail -1 terraform/masters.ip):8080/r/projects/$(cat ansible/tmp/kubernetes_environment.id)/kubernetes-dashboard:9090/)
            if [ $(echo $cRetVal | grep -i kubernetes | wc -l) -ne 0 ]; then
                KUBERNETES_DASHBOARD_UP=true
            # BUG kill stuck dashboard container
            else
                for node in $(cat terraform/hosts.ip); do
                    let DASHBOARD_CONTAINER_COUNT=$DASHBOARD_CONTAINER_COUNT+$(ssh -o StrictHostKeyChecking=no root@$node docker ps | grep kubernetes-dashboard | awk '{print $1}' | wc -w | sed "s/ //g") || true
                done
                if [[ $DASHBOARD_CONTAINER_COUNT -eq 2 && $(curl --connect-timeout 5 --max-time 5 -s http://$(tail -1 terraform/masters.ip):8080/r/projects/$(cat ansible/tmp/kubernetes_environment.id)/kubernetes-dashboard:9090/ | grep -i "Service Unavailable" | wc -l) -ne 0 ]]; then
                    for node in $(cat terraform/hosts.ip); do
                        dashboard_container="$(ssh -o StrictHostKeyChecking=no root@$node docker ps | grep k8s_POD.*dashboard | awk '{print $1}')" || true
                        if [ "${dashboard_container:-}" != "" ]; then
                            ssh -o StrictHostKeyChecking=no root@$node "docker stop -t0 $dashboard_container" >> /dev/null 2>&1
                        fi
                        echo -ne "."
                    done
                fi
            fi
        fi
    done

    echo ""
    echo "----> Kubernetes dashboard is at http://$(tail -1 terraform/masters.ip):8080/r/projects/$(cat ansible/tmp/kubernetes_environment.id)/kubernetes-dashboard:9090/"
    echo "----> Kubernetes CLI config is at http://$(tail -1 terraform/masters.ip):8080/env/$(cat ansible/tmp/kubernetes_environment.id)/kubernetes/kubectl"
    echo ""
    echo "    CONGRATULATIONS, YOU HAVE CONFIGURED YOUR KUBERNETES ENVIRONMENT!"
}

getArgument() {
    # $1 message
    # $2 default
    while true; do
        if [ -z ${2+x} ]; then
            read -p "$1 " theargument
        else
            read -p "$1 ($2) " theargument
            if [ -z "$theargument" ]; then
                echo $2
            else
                echo $theargument
            fi
            break
        fi
    done
}
runAnsible() {
    cd ansible
    ansible-playbook -i hosts dockerStorageSetup.yml
    ansible-playbook -i hosts dockerSetup.yml
    if $RANCHER_HA; then
        ansible-playbook -i hosts rancherServers.yml
    else
        ansible-playbook -i hosts rancherServer.yml
    fi
    echo "Waiting for Rancher to come up ..."
    echo ""
    while [[ $(curl --connect-timeout 5 --max-time 5 -s http://$(tail -1 ../terraform/masters.ip):8080/ | grep $(tail -1 ../terraform/masters.ip) | wc -l) -eq 0 ]]; do
        echo -ne "."
        sleep 1
    done
    echo ""
    sleep 5 # wait few seconds before creating environment
    if $SEPARATE_PLANE; then
        curl -X PUT -H 'Accept: application/json' -H 'Content-Type: application/json' -d '{"accountId":null, "data":{"fields":{"stacks":[{"name":"healthcheck", "templateId":"library:infra*healthcheck"}, {"name":"kubernetes", "templateId":"library:infra*k8s", "answers":{"CONSTRAINT_TYPE":"required", "CLOUD_PROVIDER":"rancher", "REGISTRY":"", "DISABLE_ADDONS":"false", "POD_INFRA_CONTAINER_IMAGE":"gcr.io/google_containers/pause-amd64:3.0", "INFLUXDB_HOST_PATH":"", "EMBEDDED_BACKUPS":true, "BACKUP_PERIOD":"15m0s", "BACKUP_RETENTION":"24h", "ETCD_HEARTBEAT_INTERVAL":"500", "ETCD_ELECTION_TIMEOUT":"5000"}}, {"name":"network-services", "templateId":"library:infra*network-services"}, {"name":"ipsec", "templateId":"library:infra*ipsec"}]}}, "description":"Default Kubernetes template", "externalId":"catalog://library:project*kubernetes:0", "id":0, "isPublic":true, "kind":"projectTemplate", "name":"Kubernetes", "removeTime":null, "removed":null, "state":"active", "transitioning":"no", "transitioningMessage":null, "transitioningProgress":0, "stacks":[{"type":"catalogTemplate", "name":"healthcheck", "templateId":"library:infra*healthcheck"}, {"type":"catalogTemplate", "answers":{"CONSTRAINT_TYPE":"required", "CLOUD_PROVIDER":"rancher", "REGISTRY":"", "DISABLE_ADDONS":"false", "POD_INFRA_CONTAINER_IMAGE":"gcr.io/google_containers/pause-amd64:3.0", "INFLUXDB_HOST_PATH":"", "EMBEDDED_BACKUPS":true, "BACKUP_PERIOD":"15m0s", "BACKUP_RETENTION":"24h", "ETCD_HEARTBEAT_INTERVAL":"500", "ETCD_ELECTION_TIMEOUT":"5000"}, "name":"kubernetes", "templateId":"library:infra*k8s"}, {"type":"catalogTemplate", "name":"network-services", "templateId":"library:infra*network-services"}, {"type":"catalogTemplate", "name":"ipsec", "templateId":"library:infra*ipsec"}]}' "http://$(tail -1 ../terraform/masters.ip):8080/v2-beta/projecttemplates/$(cat tmp/kubernetes_template_id.id)" > /dev/null 2>&1
    fi
    ansible-playbook -i hosts clusterUp.yml
    cd ..
}
createAnsibleConfigs() {
    if [[ ! -e terraform/masters.ip || ! -e terraform/hosts.ip ]]; then
        echo "Terraform had too many errors. Make sure you haven't reached your provisioning limit."
        exit 1
    fi
    echo "Creating ansible hosts file and variable files"
    rm -f ansible/hosts 2> /dev/null
    if $RANCHER_HA; then
        echo "[MYSQLDB]" >> ansible/hosts
        cat terraform/mysqldb.ip >> ansible/hosts
    fi
    echo "[MASTER]" >> ansible/hosts
    cat terraform/masters.ip >> ansible/hosts
    echo "[HOST]" >> ansible/hosts
    cat terraform/hosts.ip >> ansible/hosts
    if $SEPARATE_PLANE; then
        echo "[K8SSRVS]" >> ansible/hosts
        cat terraform/k8sha.ip >> ansible/hosts
        echo "[K8SETCD]" >> ansible/hosts
        cat terraform/k8setcd.ip >> ansible/hosts
    fi
    echo "    created: ansible/hosts"
    master=$(tail -1 terraform/masters.ip)
    echo "master: $master" > ansible/roles/ranchermaster/vars/vars.yml
    if $RANCHER_HA; then
        mysqldb=$(tail -1 terraform/mysqldb.ip)
        echo "mysqldb: $mysqldb" >> ansible/roles/ranchermaster/vars/vars.yml
        echo "db_user: $db_user" >> ansible/roles/ranchermaster/vars/vars.yml
        echo "db_pass: $db_pass" >> ansible/roles/ranchermaster/vars/vars.yml
        echo "db_name: $db_name" >> ansible/roles/ranchermaster/vars/vars.yml
    fi
    echo "kubernetes_name: \"$(echo $KUBERNETES_NAME | sed 's/"//g')\"" >> ansible/roles/ranchermaster/vars/vars.yml
    echo "kubernetes_description: \"$(echo $KUBERNETES_DESCRIPTION | sed 's/"//g')\"" >> ansible/roles/ranchermaster/vars/vars.yml
    cd ansible
    sed "s;private_key_file = .*$;private_key_file = $(echo $SDC_KEY | sed 's/"//g');g" ansible.cfg > tmp.cfg && mv tmp.cfg ansible.cfg
    cd ..

    echo "    created: ansible/roles/ranchermaster/vars/vars.yml"
}
runTerraformTasks() {
    if [ -e terraform/rancher.tf ]
    then
        echo "warning: a previous terraform configuration has been found"
        echo "    skipping terraform configuration and execution..."
    else
        echo "Generating terraform configs for environment..."
        updateTerraformConfig provider triton

        if $RANCHER_HA; then
            echo "    Master hostname: ${RANCHER_MASTER_HOSTNAME}1"
            echo "                     ${RANCHER_MASTER_HOSTNAME}2"
            echo "                     ${RANCHER_MASTER_HOSTNAME}db"
            updateTerraformConfig master $(echo ${RANCHER_MASTER_HOSTNAME}1 | sed 's/"//g')
            updateTerraformConfig master $(echo ${RANCHER_MASTER_HOSTNAME}2 | sed 's/"//g')
            updateTerraformConfig mysqldb $(echo ${RANCHER_MASTER_HOSTNAME}db | sed 's/"//g')
        else
            echo "    Master hostname: $RANCHER_MASTER_HOSTNAME"
            updateTerraformConfig master $(echo $RANCHER_MASTER_HOSTNAME | sed 's/"//g')
        fi

        for (( i = 1; i <= $KUBERNETES_NUMBER_OF_NODES; i++ ))
        do
            echo "    Kubernetes node $i: $(echo $KUBERNETES_NODE_HOSTNAME_BEGINSWITH | sed 's/"//g')$i"
            updateTerraformConfig host $(echo $KUBERNETES_NODE_HOSTNAME_BEGINSWITH$i | sed 's/"//g')
        done

        if $SEPARATE_PLANE; then
            updateTerraformConfig k8setcd $(echo ${KUBERNETES_NODE_HOSTNAME_BEGINSWITH}etcd1 | sed 's/"//g')
            updateTerraformConfig k8setcd $(echo ${KUBERNETES_NODE_HOSTNAME_BEGINSWITH}etcd2 | sed 's/"//g')
            updateTerraformConfig k8setcd $(echo ${KUBERNETES_NODE_HOSTNAME_BEGINSWITH}etcd3 | sed 's/"//g')
            updateTerraformConfig k8sha $(echo ${KUBERNETES_NODE_HOSTNAME_BEGINSWITH}srvs1 | sed 's/"//g')
            updateTerraformConfig k8sha $(echo ${KUBERNETES_NODE_HOSTNAME_BEGINSWITH}srvs2 | sed 's/"//g')
            updateTerraformConfig k8sha $(echo ${KUBERNETES_NODE_HOSTNAME_BEGINSWITH}srvs3 | sed 's/"//g')
        fi
        cd terraform
        echo "Starting terraform tasks"
        # terraform init --plugin-dir=$GOBIN
        $TERRAFORM get
        $TERRAFORM apply
        echo "    terraform tasks completed"
        cd ..
    fi
}
updateTerraformConfig() {
    if [ $1 == "k8sha" ]; then
        echo ""  >> terraform/rancher.tf
        echo "module \"$2\" {" >> terraform/rancher.tf
        echo "    source = \"k8sha\"" >> terraform/rancher.tf
        echo "    hostname = \"$2\"" >> terraform/rancher.tf
        echo "    networks = [\"$(echo $RANCHER_MASTER_NETWORKS | sed 's/,/","/g')\"]" >> terraform/rancher.tf
        echo "    root_authorized_keys = \"\${file(\"$(echo $SDC_KEY | sed 's/"//g')\")}\"" >> terraform/rancher.tf
        # echo "    image = \"0867ef86-e69d-4aaa-ba3b-8d2aef0c204e\"" >> terraform/rancher.tf
        echo "    package = \"$(echo $HOST_PACKAGE | sed 's/"//g')\"" >> terraform/rancher.tf
        echo "}" >> terraform/rancher.tf
        return
    fi
    if [ $1 == "k8setcd" ]; then
        echo ""  >> terraform/rancher.tf
        echo "module \"$2\" {" >> terraform/rancher.tf
        echo "    source = \"k8setcd\"" >> terraform/rancher.tf
        echo "    hostname = \"$2\"" >> terraform/rancher.tf
        echo "    networks = [\"$(echo $RANCHER_MASTER_NETWORKS | sed 's/,/","/g')\"]" >> terraform/rancher.tf
        echo "    root_authorized_keys = \"\${file(\"$(echo $SDC_KEY | sed 's/"//g')\")}\"" >> terraform/rancher.tf
        # echo "    image = \"0867ef86-e69d-4aaa-ba3b-8d2aef0c204e\"" >> terraform/rancher.tf
        echo "    package = \"$(echo $HOST_PACKAGE | sed 's/"//g')\"" >> terraform/rancher.tf
        echo "}" >> terraform/rancher.tf
        return
    fi
    if [ $1 == "master" ]; then
        echo ""  >> terraform/rancher.tf
        echo "module \"$2\" {" >> terraform/rancher.tf
        echo "    source = \"master\"" >> terraform/rancher.tf
        echo "    hostname = \"$2\"" >> terraform/rancher.tf
        echo "    networks = [\"$(echo $RANCHER_MASTER_NETWORKS | sed 's/,/","/g')\"]" >> terraform/rancher.tf
        echo "    root_authorized_keys = \"\${file(\"$(echo $SDC_KEY | sed 's/"//g')\")}\"" >> terraform/rancher.tf
        # echo "    image = \"0867ef86-e69d-4aaa-ba3b-8d2aef0c204e\"" >> terraform/rancher.tf
        echo "    package = \"$(echo $HOST_PACKAGE | sed 's/"//g')\"" >> terraform/rancher.tf
        echo "}" >> terraform/rancher.tf
        return
    fi
    if [ $1 == "mysqldb" ]; then
        echo ""  >> terraform/rancher.tf
        echo "module \"$2\" {" >> terraform/rancher.tf
        echo "    source = \"mysqldb\"" >> terraform/rancher.tf
        echo "    hostname = \"$2\"" >> terraform/rancher.tf
        echo "    networks = [\"$(echo $RANCHER_MASTER_NETWORKS | sed 's/,/","/g')\"]" >> terraform/rancher.tf
        echo "    root_authorized_keys = \"\${file(\"$(echo $SDC_KEY | sed 's/"//g')\")}\"" >> terraform/rancher.tf
        # echo "    image = \"0867ef86-e69d-4aaa-ba3b-8d2aef0c204e\"" >> terraform/rancher.tf
        echo "    package = \"$(echo $HOST_PACKAGE | sed 's/"//g')\"" >> terraform/rancher.tf
        echo "}" >> terraform/rancher.tf
        return
    fi
    if [ $1 == "host" ]; then
        echo ""  >> terraform/rancher.tf
        echo "module \"$2\" {" >> terraform/rancher.tf
        echo "    source = \"host\"" >> terraform/rancher.tf
        echo "    hostname = \"$2\"" >> terraform/rancher.tf
        echo "    networks = [\"$(echo $KUBERNETES_NODE_NETWORKS | sed 's/,/","/g')\"]" >> terraform/rancher.tf
        echo "    root_authorized_keys = \"\${file(\"$(echo $SDC_KEY | sed 's/"//g')\")}\"" >> terraform/rancher.tf
        # echo "    image = \"0867ef86-e69d-4aaa-ba3b-8d2aef0c204e\"" >> terraform/rancher.tf
        echo "    package = \"$(echo $HOST_PACKAGE | sed 's/"//g')\"" >> terraform/rancher.tf
        echo "}" >> terraform/rancher.tf
        return
    fi
    if [ $1 == "provider" ]; then
        echo "provider \"triton\" {" > terraform/rancher.tf
        echo "    account = \"$(echo $SDC_ACCOUNT | sed 's/"//g')\"" >> terraform/rancher.tf
        echo "    key_material = \"\${file(\"$(echo $SDC_KEY | sed 's/"//g')\")}\"" >> terraform/rancher.tf
        echo "    key_id = \"$(echo $SDC_KEY_ID | sed 's/"//g')\"" >> terraform/rancher.tf
        echo "    url = \"$(echo $SDC_URL | sed 's/"//g')\"" >> terraform/rancher.tf
        echo "}" >> terraform/rancher.tf
        return
    fi
    echo "error: problem updating terraform configuration..."
    exit 1
}
setConfigToFile() {
    echo "RANCHER_MASTER_NETWORKS=$RANCHER_MASTER_NETWORKS" >> config
    echo "KUBERNETES_NODE_NETWORKS=$KUBERNETES_NODE_NETWORKS" >> config
    echo "KUBERNETES_NUMBER_OF_NODES=$KUBERNETES_NUMBER_OF_NODES" >> config
    echo "KUBERNETES_NAME=\"$KUBERNETES_NAME\"" >> config
    echo "KUBERNETES_DESCRIPTION=\"$KUBERNETES_DESCRIPTION\"" >> config
    echo "RANCHER_MASTER_HOSTNAME=\"$RANCHER_MASTER_HOSTNAME\"" >> config
    echo "KUBERNETES_NODE_HOSTNAME_BEGINSWITH=\"$KUBERNETES_NODE_HOSTNAME_BEGINSWITH\"" >> config
    echo "HOST_PACKAGE=\"$HOST_PACKAGE\"" >> config
}
setConfigFromTritonENV() {
    eval "$(triton env)"
    echo "SDC_URL=\"$SDC_URL\"" >> config
    echo "SDC_ACCOUNT=\"$SDC_ACCOUNT\"" >> config
    echo "SDC_KEY_ID=\"$SDC_KEY_ID\"" >> config

    local foundKey=false
    for f in $(ls ~/.ssh); do
        if [[ "$(ssh-keygen -E md5 -lf ~/.ssh/$(echo $f | sed 's/.pub$//') 2> /dev/null | awk '{print $2}' | sed 's/^MD5://')" == "$SDC_KEY_ID" ]]; then
            echo "SDC_KEY=\"~/.ssh/$(echo $f | sed 's/.pub$//')\"" >> config
            foundKey=true
            break
        fi
        # old version of ssh-keygen, defaults to md5
        if [ "$(ssh-keygen -l -f ~/.ssh/$(echo $f | sed 's/.pub$//') 2> /dev/null | awk '{print $2}' | grep ^SHA256)" == "" ]; then
            if [[ "$(ssh-keygen -l -f ~/.ssh/$(echo $f | sed 's/.pub$//') 2> /dev/null | awk '{print $2}')" == "$SDC_KEY_ID" ]]; then
                echo "SDC_KEY=\"~/.ssh/$(echo $f | sed 's/.pub$//')\"" >> config
                foundKey=true
                break
            fi
        fi
    done
    if ! "$foundKey" ; then
        echo "error: couldn't find the ssh key associated with fingerprint $SDC_KEY_ID in ~/.ssh/ directory..."
        echo "    Clean the setup and make sure your triton profile is set up."
        echo "    To confirm your profile is set up, try \`triton info\`."
        cleanRunner
        exit 1
    fi
    echo "" >> config
}
setVarDefaults() {
    if [ -e config ]; then
        echo "error: old configuration found"
        cleanRunner
    fi
    KUBERNETES_NAME="k8s dev"
    KUBERNETES_DESCRIPTION=$KUBERNETES_NAME
    RANCHER_MASTER_HOSTNAME="kubemaster"
    KUBERNETES_NODE_HOSTNAME_BEGINSWITH="kubenode"
    KUBERNETES_NUMBER_OF_NODES=1
    RANCHER_MASTER_NETWORKS=
    KUBERNETES_NODE_NETWORKS=
    HOST_PACKAGE=
    echo "ANSIBLE_HOST_KEY_CHECKING=False" >> config
}
getConfigFromUser() {
    # get networks from the current triton profile to prompt
    local networks=$(triton networks -oname,id | grep -v "^NAME.*ID$" | tr -s " " | tr " " "=" | sort)
    # get packages for the current triton profile to prompt
    local packages=$(triton packages -oname,id | grep "\-kvm-" | grep -v "^NAME.*ID$" | tr -s " " | tr " " "=" | sort)

    local tmp=0
    local gotValidInput=false
    local tmp_ValidatedInput
    echo "---------------"
    KUBERNETES_NAME=$(getArgument "Name your Kubernetes environment:" "$(echo $KUBERNETES_NAME | sed 's/"//g')")
    echo "---------------"
    if [[ $KUBERNETES_DESCRIPTION == "" ]]; then
        KUBERNETES_DESCRIPTION=$(getArgument "Describe this Kubernetes environment:" "$(echo $KUBERNETES_NAME | sed 's/"//g')")
    else
        KUBERNETES_DESCRIPTION=$(getArgument "Describe this Kubernetes environment:" "$(echo $KUBERNETES_DESCRIPTION | sed 's/"//g')")
    fi
    echo "---------------"
    gotValidInput=false
    while ! $gotValidInput; do
        read -p "Would you like HA for Kubernetes Cluster Manager (+3 VMs) (yes | no)? " yn
        case $yn in
            yes )
                RANCHER_HA=true
                gotValidInput=true
                ;;
            no )
                RANCHER_HA=false
                gotValidInput=true
                ;;
            * ) echo "Please answer yes or no.";;
        esac
    done
    echo "---------------"
    gotValidInput=false
    while ! $gotValidInput; do
        read -p "Run Kubernetes Management Services on dedicated nodes (+3 VMs for etcd, +3 VMs for K8s services - apiserver/scheduler/controllermanager...) (yes | no)? " yn
        case $yn in
            yes )
                SEPARATE_PLANE=true
                gotValidInput=true
                # echo "SEPARATE_PLANE: true" >> ansible/roles/ranchermaster/vars/vars.yml
                ;;
            no )
                SEPARATE_PLANE=false
                gotValidInput=true
                ;;
            * ) echo "Please answer yes or no.";;
        esac
    done
    echo "---------------"
    gotValidInput=false
    while ! $gotValidInput; do
        tmp_ValidatedInput=$(getArgument "Hostname of the master:" "$(echo $RANCHER_MASTER_HOSTNAME | sed 's/"//g')")
        if [[ $tmp_ValidatedInput =~ ^[a-zA-Z][0-9a-zA-Z]+$ ]]; then
            gotValidInput=true
        else
            echo "error: Enter a valid hostname or leave blank to use the default."
            echo "    Must start with a letter and can only include letters and numbers"
        fi
    done
    RANCHER_MASTER_HOSTNAME=$tmp_ValidatedInput
    echo "---------------"
    gotValidInput=false
    while ! $gotValidInput; do
        tmp_ValidatedInput=$(getArgument "Enter a string to use for appending to hostnames of all the nodes:" "$(echo $KUBERNETES_NODE_HOSTNAME_BEGINSWITH | sed 's/"//g')")
        if [[ $tmp_ValidatedInput =~ ^[a-zA-Z][0-9a-zA-Z]+$ ]]; then
            gotValidInput=true
        else
            echo "error: Enter a valid value or leave blank to use the default."
            echo "    Must start with a letter and can only include letters and numbers"
        fi
    done
    KUBERNETES_NODE_HOSTNAME_BEGINSWITH=$tmp_ValidatedInput
    echo "---------------"
    # HARD LIMIT: 1-9 nodes allowed only since this setup has no HA
    gotValidInput=false
    while ! $gotValidInput; do
        tmp_ValidatedInput=$(getArgument "How many nodes should this Kubernetes cluster have:" "$(echo $KUBERNETES_NUMBER_OF_NODES | sed 's/"//g')")
        if [[ $tmp_ValidatedInput =~ ^[1-9]$ ]]; then
            gotValidInput=true
        else
            echo "error: Enter a valid value (1-9) or leave blank to use the default."
        fi
    done
    KUBERNETES_NUMBER_OF_NODES=$tmp_ValidatedInput
    echo "---------------"
    echo "From the networks below:"
    # print options and find location for "Joyent-SDC-Public"
    local publicNetworkLocation=1
    local countNetwork
    tmp=0
    for network in $networks; do
        tmp=$((tmp + 1))
        echo -e "$tmp.\t$(echo $network | sed 's/=/  /g')"
        # get default location of public network
        if [[ "$network" == "Joyent-SDC-Public="* ]]; then
            publicNetworkLocation=$tmp
        fi
    done
    countNetwork=$tmp

    # set publicNetworkLocation to RANCHER_MASTER_NETWORKS, if it wasn't set already
    if [[ $RANCHER_MASTER_NETWORKS == "" ]]; then
        RANCHER_MASTER_NETWORKS=$(getNetworkIDs $publicNetworkLocation)
    fi

    # get input for network and validate to make sure the input provided is within the limit (number of networks)
    gotValidInput=false
    while ! $gotValidInput; do
        tmp_RANCHER_MASTER_NETWORKS=$(getArgument "What networks should the master be a part of, provide comma separated values:" "$(echo $RANCHER_MASTER_NETWORKS | sed 's/"//g')")
        RANCHER_MASTER_NETWORKS=$(echo $RANCHER_MASTER_NETWORKS | tr ',' '\n' | sort | uniq | tr '\n' ',' | sed 's/\(.*\),$/\1/')
        tmp_RANCHER_MASTER_NETWORKS=$(echo $tmp_RANCHER_MASTER_NETWORKS | tr ',' '\n' | sort | uniq | tr '\n' ',' | sed 's/\(.*\),$/\1/')

        # if valid input was given, move forward, else quit
        if [[ $(echo $tmp_RANCHER_MASTER_NETWORKS | grep '^[1-9][0-9]\?\(,[1-9][0-9]\?\)*$' 2> /dev/null) ]]; then
            gotValidInput=true
            for network in $(echo $tmp_RANCHER_MASTER_NETWORKS | tr "," "\n"); do
                if [[ "$network" -gt "$countNetwork" || "$network" -lt 1 ]]; then
                    echo "error: Enter a valid option or leave blank to use the default."
                    echo "    Values should be comma separated between 1 and $countNetwork."
                    gotValidInput=false
                fi
            done

            if $gotValidInput; then
                RANCHER_MASTER_NETWORKS=$(getNetworkIDs $tmp_RANCHER_MASTER_NETWORKS)
            fi

        elif [[ $tmp_RANCHER_MASTER_NETWORKS == $RANCHER_MASTER_NETWORKS ]]; then
            gotValidInput=true
        else
            echo "error: Enter a valid option or leave blank to use the default."
            echo "    Values should be comma separated between 1 and $countNetwork."
        fi
    done
    echo "---------------"
    echo "From the networks below:"
    # print options
    tmp=0
    for network in $networks; do
        tmp=$((tmp + 1))
        echo -e "$tmp.\t$(echo $network | sed 's/=/  /g')"
    done

    # set publicNetworkLocation to KUBERNETES_NODE_NETWORKS, if it wasn't set already
    if [[ $KUBERNETES_NODE_NETWORKS == "" ]]; then
        KUBERNETES_NODE_NETWORKS=$(getNetworkIDs $publicNetworkLocation)
    fi

    # get input for network and validate to make sure the input provided is within the limit (number of networks)
    gotValidInput=false
    while ! $gotValidInput; do
        tmp_KUBERNETES_NODE_NETWORKS=$(getArgument "What networks should the nodes be a part of, provide comma separated values:" "$(echo $KUBERNETES_NODE_NETWORKS | sed 's/"//g')")
        KUBERNETES_NODE_NETWORKS=$(echo $KUBERNETES_NODE_NETWORKS | tr ',' '\n' | sort | uniq | tr '\n' ',' | sed 's/\(.*\),$/\1/')
        tmp_KUBERNETES_NODE_NETWORKS=$(echo $tmp_KUBERNETES_NODE_NETWORKS | tr ',' '\n' | sort | uniq | tr '\n' ',' | sed 's/\(.*\),$/\1/')

        # if valid input was given, move forward, else quit
        if [[ $(echo $tmp_KUBERNETES_NODE_NETWORKS | grep '^[1-9][0-9]\?\(,[1-9][0-9]\?\)*$' 2> /dev/null) ]]; then
            gotValidInput=true
            for network in $(echo $tmp_KUBERNETES_NODE_NETWORKS | tr "," "\n"); do
                if [[ "$network" -gt "$countNetwork" || "$network" -lt 1 ]]; then
                    echo "error: Enter a valid option or leave blank to use the default."
                    echo "    Values should be comma separated between 1 and $countNetwork."
                    gotValidInput=false
                fi
            done

            if $gotValidInput; then
                KUBERNETES_NODE_NETWORKS=$(getNetworkIDs $tmp_KUBERNETES_NODE_NETWORKS)
            fi

        elif [[ $tmp_KUBERNETES_NODE_NETWORKS == $KUBERNETES_NODE_NETWORKS ]]; then
            gotValidInput=true
        else
            echo "error: Enter a valid option or leave blank to use the default."
            echo "    Values should be comma separated between 1 and $countNetwork."
        fi
    done
    echo "---------------"
    echo "From the packages below:"
    # print options and find location for "k4-highcpu-kvm-7.75G"
    local packageLocation=1
    local countPackages
    tmp=0
    for package in $packages; do
        tmp=$((tmp + 1))
        echo -e "$tmp.\t$(echo $package | sed 's/=/  /g')"
        # get default location of package
        if [[ "$package" == "k4-highcpu-kvm-7.75G="* ]]; then
            packageLocation=$tmp
        fi
    done
    countPackages=$tmp

    # set packageLocation to HOST_PACKAGE, if it wasn't set already
    if [[ $HOST_PACKAGE == "" ]]; then
        HOST_PACKAGE=$(getPackageID $packageLocation)
    fi

    # get input for package and validate to make sure the input provided is within the limit (number of packages)
    gotValidInput=false
    while ! $gotValidInput; do
        tmp_HOST_PACKAGE=$(getArgument "What KVM package should the master and nodes run on:" "$(echo $HOST_PACKAGE | sed 's/"//g')")

        # if valid input was given, move forward, else quit
        if [[ $(echo $tmp_HOST_PACKAGE | grep '^[1-9][0-9]*$' 2> /dev/null) ]]; then
            gotValidInput=true
            for package in $(echo $tmp_HOST_PACKAGE | tr "," "\n"); do
                if [[ "$package" -gt "$countPackages" || "$package" -lt 1 ]]; then
                    echo "error: Enter a valid option or leave blank to use the default."
                    echo "    Value should be between 1 and $countPackages."
                    gotValidInput=false
                fi
            done

            if $gotValidInput; then
                HOST_PACKAGE=$(getPackageID $tmp_HOST_PACKAGE)
                echo "entered $tmp_HOST_PACKAGE and got $HOST_PACKAGE"
            fi

        elif [[ $tmp_HOST_PACKAGE == $(echo $HOST_PACKAGE | sed 's/"//g') ]]; then
            gotValidInput=true
        else
            echo "error: Enter a valid option or leave blank to use the default."
            echo "    Value should be between 1 and $countPackages."
        fi
    done
    HOST_PACKAGE=$(echo $HOST_PACKAGE | sed 's/"//g')
}
verifyConfig() {
    echo "################################################################################"
    echo "Verify that the following configuration is correct:"
    echo ""
    echo "Name of kubernetes environment: $KUBERNETES_NAME"
    echo "Kubernetes environment description: $KUBERNETES_DESCRIPTION"
    if $RANCHER_HA; then
        echo "Cluster Manager hosts: ${RANCHER_MASTER_HOSTNAME}1, ${RANCHER_MASTER_HOSTNAME}2, ${RANCHER_MASTER_HOSTNAME}db"
    else
        echo "Master hostname: $RANCHER_MASTER_HOSTNAME"
    fi
    if $SEPARATE_PLANE; then
        echo "    There will be three nodes created for etcd and three for Kubernetes HA."
    fi
    echo "All node hostnames will start with: $KUBERNETES_NODE_HOSTNAME_BEGINSWITH"
    echo "Kubernetes environment will have $KUBERNETES_NUMBER_OF_NODES nodes"
    echo "Master server will be part of these networks: $RANCHER_MASTER_NETWORKS"
    echo "Kubernetes nodes will be a part of these networks: $KUBERNETES_NODE_NETWORKS"
    echo "This package will be used for all the hosts: $HOST_PACKAGE"
    echo ""
    echo "Make sure the above information is correct before answering:"
    echo "    to view list of networks call \"triton networks -l\""
    echo "    to view list of packages call \"triton packages -l\""
    echo "WARN: Make sure that the nodes and master are part of networks that can communicate with each other and this system from which the setup is running."


    while true; do
    read -p "Is the above config correct (yes | no)? " yn
    case $yn in
        yes )
            break
            ;;
        no )
            exit 0
            ;;
        * ) echo "Please answer yes or no.";;
    esac
    done
}
cleanRunner() {
    echo "Clearing settings...."
    while true; do
        if [ -e terraform/masters.ip ]; then
            echo "WARNING: You are about to destroy KVMs associated with this Rancher cluster."

            read -p "Do you wish to destroy the KVMs and reset configuration (yes | no)? " yn
        else
            read -p "Do you wish to reset configuration (yes | no)? " yn
        fi
        case $yn in
            yes )
                if [ -e terraform/rancher.tf ]; then
                    cd terraform
                    echo "    destroying images..."
                    $TERRAFORM destroy -force 2> /dev/null || true
                    cd ..
                fi
                if [[ -e terraform/hosts.ip  &&  -e terraform/masters.ip  &&  -e ~/.ssh/known_hosts ]]; then
                    for host_key in $(cat terraform/hosts.ip terraform/masters.ip); do
                        ssh-keygen -R $host_key 2>&1 >> /dev/null
                    done
                fi
                rm -rf terraform/hosts.ip terraform/masters.ip terraform/k8setcd.ip terraform/k8sha.ip terraform/mysqldb.ip terraform/terraform.* terraform/.terraform* terraform/rancher.tf 2>&1 >> /dev/null
                rm -rf ansible/roles/ranchermaster/vars/vars.yml ansible/tmp/kubernetes_* 2>&1 >> /dev/null
                echo 'master: 64.30.129.229' > ansible/roles/ranchermaster/vars/vars.yml
                echo 'kubernetes_name: "k8s dev"' >> ansible/roles/ranchermaster/vars/vars.yml
                echo 'kubernetes_description: "k8s dev"' >> ansible/roles/ranchermaster/vars/vars.yml
                sed "s~private_key_file = .*$~private_key_file = ~g" ansible/ansible.cfg > tmp && mv tmp ansible/ansible.cfg
                rm -f  ansible/hosts ansible/*retry ansible/ansible.cfg.tmp 2>&1 >> /dev/null
                rm -rf config tmp/* 2>&1 >> /dev/null

                echo "    All clear!"
                return;;
            no ) exit;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}
debugVars() {
    echo "KUBERNETES_NAME=$KUBERNETES_NAME"
    echo "KUBERNETES_DESCRIPTION=$KUBERNETES_DESCRIPTION"
    echo "RANCHER_MASTER_HOSTNAME=$RANCHER_MASTER_HOSTNAME"
    echo "KUBERNETES_NODE_HOSTNAME_BEGINSWITH=$KUBERNETES_NODE_HOSTNAME_BEGINSWITH"
    echo "KUBERNETES_NUMBER_OF_NODES=$KUBERNETES_NUMBER_OF_NODES"
    echo "RANCHER_MASTER_NETWORKS=$RANCHER_MASTER_NETWORKS"
    echo "KUBERNETES_NODE_NETWORKS=$KUBERNETES_NODE_NETWORKS"
    echo "HOST_PACKAGE=$HOST_PACKAGE"
}
getNetworkIDs() {
    values=$(echo $1 | tr "," " ")
    local networks
    for network in $values; do
        networks="$networks,$(triton networks -oname,id | sort | grep -v "^NAME *ID$" | sed -n "$network"p | awk 'NF>1{print $NF}')"
    done
    echo "$networks" | sed 's/^,\(.*\)$/\1/' | sed 's/\(.*\),$/\1/'
}
getPackageID() {
    echo "$(triton packages -oname,id | grep "\-kvm-" | grep -v "^NAME.*ID$" | tr -s " " | sort | sed -n "$1"p | awk 'NF>1{print $NF}')"
}
exportVars() {
    grep -v "^$" config > config.tmp
    while read line; do
        export "$line"
    done <config.tmp
    rm config.tmp
    export SEPARATE_PLANE
    export RANCHER_HA
}

main "$@"
