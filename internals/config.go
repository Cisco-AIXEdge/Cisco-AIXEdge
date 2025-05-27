package internals

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/cisco"
)

type configFile struct {
	Apikey       string `json:"api_key"`
	PID          string `json:"pid,omitempty"`
	SerialNumber string `json:"serialnumber,omitempty"`
	Eula         bool   `json:"eula"`
	SwVer        string `json:"swVer"`
	Platform     string `json:"platform"`
}

// Function opens the configuration file
func (c *Client) configRead() (configFile, error) {
	file, err := os.Open(".config.json")
	if err != nil {
		return configFile{}, errors.New("No API Key")
	}
	defer file.Close()
	return c.configJSON(file), nil
}

// Function writes SN, PN, API key into .config.json
// It triggers cmd.py to get SN, PN from device
// With these and API key it writes the configuration json locally on the device.
func (c *Client) ConfigWrite(api string) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Println("Missing/Corrupted dependecy! Python module is not able to collect SN/PID of device")
		}
	}()

	// Download python module
	c.getCLI("cmd.py")

	var cfgJson string
	var err error
	cfg := &configFile{}
	cfg.Apikey = api
	cfg.Eula = true
	iosxe := cisco.IOSXE{}
	cfg.PID, cfg.SerialNumber, cfg.SwVer, cfg.Platform, err = iosxe.Device()
	if err != nil {
		panic(err)
	}

	if _, err := os.Create(".config.json"); err != nil {
		panic(err)
	}
	if b, err := json.MarshalIndent(cfg, "", "\t"); err == nil {
		cfgJson = string(b)
	} else {
		panic(err)
	}
	if err := os.WriteFile(".config.json", []byte(cfgJson), 0644); err != nil {
		panic(err)
	}
	fmt.Println("AIXEdge configured")

}

// Function is JSON decoder
func (c *Client) configJSON(file *os.File) configFile {
	decoder := json.NewDecoder(file)
	cfg := configFile{}
	err := decoder.Decode(&cfg)
	if err != nil {
		fmt.Print(err)
	}
	return cfg
}

// Functions downloads cmd.py
func (c *Client) getCLI(file string) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Println("Error downloading python module")
		}
	}()
	var latestVersion string
	url := fmt.Sprintf("%vlatest.txt", c.SoftwareURL)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching latest version:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		latestVersion = strings.TrimSpace(string(body))
	}
	url = fmt.Sprintf("%v%v/%v", c.SoftwareURL, latestVersion, file)
	c.download(url, file)
}
