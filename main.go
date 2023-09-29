package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateDescriptionRequest struct {
	ProductName     string   `json:"productName"`
	ProductWarnings []string `json:"productWarnings"`
}

type OpenAIRequestBody struct {
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Model       string    `json:"model,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type AmazonProduct struct {
	Price    string `json:"price"`
	ImageURL string `json:"image_url"`
	Title    string `json:"title"`
}

func handleGenerateDescription(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request for generating product description.")

	var requestBody GenerateDescriptionRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("Error in decoding the request body:", err)
		return
	}
	log.Println("Successfully decoded the request body.")

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("メルカリに出品するための商品説明文を生成してください。商品名は%sです。商品に関する注意事項として、%sを、商品説明に含めるようにしてください。文体はシンプルにして、商品説明は短めにお願いします。", requestBody.ProductName, requestBody.ProductWarnings),
		},
	}

	openAIRequest := OpenAIRequestBody{
		MaxTokens:   500,
		Temperature: 0.8,
		Model:       "gpt-3.5-turbo-0613",
		Messages:    messages,
	}

	requestBodyBytes, err := json.Marshal(openAIRequest)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error in marshaling the OpenAI request body:", err)
		return
	}
	log.Println("Successfully marshaled the OpenAI request body.")
	log.Println("OpenAI API Request:", openAIRequest)

	apiEndpoint := "https://api.openai.com/v1/chat/completions"
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error in creating the OpenAI API request:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+openAIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	log.Println("Sending request to OpenAI API...")
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error in sending the request to OpenAI API:", err)
		return
	}
	log.Println("Received response from OpenAI API.")
	defer resp.Body.Close()

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
		log.Println("Error in decoding the OpenAI API response:", err)
		return
	}
	log.Println("Successfully decoded the OpenAI API response.")
	log.Println("OpenAI API Response:", openAIResponse)

	if len(openAIResponse.Choices) > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openAIResponse.Choices[0].Message.Content)
		log.Println("Generated product description:", openAIResponse.Choices[0].Message.Content)
	} else {
		http.Error(w, "OpenAI API returned empty choices", http.StatusInternalServerError)
		log.Println("Error: OpenAI API returned empty choices")
		return
	}
}

func handleGetPrice(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("python3", "scrape_amazon.py")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Printf("Error running the python script: %s", err)
		log.Printf("Stderr: %s", stderr.String())
		log.Printf("Stdout: %s", stdout.String())
		return
	}
	var product AmazonProduct
	err = json.Unmarshal(stdout.Bytes(), &product)
	if err != nil {
		log.Printf("Error unmarshalling the python script output: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error encoding product to JSON:", err)
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	http.HandleFunc("/generate-description", handleGenerateDescription)

	http.HandleFunc("/get-price", handleGetPrice)

	log.Println("Server is starting at :8080")
	http.ListenAndServe(":8080", nil)
}