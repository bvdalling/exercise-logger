# Build GO Backend
FROM golang:alpine AS go-builder

# Install build dependencies for CGO (required for sqlite3)
RUN apk add gcc musl-dev

WORKDIR /app/backend

# Set GOTOOLCHAIN to auto to allow downloading required toolchain version
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy backend source code
COPY backend/ .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o gym-app-backend .

# Runtime stage
FROM node:20-alpine

# Install ca-certificates and sqlite for CGO (for Go binary)
RUN apk --no-cache add ca-certificates sqlite wget

WORKDIR /app

# Copy backend binary
COPY --from=go-builder /app/backend/gym-app-backend ./backend/

# Copy frontend source code and install dependencies
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .

# Create data directory for database
RUN mkdir -p /app/data

# Set environment variables
ENV DATA_DIR=/app/data
ENV PORT=3001
ENV HOST=0.0.0.0
ENV FRONTEND_PORT=5173

# Copy startup script
COPY startup.sh /app/startup.sh
RUN chmod +x /app/startup.sh

# Expose port 5173 for Vite dev server (can be overridden with FRONTEND_PORT env var at runtime)
# Note: EXPOSE doesn't support variable substitution, use -p flag at runtime to map custom ports
EXPOSE 5173

# Health check (uses FRONTEND_PORT env var or defaults to 5173)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD sh -c 'PORT=${FRONTEND_PORT:-5173}; wget --no-verbose --tries=1 --spider http://localhost:$PORT/api/health || exit 1'

# Start both services
CMD ["/app/startup.sh"]
