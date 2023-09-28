package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type GenerateDescriptionRequest struct {
	ProductName     string   `json:"productName"`
	ProductWarnings []string `json:"productWarnings"`
}

type OpenAIRequestBody struct {
	Prompt      string   `json:"prompt"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	Model       string   `json:"model,omitempty"`
	Messages    []string `json:"messages,omitempty"`
}

type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func handleGenerateDescription(w http.ResponseWriter, r *http.Request) {
	// Decode the request body
	var requestBody GenerateDescriptionRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("Error:", err)
		return
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	// Create a request body for OpenAI API
	openAIRequest := OpenAIRequestBody{
		Prompt:      fmt.Sprintf("フリマ向けの商品説明文を生成してください。商品名は%sで、商品に関する注意事項は、%sです。", requestBody.ProductName, requestBody.ProductWarnings),
		MaxTokens:   10,
		Temperature: 0.7,
		Model:       "gpt-3.5-turbo",
		Messages:    []string{},
	}

	requestBodyBytes, err := json.Marshal(openAIRequest)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error:", err)
		return
	}

	apiEndpoint := "https://api.openai.com/v1/chat/completions"
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error:", err)
		return
	}

	// Set API Key and content type
	req.Header.Set("Authorization", "Bearer "+openAIKey)
	req.Header.Set("Content-Type", "application/json")

	// Add Messages field to the request body
	reqBody := make(map[string]interface{})
	err = json.Unmarshal(requestBodyBytes, &reqBody)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error:", err)
		return
	}
	reqBody["messages"] = []string{}
	requestBodyBytes, err = json.Marshal(reqBody)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error:", err)
		return
	}

	req.Body = ioutil.NopCloser(bytes.NewReader(requestBodyBytes))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	// Check for errors from OpenAI API
	if resp.StatusCode >= 400 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		http.Error(w, "OpenAI API returned an error", http.StatusInternalServerError)
		log.Printf("Error: OpenAI API returned status code %d, body: %s", resp.StatusCode, bodyString)
		return
	}

	var openAIResponse OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResponse)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error:", err)
		return
	}

	if len(openAIResponse.Choices) > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openAIResponse.Choices[0].Text)
	} else {
		http.Error(w, "OpenAI API returned empty choices", http.StatusInternalServerError)
		log.Println("Error: OpenAI API returned empty choices")
		return
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static")) // `./static` は静的ファイルが格納されているディレクトリへのパス
	http.Handle("/static/", http.StripPrefix("/static", fs))

	http.HandleFunc("/generate-description", handleGenerateDescription) // あなたの既存のエンドポイント

	http.ListenAndServe(":8080", nil)

}
