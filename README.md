## Code Structure

```
.
├── README.md              # Project documentation
├── ARCHITECTURE.md        # Architecture documentation
├── DEPLOYMENT.md          # Deployment guide
├── MIGRATION_SUMMARY.md   # Migration guide from SQLite to PostgreSQL
├── ansible/               # Ansible deployment automation
│   ├── README.md          # Ansible documentation
│   ├── deploy-vcf-docker.yaml  # Docker deployment playbook
│   ├── inventory-vcf.ini  # Inventory example
│   └── *.deprecated       # Legacy files (old 2-tier architecture)
├── deploy/                # Deployment scripts and configs
│   ├── VM_DEPLOYMENT.md   # VM deployment guide (3-tier)
│   ├── deploy-ec2.sh      # EC2 deployment script
│   ├── deploy-vcf.sh      # VCF deployment script
│   ├── .env.docker        # Docker environment config
│   └── .env.vm.example    # VM environment template
├── docker/                # Docker deployment files
│   ├── Dockerfile         # Backend container (Go + PostgreSQL driver)
│   ├── Dockerfile.frontend # Frontend container (Nginx)
│   ├── compose.yaml       # Docker Compose (3 services)
│   └── nginx-docker.conf  # Nginx config for Docker
└── app/                   # Application source code
    ├── go.mod             # Go dependencies (PostgreSQL driver)
    ├── go.sum             # Dependency checksums
    ├── pizza-mix.go       # Main application
    ├── db.go              # Database connection & migrations
    ├── repository.go      # Data access layer
    └── static/            # Frontend static assets
        ├── index.html
        ├── css/
        │   └── style.css
        └── js/
            └── script.js
```

## Running with Docker (Three-Tier Architecture)

Files for this setup are located in the `docker/` directory.

The application uses a three-tier architecture with separate services:
- **Database**: PostgreSQL 16 in a container with persistent volume
- **Backend**: Go API server in a container
- **Frontend**: Nginx serving static files in a container

**Requirements:**
* Docker and Docker Compose installed

**Steps:**
1. **Build and start the application:**
    *   From the project root directory, run:
        ```bash
        docker compose -f docker/compose.yaml up --build -d 
        ```
    *   This command uses the configuration in `docker/compose.yaml`, builds both images using the Dockerfiles in `docker/`, and starts the services.

2. **Access the application:**
    - Open your browser and go to `http://localhost:8000`.

**Details:**
- The Docker Compose setup defines three services in `docker/compose.yaml`:
  - `database`: PostgreSQL with data persistence
  - `backend`: Go API connected to database
  - `frontend`: Nginx proxying API requests to backend
- Database data is stored in a Docker volume `pizza-db-data` for persistence
- All services communicate over a dedicated Docker network (`pizza-net`)
- The `backend` waits for database health check before starting
- The `frontend` depends on the `backend` service
- Only the frontend port is exposed: port `80` (container) → `8000` (host)
- Backend and database ports remain internal to the Docker network
- The frontend JavaScript (`static/js/script.js`) is configured to make API calls to `http://backend:8080`, using Docker's internal DNS to resolve the `backend` service name.
- Both containers run as non-root users where applicable.

**Stopping the application:**
*   From the project root directory, run:
    ```bash
    docker compose -f docker/compose.yaml down
    ```

## Deployment Options

### Option 1: Docker Compose (Recommended for Development/Single Host)
Use the Docker setup above for easy deployment on a single machine with all services containerized.

### Option 2: Separate VMs (Recommended for Production)
Deploy the three tiers (database, backend, frontend) across separate VMs for better isolation, scalability, and resilience.

**See detailed guide**: [`deploy/VM_DEPLOYMENT.md`](deploy/VM_DEPLOYMENT.md)

**Quick overview**: The VM deployment requires:
- PostgreSQL VM: Database server
- Backend VM: Go application with environment variables pointing to database
- Frontend VM: Nginx serving static files and proxying API requests

## Running Manually on Separate VMs (Legacy Two-Tier)

**Note**: This is the legacy two-tier setup. For the new three-tier architecture with PostgreSQL, see [`deploy/VM_DEPLOYMENT.md`](deploy/VM_DEPLOYMENT.md).

This section describes the old manual deployment onto two separate Virtual Machines. One VM hosts the Go backend API with embedded SQLite, and the other hosts the static frontend files served by Nginx.

**Prerequisites:**
*   Two VMs created (e.g., in vSphere), running a Linux distribution (examples use Ubuntu/Debian commands).
*   Network connectivity between the VMs.
*   Known IP addresses for both VMs (referred to as `<backend_vm_ip>` and `<frontend_vm_ip>`).
*   SSH access or console access to both VMs.
*   Firewalls configured on both VMs to allow necessary traffic (see steps below).

**1. Backend VM Setup**

*   **Goal:** Build and run the Go API service.

    1.  **Install Go & Git:**
        ```bash
        sudo apt update
        sudo apt install -y golang-go git
        # Verify installation
        go version
        ```
    2.  **Copy Backend Files:** Copy the entire `app/` directory content related to the backend (`app/pizza-mix.go`, `app/go.mod`, `app/go.sum`) to the backend VM (e.g., `/opt/pizza-backend/app`).
        ```bash
        # Example using scp from your local machine:
        # scp app/go.mod app/go.sum app/pizza-mix.go user@<backend_vm_ip>:/tmp/
        # On the VM:
        sudo mkdir -p /opt/pizza-backend/app
        sudo mv /tmp/go.mod /tmp/go.sum /tmp/pizza-mix.go /opt/pizza-backend/app/
        cd /opt/pizza-backend/app # Change directory to app/
        ```
    3.  **Fetch Dependencies:**
        ```bash
        # Run from /opt/pizza-backend/app
        sudo -u <your_non_root_user> go mod tidy 
        ```
    4.  **Build Backend:**
        ```bash
        # Run from /opt/pizza-backend/app
        sudo -u <your_non_root_user> go build -o ../pizza-backend pizza-mix.go # Output binary to parent dir
        ```
    5.  **Create Systemd Service:** Update paths in `/etc/systemd/system/pizza-backend.service`:
        ```ini
        [Unit]
        Description=Pizza Recommender Backend API Service
        After=network.target

        [Service]
        User=pizzauser
        Group=pizzauser
        WorkingDirectory=/opt/pizza-backend/app
        ExecStart=/opt/pizza-backend/pizza-backend
        Restart=on-failure
        RestartSec=5s
        # Optional: Add environment variables if needed later
        # Environment="GIN_MODE=release"

        [Install]
        WantedBy=multi-user.target
        ```
        *Adjust `User` and `Group` as needed. Ensure the user has permissions for the `WorkingDirectory`.* 
    6.  **Enable & Start Service:**
        ```bash
        sudo systemctl daemon-reload
        sudo systemctl enable pizza-backend.service
        sudo systemctl start pizza-backend.service
        # Check status:
        sudo systemctl status pizza-backend.service
        ```
    7.  **Configure Firewall:** Allow incoming connections on port 8080.
        ```bash
        # Example using ufw:
        sudo ufw allow 8080/tcp
        sudo ufw reload
        ```

**2. Frontend VM Setup**

*   **Goal:** Serve static files (`index.html`, CSS, JS) using Nginx and proxy API calls to the backend VM.

    1.  **Install Nginx:**
        ```bash
        sudo apt update
        sudo apt install -y nginx
        sudo systemctl enable nginx
        sudo systemctl start nginx
        ```
    2.  **Copy Frontend Files:** Copy the *contents* of the `app/static/` directory to the frontend VM's web root (e.g., `/var/www/pizza-app`).
        ```bash
        # Example using scp from your local machine:
        # scp -r app/static/* user@<frontend_vm_ip>:/tmp/
        # On the VM:
        sudo mkdir -p /var/www/pizza-app
        sudo mv /tmp/* /var/www/pizza-app/ # Move contents of static
        # Ensure correct permissions 
        sudo chown -R www-data:www-data /var/www/pizza-app
        sudo find /var/www/pizza-app -type d -exec chmod 755 {} \;
        sudo find /var/www/pizza-app -type f -exec chmod 644 {} \;
        ```
    3.  **Configure Nginx:** Create an Nginx configuration file (e.g., `/etc/nginx/sites-available/pizza-app.conf`). You can adapt the template found in `ansible/templates/nginx.conf.j2`, manually replacing `{{ backend_ip_address }}` with the actual backend IP and `{{ ansible_fqdn | default(inventory_hostname) }}` with the frontend IP or domain.
        ```nginx
        # Example structure (adapt from ansible/templates/nginx.conf.j2)
        server {
            listen 80; 
            server_name <frontend_vm_ip_or_domain>;
            root /var/www/pizza-app; 
            index index.html;
            location /api/ {
                rewrite ^/api/(.*)$ /$1 break;
                proxy_pass http://<backend_vm_ip>:8080; # Manual replacement needed
                # ... proxy headers ...
            }
            location /static/ { }
            location / { try_files $uri /index.html; }
        }
        ```
    4.  **Enable Nginx Site & Test Config:**
        ```bash
        sudo ln -s /etc/nginx/sites-available/pizza-app.conf /etc/nginx/sites-enabled/
        # Optional: Remove default site if it conflicts on port 80
        # sudo rm /etc/nginx/sites-enabled/default 
        sudo nginx -t # Test configuration
        sudo systemctl reload nginx
        ```
    5.  **Verify Frontend Script:** Ensure `static/js/script.js` (on the frontend VM at `/var/www/pizza-app/js/script.js`) uses the relative path for API calls:
        ```javascript
        const apiBaseUrl = '/api'; // Should already be like this
        // ... fetch calls use `${apiBaseUrl}/ingredients`, etc.
        ```
    6.  **Configure Firewall:** Allow incoming connections on the port Nginx is listening on (e.g., 80 or 8000).
        ```bash
        # Example using ufw (if Nginx listens on port 80):
        sudo ufw allow 'Nginx Full'
        # Or if on port 8000:
        # sudo ufw allow 8000/tcp
        sudo ufw reload
        ```

**3. Accessing the Application**

*   Open your web browser and navigate to the frontend VM's IP address or domain name (e.g., `http://<frontend_vm_ip>` if using port 80, or `http://<frontend_vm_ip>:8000` if using port 8000).
*   The frontend should load, fetch ingredients from the backend via the Nginx proxy, and allow you to get recommendations.

## Automating Deployment with Ansible (Packages + Code + Nginx Config)

This section provides an enhanced playbook (`ansible/deploy.yaml`) that installs packages, deploys code via Git, and configures/enables the Nginx site using a template.

**Assumptions & Scope:**
*   Installs Go, Git, Nginx; deploys code; ensures directories exist; templates and enables the Nginx site config; reloads Nginx on changes.
*   **Does not** build the Go binary, configure/manage the Go systemd service, manage firewalls, or initially create users/groups (`pizzauser`, `www-data`).
*   Requires Ansible, SSH access, and network access as previously noted.

**Files (inside `ansible/` directory):**
*   `inventory.ini`: Defines hosts and variables (including the *new* `backend_ip_address` variable for the backend host).
*   `deploy.yaml`: The Ansible playbook.
*   `templates/nginx.conf.j2`: Jinja2 template for the Nginx site configuration.

**Steps:**
1.  **Prepare Inventory (`ansible/inventory.ini`):**
    *   Edit `ansible/inventory.ini`.
    *   Fill in IPs, user, Git URL, Git version.
    *   **Crucially**, add the `backend_ip_address=<backend_vm_ip>` variable to the backend host line.
    *   Configure SSH key authentication if needed.

2.  **Review Playbook & Template (`ansible/`):**
    *   Review `deploy.yaml`. It now includes tasks using the `template` module to deploy the Nginx config from `templates/nginx.conf.j2` and a handler to reload Nginx.
    *   Review `templates/nginx.conf.j2` to understand the variables used.

3.  **Run the Playbook:**
    *   From your local machine (in the project root directory), run:
        ```bash
        ansible-playbook -i ansible/inventory.ini ansible/deploy.yaml
        ```

4.  **Post-Deployment Steps (Manual/Separate Automation):**
    *   **Backend VM:** Create `pizzauser` user/group. `cd /opt/pizza-backend/app`, run `go build -o ../pizza-backend pizza-mix.go`. Set up and start the `pizza-backend.service` (systemd).
    *   **Frontend VM:** Ensure `www-data` user/group exists. The Nginx site should now be configured and enabled by Ansible.
    *   Configure firewalls on both VMs.

**Further Improvements:**
*   Use Ansible roles for better organization (e.g., roles for `common`, `backend-app`, `frontend-webserver`).
*   Use the Ansible `template` module to manage configuration files (`nginx.conf`, systemd service file).
*   Use Ansible `service` module and Handlers to manage services (start, restart, reload Nginx/pizza-backend).
*   Use Ansible `user` and `group` modules to ensure users/groups exist.
*   Consider separate Git repositories for frontend and backend code.
