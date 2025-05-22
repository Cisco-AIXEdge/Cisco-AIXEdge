package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/cisco"
	"github.com/chzyer/readline"
	"github.com/sashabaranov/go-openai"
)

func (a *Client) Interactive(sn string) {

	client := openai.NewClient(a.API)

	req := openai.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a Cisco IOS XE configuration assistant and you answer only to IOS-XE related questions. Put all commands that you suggest in code blocks",
			},
		},
		Tools: []openai.Tool{cisco.Show_cdp_tool, cisco.Show_ip_route_tool, cisco.Show_ip_int_br_tool, cisco.Show_vlan_tool, cisco.Show_stp_tool, cisco.Show_mac_address_tool, cisco.Show_arp_tool, cisco.Review_config_tool},
	}
	cnt := context.Background()

	cisco.Rl, _ = readline.NewEx(&readline.Config{
		Prompt:          "> ",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		HistoryFile:     "/tmp/readline.tmp",
	})
	defer cisco.Rl.Close()

	printInstructions()

	for {
		line, err := cisco.Rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		switch strings.ToLower(line) {
		case "exit":
			// printChatHistory(req.Messages[1:])
			writeChatHistoryToFile(req.Messages[1:])
			return
		default:
			handleChatCompletion(client, &req, line, &cnt)
		}
	}
}

func printInstructions() {
	fmt.Println("IOS-XE AI Assistant")
	fmt.Println("Type 'exit' to end the conversation.")
}

func handleChatCompletion(client *openai.Client, req *openai.ChatCompletionRequest, line string, cnt *context.Context) {
	req.Messages = append(req.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: line,
	})
	resp, err := client.CreateChatCompletion(*cnt, *req)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) == 0 {

		cisco.CodeBlocks = printFormattedContent(resp.Choices[0].Message.Content)
		req.Messages = append(req.Messages, resp.Choices[0].Message)
	} else {
		funcName := msg.ToolCalls[0].Function.Name
		answer, err := cisco.CallFunctionByName(funcName)
		if err != nil {
			fmt.Printf("Function calling error: %v\n", err)
			return
		}
		if funcName != "ReviewConfig" {
			resp, err = client.CreateChatCompletion(*cnt,
				openai.ChatCompletionRequest{
					Model: "gpt-4o",
					Messages: []openai.ChatCompletionMessage{
						{
							Role: openai.ChatMessageRoleSystem,
							// here i added the response as context and i ask again the question
							Content: answer.(string),
						},
						{
							Role:    openai.ChatMessageRoleUser,
							Content: line,
						},
					},
				},
			)
			if err != nil {
				fmt.Printf("ChatCompletion error: %v\n", err)
				return
			}
			cisco.CodeBlocks = printFormattedContent(resp.Choices[0].Message.Content)
			req.Messages = append(req.Messages, resp.Choices[0].Message)
		} else {
			resp, err = client.CreateChatCompletion(*cnt,
				openai.ChatCompletionRequest{
					Model: "gpt-4o",
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleUser,
							Content: "Thank you!",
						},
					},
				},
			)
			if err != nil {
				fmt.Printf("ChatCompletion error: %v\n", err)
				return
			}
			cisco.CodeBlocks = printFormattedContent(resp.Choices[0].Message.Content)
			req.Messages = append(req.Messages, resp.Choices[0].Message)
		}
	}
}

func printFormattedContent(content string) []string {
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	codeBlock := strings.Builder{}
	codeBlocks := []string{}

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				codeBlocks = append(codeBlocks, codeBlock.String())
				codeBlock.Reset()
			}
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			fmt.Printf(cisco.Yellow+"    %s\n"+cisco.Reset, line)
			codeBlock.WriteString(line + "\n")
		} else {
			printHighlightedLine(line)
		}
	}

	return codeBlocks
}

func printHighlightedLine(line string) {
	var result strings.Builder
	inBackticks := false
	inAsterisks := false
	i := 0

	for i < len(line) {
		switch {
		case i+1 < len(line) && line[i:i+2] == "**":
			inAsterisks = !inAsterisks
			if inAsterisks {
				result.WriteString(cisco.Green)
			} else {
				result.WriteString(cisco.Reset)
			}
			i += 2
		case line[i] == '`':
			inBackticks = !inBackticks
			if inBackticks {
				result.WriteString(cisco.Green)
			} else {
				result.WriteString(cisco.Reset)
			}
			i++
		default:
			result.WriteByte(line[i])
			i++
		}
	}

	// Ensure we reset the color at the end of the line
	result.WriteString(cisco.Reset)
	fmt.Println(result.String())
}

func writeChatHistoryToFile(messages []openai.ChatCompletionMessage) {
	// Open the file in append mode, or create it if it doesn't exist
	file, _ := os.OpenFile("chat.telemetry", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	cmd := exec.Command("python3", "cmd.py", "-c", "show clock")
	out, _ := cmd.Output()

	timestamp := string(out)
	file.WriteString(fmt.Sprintf("\nTime: %s\n\n", timestamp))
	// Write chat history to the file
	for _, msg := range messages {
		switch msg.Role {
		case openai.ChatMessageRoleUser:
			_, _ = file.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
		case openai.ChatMessageRoleAssistant:
			_, _ = file.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
		}
	}
	// uploadFileToS3(filename, bucketName, objectKey)

	// Delete local file
	_ = os.Remove("chat.telemetry")

}
