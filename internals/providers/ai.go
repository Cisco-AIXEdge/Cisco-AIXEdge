package providers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/cisco"
	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
)

type Engine struct {
	Provider string
	Version  string
}

type Client struct {
	API    string
	Engine Engine
}

func geminiPrintResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {

		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
}

func isValidShowCommand(command string) bool {
	// Check if the command starts with "show"
	if !strings.HasPrefix(command, "show") {
		return false
	}

	// Define a blacklist of show commands
	blacklist := []string{
		"show tech",
		"show interfaces",
		"show tech-support",
	}

	// Check if the command is in the blacklist
	for _, blacklistedCommand := range blacklist {
		if strings.TrimSpace(strings.ToLower(command)) == blacklistedCommand {
			return false
		}
	}

	// If the command starts with "show" and is not in the blacklist, return true
	return true
}

// Function handles interaction between app and OpenAI
func (a *Client) Prompt(content string) (string, string, int) {
	var prompt string
	var cmd string
	var error_code int
	error_code = 200
	//To separate cisco command and AI query '@' is used
	promptSeparator := "@"
	cli := cisco.IOSXE{}
	switch a.Engine.Provider {
	case "openai":
		maxToken := 500
		client := openai.NewClient(a.API)
		var resp openai.ChatCompletionResponse
		var err error

		//Based on existance of separator the API call is selected
		if strings.Contains(content, promptSeparator) {
			contents := strings.Split(content, promptSeparator)
			cmd = contents[0]
			prompt = contents[1]
			if isValidShowCommand(cmd) {
				output, err := cli.Command(cmd)
				if err == nil {
					resp, err = client.CreateChatCompletion(
						context.Background(),
						openai.ChatCompletionRequest{
							Model:     a.Engine.Version,
							MaxTokens: maxToken,
							Messages: []openai.ChatCompletionMessage{
								{
									Role:    openai.ChatMessageRoleSystem,
									Content: "You are Cisco network engineer assistant and you respond only to questions about Cisco IOS-XE",
								},
								{
									Role:    openai.ChatMessageRoleUser,
									Content: "You have the following output: " + output,
								},
								{
									Role:    openai.ChatMessageRoleUser,
									Content: prompt,
								},
							},
						},
					)
					if err != nil {
						switch {
						case strings.Contains(err.Error(), "400"):
							fmt.Print("Copilot cannot understand at this moment this output! :(\n")
							error_code = 400
						case strings.Contains(err.Error(), "401"):
							fmt.Printf("Invalid API key!\nUse copilot-cfg <API KEY> to update the API key\n")

						case strings.Contains(err.Error(), "429"):
							fmt.Printf("Copilot has hit limit..Please try again later!")
							error_code = 429

						default:
							fmt.Printf("Copilot has encountered a server error.")
							error_code = 500
						}

					} else {
						fmt.Println(resp.Choices[0].Message.Content)
					}
				} else {
					error_code = 402
					fmt.Print("There is a typo in you show command. Fix it and try again! :)\n")
				}
			} else {
				error_code = 402
				fmt.Print("The command is not supported yet. :)\n")
			}
		} else {
			prompt = content
			resp, err = client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:     a.Engine.Version,
					MaxTokens: maxToken,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleSystem,
							Content: "You are Cisco network engineer assistant and you respond only to questions about Cisco IOS-XE devices.",
						},
						{
							Role:    openai.ChatMessageRoleUser,
							Content: "You are Cisco network engineer assistant and you respond only to questions about Cisco IOS-XE devices.",
						},
						{
							Role:    openai.ChatMessageRoleUser,
							Content: prompt,
						},
					},
				},
			)
			if err != nil {
				switch {

				case strings.Contains(err.Error(), "400"):
					fmt.Print("Copilot cannot understand at this moment this output! :(\n")
					error_code = 400
				case strings.Contains(err.Error(), "401"):
					fmt.Printf("Invalid API key!\nUse copilot-cfg <API KEY> to update the API key\n")

				case strings.Contains(err.Error(), "429"):
					fmt.Printf("Copilot has hit limit..Please try again later!")
					error_code = 429

				default:
					fmt.Printf("Copilot has encountered a server error.")
					error_code = 500
				}

			} else {
				fmt.Println(resp.Choices[0].Message.Content)
			}
		}
	case "gemini":
		ctx := context.Background()
		// client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyC3FBUNpldvriXbdKlhcvTZivydElcKH-I"))
		client, err := genai.NewClient(ctx, option.WithAPIKey(a.API))
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		model := client.GenerativeModel(a.Engine.Version)
		model.SetMaxOutputTokens(500)

		//Based on existance of separator the API call is selected
		if strings.Contains(content, promptSeparator) {
			contents := strings.Split(content, promptSeparator)
			cmd = contents[0]
			prompt = contents[1]
			if isValidShowCommand(cmd) {
				output, err := cli.Command(cmd)
				if err == nil {
					model.SystemInstruction = &genai.Content{
						Parts: []genai.Part{genai.Text(`You are Cisco network engineer assistant and you respond only to questions about Cisco IOS-XE. You have the following output:
					` + output)},
					}

					resp, err := model.GenerateContent(ctx, genai.Text(prompt))
					if err != nil {
						switch {
						case strings.Contains(err.Error(), "400"):
							fmt.Print("Copilot cannot understand at this moment this output! :(\n")
							error_code = 400
						case strings.Contains(err.Error(), "401"):
							fmt.Printf("Invalid API key!\nUse copilot-cfg <API KEY> to update the API key\n")

						case strings.Contains(err.Error(), "429"):
							fmt.Printf("Copilot has hit limit..Please try again later!")
							error_code = 429

						default:
							fmt.Printf("Copilot has encountered a server error.")
							error_code = 500
						}

					} else {
						geminiPrintResponse(resp)
					}
				} else {
					error_code = 402
					fmt.Print("There is a typo in you show command. Fix it and try again! :)\n")
				}
			} else {
				error_code = 402
				fmt.Print("The command is not supported yet. :)\n")
			}
		} else {
			prompt = content
			model.SystemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(`You are Cisco network engineer assistant and you respond only to questions about Cisco IOS-XE.`)},
			}

			resp, err := model.GenerateContent(ctx, genai.Text(prompt))
			if err != nil {
				switch {

				case strings.Contains(err.Error(), "400"):
					fmt.Print("Copilot cannot understand at this moment this output! :(\n")
					error_code = 400
				case strings.Contains(err.Error(), "401"):
					fmt.Printf("Invalid API key!\nUse copilot-cfg <API KEY> to update the API key\n")

				case strings.Contains(err.Error(), "429"):
					fmt.Printf("Copilot has hit limit..Please try again later!")
					error_code = 429

				default:
					fmt.Printf("Copilot has encountered a server error.")
					error_code = 500
				}

			} else {
				geminiPrintResponse(resp)
			}
		}
	}
	return prompt, cmd, error_code
}
