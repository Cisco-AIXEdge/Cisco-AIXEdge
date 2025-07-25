package internals

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"google.golang.org/api/option"
)

type AnswerFeature struct {
	Query string `json:"feature_name"`
}

type Question struct {
	Question string `json:"question"`
	Pid      string `json:"pid"`
	SwVer    string `json:"swVer"`
	ApiKey   string `json:"api_key"`
}

type Feature struct {
	FeatureDesc    string `json:"feature_desc"`
	FeatureName    string `json:"feature_name"`
	FeatureSetDesc string `json:"feature_set_desc"`
}

type Platform struct {
	PlatformID   int    `json:"platform_id"`
	PlatformName string `json:"platform_name"`
}

type Release struct {
	ReleaseID     int    `json:"release_id"`
	ReleaseNumber string `json:"release_number"`
}

func getPlatformID(platformName string) (int, error) {
	url := "https://cfnngws.cisco.com/api/v1/platform"
	payload := strings.NewReader(`{"mdf_product_type":"Switches"}`)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	var platforms []Platform
	err = json.Unmarshal(body, &platforms)
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	for _, platform := range platforms {
		if strings.Contains(strings.ToLower(platform.PlatformName), strings.ToLower(platformName)) {
			return platform.PlatformID, nil
		}
	}

	return 0, fmt.Errorf("platform not found: %s", platformName)
}

func getReleaseID(platformID int, releaseNumber string) (int, error) {
	url := "https://cfnngws.cisco.com/api/v1/release"
	payload := fmt.Sprintf(`{"platform_id":%d,"mdf_product_type":"Switches","release_id":null,"feature_set_id":null}`, platformID)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	var releases []Release
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	for _, release := range releases {
		if strings.Contains(strings.ToLower(release.ReleaseNumber), strings.ToLower(releaseNumber)) {
			return release.ReleaseID, nil
		}
	}

	return 0, fmt.Errorf("release not found: %s", releaseNumber)
}

func find_feature(query string, platformName, softwareVersion string) (string, error) {

	var platformID int
	var releaseID int

	//getting platform ID
	platformID, err := getPlatformID(platformName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	releaseID, err = getReleaseID(platformID, softwareVersion)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	url := "https://cfnngws.cisco.com/api/v1/by_product_result"
	method := "POST"

	payload := fmt.Sprintf(`{"platform_id":%d,"mdf_product_type":"Switches","release_id":%d,"feature_set_id":null}`, platformID, releaseID)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var features []Feature
	err = json.NewDecoder(res.Body).Decode(&features)
	if err != nil {
		return "", err
	}

	var results []string
	swg := metrics.NewSmithWatermanGotoh()
	swg.CaseSensitive = false
	swg.GapPenalty = -0.1
	swg.Substitution = metrics.MatchMismatch{
		Match:    1,
		Mismatch: -0.5,
	}

	for _, feature := range features {
		similarity := strutil.Similarity(query, feature.FeatureName, swg)
		if similarity >= 0.90 {
			result := fmt.Sprintf("Feature: %s (Similarity: %.2f), Available in: %s",
				feature.FeatureName, similarity, feature.FeatureSetDesc)
			results = append(results, result)
		}
	}

	result := strings.Join(results, "\n")
	result = result + "\n Advise the customer to go to Cisco Feature Navigator for more information. Here is the link: https://cfnng.cisco.com/"

	return result, nil
}

func (c *Client) FeaturePrompt(content string) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Println("API key non-existant. Please do copilot-cfg <API KEY>")
		}
	}()
	cfg, err := c.configRead()
	if err != nil {
		panic(err)
	}

	switch cfg.Engine {
	case "openai":
		ctx := context.Background()
		params := jsonschema.Definition{
			Type: jsonschema.Object,
			// Here after e.g. we can put the PID of the device to auto complete if the customer asks "tell me which sfps are compatible with this device"
			Properties: map[string]jsonschema.Definition{
				"feature_name": {
					Type:        jsonschema.String,
					Description: "The name of the Cisco IOS-XE feature that the user is asking about",
				},
			},
			Required: []string{"feature_name"},
		}
		// To improve accuracy we need to be sure that this is a good description...this is how openai knows that is a god moment to call the function when a q is asked
		f := openai.FunctionDefinition{
			Name:        "find_feature",
			Description: "Retrieves information about existance of a certain feature for " + cfg.Platform + " with software version " + cfg.SwVer,
			Parameters:  params,
		}
		// added the function to the tool chain
		t := openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: f,
		}
		maxToken := 1000
		//this is my api key not monitored
		client := openai.NewClient(cfg.Apikey)
		var resp openai.ChatCompletionResponse
		resp, err = client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:     cfg.EngineVERSION,
				MaxTokens: maxToken,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: "You are Cisco network engineer assistant and you respond only to questions about Cisco IOS-XE devices.",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: content,
					},
				},
				// here i give it the tool chain to use in case it needs it
				Tools: []openai.Tool{t},
			},
		)
		if err != nil || len(resp.Choices) != 1 {
			panic("No API Key")
		}
		msg := resp.Choices[0].Message

		if len(msg.ToolCalls) == 0 {
			//if the tool chain is not used i just print the answer (general question) (we can do prompt limiting only for optics)

			fmt.Print(resp.Choices[0].Message.Content)

		} else {
			// i get the function name
			funcName := msg.ToolCalls[0].Function.Name
			// i get the function arguments
			var answer2 AnswerFeature
			json.Unmarshal([]byte(msg.ToolCalls[0].Function.Arguments), &answer2)
			// get the reponse from the function

			answer := callFeatureByName(funcName, cfg.SwVer, cfg.Platform, answer2.Query)

			resp, err := client.CreateChatCompletion(ctx,
				openai.ChatCompletionRequest{
					Model:     cfg.EngineVERSION,
					MaxTokens: maxToken,
					Messages: []openai.ChatCompletionMessage{
						{
							Role: openai.ChatMessageRoleSystem,
							// here i added the response as context and i ask again the question
							Content: answer.(string),
						},
						{
							Role:    openai.ChatMessageRoleUser,
							Content: content,
						},
					},
				},
			)
			_ = err
			// _ = resp
			fmt.Print(resp.Choices[0].Message.Content)

		}

	case "gemini":
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.Apikey))
		if err != nil {
			panic("No API Key")
		}
		defer client.Close()
		params := &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"feature_name": {
					Type:        genai.TypeString,
					Description: "The name of the Cisco IOS-XE feature that the user is asking about",
				},
			},
			Required: []string{"feature_name"},
		}

		f := &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{{
				Name:        "find_feature",
				Description: "Retrieves information about existance of a certain feature for " + cfg.Platform + " with software version " + cfg.SwVer,
				Parameters:  params,
			}},
		}

		model := client.GenerativeModel(cfg.EngineVERSION)
		model.Tools = []*genai.Tool{f}

		session := model.StartChat()
		res, err := session.SendMessage(ctx, genai.Text(content))
		if err != nil {
			panic("No API Key")
		}
		part := res.Candidates[0].Content.Parts[0]
		funcall, ok := part.(genai.FunctionCall)
		if !ok {
			resp, err := model.GenerateContent(ctx, genai.Text(content))
			if err != nil {
				log.Fatal(err)
			}
			geminiPrintResponse(resp)
		}

		// Expect the model to pass a proper string "location" argument to the tool.
		locArg, ok := funcall.Args["feature_name"].(string)

		if !ok {
			log.Fatalf("expected string: %v", funcall.Args["feature_name"])
		}

		answer, err := find_feature(locArg, cfg.Platform, cfg.SwVer)
		if err != nil {
			log.Fatal(err)
		}
		res, err = session.SendMessage(ctx, genai.FunctionResponse{
			Name: f.FunctionDeclarations[0].Name,
			Response: map[string]any{
				"answer": answer,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		geminiPrintResponse(res)
	}
}

func callFeatureByName(functionName string, softwareVersion string, platformName string, query string) interface{} {
	functions := map[string]interface{}{
		"find_feature": find_feature,
	}

	if fn, ok := functions[functionName]; ok {
		resultValue := reflect.ValueOf(fn)
		if resultValue.Kind() != reflect.Func {
			return nil
		}

		queryValue := reflect.ValueOf(query)
		platformNameValue := reflect.ValueOf(platformName)
		softwareVersionValue := reflect.ValueOf(softwareVersion)

		resultValues := resultValue.Call([]reflect.Value{queryValue, platformNameValue, softwareVersionValue})

		if len(resultValues) > 0 {
			// Check if there's an error (second return value)
			if len(resultValues) > 1 && !resultValues[1].IsNil() {
				return fmt.Sprintf("Error: %v", resultValues[1].Interface())
			}
			return resultValues[0].Interface()
		}
	} else {
		return "Error: Function not found"
	}
	return nil
}
