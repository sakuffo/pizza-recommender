#!/bin/bash
# EC2 Deployment Script for Pizza Mixing App
# This script installs Docker, clones the repo, and runs the application

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Pizza Mixing App Deployment on EC2...${NC}"

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
    git

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

# Check if services are running
if sudo docker compose ps | grep -q "Up"; then
    echo -e "${GREEN}✓ Services are running!${NC}"
    
    # Get the public IP
    PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null || echo "localhost")
    
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Deployment completed successfully!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e "${YELLOW}Application is accessible at:${NC}"
    echo -e "  ${GREEN}http://${PUBLIC_IP}:8000${NC}"
    echo ""
    echo -e "${YELLOW}To check logs:${NC}"
    echo -e "  cd ${APP_DIR}/docker && sudo docker compose logs -f"
    echo ""
    echo -e "${YELLOW}To stop the application:${NC}"
    echo -e "  cd ${APP_DIR}/docker && sudo docker compose down"
else
    echo -e "${RED}✗ Failed to start services. Check logs:${NC}"
    echo -e "  cd ${APP_DIR}/docker && sudo docker compose logs"
    exit 1
fi