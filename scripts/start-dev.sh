#!/bin/bash

# LAN Relay Development Startup Script

echo "🛰️  Starting LAN Relay Development Environment"
echo "============================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js first."
    exit 1
fi

# Function to cleanup processes
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    if [ ! -z "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
        kill "$BACKEND_PID" 2>/dev/null
        echo "   ✅ Backend stopped"
    fi
    if [ ! -z "$FRONTEND_PID" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
        kill "$FRONTEND_PID" 2>/dev/null
        echo "   ✅ Frontend stopped"
    fi
    exit 0
}

# Set up signal handling
trap cleanup INT TERM

# Start backend in background
echo "🚀 Starting Go backend server..."
if ! cd backend; then
    echo "❌ Failed to change to backend directory"
    exit 1
fi

go run . &
BACKEND_PID=$!
cd ..

# Wait for backend to start and check if it's running
sleep 3
if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
    echo "❌ Backend failed to start"
    exit 1
fi

# Start frontend
echo "🎨 Starting React frontend..."
if ! cd frontend; then
    echo "❌ Failed to change to frontend directory"
    cleanup
    exit 1
fi

npm start &
FRONTEND_PID=$!
cd ..

# Wait for frontend to start and check if it's running
sleep 3
if ! kill -0 "$FRONTEND_PID" 2>/dev/null; then
    echo "❌ Frontend failed to start"
    cleanup
    exit 1
fi

echo ""
echo "✅ Services started successfully!"
echo "📊 Dashboard: http://localhost:3000"
echo "🔧 Backend API: http://localhost:8080"
echo ""
echo "To stop services, press Ctrl+C"

# Wait for processes to finish or be interrupted
wait 