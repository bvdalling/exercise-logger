#!/bin/sh

# Start the Go backend server in the background
cd /app && ./backend/gym-app-backend &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Get the frontend port from environment (defaults to 5173)
FRONTEND_PORT=${FRONTEND_PORT:-5173}
echo "Starting Vue dev server on port $FRONTEND_PORT"

# Start the Vue dev server in the background
# Unset PORT to avoid conflict with backend (PORT=3001), explicitly set FRONTEND_PORT
cd /app/frontend
(
  unset PORT
  export FRONTEND_PORT=$FRONTEND_PORT
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
          npm run dev
        ) &
        FRONTEND_PID=$!
    fi

    # Wait 10 seconds before next check
    sleep 10
done
