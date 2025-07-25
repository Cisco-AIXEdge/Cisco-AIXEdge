package providers

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/cisco"
	"github.com/chzyer/readline"
	"github.com/sashabaranov/go-openai"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ChatMessage struct {
	Role	string `json:"role"` // "User" or "Assistant"
	Content string `json:"content"`
}

func (a *Client) Interactive(sn string) {

	switch a.Engine.Provider {
	case "openai":
		client := openai.NewClient(a.API)

		req := openai.ChatCompletionRequest{
			Model: a.Engine.Version,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a Cisco IOS-XE configuration assistant and you answer only to IOS-XE related questions. Put all commands that you suggest in code blocks",
				},
			},
			Tools: []openai.Tool{cisco.Show_cdp_tool_openai, cisco.Show_ip_route_tool_openai, cisco.Show_ip_int_br_tool_openai, cisco.Show_vlan_tool_openai, cisco.Show_stp_tool_openai, cisco.Show_mac_address_tool_openai, cisco.Show_arp_tool_openai, cisco.Review_config_tool_openai},
		}
		ctx := context.Background()

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
				handleOpenAIChatCompletion(client, &req, line, &ctx)
			}
		}
	case "gemini":
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(a.API))
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		model := client.GenerativeModel(a.Engine.Version)
		model.Tools = []*genai.Tool{&cisco.Tools_gemini}
		model.SetMaxOutputTokens(1024)

		// Initialize chat history for Gemini
        var chatHistory []ChatMessage

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
				writeChatHistoryToFile(chatHistory[1:])
				return
			default:
				handleGeminiChatCompletion(model, &chatHistory, line, &ctx)
			}
		}
	}
}

func printInstructions() {
	fmt.Println("IOS-XE AI Assistant")
	fmt.Println("Type 'exit' to end the conversation.")
}

func handleOpenAIChatCompletion(client *openai.Client, req *openai.ChatCompletionRequest, line string, ctx *context.Context) {
	req.Messages = append(req.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: line,
	})
	resp, err := client.CreateChatCompletion(*ctx, *req)
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
			resp, err = client.CreateChatCompletion(*ctx,
				openai.ChatCompletionRequest{
					Model: req.Model,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleSystem,
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
			resp, err = client.CreateChatCompletion(*ctx,
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

func handleGeminiChatCompletion(model *genai.GenerativeModel, chatHistory *[]ChatMessage, line string, ctx *context.Context) {
    // Add user message to chat history first
    *chatHistory = append(*chatHistory, ChatMessage{Role: "User", Content: line})
    
    // Start a chat session to maintain conversation history
    cs := model.StartChat()
    
    // Add previous messages to the chat session (excluding the current user message we just added)
    for i, msg := range *chatHistory {
        if i == len(*chatHistory)-1 {
            break // Skip the current message we just added
        }
        if msg.Role == "User" {
            cs.History = append(cs.History, &genai.Content{
                Parts: []genai.Part{genai.Text(msg.Content)},
                Role:  "user",
            })
        } else if msg.Role == "Assistant" {
            cs.History = append(cs.History, &genai.Content{
                Parts: []genai.Part{genai.Text(msg.Content)},
                Role:  "model",
            })
        }
    }
    
    resp, err := cs.SendMessage(*ctx, genai.Text(line))
    if err != nil {
        log.Fatal(err)
    }

    // Extract and handle the response content
    if resp.Candidates != nil && len(resp.Candidates) > 0 {
        candidate := resp.Candidates[0]
        if candidate.Content != nil && candidate.Content.Parts != nil {
            
            var hasToolCall bool
            var toolCallResults []genai.Part
            
            // First pass: check for tool calls and execute them
            for _, part := range candidate.Content.Parts {
                if fnCall, ok := part.(genai.FunctionCall); ok {
                    hasToolCall = true
                    
                    // Execute the function call
                    answer, err := cisco.CallFunctionByName(fnCall.Name)
                    if err != nil {
                        fmt.Printf("Function calling error: %v\n", err)
                        return
                    }
                    
                    // Create function response part
                    toolCallResults = append(toolCallResults, genai.FunctionResponse{
                        Name:     fnCall.Name,
                        Response: map[string]any{"result": answer},
                    })
                }
            }
            
            if hasToolCall {
                // Send function responses back through the chat session
                finalResp, err := cs.SendMessage(*ctx, toolCallResults...)
                if err != nil {
                    log.Fatal(err)
                }
                
                // Process the final response
                if finalResp.Candidates != nil && len(finalResp.Candidates) > 0 {
                    finalCandidate := finalResp.Candidates[0]
                    if finalCandidate.Content != nil && finalCandidate.Content.Parts != nil {
                        for _, part := range finalCandidate.Content.Parts {
                            if textPart, ok := part.(genai.Text); ok {
                                *chatHistory = append(*chatHistory, ChatMessage{Role: "Assistant", Content: string(textPart)})
                                cisco.CodeBlocks = printFormattedContent(string(textPart))
                            }
                        }
                    }
                }
            } else {
                // No tool calls, process text normally
                for _, part := range candidate.Content.Parts {
                    if textPart, ok := part.(genai.Text); ok {
                        *chatHistory = append(*chatHistory, ChatMessage{Role: "Assistant", Content: string(textPart)})
                        cisco.CodeBlocks = printFormattedContent(string(textPart))
                    }
                }
            }
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

func writeChatHistoryToFile(messages interface{}) {
	// Open the file in append mode, or create it if it doesn't exist
	file, _ := os.OpenFile("chat.telemetry", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	cmd := exec.Command("python3", "cmd.py", "-c", "show clock")
	out, _ := cmd.Output()

	timestamp := string(out)
	file.WriteString(fmt.Sprintf("\nTime: %s\n\n", timestamp))
	
	// Write chat history to the file based on type
	switch msgs := messages.(type) {
	case []openai.ChatCompletionMessage:
		for _, msg := range msgs {
			switch msg.Role {
			case openai.ChatMessageRoleUser:
				_, _ = file.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
			case openai.ChatMessageRoleAssistant:
				_, _ = file.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
			}
		}
	case []*ChatMessage:
		for _, msg := range msgs {
			switch msg.Role {
			case "User":
				_, _ = file.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
			case "Assistant":
				_, _ = file.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
			}
		}
	default:
		_, _ = file.WriteString("Error: Unsupported message type\n")
	}

	// Delete local file
	_ = os.Remove("chat.telemetry")
}
