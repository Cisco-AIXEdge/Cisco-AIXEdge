package internals

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func validateProvider(provider string) (bool, error) {
	supportedProviders := []string{"openai", "gemini"}
	for _, p := range supportedProviders {
		if p == provider {
			return true, nil
		}
	}
	return false, fmt.Errorf("unsupported provider: %s", provider)
}

func validateModel(provider string, model string, apiKey string) (bool, error) {
	ctx := context.Background()
	switch provider {
	case "openai":
		client := openai.NewClient(apiKey)
		model, err := client.GetModel(ctx, model)
		if err != nil {
			return false, fmt.Errorf("Invalid model '%s'\nRefer to https://platform.openai.com/docs/models for available models\n", model)
		}
		return true, nil
	case "gemini":
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			return false, fmt.Errorf("Error creating Gemini client: %v\n", err)
		}
		defer client.Close()
		
		model := client.GenerativeModel(model)
		_, err = model.Info(ctx)
		if err != nil {			
			return false, fmt.Errorf("Error retrieving model info: %v\nRefer to https://ai.google.dev/gemini-api/docs/models for available models\n", err)
		}

		return true, nil
	default:
		return false, fmt.Errorf("Unsupported provider: %s", provider)
	}
	return false, fmt.Errorf("Unsupported provider: %s", provider)
}
