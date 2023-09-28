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
	ProductName     string `json:"productName"`
	ProductWarnings string `json:"productWarnings"`
}

type GenerateDescriptionResponse struct {
	Description string `json:"description"`
}

type OpenAIRequestBody struct {
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

func handleGenerateDescription(w http.ResponseWriter, r *http.Request) {
	// リクエストボディをパースする
	var requestBody GenerateDescriptionRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("Error: ", err)
		return
	}

// OpenAI APIのためのリクエストボディを作成する
openAIRequest := OpenAIRequestBody{
	Prompt:      fmt.Sprintf("フリマ向けの商品説明文を生成してください。商品名は%sで、商品に関する注意事項は、%sです。", requestBody.ProductName, requestBody.ProductWarnings),
	MaxTokens:   100,  // 生成するテキストの最大トークン数。必要に応じて調整してください。
	Temperature: 0.7,  // 生成するテキストの「ランダム性」を制御するパラメータ。0～1の範囲で設定。
}



	requestBodyBytes, err := json.Marshal(openAIRequest)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error: ", err)
		return
	}

	apiEndpoint := "https://api.openai.com/v1/chat/completions"

	// デバッグ: 送信するリクエストボディを出力
	log.Printf("Sending this request body to OpenAI API: %s\n", string(requestBodyBytes))

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error: ", err)
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		http.Error(w, "OpenAI API returned an error", http.StatusInternalServerError)
		log.Printf("Error: OpenAI API returned status code %d, body: %s", resp.StatusCode, bodyString)
		return
	}

	// OpenAI APIから返されたレスポンスをパースする
	var responseBody GenerateDescriptionResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error: ", err)
		return
	}

	// レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseBody)
}

func main() {
	http.HandleFunc("/generate-description", handleGenerateDescription)

	fmt.Println("Server is running on http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("Error starting the server: ", err)
		return
	}
}
