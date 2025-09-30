# Cleanup Summary - SQLite to PostgreSQL Migration

This document summarizes all the cleanup performed after migrating from SQLite to PostgreSQL.

## Files Removed

### 1. `app/db.sql`
**Reason**: Old SQLite schema file. The schema is now defined in `app/db.go` using PostgreSQL syntax.

**What was in it**: SQLite table definitions with `AUTOINCREMENT` syntax.

**Replacement**: Schema migration is handled in `app/db.go` with PostgreSQL `SERIAL` syntax.

## Files Renamed/Deprecated

### 1. `ansible/deploy.yaml` → `ansible/deploy.yaml.deprecated`
**Reason**: This playbook deploys the old two-tier architecture (backend with embedded SQLite + frontend on separate VMs).

**Why deprecated**:
- Uses SQLite (embedded database)
- Doesn't support the new three-tier architecture
- Missing database service deployment

**Replacements**:
- For Docker deployment: Use `ansible/deploy-vcf-docker.yaml`
- For VM deployment: Follow manual steps in `deploy/VM_DEPLOYMENT.md`

### 2. `ansible/inventory.ini` → `ansible/inventory.ini.deprecated`
**Reason**: Inventory file for the deprecated `deploy.yaml` playbook.

**Replacement**: Use `ansible/inventory-vcf.ini` for Docker deployments.

### 3. `ansible/templates/` → `ansible/templates.deprecated/`
**Reason**: Nginx template used by the deprecated playbook.

**Note**: The template itself is still valid and can be reused if needed.

## Code Updates

### 1. `app/db.go`
**Changes**:
- Import changed from `github.com/mattn/go-sqlite3` to `github.com/lib/pq`
- `connectDB()` now takes connection string instead of file path
- SQL syntax updated for PostgreSQL:
  - `INTEGER PRIMARY KEY AUTOINCREMENT` → `SERIAL PRIMARY KEY`
  - `INSERT OR IGNORE` → `INSERT ... ON CONFLICT DO NOTHING`
  - `?` placeholders → `$1, $2, $3` placeholders
- Error checking updated for PostgreSQL constraint errors

### 2. `app/repository.go`
**Changes**:
- `GROUP_CONCAT()` → `STRING_AGG()` (PostgreSQL function)
- Placeholder syntax updated from `?` to `$1, $2, $3`

### 3. `app/pizza-mix.go`
**Changes**:
- Added environment variable configuration
- Removed file path logic for SQLite
- Added `getEnv()` helper function
- Updated to use PostgreSQL connection string

### 4. `app/go.mod` and `app/go.sum`
**Changes**:
- Removed: `github.com/mattn/go-sqlite3 v1.14.28`
- Added: `github.com/lib/pq v1.10.9`
- Ran `go mod tidy` to clean up dependencies

### 5. `docker/Dockerfile`
**Changes**:
- Removed CGO requirement (PostgreSQL driver is pure Go)
- Removed `build-base` and `sqlite-libs` dependencies
- Simplified build process
- Changed `CGO_ENABLED=1` to `CGO_ENABLED=0`

### 6. `docker/compose.yaml`
**Changes**:
- Added `database` service (PostgreSQL 16)
- Added environment variables to `backend` service
- Added health check for database
- Added volume for database persistence
- Backend now depends on database health

## Documentation Updates

### 1. `README.md`
**Changes**:
- Updated "Code Structure" section with new files
- Updated Docker section to mention three-tier architecture
- Added deployment options section
- Marked legacy VM deployment as deprecated

### 2. `.gitignore`
**Changes**:
- Added `.env` and environment files
- Added `*.db`, `*.sqlite*` to ignore old database files
- Added `/app/data/` directory (no longer needed)
- Added `/app/pizza-backend` binary

## New Files Created

### Documentation
1. `ARCHITECTURE.md` - Complete architecture documentation
2. `MIGRATION_SUMMARY.md` - Migration guide and testing steps
3. `CLEANUP_SUMMARY.md` - This file
4. `ansible/README.md` - Ansible documentation

### Deployment
1. `deploy/VM_DEPLOYMENT.md` - Three-tier VM deployment guide
2. `deploy/.env.docker` - Docker environment configuration
3. `deploy/.env.vm.example` - VM environment template

## Files That Remain Unchanged

These files work with both old and new architecture:

### Deployment Scripts
- `deploy/deploy-ec2.sh` - Deploys Docker Compose (now with PostgreSQL)
- `deploy/deploy-vcf.sh` - Deploys Docker Compose (now with PostgreSQL)
- `ansible/deploy-vcf-docker.yaml` - Deploys Docker Compose (now with PostgreSQL)

### Frontend
- `app/static/index.html`
- `app/static/css/style.css`
- `app/static/js/script.js`
- `docker/Dockerfile.frontend`
- `docker/nginx-docker.conf`

**Note**: Frontend is unchanged because API endpoints remain the same.

## Verification Checklist

After cleanup, verify:

- [x] No SQLite references in active code
- [x] No `go-sqlite3` imports in Go files
- [x] go.mod and go.sum cleaned up
- [x] Old database schema file removed
- [x] Deprecated playbooks renamed
- [x] Documentation updated
- [x] .gitignore updated
- [x] New documentation created

## What Was NOT Removed

We kept these deprecated files for reference:

1. **`ansible/deploy.yaml.deprecated`** - Shows how the old 2-tier deployment worked
2. **`ansible/inventory.ini.deprecated`** - Inventory example for old playbook
3. **`ansible/templates.deprecated/`** - Nginx template (still useful)

### Why Keep Them?

- Historical reference
- May help users migrating from old version
- Templates can be reused
- Shows evolution of the project

## Migration Path for Users

If users have the old SQLite version deployed:

### Option 1: Fresh Deployment (Recommended)
1. Deploy new three-tier architecture
2. Manually migrate custom data if any
3. Update DNS/load balancer to point to new deployment
4. Decommission old deployment

### Option 2: In-Place Update (Complex)
1. Export SQLite data
2. Set up PostgreSQL database
3. Import data to PostgreSQL
4. Update backend code
5. Update configuration
6. Restart services

**Note**: The seed data (15 pizzas) will be automatically recreated, so only custom additions need migration.

## Summary

**Removed**: 1 file (db.sql)
**Deprecated**: 3 items (deploy.yaml, inventory.ini, templates/)
**Updated**: 8 files (Go code, Dockerfile, compose.yaml, documentation)
**Created**: 7 new files (documentation and configs)

The codebase is now clean, consistent, and fully PostgreSQL-based with no SQLite remnants in active code.
