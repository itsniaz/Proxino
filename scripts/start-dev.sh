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

# Start backend in background
echo "🚀 Starting Go backend server..."
cd backend
go run . &
BACKEND_PID=$!
cd ..

# Wait a moment for backend to start
sleep 2

# Start frontend
echo "🎨 Starting React frontend..."
cd frontend
npm start &
FRONTEND_PID=$!
cd ..

echo ""
echo "✅ Services started successfully!"
echo "📊 Dashboard: http://localhost:3000"
echo "🔧 Backend API: http://localhost:8080"
echo ""
echo "To stop services, press Ctrl+C"

# Wait for interrupt
trap "echo ''; echo '🛑 Stopping services...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0" INT

wait 