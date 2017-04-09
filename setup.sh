#!/bin/bash

main() {
    if [[ ! -z "$1" && $1 == "-c" ]]; then
        cleanRunner
    fi

    if [ -e terraform/rancher.tf ]; then
        echo "error: configuration for a previous run has been found"
        echo "    clean the configuration (./setup -c)"
        exit
    fi

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
    echo "Congradulations, your Kubernetes cluster setup has been complete."
    echo "----> Rancher dashboard is at http://$(cat terraform/masters.ip):8080"
    echo ""

    echo "It will take a few minutes for all the Kubernetes process to start up before you can access Kubernetes Dashboard"
    echo "----> To check what processes/containers are coming up, go to http://$(cat terraform/masters.ip):8080/env/$(cat ansible/tmp/kubernetes_environment.id)/infra/containers"
    echo "    once all these containers are up, you should be able to access Kubernetes by its dashboard or using CLI"

    echo "Waiting on Kubernetes dashboard to come up."
    echo ""

    KUBERNETES_DASHBOARD_UP=
    while [ ! $KUBERNETES_DASHBOARD_UP ]; do
        echo -ne "."
        sleep 2
        if [ $(curl -s http://$(cat terraform/masters.ip):8080/r/projects/$(cat ansible/tmp/kubernetes_environment.id)/kubernetes-dashboard:9090/ | grep -i kubernetes | wc -l) -ne 0 ]; then
            KUBERNETES_DASHBOARD_UP=true
        fi
    done
    echo ""
    echo "----> Kubernetes dashboard is at http://$(cat terraform/masters.ip):8080/r/projects/$(cat ansible/tmp/kubernetes_environment.id)/kubernetes-dashboard:9090/"
    echo "----> Kubernetes CLI config is at http://$(cat terraform/masters.ip):8080/env/$(cat ansible/tmp/kubernetes_environment.id)/kubernetes/kubectl"
    echo ""
    echo "    CONGRATULATIONS, YOU HAVE CONFIGURED YOUR KUBERNETES ENVIRONMENT!"
}

function getArgument {
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
function runAnsible {
    cd ansible
    ansible-playbook -i hosts clusterUp.yml
    cd ..
}
function createAnsibleConfigs {
    echo "Creating ansible hosts file and variable files"
    rm ansible/hosts 2> /dev/null
    echo "[MASTER]" >> ansible/hosts
    cat terraform/masters.ip >> ansible/hosts
    echo "[HOST]" >> ansible/hosts
    cat terraform/hosts.ip >> ansible/hosts
    echo "    created: ansible/hosts"
    master=$(tail -1 terraform/masters.ip)
    echo "master: $master" > ansible/roles/ranchermaster/vars/vars.yml
    echo "kubernetes_name: \"$(echo $KUBERNETES_NAME | sed 's/"//g')\"" >> ansible/roles/ranchermaster/vars/vars.yml
    echo "kubernetes_description: \"$(echo $KUBERNETES_DESCRIPTION | sed 's/"//g')\"" >> ansible/roles/ranchermaster/vars/vars.yml
    cd ansible
    sed -i.tmp -e "s;private_key_file = .*$;private_key_file = $(echo $SDC_KEY | sed 's/"//g');g" ansible.cfg
    cd ..
    rm ansible/ansible.cfg.tmp 2>&1 >> /dev/null

    echo "    created: ansible/roles/ranchermaster/vars/vars.yml"
}
function runTerraformTasks {
    if [ -e terraform/rancher.tf ]
    then
        echo "warning: a previous terraform configuration has been found"
        echo "    skipping terraform configuration and execution..."
    else
        echo "Generating terraform configs for environment..."
        updateTerraformConfig provider triton
        echo "    Master hostname: $RANCHER_MASTER_HOSTNAME"
        updateTerraformConfig master $(echo $RANCHER_MASTER_HOSTNAME | sed 's/"//g')
        for (( i = 1; i <= $KUBERNETES_NUMBER_OF_NODES; i++ ))
        do
            echo "    Kubernetes node $i: $(echo $KUBERNETES_NODE_HOSTNAME_BEGINSWITH | sed 's/"//g')$i"
            updateTerraformConfig host $(echo $KUBERNETES_NODE_HOSTNAME_BEGINSWITH$i | sed 's/"//g')
        done

        cd terraform
        echo "Starting terraform tasks"
        terraform get
        terraform apply
        echo "    terraform tasks completed"
        cd ..
    fi
}
function updateTerraformConfig {
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
function setConfigToFile {
    sed -i.tmp -e "s~RANCHER_MASTER_NETWORKS=.*$~RANCHER_MASTER_NETWORKS=$RANCHER_MASTER_NETWORKS~g" config
    sed -i.tmp -e "s~KUBERNETES_NODE_NETWORKS=.*$~KUBERNETES_NODE_NETWORKS=$KUBERNETES_NODE_NETWORKS~g" config
    sed -i.tmp -e "s~KUBERNETES_NUMBER_OF_NODES=.*$~KUBERNETES_NUMBER_OF_NODES=$KUBERNETES_NUMBER_OF_NODES~g" config
    sed -i.tmp -e "s~KUBERNETES_NAME=.*$~KUBERNETES_NAME=\"$KUBERNETES_NAME\"~g" config
    sed -i.tmp -e "s~KUBERNETES_DESCRIPTION=.*$~KUBERNETES_DESCRIPTION=\"$KUBERNETES_DESCRIPTION\"~g" config
    sed -i.tmp -e "s~RANCHER_MASTER_HOSTNAME=.*$~RANCHER_MASTER_HOSTNAME=\"$RANCHER_MASTER_HOSTNAME\"~g" config
    sed -i.tmp -e "s~KUBERNETES_NODE_HOSTNAME_BEGINSWITH=.*$~KUBERNETES_NODE_HOSTNAME_BEGINSWITH=\"$KUBERNETES_NODE_HOSTNAME_BEGINSWITH\"~g" config
    sed -i.tmp -e "s~HOST_PACKAGE=.*$~HOST_PACKAGE=\"$HOST_PACKAGE\"~g" config
    rm config.tmp 2>&1 >> /dev/null
}
function setConfigFromTritonENV {
    eval "$(triton env)"
    sed -i.tmp -e "s~SDC_URL=.*$~SDC_URL=\"$SDC_URL\"~g" config
    sed -i.tmp -e "s~SDC_ACCOUNT=.*$~SDC_ACCOUNT=\"$SDC_ACCOUNT\"~g" config
    sed -i.tmp -e "s~SDC_KEY_ID=.*$~SDC_KEY_ID=\"$SDC_KEY_ID\"~g" config

    local foundKey=false
    for f in $(ls ~/.ssh); do
        if [[ "$(ssh-keygen -E md5 -lf ~/.ssh/$(echo $f | sed 's/.pub$//') 2> /dev/null | awk '{print $2}' | sed 's/^MD5://')" == "$SDC_KEY_ID" ]]; then
            sed -i.tmp -e "s~SDC_KEY=.*$~SDC_KEY=\"\~/.ssh/$(echo $f | sed 's/.pub$//')\"~g" config
            foundKey=true
        fi
    done
    rm config.tmp 2>&1 >> /dev/null
    if [[ ! foundKey ]]; then
        echo "error: couldn't find the ssh key associated with fingerprint $SDC_KEY_ID in ~/.ssh/ directory..."
        exit 1
    fi
}
function setVarDefaults {
    if [ ! $(grep KUBERNETES_NAME=$ config) ]; then
        echo "warn: old configuration found"
        echo "    Make sure you are not using the defaults if you already have the old environment up and running"
        KUBERNETES_NAME=$(grep KUBERNETES_NAME= config | sed "s/KUBERNETES_NAME=//g")
        KUBERNETES_DESCRIPTION=$(grep KUBERNETES_DESCRIPTION= config | sed "s/KUBERNETES_DESCRIPTION=//g")
        RANCHER_MASTER_HOSTNAME=$(grep RANCHER_MASTER_HOSTNAME= config | sed "s/RANCHER_MASTER_HOSTNAME=//g")
        KUBERNETES_NODE_HOSTNAME_BEGINSWITH=$(grep KUBERNETES_NODE_HOSTNAME_BEGINSWITH= config | sed "s/KUBERNETES_NODE_HOSTNAME_BEGINSWITH=//g")
        KUBERNETES_NUMBER_OF_NODES=$(grep KUBERNETES_NUMBER_OF_NODES= config | sed "s/KUBERNETES_NUMBER_OF_NODES=//g")
        RANCHER_MASTER_NETWORKS=$(grep RANCHER_MASTER_NETWORKS= config | sed "s/RANCHER_MASTER_NETWORKS=//g")
        KUBERNETES_NODE_NETWORKS=$(grep KUBERNETES_NODE_NETWORKS= config | sed "s/KUBERNETES_NODE_NETWORKS=//g")
        HOST_PACKAGE=$(grep HOST_PACKAGE= config | sed "s/HOST_PACKAGE=//g")
    else
        KUBERNETES_NAME="k8s dev"
        KUBERNETES_DESCRIPTION=$KUBERNETES_NAME
        RANCHER_MASTER_HOSTNAME="kubemaster"
        KUBERNETES_NODE_HOSTNAME_BEGINSWITH="kubenode"
        KUBERNETES_NUMBER_OF_NODES=2
        RANCHER_MASTER_NETWORKS=
        KUBERNETES_NODE_NETWORKS=
        HOST_PACKAGE=
    fi
}
function getConfigFromUser {
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
            echo "error: Enter a valid value (number between 1-9) or leave blank to use the default."
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
function verifyConfig {
    echo "################################################################################"
    echo "Verify that the following configuration is correct:"
    echo ""
    echo "Name of kubernetes environment: $KUBERNETES_NAME"
    echo "Kubernetes environment description: $KUBERNETES_DESCRIPTION"
    echo "Master hostname: $RANCHER_MASTER_HOSTNAME"
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
            return 1
            ;;
        no )
            exit 0
            ;;
        * ) echo "Please answer yes or no.";;
    esac
    done
}
function cleanRunner {
    echo "Clearing settings...."
    while true; do
        if [ -e terraform/masters.ip ]; then
            echo "WARNING: You are about to destroy the following KVMs associated with Rancher cluster:"
            cat terraform/masters.ip 2> /dev/null
            cat terraform/hosts.ip 2> /dev/null

            read -p "Do you wish to destroy the KVMs and reset configuration (yes | no)? " yn
        else
            read -p "Do you wish to reset configuration (yes | no)? " yn
        fi
    case $yn in
        yes )
            if [ -e terraform/rancher.tf ]; then
                cd terraform
                echo "    destroying images..."
                terraform destroy -force 2> /dev/null
                cd ..
            fi
            rm -rf terraform/hosts.ip terraform/masters.ip terraform/terraform.* terraform/.terraform* terraform/rancher.tf 2>&1 >> /dev/null

            sed -i.tmp -e "s~private_key_file = .*$~private_key_file = ~g" ansible/ansible.cfg
            rm -f  ansible/hosts ansible/*retry ansible/ansible.cfg.tmp 2>&1 >> /dev/null

            rm -rf ~/.ssh/known_hosts tmp/* containers.json 2>&1 >> /dev/null

            # blank out config
            echo "SDC_URL=" > config
            echo "SDC_ACCOUNT=" >> config
            echo "SDC_KEY_ID=" >> config
            echo "SDC_KEY=" >> config
            echo "" >> config
            echo "RANCHER_MASTER_HOSTNAME=" >> config
            echo "RANCHER_MASTER_NETWORKS=" >> config
            echo "KUBERNETES_NAME=" >> config
            echo "KUBERNETES_DESCRIPTION=" >> config
            echo "KUBERNETES_NODE_HOSTNAME_BEGINSWITH=" >> config
            echo "KUBERNETES_NUMBER_OF_NODES=" >> config
            echo "KUBERNETES_NODE_NETWORKS=" >> config
            echo "HOST_PACKAGE=" >> config
            echo "" >> config
            echo "ANSIBLE_HOST_KEY_CHECKING=False" >> config
            echo "    All clear!"
            exit 1;;
        no ) exit;;
        * ) echo "Please answer yes or no.";;
    esac
    done
}
function debugVars {
    echo "KUBERNETES_NAME=$KUBERNETES_NAME"
    echo "KUBERNETES_DESCRIPTION=$KUBERNETES_DESCRIPTION"
    echo "RANCHER_MASTER_HOSTNAME=$RANCHER_MASTER_HOSTNAME"
    echo "KUBERNETES_NODE_HOSTNAME_BEGINSWITH=$KUBERNETES_NODE_HOSTNAME_BEGINSWITH"
    echo "KUBERNETES_NUMBER_OF_NODES=$KUBERNETES_NUMBER_OF_NODES"
    echo "RANCHER_MASTER_NETWORKS=$RANCHER_MASTER_NETWORKS"
    echo "KUBERNETES_NODE_NETWORKS=$KUBERNETES_NODE_NETWORKS"
    echo "HOST_PACKAGE=$HOST_PACKAGE"
}
function getNetworkIDs {
    values=$(echo $1 | tr "," " ")
    local networks
    for network in $values; do
        networks="$networks,$(triton networks -oname,id | sort | grep -v "^NAME *ID$" | sed -n "$network"p | awk 'NF>1{print $NF}')"
    done
    echo "$networks" | sed 's/^,\(.*\)$/\1/' | sed 's/\(.*\),$/\1/'
}
function getPackageID {
    echo "$(triton packages -oname,id | grep "\-kvm-" | grep -v "^NAME.*ID$" | tr -s " " | sort | sed -n "$1"p | awk 'NF>1{print $NF}')"
}
function exportVars {
    grep -v "^$" config > config.tmp
    while read line; do
        export "$line"
    done <config.tmp
    rm config.tmp
}

main "$@"
