# 🛰️ LAN Relay with Tunnel & Monitoring Dashboard

A powerful system that allows sending **HTTP/HTTPS** requests from anywhere (e.g., office) to **devices inside a home LAN** that don't have public IPs. The system uses a **Go backend server** exposed via **ngrok** as a secure relay, with a beautiful **React dashboard** for monitoring.

## 🚀 Features

- **🔄 HTTP Proxy**: Forward requests from external sources to internal LAN devices
- **📊 Real-time Dashboard**: Monitor all proxy requests with detailed logging
- **🛡️ Security**: Only allows proxying to private IP ranges (192.168.x.x, 10.x.x.x, etc.)
- **📝 Request Logging**: SQLite database stores all request details (IP, method, target, status, duration)
- **🔍 Monitoring**: System status, uptime tracking, and health checks
- **🎨 Modern UI**: Clean, responsive React dashboard with Tailwind CSS
- **⚡ Real-time Updates**: Auto-refreshing dashboard every 30 seconds

## 🏗️ Architecture

```
[Client (Office)] 
   ↓ HTTP
[ngrok public URL] 
   ↓
[Go Relay Server @ Home (192.168.X.X)]
   ↓
[Local LAN Devices]
↑
[React Dashboard]
```

## 🛠️ Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go + Gin Framework |
| Database | SQLite |
| Frontend | React + TypeScript + Tailwind CSS |
| Tunnel | ngrok |
| Proxy | Go `net/http` + `httputil.ReverseProxy` |

## 📋 Prerequisites

- **Go** 1.19 or higher
- **Node.js** 16 or higher
- **npm** or **yarn**
- **ngrok** (for external access)

## 🚀 Quick Start

### 1. Clone and Setup

```bash
git clone <your-repo>
cd local_router

# Install Go dependencies
cd backend
go mod tidy
cd ..

# Install React dependencies
cd frontend
npm install
cd ..
```

### 2. Development Mode

Use the provided development script:

```bash
./scripts/start-dev.sh
```

This will start both the Go backend (port 8080) and React frontend (port 3000).

### 3. Manual Setup

**Backend:**
```bash
cd backend
go run .
```

**Frontend:**
```bash
cd frontend
npm start
```

### 4. Expose with ngrok

```bash
ngrok http 8080
```

Copy the provided URL (e.g., `https://abc123.ngrok.io`)

## 📖 Usage

### Making Proxy Requests

Send requests to your ngrok URL using this format:

```
https://your-ngrok-url.ngrok.io/proxy/TARGET_IP:PORT/path
```

**Examples:**

```bash
# Proxy to a local web server
curl https://abc123.ngrok.io/proxy/192.168.0.100:8080/api/health

# Proxy to a local API with query parameters
curl "https://abc123.ngrok.io/proxy/192.168.0.150:3000/api/users?limit=10"

# POST request with data
curl -X POST https://abc123.ngrok.io/proxy/192.168.0.200:5000/webhook \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello from remote!"}'
```

### Dashboard Features

Access the dashboard at `http://localhost:3000`:

- **📊 Status Cards**: System online status, total requests, uptime, ngrok status
- **📝 Request Logs**: Real-time table showing all proxy requests with details
- **🔄 Controls**: Refresh data and clear logs buttons
- **📱 Responsive**: Works on desktop, tablet, and mobile devices

## ⚙️ Configuration

### Environment Variables

Create a `.env` file in the backend directory:

```bash
cp config/env.example backend/.env
```

Available options:

```env
PORT=8080                    # Server port
ENVIRONMENT=development      # development/production
DATABASE_PATH=relay.db      # SQLite database file
LOG_LEVEL=info              # debug/info/warn/error
NGROK_TOKEN=your_token      # Optional: ngrok auth token
NGROK_DOMAIN=custom.ngrok.io # Optional: custom ngrok domain
```

## 🏗️ Project Structure

```
local_router/
├── backend/                 # Go backend server
│   ├── internal/
│   │   ├── config/         # Configuration management
│   │   ├── database/       # SQLite database operations
│   │   ├── handlers/       # HTTP request handlers
│   │   ├── logger/         # Logging utilities
│   │   └── models/         # Data models
│   ├── main.go             # Application entry point
│   └── go.mod              # Go dependencies
├── frontend/               # React dashboard
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── services/       # API services
│   │   ├── types/          # TypeScript types
│   │   └── App.tsx         # Main app component
│   ├── public/             # Static assets
│   └── package.json        # Node.js dependencies
├── config/                 # Configuration files
├── scripts/                # Utility scripts
└── README.md               # This file
```

## 🔒 Security Features

- **IP Validation**: Only private IP ranges are allowed as targets
- **Request Logging**: All requests are logged for monitoring
- **CORS Protection**: Frontend-backend communication is properly secured
- **Header Filtering**: Hop-by-hop headers are properly handled
- **Timeout Protection**: 30-second request timeout prevents hanging

## 🚀 Production Deployment

### 1. Build for Production

```bash
# Build Go binary
cd backend
go build -o bin/lan-relay .

# Build React app
cd ../frontend
npm run build
```

### 2. Run Production Server

```bash
cd backend
ENVIRONMENT=production ./bin/lan-relay
```

### 3. Serve Frontend

You can serve the React build files using any web server (nginx, Apache, etc.) or serve them from the Go server by adding static file handling.

## 🐛 Troubleshooting

### Common Issues

**Backend won't start:**
- Check if port 8080 is available
- Ensure Go dependencies are installed: `go mod tidy`
- Check database permissions for SQLite file creation

**Frontend can't connect to backend:**
- Verify backend is running on port 8080
- Check CORS configuration in backend
- Ensure `REACT_APP_API_URL` is set correctly

**Proxy requests failing:**
- Verify target IP is in private range (192.168.x.x, 10.x.x.x, 172.16-31.x.x)
- Check if target service is actually running
- Verify network connectivity from relay server to target

**ngrok issues:**
- Install ngrok: `brew install ngrok` (macOS) or download from website
- Authenticate: `ngrok authtoken YOUR_TOKEN`
- Check firewall settings

### Debug Mode

Enable debug logging:

```bash
LOG_LEVEL=debug go run .
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make changes and test thoroughly
4. Commit: `git commit -m "Add feature-name"`
5. Push: `git push origin feature-name`
6. Create a Pull Request

## 📄 License

This project is licensed under the MIT License. See LICENSE file for details.

## 🔮 Future Enhancements

- [ ] ngrok API integration for automated tunnel management
- [ ] Request/response body logging (optional)
- [ ] Rate limiting and throttling
- [ ] User authentication and access control
- [ ] WebSocket proxy support
- [ ] Docker containerization
- [ ] Metrics and analytics
- [ ] Custom domain support
- [ ] Load balancing for multiple targets

## 📞 Support

If you encounter any issues or have questions:

1. Check the troubleshooting section above
2. Search existing issues in the repository
3. Create a new issue with detailed information
4. Include logs, configuration, and steps to reproduce

---

**Happy relaying! 🛰️** 