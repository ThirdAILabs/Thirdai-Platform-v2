#!/bin/bash

function install_ansible() {
    # Check if ansible-playbook is installed
    if ! command -v ansible-playbook &> /dev/null; then
        echo "Ansible not found, installing..."

        # Install Ansible based on the OS
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            if [ -f /etc/debian_version ]; then
                echo "Installing Ansible on Ubuntu/Debian..."
                sudo apt update
                sudo apt install -y software-properties-common
                sudo add-apt-repository --yes --update ppa:ansible/ansible
                sudo apt install -y ansible
            elif [ -f /etc/system-release ] && grep -q "Amazon Linux release 2023" /etc/system-release; then
                echo "Installing Ansible on CentOS/RHEL..."
                sudo dnf install -y ansible
            elif [ -f /etc/system-release ] && grep -q "Amazon Linux release 2" /etc/system-release; then
                echo "Installing Ansible on Amazon Linux 2..."
                sudo amazon-linux-extras install -y ansible2
            else
                echo "Unsupported OS: $OSTYPE"
                echo "Please download and install Ansible manually for your operating system."
                exit 1
            fi
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            echo "Installing Ansible on macOS..."
            brew install ansible
        else
            echo "Unsupported OS: $OSTYPE"
            echo "Please download and install Ansible manually for your operating system."
            exit 1
        fi
    else
        echo "Ansible is already installed"
    fi

    echo "Installing required Ansible Galaxy collections..."
    ansible-galaxy collection install community.general
    ansible-galaxy collection install ansible.posix
    # From v4.0.0 commuinty docker support is not there for amazon linux 2
    if [ -f /etc/system-release ] && grep -q "Amazon Linux release 2" /etc/system-release; then
        ansible-galaxy collection install community.docker:==3.13.1 --force
    else
        ansible-galaxy collection install community.docker
    fi
    ansible-galaxy collection install community.postgresql

    echo "All required Ansible Galaxy collections are installed."
}

install_ansible

VERBOSE=0  # Default: No verbose mode
CLEANUP=0  # Flag for cleanup mode
ONBOARD_CLIENTS=0  # Flag for onboard_clients mode
NEW_CLIENT_CONFIG_PATH=""   # Declare globally, default empty
    
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -v|--verbose) VERBOSE=1 ;;   # Enable verbose mode if -v or --verbose is passed
        --cleanup) CLEANUP=1 ;;  # Enable cleanup mode if --cleanup is passed
        --onboard_clients) ONBOARD_CLIENTS=1 ;;  # Enable onboard_clients if --onboard_clients is passed
        *) CONFIG_PATH=$(realpath "$1") ;;  # Treat the first argument as the config path
    esac
    shift
done

# Check if config file is provided
if [ -z "$CONFIG_PATH" ]; then
    echo "Usage: $0 [-v|--verbose] [-b|--branch <branch_name>] [--cleanup] /path/to/your/config.yml"
    exit 1
fi

# Check if config file exists
if [ ! -f "$CONFIG_PATH" ]; then
    echo "Config file not found at $CONFIG_PATH"
    exit 1
fi
echo "Using config file at $CONFIG_PATH"

if [ "$ONBOARD_CLIENTS" -eq 1 ]; then
    read -p "Enter the absolute path to the new client config file (e.g., /path/to/new_client_config.yml): " NEW_CLIENT_CONFIG_PATH

    # Check if a valid path was provided
    if [ -z "$NEW_CLIENT_CONFIG_PATH" ]; then
        echo "Error: No file location provided for new client configuration."
        exit 1
    fi

    # Check if config file exists
    if [ ! -f "$NEW_CLIENT_CONFIG_PATH" ]; then
        echo "New client config file not found at $NEW_CLIENT_CONFIG_PATH"
        exit 1
    fi

    echo "Using new client config file at $NEW_CLIENT_CONFIG_PATH"
fi

# Model path
MODEL_FOLDER="pretrained-models/"

# Warn if model file is not found
if [ ! -d "$MODEL_FOLDER" ]; then
    echo "WARNING: Model file not found at $GENERATIVE_MODEL_FOLDER. The playbook will proceed without it."
fi

MODEL_FOLDER=$(realpath "$MODEL_FOLDER")

# Search for docker_images folder with a prefix
DOCKER_IMAGES_PATH=$(find . -type d -name "docker_images-*" | head -n 1)

if [ -z "$DOCKER_IMAGES_PATH" ]; then
    echo "WARNING: No docker_images folder found with prefix 'docker_images-'. The playbook will proceed without it."
else
    DOCKER_IMAGES_PATH=$(realpath "$DOCKER_IMAGES_PATH")
    echo "Found docker images folder at $DOCKER_IMAGES_PATH"
fi

# Change directory to platform directory
cd "$(dirname "$0")/platform" || exit 1

# Run the appropriate playbook based on the cleanup flag
if [ "$CLEANUP" -eq 1 ]; then
    echo "Running cleanup playbook..."
    if [ "$VERBOSE" -eq 1 ]; then
        ansible-playbook playbooks/test_cleanup.yml --extra-vars "config_path=$CONFIG_PATH model_folder=$MODEL_FOLDER  docker_images=$DOCKER_IMAGES_PATH" -vvvv
    else
        ansible-playbook playbooks/test_cleanup.yml --extra-vars "config_path=$CONFIG_PATH model_folder=$MODEL_FOLDER  docker_images=$DOCKER_IMAGES_PATH"
    fi
elif [ "$ONBOARD_CLIENTS" -eq 1 ]; then
    echo "Running onboarding playbook..."
    if [ "$VERBOSE" -eq 1 ]; then
        ansible-playbook playbooks/onboard_clients.yml --extra-vars "config_path=$CONFIG_PATH new_client_config_path=$NEW_CLIENT_CONFIG_PATH model_folder=$MODEL_FOLDER  docker_images=$DOCKER_IMAGES_PATH" -vvvv
    else
        ansible-playbook playbooks/onboard_clients.yml --extra-vars "config_path=$CONFIG_PATH new_client_config_path=$NEW_CLIENT_CONFIG_PATH model_folder=$MODEL_FOLDER  docker_images=$DOCKER_IMAGES_PATH"
    fi
else
    echo "Running deployment playbook..."
    if [ "$VERBOSE" -eq 1 ]; then
        ansible-playbook playbooks/test_deploy.yml --extra-vars "config_path=$CONFIG_PATH model_folder=$MODEL_FOLDER  docker_images=$DOCKER_IMAGES_PATH" -vvvv
    else
        ansible-playbook playbooks/test_deploy.yml --extra-vars "config_path=$CONFIG_PATH model_folder=$MODEL_FOLDER  docker_images=$DOCKER_IMAGES_PATH"
    fi
fi
