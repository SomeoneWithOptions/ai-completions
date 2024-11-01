package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type Model string

type ClaudeResponse struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Role    string `json:"role,omitempty"`
	Model   string `json:"model,omitempty"`
	Content []struct {
		Type string `json:"type,omitempty"`
		Text string `json:"text,omitempty"`
	} `json:"content,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence any    `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens,omitempty"`
		OutputTokens int `json:"output_tokens,omitempty"`
	} `json:"usage,omitempty"`
}

func main() {

	var flags []string
	var input string

	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, "-") && len(arg) < 3 {
			flags = append(flags, arg)
			continue
		}
		input = arg
		break
	}

	switch {
	case slices.Contains(flags, "-s"):
		t, err := CheckText(input)
		if err != nil {
			panic(err)
		}

		fmt.Println(t)
		cmd := exec.Command("pbcopy")
		stdin, _ := cmd.StdinPipe()
		stdin.Write([]byte(t))
		stdin.Close()
		cmd.Run()
	case slices.Contains(flags, "-q"):
		t, err := AskProgrammingQuestion(input)
		if err != nil {
			panic(err)
		}
		fmt.Println(t)
	}

}

func CheckText(userMessage string) (string, error) {

	var response ClaudeResponse
	url := "https://api.anthropic.com/v1/messages"
	apiKey := os.Getenv("API_KEY")

	requestBody := map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 1024,
		"system": `Return only the enhanced version of the input text with the following improvements:
1. Fix spelling and grammar errors
2. Improve word choice with more precise and sophisticated vocabulary where appropriate
3. Correct capitalization and punctuation
4. Enhance sentence structure for better readability
4. Format paragraphs properly

Do not include any explanations, comments, or other text besides the corrected version. Output only the improved text.`,

		"messages": []map[string]string{
			{
				"role":    "user",
				"content": fmt.Sprintf("Check this text: \n %s", userMessage),
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}

	return response.Content[0].Text, nil
}

func AskProgrammingQuestion(userMessage string) (string, error) {

	var response ClaudeResponse
	url := "https://api.anthropic.com/v1/messages"
	apiKey := os.Getenv("API_KEY")

	requestBody := map[string]interface{}{
		"model":      "claude-3-5-sonnet-latest",
		"max_tokens": 1024,
		"system":     "you are an expert software and devops engineer, give short and concise answers except if explicitly asked for explanations",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": userMessage,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: \"%v\"", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}

	return response.Content[0].Text, nil
}
