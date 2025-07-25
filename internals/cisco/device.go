package cisco

import (
	"fmt"
	"os"

	"os/exec"
	"reflect"
	"strings"

	"github.com/chzyer/readline"
	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
)

var CodeBlocks []string
var (
	Rl *readline.Instance
)

type DEVICE struct {
}

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

/////////////////////////////////////////
// THIS SECTION IS FOR TOOL DEFINITION //
/////////////////////////////////////////

// OpenAI tools
var F1_openai = openai.FunctionDefinition{
	Name:        "Show_cdp",
	Description: "Get information about what devices are connected to this device. Has information about neighbouring devices.",
}
var Show_cdp_tool_openai = openai.Tool{
	Type:     openai.ToolTypeFunction,
	Function: F1_openai,
}

var F2_openai = openai.FunctionDefinition{
	Name:        "Show_ip_route",
	Description: "Get information about what IPv4/IPv6 routes are defined. Routes from EIGRP, OSPF, Static routes and default gateway and many others.",
}
var Show_ip_route_tool_openai = openai.Tool{
	Type:     openai.ToolTypeFunction,
	Function: F2_openai,
}
var F3_openai = openai.FunctionDefinition{
	Name:        "Show_ip_int_br",
	Description: "Get a summary of the status of the interfaces. It gives info about the status of the interface, ip address and others.",
}

var Show_ip_int_br_tool_openai = openai.Tool{
	Type:     openai.ToolTypeFunction,
	Function: F3_openai,
}
var F4_openai = openai.FunctionDefinition{
	Name:        "Show_vlan",
	Description: "Tells what vlans are configured on the device, their name and on which interfaces are applied",
}

var Show_vlan_tool_openai = openai.Tool{
	Type:     openai.ToolTypeFunction,
	Function: F4_openai,
}
var F5_openai = openai.FunctionDefinition{
	Name:        "Show_stp",
	Description: "Tells information about Spanning Tree Protocol or STP. How it is configured and other details.",
}

var Show_stp_tool_openai = openai.Tool{
	Type:     openai.ToolTypeFunction,
	Function: F5_openai,
}
var F6_openai = openai.FunctionDefinition{
	Name:        "Show_mac_address",
	Description: "Tells what mac addresses are seen by each port of the device and other information. This is the mac address table of the device",
}
var Show_mac_address_tool_openai = openai.Tool{
	Type:     openai.ToolTypeFunction,
	Function: F6_openai,
}

var F7_openai = openai.FunctionDefinition{
	Name:        "Show_arp",
	Description: "Tells information about MAC address and IP bindings and on which interface is present",
}

var Show_arp_tool_openai = openai.Tool{
	Type:     openai.ToolTypeFunction,
	Function: F7_openai,
}

var F8_openai = openai.FunctionDefinition{
	Name:        "ReviewConfig",
	Description: "This function start the process to apply configuration or commands to the device. Also helps to review the commands in order to apply them",
}

var Review_config_tool_openai = openai.Tool{
	Type:     openai.ToolTypeFunction,
	Function: F8_openai,
}

// Gemini tools
var F1_gemini = genai.FunctionDeclaration{
	Name:        "Show_cdp",
	Description: "Get information about what devices are connected to this device. Has information about neighbouring devices.",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}

var F2_gemini = genai.FunctionDeclaration{
	Name:        "Show_ip_route",
	Description: "Get information about what IPv4/IPv6 routes are defined. Routes from EIGRP, OSPF, Static routes and default gateway and many others.",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}

var F3_gemini = genai.FunctionDeclaration{
	Name:        "Show_ip_int_br",
	Description: "Get a summary of the status of the interfaces. It gives info about the status of the interface, ip address and others.",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}		

var F4_gemini = genai.FunctionDeclaration{
	Name:        "Show_vlan",
	Description: "Tells what vlans are configured on the device, their name and on which interfaces are applied",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}		

var F5_gemini = genai.FunctionDeclaration{
	Name:        "Show_stp",
	Description: "Tells information about Spanning Tree Protocol or STP. How it is configured and other details.",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}

var F6_gemini = genai.FunctionDeclaration{
	Name:        "Show_mac_address",
	Description: "Tells what mac addresses are seen by each port of the device and other information. This is the mac address table of the device",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}

var F7_gemini = genai.FunctionDeclaration{
	Name:        "Show_arp",
	Description: "Tells information about MAC address and IP bindings and on which interface is present",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}	

var F8_gemini = genai.FunctionDeclaration{
	Name:        "ReviewConfig",
	Description: "This function start the process to apply configuration or commands to the device. Also helps to review the commands in order to apply them",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}

var func_declarations_gemini = []*genai.FunctionDeclaration{
	&F1_gemini,
	&F2_gemini,
	&F3_gemini,
	&F4_gemini,
	&F5_gemini,
	&F6_gemini,
	&F7_gemini,
	&F8_gemini,
}

var Tools_gemini = genai.Tool{
	FunctionDeclarations: func_declarations_gemini,
}

//////////////////////////////////////////////////////////////
// THIS SECTION IS FOR FUNCTIONS TO INTERACT WITH THE DEVICE//
//////////////////////////////////////////////////////////////

func ReviewConfig() string {

	if len(CodeBlocks) > 0 {
		var contentLines []string
		for _, block := range CodeBlocks {
			lines := strings.Split(block, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					contentLines = append(contentLines, line)
				}
			}
		}
		allCode := strings.Join(contentLines, "\n")
		editedContent := multiLineEdit(Rl, allCode)

		fmt.Println("\nEdited content:")
		fmt.Println(Green + editedContent + Reset)
		fmt.Println("\nDo you want to apply these changes? (yes/no)")
		confirmation, _ := Rl.Readline()
		if strings.ToLower(strings.TrimSpace(confirmation)) == "yes" {
			CodeBlocks = []string{editedContent}
			//HERE I APPLY THE CONFIG ON THE SWITCH
			singleLineContent := strings.ReplaceAll(editedContent, "\n", "%")
			cmd := exec.Command("python3", "cmd.py", "-a", singleLineContent)
			_, err := cmd.Output()
			if err != nil {
				return ""
			}
			fmt.Println("Changes saved.")
			cmd = exec.Command("python3", "cmd.py", "-c", "show clock")
			out, err := cmd.Output()
			if err != nil {
				return ""
			}
			timestamp := string(out)

			// Prepare the log entry
			logEntry := fmt.Sprintf("Timestamp: %s\nEdited Content:\n%s\n\n", timestamp, editedContent)

			// Open the file in append mode (or create it if it doesn't exist)
			file, err := os.OpenFile("configs.history", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("Error opening file:", err)
				return ""
			}
			defer file.Close()

			// Write the log entry to the file
			_, err = file.WriteString(logEntry)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return ""
			}

		} else {
			fmt.Println("Changes discarded.")
		}
	} else {
		fmt.Println("No code blocks to edit.")
	}
	return ""
}

func multiLineEdit(rl *readline.Instance, originalContent string) string {
	lines := strings.Split(originalContent, "\n")
	var editedLines []string

	fmt.Println("Review configuration. Press Enter to keep a line unchanged, type new content to replace it, or type 'skip' to remove the line.")
	fmt.Println("Type 'done' on a new line when you're finished editing.")

	for i := 0; i < len(lines); i++ {
		fmt.Printf(Blue+"%d: %s\n"+Reset, i+1, lines[i])
		rl.SetPrompt(fmt.Sprintf("%d> ", i+1))
		newLine, err := rl.Readline()
		if err != nil {
			break
		}
		newLine = strings.TrimSpace(newLine)

		if newLine == "done" {
			break
		} else if newLine == "skip" {
			continue
		} else if newLine != "" {
			editedLines = append(editedLines, newLine)
		} else if lines[i] != "" {
			editedLines = append(editedLines, lines[i])
		}
	}

	rl.SetPrompt("> ")
	return strings.Join(editedLines, "\n")
}

func CallFunctionByName(functionName string, args ...interface{}) (interface{}, error) {
	functions := map[string]interface{}{
		"Show_cdp":         Show_cdp,
		"Show_ip_route":    Show_ip_route,
		"Show_ip_int_br":   Show_ip_int_br,
		"Show_vlan":        Show_vlan,
		"Show_stp":         Show_stp,
		"Show_mac_address": Show_mac_address,
		// "show_runn_interface": show_runn_interface,
		"Show_arp":     Show_arp,
		"ReviewConfig": ReviewConfig,
	}

	if fn, ok := functions[functionName]; ok {
		fnValue := reflect.ValueOf(fn)
		fnType := fnValue.Type()

		if fnType.Kind() != reflect.Func {
			return nil, fmt.Errorf("'%s' is not a function", functionName)
		}

		// Check if the number of provided arguments matches the function's arity
		if len(args) != fnType.NumIn() {
			return nil, fmt.Errorf("invalid number of arguments for function '%s': expected %d, got %d", functionName, fnType.NumIn(), len(args))
		}

		// Convert the provided arguments to reflect.Value
		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			in[i] = reflect.ValueOf(arg)
		}

		// Call the function
		out := fnValue.Call(in)

		// Return the result if there's any
		if len(out) > 0 {
			return out[0].Interface(), nil
		}
		return nil, nil
	}

	return nil, fmt.Errorf("function '%s' not found", functionName)
}

func Show_cdp() string {
	cmd := exec.Command("python3", "cmd.py", "-c", "show cdp neighbour detail")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)

}

func Show_ip_route() string {
	cmd := exec.Command("python3", "cmd.py", "-c", "show ip route")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func Show_ip_int_br() string {
	cmd := exec.Command("python3", "cmd.py", "-c", "show ip interface brief")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func Show_vlan() string {
	cmd := exec.Command("python3", "cmd.py", "-c", "show vlan")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func Show_stp() string {
	cmd := exec.Command("python3", "cmd.py", "-c", "show spanning-tree")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func Show_mac_address() string {
	cmd := exec.Command("python3", "cmd.py", "-c", "show mac-address-table")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func Show_arp() string {
	cmd := exec.Command("python3", "cmd.py", "-c", "show arp")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
