# LAN Relay CLI Tool

🛰️ **LAN Relay** is a powerful command-line tool that allows you to expose and proxy local network services through secure tunnels with an embedded web dashboard.

## 🚀 Quick Start

### One-Line Installation

```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/local_router/main/install.sh | bash
```

Or download and inspect the script first:
```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/local_router/main/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

### Manual Installation

1. **Prerequisites**: Go 1.19+ and Node.js 16+
2. **Clone and build**:
   ```bash
   git clone https://github.com/yourusername/local_router.git
   cd local_router
   make install
   ```

## 📖 Usage

### Start the Service
```bash
# Start with default settings (port 8080)
lan-relay start

# Start on a specific port
lan-relay start --port 9090

# Start in daemon mode (background)
lan-relay start --daemon --port 8080
```

### Check Status
```bash
# Check if service is running
lan-relay status

# Check status on specific port
lan-relay status --port 9090
```

### Other Commands
```bash
# Show version
lan-relay version

# Show help
lan-relay --help
lan-relay start --help
```

## 🎯 Features

- **🖥️ Web Dashboard**: Modern React-based interface
- **🔒 Secure Proxy**: Proxy requests to internal network services
- **🌐 Ngrok Integration**: Expose services to the internet
- **📊 Request Logging**: Monitor all proxy requests
- **⚙️ Settings Management**: Configure via web interface
- **🛡️ Security**: Only allows connections to private IP ranges

## 🌐 Web Dashboard

Once started, access the dashboard at:
- **Local**: http://localhost:8080 (or your specified port)
- **Network**: http://YOUR_IP:8080

### Dashboard Features
- **Status Overview**: Server uptime, request counts, active connections
- **Request Logs**: Real-time view of all proxy requests
- **Settings**: Configure ngrok tokens, domains, and other options
- **Proxy Management**: Easy proxy setup with URL generation

## 🔗 Proxy Usage

Access any device on your local network through the relay:

```bash
# Access a local web server
http://localhost:8080/proxy/192.168.1.100:3000

# Access a local API
http://localhost:8080/proxy/192.168.1.50:8000/api/users

# Access any HTTP service
http://localhost:8080/proxy/[IP]:[PORT]/[PATH]
```

### Supported Services
- Web applications (React, Vue, Angular)
- API servers (REST, GraphQL)
- Development servers
- IoT devices with web interfaces
- Any HTTP/HTTPS service on your network

## 🛠️ Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/yourusername/local_router.git
cd local_router

# Build everything
make build

# Build for specific platforms
make package

# Run tests
make test

# Clean build artifacts
make clean
```

### Available Make Targets
- `make build` - Build for current platform
- `make package` - Build for all platforms (Linux, macOS, Windows)
- `make install` - Install locally
- `make uninstall` - Remove installation
- `make test` - Run all tests
- `make clean` - Clean build artifacts
- `make help` - Show all targets

### Project Structure
```
local_router/
├── backend/           # Go CLI and server
│   ├── cmd/          # CLI commands
│   ├── internal/     # Internal packages
│   └── main.go       # Entry point
├── frontend/         # React dashboard
├── scripts/          # Build and dev scripts
├── Makefile          # Build automation
└── install.sh        # Installation script
```

## 🔧 Configuration

### Environment Variables
- `PORT` - Server port (default: 8080)
- `DATABASE_PATH` - SQLite database path (default: relay.db)
- `LOG_LEVEL` - Logging level (default: info)
- `ENVIRONMENT` - Environment mode (default: development)
- `NGROK_TOKEN` - Ngrok authentication token
- `NGROK_DOMAIN` - Custom ngrok domain

### Configuration File
Create a `.env` file in your working directory:
```bash
PORT=8080
LOG_LEVEL=debug
NGROK_TOKEN=your_ngrok_token_here
```

## 🔐 Security

- **IP Filtering**: Only private IP ranges are accessible (192.168.x.x, 10.x.x.x, 172.16-31.x.x)
- **CORS Protection**: Configurable cross-origin policies
- **Request Logging**: All proxy requests are logged for monitoring
- **No External Access**: Only local network services are proxied

## 📦 Distribution

### Pre-built Binaries
Download platform-specific binaries from [Releases](https://github.com/yourusername/local_router/releases):

- `lan-relay-1.0.0-linux-amd64` - Linux x64
- `lan-relay-1.0.0-linux-arm64` - Linux ARM64
- `lan-relay-1.0.0-darwin-amd64` - macOS Intel
- `lan-relay-1.0.0-darwin-arm64` - macOS Apple Silicon
- `lan-relay-1.0.0-windows-amd64.exe` - Windows x64

### Package Managers
```bash
# Homebrew (macOS/Linux) - Coming Soon
brew install lan-relay

# Apt (Ubuntu/Debian) - Coming Soon
sudo apt install lan-relay

# Chocolatey (Windows) - Coming Soon
choco install lan-relay
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/local_router/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/local_router/discussions)
- **Wiki**: [Documentation Wiki](https://github.com/yourusername/local_router/wiki)

## 🙏 Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [React](https://reactjs.org/) - Frontend framework
- [Ngrok](https://ngrok.com/) - Secure tunneling service 