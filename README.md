# Secuchat-CLI v1.0.0

A secure, self-hosted communication tool designed for authorized red team operations and penetration testing activities.

## 🔒 Features

- **Self-hosted**: Complete control over your communications
- **WebSocket-based**: Real-time messaging with low latency  
- **Terms of Service**: Built-in ToS acceptance for operational compliance
- **OPSEC-focused**: Designed with operational security in mind
- **Multi-room support**: PIN-based room isolation
- **Cross-platform**: Works on Windows, Linux, and macOS

## ⚠️ Important Notice

**This tool is exclusively for authorized penetration testing and red team operations.** 
Users must accept comprehensive Terms of Service before accessing the system.

## 🚀 Quick Start

### Prerequisites
- Go 1.19+ 
- Python 3.x
- Git

### Installation

```bash
git clone <your-repo-url>
cd Secuchat-CLI
go mod tidy
```

### Usage

1. **Start the server:**
   ```bash
   go run .
   ```

2. **Accept Terms of Service** when prompted

3. **Access the chat:**
   - Open browser to `http://localhost:8080`
   - Enter a PIN to create/join a room

### Configuration

- **Port**: Set `PORT` environment variable (default: 8080)
- **Terms**: Modify `tos.py` to customize ToS content

## 🏗️ Architecture

- **main.go**: WebSocket server and room management
- **python_integration.go**: ToS integration with Python
- **tos.py**: Terms of Service display and acceptance

## 🔧 Development

```bash
# Run with custom port
PORT=3000 go run .

# Build binary
go build -o secuchat-cli
```

## 📋 Terms of Service

All users must accept comprehensive terms covering:
- Authorized personnel access only
- Engagement scope compliance  
- OPSEC protocols
- Professional conduct standards
- Data handling requirements
- Incident response procedures

## 🛡️ Security Considerations

- Deploy in secure, isolated networks
- Use strong PIN codes for rooms
- Regular security audits recommended
- Follow organizational OPSEC guidelines

## 📄 License

[Add your license here]

## 🤝 Contributing

This is a specialized tool for red team operations. Contributions should maintain focus on authorized security testing use cases.

---

**Version**: 1.0.0  
**Status**: MVP - Ready for authorized red team operations