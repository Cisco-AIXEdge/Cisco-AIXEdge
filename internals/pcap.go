package internals

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/providers"
)

type ChatGPTRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Protocol header structures
type EthernetHeader struct {
	DstMAC [6]byte
	SrcMAC [6]byte
	Type   uint16
}

type IPv4Header struct {
	VersionIHL     uint8
	TOS            uint8
	TotalLength    uint16
	Identification uint16
	FlagsFragment  uint16
	TTL            uint8
	Protocol       uint8
	Checksum       uint16
	SrcIP          [4]byte
	DstIP          [4]byte
}

type TCPHeader struct {
	SrcPort    uint16
	DstPort    uint16
	SeqNumber  uint32
	AckNumber  uint32
	DataOffset uint8
	Flags      uint8
	Window     uint16
	Checksum   uint16
	UrgPointer uint16
}

type UDPHeader struct {
	SrcPort  uint16
	DstPort  uint16
	Length   uint16
	Checksum uint16
}

type ICMPHeader struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	Rest     uint32
}

// PCAP-NG specific structures
type BlockType uint32

const (
	SectionHeaderBlock  BlockType = 0x0A0D0D0A
	InterfaceDescBlock  BlockType = 0x00000001
	EnhancedPacketBlock BlockType = 0x00000006
	SimplePacketBlock   BlockType = 0x00000003
	PacketBlock         BlockType = 0x00000002 // (obsolete)
)

type CaptureInfo struct {
	Timestamp     time.Time
	CaptureLength uint32
	Length        uint32
	InterfaceID   uint32
}

type InterfaceInfo struct {
	LinkType    uint16
	SnapLen     uint32
	Name        string
	Description string
	TimeRes     uint8
}

var file_path string

func (c *Client) SetPcapFile(path string) {
	file_path = "/flash/guest-share/" + path
}

func (c *Client) Pcap(question string) {
	cfg, err := c.configRead()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		return
	}

	engine := providers.Engine{
		Provider: c.Engine,
		Version:  c.EngineVERSION,
	}

	a := providers.Client{
		API:    cfg.Apikey,
		Engine: engine,
	}

	// Process the PCAP file
	summary, err := processPcap(file_path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing PCAP file: %v\n", err)
		return
	}

	// Prepare the prompt for ChatGPT
	prompt := fmt.Sprintf(`
You have a PCAP summary below please answer the following question: %s

PCAP Summary:
%s
`, question, summary)

	// Send to ChatGPT
	response, err := sendToChatGPT(prompt, a.API)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending to ChatGPT: %v\n", err)
		return
	}

	// Display the response
	fmt.Println(response)
}

// Process the PCAP file and build a summary
func processPcap(filePath string) (string, error) {
	var packetSummary strings.Builder

	// Open the PCAP file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening PCAP file: %v", err)
	}
	defer file.Close()

	// Check file format by reading the first 4 bytes
	magicBytes := make([]byte, 4)
	if _, err := file.Read(magicBytes); err != nil {
		return "", fmt.Errorf("error reading file header: %v", err)
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("error resetting file pointer: %v", err)
	}

	// Determine file format and process accordingly
	if bytes.Equal(magicBytes, []byte{0x0A, 0x0D, 0x0D, 0x0A}) {
		// PCAP-NG format
		packetSummary.WriteString("Processing PCAP-NG format file\n\n")
		return processPcapNg(file)
	} else {
		// Legacy PCAP format
		packetSummary.WriteString("Processing PCAP format file\n\n")
		return processPcapLegacy(file)
	}
}

// Process PCAP-NG format
func processPcapNg(file *os.File) (string, error) {
	var packetSummary strings.Builder
	var interfaces []InterfaceInfo

	packetCount := 0
	maxPackets := 100

	// Read blocks until we reach the end or max packets
	for packetCount < maxPackets {
		// Read block type
		var blockType BlockType
		if err := binary.Read(file, binary.LittleEndian, &blockType); err != nil {
			if err == io.EOF {
				break
			}
			return packetSummary.String(), fmt.Errorf("error reading block type: %v", err)
		}

		// Read block total length
		var totalLength uint32
		if err := binary.Read(file, binary.LittleEndian, &totalLength); err != nil {
			return packetSummary.String(), fmt.Errorf("error reading block length: %v", err)
		}

		// Calculate data length (minus header and footer)
		dataLength := totalLength - 12 // 4 (type) + 4 (length) + 4 (length at end)

		// Read block data
		blockData := make([]byte, dataLength)
		if _, err := io.ReadFull(file, blockData); err != nil {
			return packetSummary.String(), fmt.Errorf("error reading block data: %v", err)
		}

		// Skip the length field at the end of the block
		var endLength uint32
		if err := binary.Read(file, binary.LittleEndian, &endLength); err != nil {
			return packetSummary.String(), fmt.Errorf("error reading end block length: %v", err)
		}

		// Process block based on type
		switch blockType {
		case SectionHeaderBlock:
			// Process section header (just skip for now)
			continue

		case InterfaceDescBlock:
			// Parse interface description
			if len(blockData) < 8 {
				continue
			}

			iface := InterfaceInfo{
				LinkType: binary.LittleEndian.Uint16(blockData[0:2]),
				SnapLen:  binary.LittleEndian.Uint32(blockData[4:8]),
			}

			// Process options to get name and description
			optOffset := uint32(8)
			for optOffset < dataLength {
				if optOffset+4 > dataLength {
					break
				}

				optCode := binary.LittleEndian.Uint16(blockData[optOffset : optOffset+2])
				optLen := binary.LittleEndian.Uint16(blockData[optOffset+2 : optOffset+4])

				if optOffset+4+uint32(optLen) > dataLength {
					break
				}

				// Option code 2 is for interface name
				if optCode == 2 && optLen > 0 {
					iface.Name = string(blockData[optOffset+4 : optOffset+4+uint32(optLen)-1]) // Remove null terminator
				}

				// Option code 3 is for interface description
				if optCode == 3 && optLen > 0 {
					iface.Description = string(blockData[optOffset+4 : optOffset+4+uint32(optLen)-1]) // Remove null terminator
				}

				// Option code 9 is for time resolution
				if optCode == 9 && optLen == 1 {
					iface.TimeRes = blockData[optOffset+4]
				}

				// Move to next option (with padding)
				paddedLen := (uint32(optLen) + 3) & ^uint32(3) // This is equivalent to rounding up to a multiple of 4
				optOffset += 4 + paddedLen
			}

			interfaces = append(interfaces, iface)

		case EnhancedPacketBlock:
			if len(blockData) < 20 {
				continue
			}

			// Parse enhanced packet block
			interfaceID := binary.LittleEndian.Uint32(blockData[0:4])
			timestampHigh := binary.LittleEndian.Uint32(blockData[4:8])
			timestampLow := binary.LittleEndian.Uint32(blockData[8:12])
			captureLen := binary.LittleEndian.Uint32(blockData[12:16])
			packetLen := binary.LittleEndian.Uint32(blockData[16:20])

			// Create packet data
			var packetData []byte
			if 20+captureLen <= uint32(len(blockData)) {
				packetData = blockData[20 : 20+captureLen]
			} else {
				continue
			}

			// Create timestamp (assuming microsecond resolution)
			timestamp := time.Unix(int64(timestampHigh), int64(timestampLow)*1000)

			// Add packet info to summary
			packetSummary.WriteString(fmt.Sprintf("Packet %d:\n", packetCount+1))
			packetSummary.WriteString(fmt.Sprintf("  Time: %s\n", timestamp))
			packetSummary.WriteString(fmt.Sprintf("  Length: %d bytes\n", packetLen))
			packetSummary.WriteString(fmt.Sprintf("  Interface ID: %d\n", interfaceID))

			// Parse packet data
			parsePacketLayers(packetData, &packetSummary)

			packetSummary.WriteString("\n")
			packetCount++

		case SimplePacketBlock:
			if len(blockData) < 4 {
				continue
			}

			// Parse simple packet block
			packetLen := binary.LittleEndian.Uint32(blockData[0:4])

			// Create packet data
			var packetData []byte
			if 4+packetLen <= uint32(len(blockData)) {
				packetData = blockData[4 : 4+packetLen]
			} else {
				continue
			}

			// Add packet info to summary
			packetSummary.WriteString(fmt.Sprintf("Packet %d:\n", packetCount+1))
			packetSummary.WriteString(fmt.Sprintf("  Length: %d bytes\n", packetLen))

			// Parse packet data
			parsePacketLayers(packetData, &packetSummary)

			packetSummary.WriteString("\n")
			packetCount++

		case PacketBlock:
			// Legacy packet block (obsolete)
			if len(blockData) < 16 {
				continue
			}

			// Parse packet block
			interfaceID := binary.LittleEndian.Uint16(blockData[0:2])
			// Skip 2 bytes (drops count)
			timestampHigh := binary.LittleEndian.Uint32(blockData[4:8])
			timestampLow := binary.LittleEndian.Uint32(blockData[8:12])
			captureLen := binary.LittleEndian.Uint32(blockData[12:16])
			packetLen := binary.LittleEndian.Uint32(blockData[16:20])

			// Create packet data
			var packetData []byte
			if 20+captureLen <= uint32(len(blockData)) {
				packetData = blockData[20 : 20+captureLen]
			} else {
				continue
			}

			// Create timestamp
			timestamp := time.Unix(int64(timestampHigh), int64(timestampLow)*1000)

			// Add packet info to summary
			packetSummary.WriteString(fmt.Sprintf("Packet %d:\n", packetCount+1))
			packetSummary.WriteString(fmt.Sprintf("  Time: %s\n", timestamp))
			packetSummary.WriteString(fmt.Sprintf("  Length: %d bytes\n", packetLen))
			packetSummary.WriteString(fmt.Sprintf("  Interface ID: %d\n", interfaceID))

			// Parse packet data
			parsePacketLayers(packetData, &packetSummary)

			packetSummary.WriteString("\n")
			packetCount++
		}
	}

	return packetSummary.String(), nil
}

// Process legacy PCAP format
func processPcapLegacy(file *os.File) (string, error) {
	var packetSummary strings.Builder

	// Read global header (24 bytes)
	header := make([]byte, 24)
	if _, err := io.ReadFull(file, header); err != nil {
		return "", fmt.Errorf("error reading PCAP header: %v", err)
	}

	// Determine endianness
	var endian binary.ByteOrder = binary.LittleEndian
	magic := binary.LittleEndian.Uint32(header[0:4])
	if magic == 0xa1b2c3d4 || magic == 0xa1b23c4d {
		endian = binary.BigEndian
	} else if magic != 0xd4c3b2a1 && magic != 0x4d3cb2a1 {
		return "", fmt.Errorf("invalid PCAP magic number: %x", magic)
	}

	// Process up to 100 packets
	packetCount := 0
	for packetCount < 100 {
		// Read packet header (16 bytes)
		packetHeader := make([]byte, 16)
		if _, err := io.ReadFull(file, packetHeader); err != nil {
			if err == io.EOF {
				break
			}
			return packetSummary.String(), fmt.Errorf("error reading packet header: %v", err)
		}

		// Parse packet header
		tsSec := endian.Uint32(packetHeader[0:4])
		tsUsec := endian.Uint32(packetHeader[4:8])
		inclLen := endian.Uint32(packetHeader[8:12])
		origLen := endian.Uint32(packetHeader[12:16])

		// Read packet data
		packetData := make([]byte, inclLen)
		if _, err := io.ReadFull(file, packetData); err != nil {
			return packetSummary.String(), fmt.Errorf("error reading packet data: %v", err)
		}

		// Create timestamp
		timestamp := time.Unix(int64(tsSec), int64(tsUsec)*1000)

		// Add packet info to summary
		packetSummary.WriteString(fmt.Sprintf("Packet %d:\n", packetCount+1))
		packetSummary.WriteString(fmt.Sprintf("  Time: %s\n", timestamp))
		packetSummary.WriteString(fmt.Sprintf("  Length: %d bytes\n", origLen))

		// Parse packet layers
		parsePacketLayers(packetData, &packetSummary)

		packetSummary.WriteString("\n")
		packetCount++
	}

	return packetSummary.String(), nil
}

// Parse packet layers and add to summary
func parsePacketLayers(data []byte, summary *strings.Builder) {
	// Parse Ethernet header (if enough data)
	if len(data) < 14 {
		summary.WriteString("  Incomplete packet (too short for Ethernet header)\n")
		return
	}

	// Extract Ethernet header
	ethHeader := parseEthernetHeader(data[:14])
	summary.WriteString("  Layer: Ethernet\n")
	summary.WriteString(fmt.Sprintf("    Ethernet: %s -> %s\n",
		formatMAC(ethHeader.SrcMAC[:]), formatMAC(ethHeader.DstMAC[:])))

	// Determine next layer type based on EtherType
	etherType := binary.BigEndian.Uint16(data[12:14])

	// Handle IPv4
	if etherType == 0x0800 && len(data) >= 34 {
		ipOffset := 14
		ipHeaderLen := int((data[ipOffset] & 0x0F) * 4)
		if len(data) < ipOffset+ipHeaderLen {
			summary.WriteString("  Incomplete IPv4 header\n")
			return
		}

		// Parse IPv4 header
		ipHeader := parseIPv4Header(data[ipOffset : ipOffset+ipHeaderLen])
		summary.WriteString("  Layer: IPv4\n")
		summary.WriteString(fmt.Sprintf("    IPv4: %s -> %s\n",
			net.IP(ipHeader.SrcIP[:]).String(),
			net.IP(ipHeader.DstIP[:]).String()))
		summary.WriteString(fmt.Sprintf("    Protocol: %d\n", ipHeader.Protocol))

		// Get transport layer offset
		transportOffset := ipOffset + ipHeaderLen

		// Handle TCP
		if ipHeader.Protocol == 6 && len(data) >= transportOffset+20 {
			tcpHeader := parseTCPHeader(data[transportOffset : transportOffset+20])
			summary.WriteString("  Layer: TCP\n")
			summary.WriteString(fmt.Sprintf("    TCP: %d -> %d\n",
				tcpHeader.SrcPort, tcpHeader.DstPort))
			summary.WriteString(fmt.Sprintf("    Flags: %s\n",
				formatTCPFlags(tcpHeader.Flags)))
			summary.WriteString(fmt.Sprintf("    Seq: %d, Ack: %d\n",
				tcpHeader.SeqNumber, tcpHeader.AckNumber))

			// Calculate payload offset
			dataOffset := int((tcpHeader.DataOffset >> 4) * 4)
			payloadOffset := transportOffset + dataOffset

			// Extract application data if available
			if len(data) > payloadOffset && len(data[payloadOffset:]) > 0 {
				payload := data[payloadOffset:]
				if len(payload) > 100 {
					payload = payload[:100]
				}
				summary.WriteString("  Application Data:\n")
				summary.WriteString(fmt.Sprintf("    %s\n", formatPayload(payload)))
			}
		}

		// Handle UDP
		if ipHeader.Protocol == 17 && len(data) >= transportOffset+8 {
			udpHeader := parseUDPHeader(data[transportOffset : transportOffset+8])
			summary.WriteString("  Layer: UDP\n")
			summary.WriteString(fmt.Sprintf("    UDP: %d -> %d\n",
				udpHeader.SrcPort, udpHeader.DstPort))
			summary.WriteString(fmt.Sprintf("    Length: %d\n", udpHeader.Length))

			// Extract application data
			payloadOffset := transportOffset + 8
			if len(data) > payloadOffset && len(data[payloadOffset:]) > 0 {
				payload := data[payloadOffset:]
				if len(payload) > 100 {
					payload = payload[:100]
				}
				summary.WriteString("  Application Data:\n")
				summary.WriteString(fmt.Sprintf("    %s\n", formatPayload(payload)))
			}
		}

		// Handle ICMP
		if ipHeader.Protocol == 1 && len(data) >= transportOffset+8 {
			icmpHeader := parseICMPHeader(data[transportOffset : transportOffset+8])
			summary.WriteString("  Layer: ICMP\n")
			summary.WriteString(fmt.Sprintf("    ICMP Type: %d, Code: %d\n",
				icmpHeader.Type, icmpHeader.Code))
		}
	}

	// Handle IPv6 (basic support)
	if etherType == 0x86DD && len(data) >= 54 {
		summary.WriteString("  Layer: IPv6\n")
		summary.WriteString("    IPv6 details not fully parsed in this version\n")
	}
}

// Parse Ethernet header
func parseEthernetHeader(data []byte) EthernetHeader {
	var header EthernetHeader
	copy(header.DstMAC[:], data[0:6])
	copy(header.SrcMAC[:], data[6:12])
	header.Type = binary.BigEndian.Uint16(data[12:14])
	return header
}

// Parse IPv4 header
func parseIPv4Header(data []byte) IPv4Header {
	var header IPv4Header
	header.VersionIHL = data[0]
	header.TOS = data[1]
	header.TotalLength = binary.BigEndian.Uint16(data[2:4])
	header.Identification = binary.BigEndian.Uint16(data[4:6])
	header.FlagsFragment = binary.BigEndian.Uint16(data[6:8])
	header.TTL = data[8]
	header.Protocol = data[9]
	header.Checksum = binary.BigEndian.Uint16(data[10:12])
	copy(header.SrcIP[:], data[12:16])
	copy(header.DstIP[:], data[16:20])
	return header
}

// Parse TCP header
func parseTCPHeader(data []byte) TCPHeader {
	var header TCPHeader
	header.SrcPort = binary.BigEndian.Uint16(data[0:2])
	header.DstPort = binary.BigEndian.Uint16(data[2:4])
	header.SeqNumber = binary.BigEndian.Uint32(data[4:8])
	header.AckNumber = binary.BigEndian.Uint32(data[8:12])
	header.DataOffset = data[12]
	header.Flags = data[13]
	header.Window = binary.BigEndian.Uint16(data[14:16])
	header.Checksum = binary.BigEndian.Uint16(data[16:18])
	header.UrgPointer = binary.BigEndian.Uint16(data[18:20])
	return header
}

// Parse UDP header
func parseUDPHeader(data []byte) UDPHeader {
	var header UDPHeader
	header.SrcPort = binary.BigEndian.Uint16(data[0:2])
	header.DstPort = binary.BigEndian.Uint16(data[2:4])
	header.Length = binary.BigEndian.Uint16(data[4:6])
	header.Checksum = binary.BigEndian.Uint16(data[6:8])
	return header
}

// Parse ICMP header
func parseICMPHeader(data []byte) ICMPHeader {
	var header ICMPHeader
	header.Type = data[0]
	header.Code = data[1]
	header.Checksum = binary.BigEndian.Uint16(data[2:4])
	header.Rest = binary.BigEndian.Uint32(data[4:8])
	return header
}

// Format MAC address as string
func formatMAC(mac []byte) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

// Format TCP flags as string
func formatTCPFlags(flags uint8) string {
	var flagStrings []string

	if (flags & 0x01) != 0 {
		flagStrings = append(flagStrings, "FIN")
	}
	if (flags & 0x02) != 0 {
		flagStrings = append(flagStrings, "SYN")
	}
	if (flags & 0x04) != 0 {
		flagStrings = append(flagStrings, "RST")
	}
	if (flags & 0x08) != 0 {
		flagStrings = append(flagStrings, "PSH")
	}
	if (flags & 0x10) != 0 {
		flagStrings = append(flagStrings, "ACK")
	}
	if (flags & 0x20) != 0 {
		flagStrings = append(flagStrings, "URG")
	}
	if (flags & 0x40) != 0 {
		flagStrings = append(flagStrings, "ECE")
	}
	if (flags & 0x80) != 0 {
		flagStrings = append(flagStrings, "CWR")
	}

	if len(flagStrings) == 0 {
		return "None"
	}
	return strings.Join(flagStrings, ", ")
}

// Format binary payload
func formatPayload(payload []byte) string {
	// First try to represent as ASCII
	isASCII := true
	for _, b := range payload {
		if b < 32 || b > 126 {
			isASCII = false
			break
		}
	}

	if isASCII {
		return string(payload)
	}

	// Otherwise return a hex dump
	var hexStr strings.Builder
	for i, b := range payload {
		if i > 0 && i%16 == 0 {
			hexStr.WriteString("\n    ")
		} else if i > 0 {
			hexStr.WriteString(" ")
		}
		hexStr.WriteString(fmt.Sprintf("%02x", b))
	}
	return hexStr.String()
}

// ChatGPT communication function
func sendToChatGPT(prompt string, apiKey string) (string, error) {
	// Create the request payload
	reqBody := ChatGPTRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Convert the request to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse the response
	var chatResponse ChatGPTResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return "", err
	}

	// Check for errors
	if chatResponse.Error != nil {
		return "", fmt.Errorf("ChatGPT API error: %s", chatResponse.Error.Message)
	}

	// Check if we got any choices back
	if len(chatResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from ChatGPT")
	}

	return chatResponse.Choices[0].Message.Content, nil
}
