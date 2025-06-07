# Cisco-AIXEdge

<img src="https://github.com/user-attachments/assets/3ad32655-d8e6-47ba-9bf8-0c3f2564912b" width="300" height="300">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Cisco-AIXEdge/Cisco-AIXEdge)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg)](https://aixedge.readthedocs.io)

## 🤖 AI-Powered Assistant for Cisco Network Devices

Cisco AIXEdge is an open-source AI assistant that simplifies management and troubleshooting of Cisco IOS-XE and NX-OS devices. It leverages natural language processing to interpret commands, automate routine tasks, and provide contextual assistance to network engineers.

## ✨ Key Features

- **Natural Language Interface**: Interact with your Cisco devices using plain English commands
- **Multi-Platform Support**: Works with both IOS-XE and NX-OS environments
- **Multi-Language Translation**: Provides suggestions in your native language
- **Intelligent Troubleshooting**: Automatically diagnoses common network issues
- **Configuration Assistance**: Provides suggestions and validates configurations

## 🚀 Getting Started

### Prerequisites

- go version go1.22.4
- Network access to Cisco devices (IOS-XE or NX-OS)
- Bring your Own API key (OpenAI / Google Gemini)

### Build 
``` bash
env GOOS=linux GOARCH=386 go build -o aixedge.built ./aixedge.go
```

### Installation

```bash
Switch#copy http://bootstrap.cisco-aixedge.com/aixedge-init.cfg running-config
```

### Basic Usage

```bash
SW#aixedge-help

AI assitant for Cisco AIXEdge products.
        Arguments:
        aixedge-help                                                                    Presents options to run AI assistant
	aixedge-chat									Chat with the AI assitant
        aixedge <query>                                                                 Queries adressed to AI Assistant
        aixedge <show command> @ <query to AI assistant>                                AI Assistant helps with command's output
        aixedge-upgrade                                                                 Upgrades the AI Assistant to the latest version
        aixedge-cfg <API_KEY>                                                           Initial config of the script; Adds the OpenAI API key;
        aixedge-init                                                                    Initialization of AI assistant
        aixedge-uninstall                                                               Uninstall the AI assistant
        aixedge-version                                                                 Shows installed version
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
      
## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments
- The network engineering community for valuable feedback

## 📬 Contact

For questions, feedback, or support:
- GitHub Issues: [https://github.com/cisco/aixedge/issues](https://github.com/cisco/aixedge/issues)
- Email: scozma@cisco.com, rcsapo@cisco.com

---

*Cisco AIXEDGE is not officially endorsed by Cisco Systems, Inc. This is a community-driven open-source project.*
