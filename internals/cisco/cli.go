package cisco

import (
	"errors"
	"os/exec"
	"strings"
)

type IOSXE struct {
}

// Interactiom between app and python script cmd.py is handled here.
func (c *IOSXE) Device() (string, string, string, string, error) {
	cmd := exec.Command("python3", "cmd.py", "-d")
	out, err := cmd.Output()
	if err != nil {
		return "", "", "", "", errors.New("Missing/Corrupted dependency")
	}
	data := strings.Split(string(out), ",")
	return data[0], data[1], data[2], strings.Trim(data[3], "\n"), nil
}

func (c *IOSXE) Inventory() (string, error) {
	cmd := exec.Command("python3", "cmd.py", "-i")
	out, err := cmd.Output()
	if err != nil {
		return "", errors.New("Missing/Corrupted dependency")
	}
	return string(out), nil
}

func (c *IOSXE) Command(command string) (string, error) {
	cmd := exec.Command("python3", "cmd.py", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return "", errors.New("Bad")
	}
	return string(out), nil
}
