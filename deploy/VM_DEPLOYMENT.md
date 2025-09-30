# VM Deployment Guide

This guide explains how to deploy the Pizza Mixing application across separate VMs (Virtual Machines).

## Architecture Overview

The application is split into three independent services:

1. **Database VM**: PostgreSQL 16
2. **Backend VM**: Go API server
3. **Frontend VM**: Nginx web server

Each service runs on its own VM and communicates over the network.

## Prerequisites

- 3 Ubuntu 20.04 LTS or newer VMs
- Network connectivity between all VMs
- SSH access to all VMs
- Sudo privileges on all VMs

## VM Requirements

### Database VM
- **Minimum**: 2GB RAM, 2 vCPU, 20GB disk
- **Ports**: 5432 (PostgreSQL)

### Backend VM
- **Minimum**: 1GB RAM, 1 vCPU, 10GB disk
- **Ports**: 8080 (API)

### Frontend VM
- **Minimum**: 512MB RAM, 1 vCPU, 10GB disk
- **Ports**: 80 or 8000 (web)

## Step-by-Step Deployment

### 1. Database VM Setup

SSH into your database VM:

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install PostgreSQL 16
sudo apt-get install -y postgresql-common
sudo /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh -y
sudo apt-get install -y postgresql-16

# Configure PostgreSQL to listen on all interfaces
sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" /etc/postgresql/16/main/postgresql.conf

# Add client authentication rule
echo "host    all             all             0.0.0.0/0               md5" | sudo tee -a /etc/postgresql/16/main/pg_hba.conf

# Restart PostgreSQL
sudo systemctl restart postgresql

# Create database and user
sudo -u postgres psql -c "CREATE DATABASE pizza;"
sudo -u postgres psql -c "CREATE USER pizza WITH ENCRYPTED PASSWORD 'pizzapass';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE pizza TO pizza;"
sudo -u postgres psql -d pizza -c "GRANT ALL ON SCHEMA public TO pizza;"

# Allow firewall
sudo ufw allow 5432/tcp
```

**Security Note**: Replace 'pizzapass' with a strong password and restrict `pg_hba.conf` to specific IPs in production.

### 2. Backend VM Setup

SSH into your backend VM:

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Go 1.21+ (if not available, download from golang.org)
sudo apt-get install -y golang-go

# Clone the repository
cd /opt
sudo git clone https://github.com/your-username/pizza-mixing.git
cd pizza-mixing/app

# Create environment configuration
sudo tee /opt/pizza-mixing/.env <<EOF
DB_HOST=<DATABASE_VM_IP>
DB_PORT=5432
DB_USER=pizza
DB_PASSWORD=pizzapass
DB_NAME=pizza
DB_SSLMODE=require
EOF

# Build the application
sudo go build -o /opt/pizza-mixing/pizza-backend .

# Create systemd service
sudo tee /etc/systemd/system/pizza-backend.service > /dev/null <<EOF
[Unit]
Description=Pizza Mixing Backend API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/pizza-mixing
EnvironmentFile=/opt/pizza-mixing/.env
ExecStart=/opt/pizza-mixing/pizza-backend
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Set permissions
sudo chown -R www-data:www-data /opt/pizza-mixing

# Start and enable service
sudo systemctl daemon-reload
sudo systemctl enable pizza-backend
sudo systemctl start pizza-backend

# Check status
sudo systemctl status pizza-backend

# Allow firewall
sudo ufw allow 8080/tcp
```

### 3. Frontend VM Setup

SSH into your frontend VM:

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Nginx
sudo apt-get install -y nginx

# Clone repository (for static files)
cd /opt
sudo git clone https://github.com/your-username/pizza-mixing.git

# Create Nginx configuration
sudo tee /etc/nginx/sites-available/pizza-frontend > /dev/null <<'EOF'
server {
    listen 80;
    server_name _;

    root /opt/pizza-mixing/static;
    index index.html;

    # Serve static files
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API requests to backend VM
    location /api/ {
        proxy_pass http://<BACKEND_VM_IP>:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://<BACKEND_VM_IP>:8080/health;
    }
}
EOF

# Enable the site
sudo ln -sf /etc/nginx/sites-available/pizza-frontend /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
sudo systemctl enable nginx

# Allow firewall
sudo ufw allow 80/tcp
```

**Note**: Replace `<BACKEND_VM_IP>` with your backend VM's IP address in the Nginx config.

## Verification

### Test Database Connection

From backend VM:

```bash
psql -h <DATABASE_VM_IP> -U pizza -d pizza -c "SELECT 1;"
```

### Test Backend API

```bash
curl http://<BACKEND_VM_IP>:8080/health
```

Expected response:
```json
{"status":"healthy"}
```

### Test Frontend

Open browser to: `http://<FRONTEND_VM_IP>/`

## Ansible Automation

For automated deployment across multiple VMs, use the provided Ansible playbook:

```bash
# Update inventory with your VM IPs
cd ansible

# Edit inventory-vm.ini:
# [database]
# db-vm ansible_host=192.168.1.10
#
# [backend]
# api-vm ansible_host=192.168.1.11
#
# [frontend]
# web-vm ansible_host=192.168.1.12

# Run deployment
ansible-playbook -i inventory-vm.ini deploy-vm.yaml
```

## Monitoring

### Check Service Status

```bash
# Database VM
sudo systemctl status postgresql

# Backend VM
sudo systemctl status pizza-backend
sudo journalctl -u pizza-backend -f

# Frontend VM
sudo systemctl status nginx
sudo tail -f /var/log/nginx/access.log
```

## Troubleshooting

### Backend Can't Connect to Database

1. Check network connectivity:
   ```bash
   ping <DATABASE_VM_IP>
   telnet <DATABASE_VM_IP> 5432
   ```

2. Verify PostgreSQL is listening:
   ```bash
   # On database VM
   sudo netstat -tlnp | grep 5432
   ```

3. Check PostgreSQL logs:
   ```bash
   sudo tail -f /var/log/postgresql/postgresql-16-main.log
   ```

### Frontend Can't Reach Backend

1. Verify backend is running:
   ```bash
   curl http://<BACKEND_VM_IP>:8080/health
   ```

2. Check Nginx configuration:
   ```bash
   sudo nginx -t
   sudo systemctl status nginx
   ```

3. Check Nginx error logs:
   ```bash
   sudo tail -f /var/log/nginx/error.log
   ```

## Scaling

### Add Multiple Backend Instances

1. Deploy additional backend VMs following Step 2
2. Configure frontend Nginx with upstream load balancing:

```nginx
upstream backend_pool {
    server <BACKEND_VM_1_IP>:8080;
    server <BACKEND_VM_2_IP>:8080;
    server <BACKEND_VM_3_IP>:8080;
}

server {
    location /api/ {
        proxy_pass http://backend_pool/;
        # ... rest of proxy settings
    }
}
```

### Database High Availability

For production, consider:
- PostgreSQL replication (primary + replicas)
- Connection pooling (PgBouncer)
- Automated backups
- Monitoring (Prometheus + Grafana)

## Security Hardening

1. **Use SSL/TLS**: Configure PostgreSQL with SSL certificates
2. **Firewall Rules**: Restrict access to specific IPs only
3. **Strong Passwords**: Use generated passwords, not defaults
4. **Regular Updates**: Keep all systems patched
5. **Monitoring**: Set up alerts for failures
6. **Backups**: Automated PostgreSQL backups
