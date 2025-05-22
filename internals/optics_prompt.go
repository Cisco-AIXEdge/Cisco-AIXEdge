package internals

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/cisco"
	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"google.golang.org/api/option"
)

type ResponseBodyOptics struct {
	Message string `json:"message"`
}

type AnswerOptics struct {
	Query string `json:"query"`
}

type NetworkDevice struct {
	ProductFamily               string             `json:"productFamily"`
	NetworkAndTransceiverCompat []TransceiverEntry `json:"networkAndTransceiverCompatibility"`
}

type TransceiverEntry struct {
	ProductId    string        `json:"productId"`
	Transceivers []Transceiver `json:"transceivers"`
}

type Transceiver struct {
	TmgId           int    `json:"tmgId"`
	ProductFamilyId int    `json:"productFamilyId"`
	ProductFamily   string `json:"productFamily"`
	ProductModelId  int    `json:"productModelId"`
	ProductId       string `json:"productId"`
	Reach           string `json:"reach"`
	Media           string `json:"media"`
	DataRate        string `json:"dataRate"`
	SwVersion       string `json:"softReleaseMinVer"`
}

type ResponseOptics struct {
	TotalCount     int             `json:"totalCount"`
	ItemPerPage    *int            `json:"itemPerPage"`
	Page           *int            `json:"page"`
	NetworkDevices []NetworkDevice `json:"networkDevices"`
}

// //////////////////////////////////////
// /INITIAL
// type Response struct {
// 	Message string `json:"message"`
// }

// //////ADAUGAT ACUM
func isInList(strList []string, target string) bool {
	for _, str := range strList {
		if str == target {
			return true
		}
	}
	return false
}

func getSuggestionJSON(suggestionText string) string {

	upperStr := strings.ToUpper(suggestionText)

	result := ""
	if !isInList(INVENTORY, upperStr) {
		url := "https://tmgmatrix.cisco.com/public/api/networkdevice/search"
		method := "POST"
		//Searching by input got from OpenAI (PID)
		// I didn't use autosuggest from TMG
		payload := strings.NewReader(`{
    "cableType": [],
    "dataRate": [],
    "formFactor": [],
    "reach": [],
    "searchInput": [
        "` + suggestionText + `"
    ],
    "osType": [
        {
            "id": 3,
            "name": "IOS XE",
            "count": 0,
            "filterChecked": true,
            "filtername": "osType"
        }
    ],
    "transceiverProductFamily": [],
    "transceiverProductID": [

    ],
    "networkDeviceProductFamily": [],
    "networkDeviceProductID": []
}`)
		//Make HTTP REQ to TMG
		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return ""
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return ""
		}

		var response ResponseOptics
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		// I got multiple answers for an optic compatibility with a device (different sw versions where compatible and I kept the earliest sw version)
		minVersions := make(map[string]string)
		for _, device := range response.NetworkDevices {

			for _, entry := range device.NetworkAndTransceiverCompat {
				for _, t := range entry.Transceivers {
					//I have t.ProductID and entry.ProductID because it's a difference in JSON when I query an optic or a device.
					currentMinVer, exists := minVersions[t.ProductId]
					//just change here to ">0" to get the latest version presented on TMG
					if !exists || strings.Compare(t.SwVersion, currentMinVer) < 0 {
						minVersions[t.ProductId] = t.SwVersion

					}
					currentMinVer, exists = minVersions[entry.ProductId]

					if !exists || strings.Compare(t.SwVersion, currentMinVer) < 0 {
						minVersions[entry.ProductId] = t.SwVersion

					}
				}
			}
		}

		for _, device := range response.NetworkDevices {
			if strings.Contains(device.ProductFamily, "C9") {
				result += "  Product Family: " + device.ProductFamily + "\n"
				for _, entry := range device.NetworkAndTransceiverCompat {
					result += "  - Product ID: " + entry.ProductId + "\n"
					for _, t := range entry.Transceivers {
						// Here it generates the string for optic search
						if t.SwVersion == minVersions[t.ProductId] {
							result += " - Transceiver:" + t.ProductId + ", Reach:" + t.Reach + ", Media:" + t.Media + ", Speed:" + t.DataRate + ", Software Release:" + t.SwVersion + "\n"
						}
						//Here for device search
						if minVersions[entry.ProductId] != minVersions[t.ProductId] {
							if t.SwVersion == minVersions[entry.ProductId] {
								result += " - Transceiver:" + t.ProductId + ", Reach:" + t.Reach + ", Media:" + t.Media + ", Speed:" + t.DataRate + ", Software Release:" + t.SwVersion + "\n"

							}
						}
					}
				}
			}
		}
		result += "\n\n If you dont know the answer tell the customer that you haven't found anything and fr more info to go to https://tmgmatrix.cisco.com/"
		// fmt.Print(result)
		//Uncomment this to see how it looks and let me know if anything needs to be added as info (check in postman)
		return result
	}
	for _, str := range INVENTORY {

		url := "https://tmgmatrix.cisco.com/public/api/networkdevice/search"
		method := "POST"
		//Searching by input got from OpenAI (PID)
		// I didn't use autosuggest from TMG
		payload := strings.NewReader(`{
    "cableType": [],
    "dataRate": [],
    "formFactor": [],
    "reach": [],
    "searchInput": [
        "` + str + `"
    ],
    "osType": [
        {
            "id": 3,
            "name": "IOS XE",
            "count": 0,
            "filterChecked": true,
            "filtername": "osType"
        }
    ],
    "transceiverProductFamily": [],
    "transceiverProductID": [

    ],
    "networkDeviceProductFamily": [],
    "networkDeviceProductID": []
}`)
		//Make HTTP REQ to TMG
		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return ""
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return ""
		}

		var response ResponseOptics
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		// I got multiple answers for an optic compatibility with a device (different sw versions where compatible and I kept the earliest sw version)
		minVersions := make(map[string]string)
		for _, device := range response.NetworkDevices {

			for _, entry := range device.NetworkAndTransceiverCompat {
				for _, t := range entry.Transceivers {
					//I have t.ProductID and entry.ProductID because it's a difference in JSON when I query an optic or a device.
					currentMinVer, exists := minVersions[t.ProductId]
					//just change here to ">0" to get the latest version presented on TMG
					if !exists || strings.Compare(t.SwVersion, currentMinVer) < 0 {
						minVersions[t.ProductId] = t.SwVersion

					}
					currentMinVer, exists = minVersions[entry.ProductId]

					if !exists || strings.Compare(t.SwVersion, currentMinVer) < 0 {
						minVersions[entry.ProductId] = t.SwVersion

					}
				}
			}
		}

		for _, device := range response.NetworkDevices {
			if strings.Contains(device.ProductFamily, "C9") {
				result += "  Product Family: " + device.ProductFamily + "\n"
				for _, entry := range device.NetworkAndTransceiverCompat {
					result += "  - Product ID: " + entry.ProductId + "\n"
					for _, t := range entry.Transceivers {
						// Here it generates the string for optic search
						if t.SwVersion == minVersions[t.ProductId] {
							result += " - Transceiver:" + t.ProductId + ", Reach:" + t.Reach + ", Media:" + t.Media + ", Speed:" + t.DataRate + ", Software Release:" + t.SwVersion + "\n"
						}
						//Here for device search
						if minVersions[entry.ProductId] != minVersions[t.ProductId] {
							if t.SwVersion == minVersions[entry.ProductId] {
								result += " - Transceiver:" + t.ProductId + ", Reach:" + t.Reach + ", Media:" + t.Media + ", Speed:" + t.DataRate + ", Software Release:" + t.SwVersion + "\n"

							}
						}
					}
				}
			}
		}

	}
	result += "\n\n Advise the customer to go to Cisco Feature Navigator for more information. Here is the link: https://cfnng.cisco.com/"
	//Uncomment this to see how it looks and let me know if anything needs to be added as info (check in postman)
	fmt.Print(result)
	return result

}

// This is the function caller function
func callOpticByName(functionName string, arg string) interface{} {

	functions := map[string]func(string) string{
		"getSuggestionJSON": getSuggestionJSON,
	}
	if fn, ok := functions[functionName]; ok {
		resultValue := reflect.ValueOf(fn)
		if resultValue.Kind() != reflect.Func {
			return nil
		}
		argValue := reflect.ValueOf(arg)
		if !argValue.IsValid() {
			return nil
		}
		resultValues := resultValue.Call([]reflect.Value{argValue})
		if len(resultValues) > 0 {
			return resultValues[0].Interface()
		}
	} else {
		fmt.Print("Error")
	}
	return nil
}

var INVENTORY []string = nil

func (c *Client) OpticsPrompt(content string) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Println("API key non-existant. Please do copilot-cfg <API KEY>")
		}
	}()
	cfg, err := c.configRead()
	if err != nil {
		panic(err)
	}
	iosxe := cisco.IOSXE{}

	inventory, _ := iosxe.Inventory()

	data := strings.Split(string(inventory), ",")
	lastIndex := len(data) - 1
	data[lastIndex] = strings.TrimRight(data[lastIndex], "\n")
	INVENTORY = data

	modelType := "gpt-4o"

	switch modelType {
	case "gpt-4o":
		ctx := context.Background()
		params := jsonschema.Definition{
			Type: jsonschema.Object,
			// Here after e.g. we can put the PID of the device to auto complete if the customer asks "tell me which sfps are compatible with this device"
			Properties: map[string]jsonschema.Definition{
				"query": {
					Type:        jsonschema.String,
					Description: "The optic Product ID or Device Product ID e.g. " + cfg.PID + ". Your default value is " + cfg.PID,
				},
			},
			Required: []string{"query"},
		}
		// To improve accuracy we need to be sure that this is a good description...this is how openai knows that is a god moment to call the function when a q is asked
		f := openai.FunctionDefinition{
			Name:        "getSuggestionJSON",
			Description: "Retrieves information about compaitiblity between optic modules and IOS-XE devices their network modules. Network modules have NM in their ID",
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
				Model:     modelType,
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
			fmt.Println(resp.Choices[0].Message.Content)

		} else {
			// i get the function name
			funcName := msg.ToolCalls[0].Function.Name
			// i get the function arguments
			var answer2 AnswerOptics
			json.Unmarshal([]byte(msg.ToolCalls[0].Function.Arguments), &answer2)
			// get the reponse from the function
			answer := callOpticByName(funcName, answer2.Query)
			resp, err := client.CreateChatCompletion(ctx,
				openai.ChatCompletionRequest{
					Model:     modelType,
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
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				switch {

				case strings.Contains(err.Error(), "400"):
					fmt.Print("Copilot cannot understand at this moment this output! :(\n")

				case strings.Contains(err.Error(), "401"):
					fmt.Printf("Invalid API key!\nUse copilot-cfg <API KEY> to update the API key\n")

				case strings.Contains(err.Error(), "429"):
					fmt.Printf("Copilot has hit limit..Please try again later!")

				default:
					fmt.Printf("Copilot has encountered a server error.")

				}
			} else {
				fmt.Println(resp.Choices[0].Message.Content)
			}
		}

	case "gemini-1.0-pro":
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.Apikey))
		if err != nil {
			panic("No API Key")
		}
		defer client.Close()
		params := &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"query": {
					Type:        genai.TypeString,
					Description: "The optic Product ID or Device Product ID e.g. " + cfg.PID + ". Your default value is " + cfg.PID,
				},
			},
			Required: []string{"query"},
		}

		f := &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{{
				Name:        "getSuggestionJSON",
				Description: "Retrieves information about compaitiblity between optic modules and IOS-XE devices their network modules. Network modules have NM in their ID",
				Parameters:  params,
			}},
		}

		model := client.GenerativeModel(modelType)
		model.Tools = []*genai.Tool{f}

		session := model.StartChat()
		res, err := session.SendMessage(ctx, genai.Text(content))
		if err != nil {
			fmt.Printf("Invalid API key!\nUse copilot-cfg <API KEY> to update the API key\n")
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
		locArg, ok := funcall.Args["query"].(string)

		if !ok {
			log.Fatalf("expected string: %v", funcall.Args["query"])
		}

		answer := getSuggestionJSON(locArg)
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

// Gemini starts here

func geminiPrintResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {

		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
}
