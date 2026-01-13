# Single-stage build for Expatter Application
FROM python:3.12-slim-bookworm

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    nodejs \
    npm \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.25.5 manually (since it's required by go.mod)
# Note: Adjust ARCH for your system (amd64 or arm64). For automated builds we can use dpkg --print-architecture
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "amd64" ]; then GOARCH="amd64"; elif [ "$ARCH" = "arm64" ]; then GOARCH="arm64"; else GOARCH="amd64"; fi && \
    wget https://go.dev/dl/go1.23.4.linux-$GOARCH.tar.gz && \
    tar -C /usr/local -xzf go1.23.4.linux-$GOARCH.tar.gz && \
    rm go1.23.4.linux-$GOARCH.tar.gz

# Add Go to PATH
ENV PATH=$PATH:/usr/local/go/bin

# Install jobseek-expat CLI (Python package)
RUN pip3 install --no-cache-dir jobseek-expat

# Setup app directory
WORKDIR /app
RUN mkdir -p /app/data /app/frontend

# --- Build Frontend ---
WORKDIR /build/frontend
COPY jobseek-expat-web-fe .
RUN npm ci
RUN npm run build
# Move dist to app location
RUN cp -r dist /app/frontend/dist

# --- Build Backend ---
WORKDIR /build/backend
COPY jobseek-expat-web-be .
# Allow Go to downgrade capabilities if needed, or update toolchain
# We set GOTOOLCHAIN=auto so it downloads 1.25.5 if needed
ENV GOTOOLCHAIN=auto
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/expatter-server .

# --- Runtime Setup ---
WORKDIR /app

# Create app user
RUN groupadd -g 1000 appuser && \
    useradd -r -u 1000 -g appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose port
EXPOSE 8080

# Environment variables
ENV PORT=8080 \
    SCHEDULER_FREQUENCY="@every 1h" \
    APP_NAME="Expatter" \
    APP_DOMAIN="https://expatter.gyokhan.com" \
    FRONTEND_PATH="/app/frontend/dist"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["./expatter-server"]
