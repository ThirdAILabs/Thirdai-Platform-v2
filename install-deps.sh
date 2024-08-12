# Install miniconda
wget https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh
bash Miniconda3-latest-Linux-x86_64.sh

# Remove miniconda installer
rm Miniconda3-latest-Linux-x86_64.sh

# Install Nomad
sudo apt-get update && \
sudo apt-get install wget gpg coreutils
wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt-get update && sudo apt-get install nomad

# Install frontend frameworks
sudo apt update
sudo apt install nodejs
sudo apt install npm
sudo npm install -g pnpm

# Install postgres
sudo apt install postgresql
sudo -i -u postgres
psql -U postgres -d postgres

# Then run these inside postgres:
ALTER ROLE postgres WITH PASSWORD 'password';
create database model_bazaar;
\c model_bazaar;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

# Install Traefik
wget https://github.com/traefik/traefik/releases/download/v2.10.5/traefik_v2.10.5_linux_amd64.tar.gz
sudo tar -xf traefik_v2.10.5_linux_amd64.tar.gz -C /usr/local/bin

# Install uvicorn
sudo apt install uvicorn