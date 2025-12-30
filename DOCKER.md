# Docker Deployment Guide

This application is containerized and ready for deployment on Coolify or any Docker Compose compatible platform.

## Local Testing

To test the Docker setup locally:

```bash
# Build the images
docker-compose build

# Start the services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the services
docker-compose down
```

**Note**: If ports 80 or 3001 are already in use on your machine, you'll need to modify the port mappings in `docker-compose.yaml` before starting.

The services will be available at:
- Frontend: http://localhost (port 80)
- Backend API: http://localhost:3001/api

For local testing, you may want to set environment variables:
```bash
export VITE_API_URL=http://localhost:3001/api
export FRONTEND_URL=http://localhost
docker-compose up --build
```

## Coolify Deployment

### Prerequisites

1. A Coolify instance running
2. Git repository with this code
3. Domain names configured (optional, Coolify can provide subdomains)

### Deployment Steps

1. **Create a New Resource in Coolify**
   - Open your Coolify project
   - Click "Create New Resource"
   - Choose your Git repository (Public, GitHub App, or Deploy Key)

2. **Configure Build Pack**
   - Select "Docker Compose" as the build pack
   - **Base Directory**: `/` (root of repository)
   - **Docker Compose Location**: `docker-compose.yaml`

3. **Configure Environment Variables**

   Coolify provides predefined variables that are automatically available. You only need to set:

   ```
   SESSION_SECRET=<generate-a-strong-random-secret>
   ```

   **Coolify Predefined Variables Used** (automatically available):
   - `PORT` - Automatically set by Coolify to the first port in the `expose` section (3001 for backend)
   - `HOST` - Automatically set by Coolify (defaults to `0.0.0.0`)
   - `COOLIFY_URL` - Full URL(s) of the application (used for `FRONTEND_URL` for CORS)
   - `COOLIFY_FQDN` - Fully qualified domain name(s) of the application (used for Traefik routing)

   **Required Environment Variables** (set in Coolify):
   - `SESSION_SECRET` - Session secret for Express sessions (generate a strong random string)

   **Important**: 
   - `SESSION_SECRET` must be set in Coolify's environment variables - generate a strong random string
   - `COOLIFY_URL` is automatically used for CORS configuration
   - The frontend uses relative paths (`/api`) which are proxied to the backend via Docker networking
   - Traefik routing uses `COOLIFY_FQDN` automatically

5. **Storage Configuration**

   The database is stored in `./backend/data` directory. 

   **For Coolify**: If you want Coolify to automatically create the directory, you can update the volume configuration in `docker-compose.yaml` to use Coolify's `is_directory` feature:

   ```yaml
   volumes:
     - type: bind
       source: ./backend/data
       target: /app/data
       is_directory: true
   ```

   **Note**: The `is_directory: true` property is Coolify-specific and won't work with standard docker-compose. The current configuration works for both local testing and Coolify (Coolify will create the directory automatically if needed).

6. **Deploy**

   Click "Deploy" and Coolify will:
   - Build the Docker images
   - Start the containers
   - Set up health checks
   - Configure the reverse proxy (Traefik)

### Architecture

- **Backend**: Node.js/Express API running on port 3001 (configurable via `PORT` magic variable)
- **Frontend**: Vue.js SPA served by Nginx on port 80
- **Database**: SQLite stored in persistent volume at `/app/data`
- **Networking**: Frontend communicates with backend via Docker networking (service name `backend:3001`)
- **API Proxy**: Nginx proxies `/api` requests to the backend service internally

### Health Checks

Both services include health checks:
- Backend: `GET /api/health`
- Frontend: HTTP check on root path

### Troubleshooting

1. **Frontend can't connect to backend**
   - The frontend uses relative paths (`/api`) which are proxied by Nginx to the backend service
   - Verify the backend service is healthy in Coolify dashboard
   - Check that Nginx is properly proxying requests (check nginx logs)
   - Ensure CORS is configured correctly (`FRONTEND_URL`)
   - Verify Docker networking is working (services can communicate via service names)

2. **Database not persisting**
   - Verify volume mount is configured
   - Check Coolify storage settings
   - Ensure `DATA_DIR` environment variable is set

3. **Sessions not working**
   - Verify `SECRET` is set in Coolify (this becomes `SESSION_SECRET`)
   - Check that cookies are being sent (credentials: 'include' in frontend)
   - Ensure `secure` cookie setting matches your setup (HTTPS vs HTTP)

### Notes

- **Port Management**: Coolify automatically sets `PORT` to the first port in the `expose` section (3001 for backend, 80 for frontend)
- **Host Binding**: Coolify automatically sets `HOST` (defaults to `0.0.0.0`)
- **Docker Networking**: Services communicate internally using service names (`backend`, `frontend`)
- **API Communication**: Frontend uses relative paths (`/api`) which Nginx proxies to the backend service
- **Session Secret**: Set `SESSION_SECRET` in Coolify's environment variables (not a predefined variable)
- **Domain Routing**: Coolify automatically provides `COOLIFY_FQDN` and `COOLIFY_URL` for Traefik routing and CORS
- **Database**: SQLite stored in persistent volume, suitable for single-instance deployments
- **Session Store**: Currently uses MemoryStore - for production, consider Redis for multi-instance deployments

### Coolify Predefined Variables Reference

The following variables are automatically available in Coolify (no need to set them):

- `PORT` - Set to the first port in `expose` section
- `HOST` - Set to `0.0.0.0` (or custom if configured)
- `COOLIFY_URL` - Full URL(s) of the application
- `COOLIFY_FQDN` - Fully qualified domain name(s) of the application
- `COOLIFY_BRANCH` - Git branch name
- `COOLIFY_RESOURCE_UUID` - Unique resource identifier
- `COOLIFY_CONTAINER_NAME` - Container name
- `SOURCE_COMMIT` - Git commit hash (disabled by default for build cache)

