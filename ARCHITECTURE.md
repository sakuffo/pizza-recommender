# Pizza Mixing App - Architecture Overview

## Three-Tier Architecture

The application has been refactored into a modern three-tier architecture that separates data, business logic, and presentation layers.

```
┌─────────────────┐
│   Frontend VM   │
│     (Nginx)     │  Port 80/8000
│  Static Files   │
└────────┬────────┘
         │ HTTP/Proxy
         ▼
┌─────────────────┐
│   Backend VM    │
│   (Go API)      │  Port 8080
│  Business Logic │
└────────┬────────┘
         │ PostgreSQL Protocol
         ▼
┌─────────────────┐
│  Database VM    │
│  (PostgreSQL)   │  Port 5432
│   Data Layer    │
└─────────────────┘
```

## Components

### 1. Database Layer (PostgreSQL)
- **Technology**: PostgreSQL 16
- **Responsibility**: Persistent data storage
- **Port**: 5432
- **Data Stored**:
  - Pizza definitions
  - Ingredients catalog
  - Pizza-ingredient relationships

**Key Features**:
- Relational data model with foreign keys
- ACID compliance
- Data persistence across restarts
- Can be deployed standalone or as PostgreSQL cluster for HA

### 2. Backend Layer (Go API)
- **Technology**: Go 1.24+ with Gin framework
- **Responsibility**: Business logic and API endpoints
- **Port**: 8080
- **API Endpoints**:
  - `GET /health` - Health check
  - `GET /ingredients` - List all ingredients
  - `POST /recommend` - Get pizza recommendations

**Key Features**:
- Stateless design (scales horizontally)
- Environment-based configuration
- Database connection pooling
- CORS support for frontend
- Recommendation algorithm with scoring

### 3. Frontend Layer (Nginx + Static Files)
- **Technology**: Nginx web server
- **Responsibility**: Serve UI and proxy API requests
- **Port**: 80 (or 8000)
- **Content**:
  - HTML/CSS/JavaScript
  - Static assets
  - API proxy to backend

**Key Features**:
- Fast static file serving
- Reverse proxy to backend API
- Can serve multiple frontend instances behind load balancer
- CDN-ready static assets

## Deployment Models

### Model 1: Docker Compose (Development/Single Host)

All three tiers run as Docker containers on a single host:

```yaml
services:
  database:    # PostgreSQL container
  backend:     # Go API container
  frontend:    # Nginx container
```

**Advantages**:
- Easy setup and teardown
- Consistent across environments
- Isolated networking
- Volume persistence for database

**Use Cases**:
- Local development
- CI/CD testing
- Small deployments on single VM
- Docker Desktop on Windows/Mac

### Model 2: Separate VMs (Production)

Each tier runs on its own VM:

```
VM1: PostgreSQL + OS
VM2: Go Binary + OS
VM3: Nginx + Static Files + OS
```

**Advantages**:
- True isolation
- Independent scaling
- Resource optimization per tier
- Better security boundaries
- Easier maintenance per component

**Use Cases**:
- Production environments
- VMware vSphere/ESXi
- AWS EC2 / Azure VMs
- On-premises data centers

### Model 3: Hybrid (Containers on VMs)

Run Docker containers on separate VMs:

```
VM1: Docker + PostgreSQL container
VM2: Docker + Go container
VM3: Docker + Nginx container
```

**Advantages**:
- Combines VM isolation with container benefits
- Easier updates (just rebuild containers)
- Consistent environments
- Can use container orchestration

**Use Cases**:
- Organizations transitioning to containers
- Environments with both VM and container expertise
- Kubernetes clusters (future)

## Configuration

### Environment Variables

The backend uses environment variables for configuration:

```bash
# Database Connection
DB_HOST=192.168.1.10      # Database server hostname/IP
DB_PORT=5432               # PostgreSQL port
DB_USER=pizza              # Database username
DB_PASSWORD=pizzapass      # Database password
DB_NAME=pizza              # Database name
DB_SSLMODE=disable         # SSL mode (disable/require/verify-full)
```

### Docker Compose
Environment variables are set in `docker/compose.yaml` for the backend service.

### VM Deployment
Environment variables can be set via:
- `.env` file (loaded by systemd)
- Systemd service file (`Environment=` directive)
- Shell profile (`~/.bashrc`, `/etc/environment`)

## Data Flow

### User Request Flow

1. **User opens browser** → `http://frontend-vm/`
2. **Nginx serves** → `index.html`, CSS, JS
3. **User selects ingredients** → JavaScript captures input
4. **JS makes API call** → `POST /api/recommend`
5. **Nginx proxies** → `http://backend-vm:8080/recommend`
6. **Go API processes** → Query database for pizzas
7. **PostgreSQL returns** → Pizza data with ingredients
8. **Go calculates scores** → Based on preferences
9. **API returns JSON** → Top 3 recommendations
10. **JS renders results** → Display to user

### Database Initialization Flow

1. Backend starts → Reads `DB_HOST` environment variable
2. Connects to PostgreSQL → Using connection string
3. Runs migrations → Creates tables if not exist
4. Seeds data → Inserts initial pizza/ingredient data (idempotent)
5. Ready to serve → Application starts listening

## Scaling Strategies

### Horizontal Scaling

**Frontend**:
- Add more Nginx VMs
- Place behind load balancer (HAProxy, AWS ALB, etc.)
- Serve static files from CDN

**Backend**:
- Add more Go API VMs
- No session state to sync (stateless)
- Frontend load balances between backends
- Each backend connects to same database

**Database**:
- PostgreSQL primary + read replicas
- Backend reads from replicas
- Backend writes to primary
- Use connection pooler (PgBouncer)

### Vertical Scaling

**Frontend**: Rarely needed (static files)
**Backend**: Add more CPU/RAM for concurrent requests
**Database**: Add more RAM for caching, faster disks for I/O

## Security Considerations

### Network Segmentation
- Frontend in DMZ (public network)
- Backend in application tier (private network)
- Database in data tier (most restricted network)

### Firewall Rules
```
Frontend:  Allow 80/443 from Internet
           Allow 8080 to Backend
Backend:   Allow 8080 from Frontend only
           Allow 5432 to Database
Database:  Allow 5432 from Backend only
```

### SSL/TLS
- Frontend: HTTPS with Let's Encrypt or corporate certs
- Backend: Can use HTTP (internal network)
- Database: SSL connections (`DB_SSLMODE=require`)

### Secrets Management
- Use environment variables, not hardcoded
- Production: Use vault (HashiCorp Vault, AWS Secrets Manager)
- Rotate passwords regularly
- Never commit passwords to Git

## Monitoring

### Health Checks
- Frontend: `curl http://frontend-vm/` (200 OK)
- Backend: `curl http://backend-vm:8080/health` ({"status":"healthy"})
- Database: `pg_isready -h db-vm -U pizza`

### Logs
- Frontend: `/var/log/nginx/access.log`, `/var/log/nginx/error.log`
- Backend: `journalctl -u pizza-backend -f` (systemd) or Docker logs
- Database: `/var/log/postgresql/postgresql-16-main.log`

### Metrics
Consider adding:
- Prometheus for metrics collection
- Grafana for visualization
- Application Performance Monitoring (APM) tool

## Disaster Recovery

### Backups
- **Database**: PostgreSQL dumps (`pg_dump`) daily
- **Backend**: Code in Git, nothing to back up
- **Frontend**: Static files in Git, nothing to back up

### Recovery
1. Restore database from backup
2. Deploy backend from Git
3. Deploy frontend from Git
4. Update configuration to point to new IPs

## Future Enhancements

- **Kubernetes**: Orchestrate containers across multiple nodes
- **Service Mesh**: Istio for advanced networking
- **API Gateway**: Kong or AWS API Gateway
- **Caching**: Redis for frequent queries
- **Search**: Elasticsearch for ingredient search
- **Analytics**: Track recommendations for insights
