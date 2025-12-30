#!/bin/sh

# Ensure backend binds to localhost only (not 0.0.0.0)
export HOST=127.0.0.1

# Export backend port for Vite proxy (use PORT env var, default to 3001)
export BACKEND_PORT=${PORT:-3001}

# Start the Go backend server in the background
cd /app && ./backend/gym-app-backend &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Get the frontend port from environment (defaults to 5173)
FRONTEND_PORT=${FRONTEND_PORT:-5173}
echo "Starting Vue dev server on port $FRONTEND_PORT"

# Start the Vue dev server in the background
# Unset PORT to avoid conflict with backend, but keep BACKEND_PORT for proxy
cd /app/frontend
(
  unset PORT
  export FRONTEND_PORT=$FRONTEND_PORT
  export BACKEND_PORT=${BACKEND_PORT:-3001}
  npm run dev
) &
FRONTEND_PID=$!

# Function to check if a process is running
check_process() {
    if ! kill -0 $1 2>/dev/null; then
        return 1
    fi
    return 0
}

# Monitor and restart processes if they die
while true; do
    # Check backend
    if ! check_process $BACKEND_PID; then
        echo "Backend died, restarting..."
        export HOST=127.0.0.1
        export BACKEND_PORT=${PORT:-3001}
        cd /app && ./backend/gym-app-backend &
        BACKEND_PID=$!
    fi

    # Check frontend
    if ! check_process $FRONTEND_PID; then
        echo "Frontend died, restarting..."
        cd /app/frontend
        (
          unset PORT
          export FRONTEND_PORT=$FRONTEND_PORT
          export BACKEND_PORT=${BACKEND_PORT:-3001}
          npm run dev
        ) &
        FRONTEND_PID=$!
    fi

    # Wait 10 seconds before next check
    sleep 10
done
