#!/bin/bash
# VCF Ubuntu VM Deployment Script for Pizza Mixing App
# This script installs Docker and runs the application on VMware VCF

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Pizza Mixing App Deployment on VCF Ubuntu VM...${NC}"

# Function to check if running on Ubuntu
check_ubuntu() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        if [ "$ID" != "ubuntu" ]; then
            echo -e "${RED}This script is designed for Ubuntu. Detected: $ID${NC}"
            exit 1
        fi
    else
        echo -e "${RED}Cannot detect OS. Ensure you're running Ubuntu.${NC}"
        exit 1
    fi
}

# Check Ubuntu
check_ubuntu

# Update system
echo -e "${YELLOW}Updating system packages...${NC}"
sudo apt-get update -y

# Install required packages
echo -e "${YELLOW}Installing required packages...${NC}"
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    git \
    net-tools

# Install Docker if not present
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}Installing Docker...${NC}"
    
    # Add Docker's official GPG key
    sudo mkdir -m 0755 -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    
    # Set up the repository
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # Install Docker Engine
    sudo apt-get update -y
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    # Add current user to docker group
    sudo usermod -aG docker $USER
    
    # Start and enable Docker
    sudo systemctl start docker
    sudo systemctl enable docker
    
    echo -e "${GREEN}Docker installed successfully${NC}"
else
    echo -e "${GREEN}Docker is already installed${NC}"
fi

# Create app directory
APP_DIR="/opt/pizza-mixing-app"
echo -e "${YELLOW}Creating application directory at ${APP_DIR}...${NC}"
sudo mkdir -p ${APP_DIR}
sudo chown $USER:$USER ${APP_DIR}

# Clone or pull the repository
REPO_URL="${GIT_REPO_URL:-https://github.com/your-username/pizza-mixing.git}"
if [ -d "${APP_DIR}/.git" ]; then
    echo -e "${YELLOW}Updating existing repository...${NC}"
    cd ${APP_DIR}
    git pull origin main
else
    echo -e "${YELLOW}Cloning repository...${NC}"
    git clone ${REPO_URL} ${APP_DIR}
    cd ${APP_DIR}
fi

# Configure firewall for VCF environment
echo -e "${YELLOW}Configuring firewall...${NC}"
if command -v ufw &> /dev/null; then
    sudo ufw allow 8000/tcp comment 'Pizza App Frontend' 2>/dev/null || true
    sudo ufw allow 22/tcp comment 'SSH' 2>/dev/null || true
    echo -e "${GREEN}Firewall rules added${NC}"
fi

# Stop existing containers if any
echo -e "${YELLOW}Stopping existing containers...${NC}"
cd ${APP_DIR}/docker
sudo docker compose down 2>/dev/null || true

# Build and start containers
echo -e "${YELLOW}Building and starting Docker containers...${NC}"
sudo docker compose up -d --build

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to start...${NC}"
sleep 10

# Get VM IP address
VM_IP=$(hostname -I | awk '{print $1}')
VM_HOSTNAME=$(hostname -f)

# Check if services are running
if sudo docker compose ps | grep -q "Up"; then
    echo -e "${GREEN}✓ Services are running!${NC}"
    
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Deployment completed successfully on VCF!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e "${YELLOW}Application is accessible at:${NC}"
    echo -e "  ${GREEN}http://${VM_IP}:8000${NC}"
    echo -e "  ${GREEN}http://${VM_HOSTNAME}:8000${NC}"
    echo ""
    echo -e "${YELLOW}To check logs:${NC}"
    echo -e "  cd ${APP_DIR}/docker && sudo docker compose logs -f"
    echo ""
    echo -e "${YELLOW}To stop the application:${NC}"
    echo -e "  cd ${APP_DIR}/docker && sudo docker compose down"
    echo ""
    echo -e "${YELLOW}To restart after VM reboot:${NC}"
    echo -e "  cd ${APP_DIR}/docker && sudo docker compose up -d"
else
    echo -e "${RED}✗ Failed to start services. Check logs:${NC}"
    echo -e "  cd ${APP_DIR}/docker && sudo docker compose logs"
    exit 1
fi