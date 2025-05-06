package internals

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/iosxe-yosemite/IOS-XE-Copilot/internals/cisco"
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

func (c *Client) Eula() bool {
	eula := `By installing or using any Cisco Copilot feature within IOS XE, you agree to all the terms outlined below. If you do not agree to these terms, do not install or use any Cisco Copilot feature.

When you use any Cisco Copilot feature with an API key, you acknowledge that Cisco can terminate the API key at any time and shut down the Cisco Copilot feature. Cisco reserves the right to shut down any product feature electronically or by any available means. While Cisco may provide alerts or notifications regarding such shutdowns, it is your responsibility to monitor your usage and ensure your systems are prepared for any shutdown.
	
Cisco will not be liable for any damages resulting from the use of Cisco Copilot features or from their shutdown, including direct, indirect, special, or consequential damages. By clicking 'accept' or typing 'yes,' you confirm that you have read and agree to all the terms stated here. If you do not agree to these terms, do not proceed with installing or using any Cisco Copilot feature.`
	fmt.Println("Welcome to the Cisco IOS XE Copilot. Please read the End User License Agreement (EULA) below:")
	fmt.Println("--------------------------------------------------------------")
	fmt.Println("End User License Agreement (EULA)")
	fmt.Println("--------------------------------------------------------------")
	fmt.Print(eula)
	fmt.Println("\n--------------------------------------------------------------")
	fmt.Println("\nDo you accept the terms and conditions of the EULA? (yes/no)")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	userInput := scanner.Text()

	// Check if the user accepted the EULA
	if userInput == "yes" {
		fmt.Println("Thank you for accepting the EULA. You can now proceed to use the software.")
		// Proceed with the rest of the program...
		return true
	} else {
		fmt.Println("You must accept the EULA to use the software.")
		return false
	}
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
	if c.Eula() {
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
		fmt.Println("COPILOT configured")
	}

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
