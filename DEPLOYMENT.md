# Pizza Mixing App Deployment Guide

This guide provides instructions for deploying the Pizza Mixing application to AWS EC2 instances and VMware Cloud Foundation (VCF) Ubuntu VMs.

## Prerequisites

- Target system must be Ubuntu 20.04 LTS or newer
- SSH access to the target server
- Git repository with the application code
- Minimum 2GB RAM and 10GB disk space
- Internet connectivity for downloading Docker and dependencies

## Application Architecture

The application consists of:
- **Backend**: Go API server running on port 8080
- **Frontend**: Nginx serving static files on port 80
- **Docker Compose**: Orchestrates both services, exposed on port 8000

## Deployment Methods

### Method 1: AWS EC2 Deployment

#### Quick Deploy

1. Launch an Ubuntu EC2 instance (t2.micro or larger)
2. Configure Security Group to allow:
   - Port 22 (SSH)
   - Port 8000 (Application)
3. SSH into the instance and run:

```bash
# Download and run the deployment script
curl -O https://raw.githubusercontent.com/your-repo/pizza-mixing/main/deploy/deploy-ec2.sh
chmod +x deploy-ec2.sh

# Set your repository URL (optional, update the script default)
export GIT_REPO_URL="https://github.com/your-username/pizza-mixing.git"

# Run deployment
./deploy-ec2.sh
```

#### What the EC2 Script Does

1. Updates system packages
2. Installs Docker and Docker Compose
3. Clones your repository
4. Builds and starts the application containers
5. Displays the public IP address for access

### Method 2: VMware VCF Ubuntu VM Deployment

#### Option A: Shell Script Deployment

1. Create an Ubuntu VM in your VCF environment
2. SSH into the VM and run:

```bash
# Download and run the deployment script
curl -O https://raw.githubusercontent.com/sakuffo/pizza-recommender/refs/heads/master/deploy/deploy-vcf.sh
chmod +x deploy-vcf.sh

# Set your repository URL (optional)
export GIT_REPO_URL="https://github.com/your-username/pizza-mixing.git"

# Run deployment
./deploy-vcf.sh
```

#### Option B: Ansible Automation (Recommended for Multiple VMs)

1. On your Ansible control node, update the inventory file:

```bash
# Edit ansible/inventory-vcf.ini
# Add your VCF VM details:
[vcf_vms]
vcf-vm-01 ansible_host=192.168.1.100 ansible_user=ubuntu
vcf-vm-02 ansible_host=192.168.1.101 ansible_user=ubuntu
```

2. Run the Ansible playbook:

```bash
# Deploy to all VCF VMs
ansible-playbook -i ansible/inventory-vcf.ini ansible/deploy-vcf-docker.yaml

# Deploy with custom repo
ansible-playbook -i ansible/inventory-vcf.ini ansible/deploy-vcf-docker.yaml \
  -e "git_repo_url=https://github.com/your-username/pizza-mixing.git"
```

## Post-Deployment

### Accessing the Application

After successful deployment, access the application at:
- EC2: `http://<ec2-public-ip>:8000`
- VCF: `http://<vm-hostname>:8000` or `http://<vm-ip>:8000`

### Useful Commands

```bash
# Check application status
cd /opt/pizza-mixing-app/docker
sudo docker compose ps

# View application logs
sudo docker compose logs -f

# Stop the application
sudo docker compose down

# Start the application
sudo docker compose up -d

# Rebuild and restart (after code changes)
sudo docker compose up -d --build

# View backend logs only
sudo docker compose logs -f backend

# View frontend logs only
sudo docker compose logs -f frontend
```

### Troubleshooting

#### Application not accessible

1. Check if containers are running:
```bash
sudo docker compose ps
```

2. Verify port 8000 is listening:
```bash
sudo netstat -tlnp | grep 8000
```

3. Check firewall rules:
```bash
# For EC2, check Security Group
# For VCF Ubuntu:
sudo ufw status
```

4. Review logs for errors:
```bash
sudo docker compose logs backend
sudo docker compose logs frontend
```

#### Docker installation issues

```bash
# Remove old Docker versions
sudo apt-get remove docker docker-engine docker.io containerd runc

# Clean package cache
sudo apt-get clean
sudo apt-get update

# Retry installation
sudo apt-get install docker-ce docker-ce-cli containerd.io
```

#### Permission denied errors

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Apply group changes (logout/login or run)
newgrp docker
```

### Auto-Start on Reboot

The VCF Ansible playbook creates a systemd service for auto-start. For manual setup:

```bash
# Create service file
sudo tee /etc/systemd/system/pizza-app.service > /dev/null <<EOF
[Unit]
Description=Pizza Mixing App
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/pizza-mixing-app/docker
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# Enable service
sudo systemctl enable pizza-app
sudo systemctl daemon-reload
```

## Security Considerations

For production deployments:

1. **Use HTTPS**: Add SSL certificates and configure Nginx for HTTPS
2. **Restrict access**: Configure firewall rules to limit access
3. **Use secrets management**: Don't hardcode credentials
4. **Regular updates**: Keep Docker and system packages updated
5. **Monitor logs**: Set up log aggregation and monitoring

## Customization

### Change Port

Edit `docker/compose.yaml`:
```yaml
frontend:
  ports:
    - "YOUR_PORT:80"  # Change 8000 to your desired port
```

### Use Custom Domain

1. Point your domain to the server IP
2. Update Nginx configuration if needed
3. Consider adding SSL with Let's Encrypt

### Scale Backend

For multiple backend instances, modify `docker/compose.yaml`:
```yaml
backend:
  scale: 3  # Run 3 backend instances
```

## Support

For issues or questions:
1. Check application logs
2. Review this documentation
3. Check Docker and system logs
4. Ensure all prerequisites are met