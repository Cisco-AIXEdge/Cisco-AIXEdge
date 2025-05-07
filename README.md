# Cisco-AIXEdge

<img src="https://github.com/user-attachments/assets/3ad32655-d8e6-47ba-9bf8-0c3f2564912b" width="300" height="300">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/:user/:repo)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg)](https://aixedge.readthedocs.io)

## 🤖 AI-Powered Assistant for Cisco Network Devices

Cisco AIXEDGE is an open-source AI assistant that simplifies management and troubleshooting of Cisco IOS-XE and NX-OS devices. It leverages natural language processing to interpret commands, automate routine tasks, and provide contextual assistance to network engineers.

## ✨ Key Features

- **Natural Language Interface**: Interact with your Cisco devices using plain English commands
- **Multi-Platform Support**: Works with both IOS-XE and NX-OS environments
- **Multi-Language Translation**: Provides suggestions in your native language
- **Intelligent Troubleshooting**: Automatically diagnoses common network issues
- **Configuration Assistance**: Provides suggestions and validates configurations

## 🚀 Getting Started

### Prerequisites

- go version go1.22.4
- Network access to Cisco devices
- Bring your Own API key (OpenAI / Google Gemini)

### Build 
``` bash
env GOOS=linux GOARCH=386 go build ./copilot.go
```

### Installation

```bash
# Install via pip
pip install cisco-aixedge

# Or clone and install from source
git clone https://github.com/cisco/aixedge.git
cd aixedge
pip install -e .
```

### Basic Usage

```python
from cisco_aixedge import AIXEdge

# Initialize with your device credentials
assistant = AIXEdge(
    username="admin",
    password="your_password",
    devices=[
        {"ip": "192.168.1.1", "platform": "ios-xe"},
        {"ip": "192.168.1.2", "platform": "nx-os"}
    ]
)

# Use natural language to execute commands
response = assistant.execute("Show me interfaces with errors on device 192.168.1.1")

# Get configuration recommendations
suggestions = assistant.analyze("Check for security vulnerabilities in my ACLs")

# Automate a common task
assistant.automate("Configure OSPF area 0 on all core routers")
```

## 📚 Documentation

For comprehensive documentation, visit [https://aixedge.readthedocs.io](https://aixedge.readthedocs.io)

### Sample Use Cases

- Troubleshooting connectivity issues
- Automating device configurations
- Network performance optimization
- Training junior network engineers

## 🤝 Contributing

We welcome contributions from the community! Please see our [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📊 Project Roadmap

- [ ] TBD
      
## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Cisco DevNet community
- All our open-source contributors
- The network engineering community for valuable feedback

## 📬 Contact

For questions, feedback, or support:
- GitHub Issues: [https://github.com/cisco/aixedge/issues](https://github.com/cisco/aixedge/issues)
- Email: scozma@cisco.com

---

*Cisco AIXEDGE is not officially endorsed by Cisco Systems, Inc. This is a community-driven open-source project.*
