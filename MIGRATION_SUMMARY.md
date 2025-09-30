# Migration Summary: SQLite to PostgreSQL Three-Tier Architecture

## What Changed

Your pizza mixing application has been refactored from a two-tier architecture (frontend + backend with embedded SQLite) to a proper three-tier architecture (frontend + backend + PostgreSQL database).

## Key Changes

### 1. Database Layer (NEW)
- **Before**: SQLite database file embedded with backend (`/app/data/pizza.db`)
- **After**: Separate PostgreSQL 16 database server
- **Benefits**:
  - True separation of data layer
  - Can run on separate VM or container
  - Better for production environments
  - Supports concurrent connections
  - Can be scaled independently

### 2. Backend Application
**Files Modified**:
- `app/go.mod` - Changed from `mattn/go-sqlite3` to `lib/pq` (PostgreSQL driver)
- `app/db.go` - Updated connection logic and SQL syntax for PostgreSQL
- `app/repository.go` - Changed from `GROUP_CONCAT` (SQLite) to `STRING_AGG` (PostgreSQL)
- `app/pizza-mix.go` - Added environment variable configuration for database connection

**Key Changes**:
- Uses environment variables for database connection (no hardcoded paths)
- PostgreSQL connection string format
- Updated SQL syntax:
  - `INTEGER PRIMARY KEY AUTOINCREMENT` → `SERIAL PRIMARY KEY`
  - `INSERT OR IGNORE` → `INSERT ... ON CONFLICT DO NOTHING`
  - `?` placeholders → `$1, $2, $3` placeholders
  - `GROUP_CONCAT()` → `STRING_AGG()`

### 3. Docker Configuration
**Files Modified**:
- `docker/compose.yaml` - Added database service, updated backend with environment variables
- `docker/Dockerfile` - Removed CGO and SQLite dependencies

**New Services**:
```yaml
database:   # PostgreSQL container with persistent volume
backend:    # Go API with DB connection via env vars
frontend:   # Nginx (unchanged)
```

### 4. Deployment Configuration (NEW)
**New Files Created**:
- `deploy/.env.docker` - Environment variables for Docker deployment
- `deploy/.env.vm.example` - Template for VM deployment
- `deploy/VM_DEPLOYMENT.md` - Complete guide for deploying across VMs
- `ARCHITECTURE.md` - Architecture documentation
- `MIGRATION_SUMMARY.md` - This file

**Updated Files**:
- `README.md` - Updated with new architecture info
- `DEPLOYMENT.md` - References new VM deployment guide

## How to Use

### Docker Compose (Development/Testing)

No changes needed from before! Just run:

```bash
cd docker
docker compose up --build -d
```

The database will automatically:
1. Start PostgreSQL in a container
2. Backend will connect to it
3. Create tables on first run
4. Seed data automatically

Access at: `http://localhost:8000`

### Separate VMs (Production)

You now have full flexibility to deploy across separate VMs:

**Option 1: Three VMs (Recommended)**
- VM1: PostgreSQL database
- VM2: Go backend API
- VM3: Nginx frontend

**Option 2: Two VMs (Simplified)**
- VM1: PostgreSQL database
- VM2: Go backend + Nginx frontend

**Option 3: Hybrid**
- Database on managed service (AWS RDS, Azure Database)
- Backend + Frontend on VMs or containers

See detailed instructions in: `deploy/VM_DEPLOYMENT.md`

## Configuration

### Environment Variables

The backend now uses these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | Database hostname or IP |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | pizza | Database username |
| `DB_PASSWORD` | pizzapass | Database password |
| `DB_NAME` | pizza | Database name |
| `DB_SSLMODE` | disable | SSL mode for connection |

### Docker Deployment
Environment variables are set in `docker/compose.yaml` automatically.

### VM Deployment
Set environment variables using one of these methods:

1. **Environment file** (recommended):
   ```bash
   # Create .env file
   cat > /opt/pizza-backend/.env <<EOF
   DB_HOST=192.168.1.10
   DB_PORT=5432
   DB_USER=pizza
   DB_PASSWORD=your_password
   DB_NAME=pizza
   DB_SSLMODE=require
   EOF

   # Load in systemd service
   [Service]
   EnvironmentFile=/opt/pizza-backend/.env
   ```

2. **Systemd service file**:
   ```ini
   [Service]
   Environment="DB_HOST=192.168.1.10"
   Environment="DB_PORT=5432"
   ```

3. **Export in shell**:
   ```bash
   export DB_HOST=192.168.1.10
   export DB_PORT=5432
   # ... then run application
   ```

## Compatibility

### Breaking Changes
- **SQLite database files are no longer used**
  - Old data needs to be migrated manually if you had custom data
  - The seed data (15 pizzas) will be recreated in PostgreSQL automatically

### Non-Breaking Changes
- API endpoints remain the same
- Frontend code unchanged
- Response formats unchanged

## Testing the Migration

### 1. Test Docker Deployment

```bash
cd /mnt/c/Users/saaku/the-lab/experiment-dome/pizza-mixing/docker
docker compose up --build -d

# Check all services are running
docker compose ps

# Check backend logs
docker compose logs backend

# Test health endpoint
curl http://localhost:8080/health

# Test frontend
curl http://localhost:8000
```

### 2. Test Database Separately

```bash
# Connect to database container
docker exec -it pizza-database psql -U pizza -d pizza

# Check tables
\dt

# Check data
SELECT * FROM pizzas;
SELECT * FROM ingredients LIMIT 5;

# Exit
\q
```

### 3. Test API

```bash
# Get ingredients
curl http://localhost:8080/ingredients

# Get recommendations
curl -X POST http://localhost:8080/recommend \
  -H "Content-Type: application/json" \
  -d '{"disliked":["pineapple"],"preferred":["pepperoni","mushrooms"]}'
```

## Rollback Plan

If you need to revert to the old SQLite version:

```bash
# Checkout the previous commit
git log --oneline  # Find commit before migration
git checkout <commit-hash>

# Or revert specific files
git checkout HEAD~1 -- app/
git checkout HEAD~1 -- docker/
```

## Next Steps

1. **Test the Docker deployment** to ensure everything works
2. **Update any CI/CD pipelines** with new docker-compose commands
3. **Plan VM deployment** if targeting separate infrastructure
4. **Set up database backups** for PostgreSQL
5. **Configure monitoring** for the database service
6. **Update secrets management** for production database passwords

## Benefits of This Architecture

✅ **True separation of concerns**: Database, backend, and frontend are independent
✅ **Scalable**: Each tier can scale independently
✅ **Production-ready**: PostgreSQL is enterprise-grade
✅ **Flexible deployment**: Works with Docker OR separate VMs
✅ **Better for teams**: Database can be managed separately
✅ **Cloud-ready**: Easy to use managed database services (RDS, Cloud SQL, etc.)
✅ **No data loss on backend restarts**: Database is separate from application

## Questions?

- Docker issues? Check `docker compose logs`
- VM deployment? See `deploy/VM_DEPLOYMENT.md`
- Architecture questions? See `ARCHITECTURE.md`
- PostgreSQL help? See PostgreSQL docs at postgresql.org
