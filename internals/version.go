package internals

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Function checks the latest version defined in cloud.
// Compares it with the one on local device. If it differs it triggers the download from the remote repository.
func (c *Client) CheckVersion() {
	var latestVersion string
	url := fmt.Sprintf("%vlatest.txt", c.SoftwareURL)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Error fetching latest version")
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
	//Version checking
	if string(c.Version) != string(latestVersion) {
		fmt.Printf("Installed: %s\tLatest: %s\n", c.Version, latestVersion)
		fmt.Println("Upgrading...")
		//Download triggered
		copilotFile := fmt.Sprintf("%v%v/aixedge.built", c.SoftwareURL, latestVersion)
		c.download(copilotFile, "aixedge.built")
		ciscoCmd := fmt.Sprintf("%v%v/cmd.py", c.SoftwareURL, latestVersion)
		c.download(ciscoCmd, "cmd.py")
		fmt.Println("AIXEdge Upgraded:", latestVersion)
	} else {
		fmt.Println("No upgrade needed")
		fmt.Printf("Installed: %s\tLatest: %s\n", c.Version, latestVersion)
	}

}

func (c *Client) ShowVersion() {
	fmt.Printf("Current version: %s\n", c.Version)
}

func (c *Client) download(url, destination string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	c.copyFile(destination, resp.Body)
}

func (*Client) copyFile(dest string, body io.Reader) {
	file, err := os.Create(dest)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		panic(err)
	}
}
