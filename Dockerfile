# Multi-stage Dockerfile for Expatter Application
# Assumes repos are already cloned locally

# Stage 1: Build Frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /build
COPY ../jobseek-web-fe ./frontend

# Build frontend
WORKDIR /build/frontend
RUN npm ci
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.21-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build
COPY . ./backend

# Build backend
WORKDIR /build/backend
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o expatter-server .

# Stage 3: Production Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs python3 py3-pip

# Install jobseek-expat CLI (Python package)
RUN pip3 install --no-cache-dir jobseek-expat --break-system-packages

# Create app user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create directories
RUN mkdir -p /app/data /app/frontend && \
    chown -R appuser:appuser /app

# Copy backend binary
COPY --from=backend-builder --chown=appuser:appuser /build/backend/expatter-server /app/

# Copy frontend build
COPY --from=frontend-builder --chown=appuser:appuser /build/frontend/dist /app/frontend/dist

# Switch to app user
USER appuser
WORKDIR /app

# Expose port
EXPOSE 8080

# Volume for persistent data (database)
VOLUME ["/app/data"]

# Environment variables
ENV PORT=8080 \
    SCHEDULER_FREQUENCY="@every 1h" \
    APP_NAME="Expatter" \
    APP_DOMAIN="http://localhost:8080"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the application
CMD ["./expatter-server"]
