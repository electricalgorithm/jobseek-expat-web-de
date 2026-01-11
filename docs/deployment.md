# Deployment Guide

## Prerequisites

1. **Docker** and **Docker Compose** installed
2. **SSH Deploy Key** for accessing private GitHub repositories
3. **Resend API Key** for sending emails

## Setup Instructions

### 1. Generate SSH Deploy Key

```bash
# Generate a new SSH key for deployment
ssh-keygen -t ed25519 -C "deploy@expatter" -f ~/.ssh/expatter_deploy_key

# Add the public key to both GitHub repositories
cat ~/.ssh/expatter_deploy_key.pub
# Go to GitHub repo → Settings → Deploy keys → Add deploy key
# Paste the public key and grant read access
```

### 2. Configure Environment Variables

```bash
# Copy the example environment file
cp .env.docker.example .env

# Edit .env and add your Resend API key
nano .env
```

Required environment variables:
- `RESEND_API_KEY`: Your Resend API key for sending emails

### 3. Build and Run

```bash
# Build the Docker image (with SSH key for private repos)
DOCKER_BUILDKIT=1 docker-compose build --ssh default=$HOME/.ssh/expatter_deploy_key

# Start the container
docker-compose up -d

# Check logs
docker-compose logs -f
```

### 4. Verify Deployment

```bash
# Check if container is running
docker-compose ps

# Test the application
curl http://localhost:8080

# Check database
docker exec -it expatter-server ls -la /app/data
```

## Data Persistence

The database is stored in a Docker volume named `expatter-data`. This ensures:
- ✅ Data persists across container restarts
- ✅ Data persists across container rebuilds
- ✅ Data is not lost when container is stopped

### Backup Database

```bash
# Backup the database
docker cp expatter-server:/app/data/jobseek.db ./backup-$(date +%Y%m%d).db

# Restore from backup
docker cp ./backup-20260111.db expatter-server:/app/data/jobseek.db
docker-compose restart
```

## Container Management

```bash
# Stop the container
docker-compose stop

# Start the container
docker-compose start

# Restart the container
docker-compose restart

# View logs
docker-compose logs -f

# Remove container (keeps data)
docker-compose down

# Remove container AND data volume (⚠️ DESTRUCTIVE)
docker-compose down -v
```

## Updating the Application

```bash
# Pull latest code and rebuild
DOCKER_BUILDKIT=1 docker-compose build --ssh default=$HOME/.ssh/expatter_deploy_key --no-cache

# Restart with new image
docker-compose up -d

# Check logs
docker-compose logs -f
```

## Production Deployment

For production, update `docker-compose.yml`:

1. Change `APP_DOMAIN` to your actual domain
2. Change `EMAIL_FROM` to your email address
3. Add reverse proxy (nginx/traefik) for HTTPS
4. Set `SCHEDULER_FREQUENCY` as needed

Example nginx config:
```nginx
server {
    listen 80;
    server_name yourdomain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker-compose logs

# Check if port 8080 is already in use
lsof -i :8080
```

### Database issues
```bash
# Access container shell
docker exec -it expatter-server sh

# Check database
cd /app/data
ls -la
```

### SSH key issues during build
```bash
# Ensure SSH agent is running
eval $(ssh-agent)
ssh-add ~/.ssh/expatter-deploy-key

# Verify key is added
ssh-add -l

# Test GitHub access
ssh -T git@github.com
```

## Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `APP_NAME` | `Expatter` | Application name |
| `APP_DOMAIN` | `http://localhost:8080` | Public domain |
| `RESEND_API_KEY` | - | Resend API key (required) |
| `EMAIL_FROM` | `jobs@yourdomain.com` | Sender email |
| `SCHEDULER_FREQUENCY` | `@every 1h` | Job check frequency |
| `DB_PATH` | `./data/jobseek.db` | Database file path |
