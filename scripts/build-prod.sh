#!/bin/bash

# LAN Relay Production Build Script

echo "🏗️  Building LAN Relay for Production"
echo "===================================="

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

# Build backend
echo "🚀 Building Go backend..."
cd backend
mkdir -p bin
go build -ldflags "-w -s" -o bin/lan-relay .
if [ $? -ne 0 ]; then
    echo "❌ Backend build failed"
    exit 1
fi
echo "✅ Backend built successfully"
cd ..

# Build frontend
echo "🎨 Building React frontend..."
cd frontend
npm run build
if [ $? -ne 0 ]; then
    echo "❌ Frontend build failed"
    exit 1
fi
echo "✅ Frontend built successfully"
cd ..

# Create production package
echo "📦 Creating production package..."
mkdir -p dist
cp -r backend/bin dist/
cp -r frontend/build dist/frontend
cp config/env.example dist/.env.example
cp README.md dist/

echo ""
echo "✅ Production build completed!"
echo "📁 Files are in the 'dist' directory"
echo "🚀 To run: cd dist && ./bin/lan-relay"
echo "" 