#!/bin/bash

# Add Hosts to Kubernetes Environment
# etcd
if [ ${etcd_node_count} -gt 0 ]; then
    for x in $(seq 1 ${etcd_node_count}); do
    	curl -X POST \
    		-H 'Accept: application/json' \
    		-H 'Content-Type: application/json' \
    		-d '{"hostname":"${name}-etcd-'$x'", "labels": {"etcd": "true"}, "engineInstallUrl":"${docker_engine_install_url}", "tritonConfig":{"account": "${triton_account}", "image": "${triton_image_name}@${triton_image_version}", "keyId": "${triton_key_id}", "keyPath": "${triton_key_path}", "package": "${etcd_triton_machine_package}", "sshUser": "${triton_ssh_user}", "url": "${triton_url}"}}' \
    		'${rancher_host}/v2-beta/projects/'${environment_id}'/hosts'
    
    	sleep 1;
    done
fi

# orchestration
if [ ${orchestration_node_count} -gt 0 ]; then
    for x in $(seq 1 ${orchestration_node_count}); do
    	curl -X POST \
    		-H 'Accept: application/json' \
    		-H 'Content-Type: application/json' \
    		-d '{"hostname":"${name}-orchestration-'$x'", "labels": {"orchestration": "true"}, "engineInstallUrl":"${docker_engine_install_url}", "tritonConfig":{"account": "${triton_account}", "image": "${triton_image_name}@${triton_image_version}", "keyId": "${triton_key_id}", "keyPath": "${triton_key_path}", "package": "${orchestration_triton_machine_package}", "sshUser": "${triton_ssh_user}", "url": "${triton_url}"}}' \
    		'${rancher_host}/v2-beta/projects/'${environment_id}'/hosts';
    
    	sleep 1;
    done
fi

# compute
for x in $(seq 1 ${compute_node_count}); do
	curl -X POST \
		-H 'Accept: application/json' \
		-H 'Content-Type: application/json' \
		-d '{"hostname":"${name}-compute-'$x'", "labels": {"compute": "true"}, "engineInstallUrl":"${docker_engine_install_url}", "tritonConfig":{"account": "${triton_account}", "image": "${triton_image_name}@${triton_image_version}", "keyId": "${triton_key_id}", "keyPath": "${triton_key_path}", "package": "${compute_triton_machine_package}", "sshUser": "${triton_ssh_user}", "url": "${triton_url}"}}' \
		'${rancher_host}/v2-beta/projects/'${environment_id}'/hosts';

	sleep 1;
done
