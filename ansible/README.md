# Ansible Deployment Playbooks

This directory contains Ansible playbooks for automating deployment of the Pizza Mixing application.

## Available Playbooks

### `deploy-vcf-docker.yaml` - Docker Compose on VMs (Recommended)

Deploys the complete three-tier application using Docker Compose on VMware VCF Ubuntu VMs.

**What it deploys:**
- PostgreSQL database container
- Go backend API container
- Nginx frontend container
- All on the same VM using Docker Compose

**Use case:** Quick deployment for development, testing, or single-VM production environments.

**Requirements:**
- Ubuntu 20.04+ VMs
- SSH access
- Internet connectivity

**Usage:**
```bash
# Edit inventory
cp inventory-vcf.ini.example inventory-vcf.ini
# Add your VM details

# Deploy
ansible-playbook -i inventory-vcf.ini deploy-vcf-docker.yaml
```

## Deprecated Playbooks

### `deploy.yaml.deprecated` - Legacy 2-tier deployment

**⚠️ DEPRECATED**: This playbook deploys the old two-tier architecture (backend with SQLite + frontend).

The application has been refactored to use PostgreSQL instead of SQLite. This playbook is kept for reference only.

**For new deployments**, use one of these options:
1. **Docker Compose**: Use `deploy-vcf-docker.yaml` for containerized deployment
2. **Separate VMs**: Follow the manual guide in `deploy/VM_DEPLOYMENT.md` for true three-tier deployment

## Creating Your Own Playbooks

If you need a custom Ansible playbook for the three-tier VM deployment (database, backend, frontend on separate VMs), refer to:
- `deploy/VM_DEPLOYMENT.md` for the manual steps
- Convert those steps to Ansible tasks

Example structure:
```yaml
# Playbook for 3-tier VM deployment
- hosts: database_vms
  tasks:
    - Install PostgreSQL
    - Configure database
    - Create pizza database

- hosts: backend_vms
  tasks:
    - Install Go
    - Deploy backend binary
    - Configure environment variables
    - Set up systemd service

- hosts: frontend_vms
  tasks:
    - Install Nginx
    - Deploy static files
    - Configure reverse proxy
```

## Inventory Files

### `inventory-vcf.ini`
Example inventory for VMware VCF deployments with Docker.

### `inventory.ini`
Legacy inventory file (deprecated).

## Templates Directory

Contains Jinja2 templates for configuration files used by the playbooks.

## Additional Resources

- **VM Deployment Guide**: `../deploy/VM_DEPLOYMENT.md`
- **Architecture Documentation**: `../ARCHITECTURE.md`
- **Main README**: `../README.md`
