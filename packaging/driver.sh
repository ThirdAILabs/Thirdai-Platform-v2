#!/bin/bash

function install_ansible() {
    if ! command -v ansible-playbook &> /dev/null
    then
        echo "Ansible not found, installing..."
        
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            if [ -f /etc/debian_version ]; then
                sudo apt update && sudo apt install -y ansible
            elif [ -f /etc/redhat-release ]; then
                sudo yum install -y ansible
            fi
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            brew install ansible
        fi
    else
        echo "Ansible is already installed"
    fi
}

install_ansible

if [ -z "$1" ]; then
    echo "Usage: $0 /path/to/your/config.yml"
    exit 1
fi

CONFIG_PATH="$1"

if [ ! -f "$CONFIG_PATH" ]; then
    echo "Config file not found at $CONFIG_PATH"
    exit 1
fi

cd "$(dirname "$0")/platform" || exit 1

ansible-playbook ../playbooks/test_deploy.yml --extra-vars "config_path=$CONFIG_PATH"
