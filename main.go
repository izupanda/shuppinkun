package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
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

type GeneratedContent struct {
	ProductTitle       string `json:"productTitle"`
	ProductDescription string `json:"productDescription"`
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

type GenerateDescriptionRequest struct {
	ProductName     string   `json:"productName"`
	ProductWarnings []string `json:"productWarnings"` // 変更点
}

func handleGenerateDescription(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request for generating a product description.")

	var requestBody GenerateDescriptionRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("Error decoding the request body:", err)
		return
	}
	log.Println("Successfully decoded the request body.")

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		http.Error(w, "OPENAI_API_KEY is not set", http.StatusInternalServerError)
		log.Fatal("OPENAI_API_KEY is not set")
		return
	}

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("メルカリに出品するための商品説明文を生成してください。商品名は%sです。商品に関する注意事項として、%sを、商品説明に含めるようにしてください。文体はシンプルな箇条書きで、商品事態に関する説明は短めにお願いします。最初に【商品タイトル】とつけた後に、商品タイトルを記述してください。それ以降に【商品説明文】とつけた後に、商品説明文を記述してください。", requestBody.ProductName, requestBody.ProductWarnings),
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
		log.Println("Error marshaling the OpenAI request body:", err)
		return
	}
	log.Println("Successfully marshaled the OpenAI request body.")

	apiEndpoint := "https://api.openai.com/v1/chat/completions"
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error creating the OpenAI API request:", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openAIKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error sending the request to OpenAI API:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		http.Error(w, "OpenAI API returned an error", http.StatusInternalServerError)
		log.Printf("OpenAI API returned status code %d, body: %s", resp.StatusCode, bodyString)
		return
	}

	var openAIResponse OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResponse)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error decoding the OpenAI API response:", err)
		return
	}

	if len(openAIResponse.Choices) > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openAIResponse.Choices[0].Message.Content)
		log.Println("Generated product description:", openAIResponse.Choices[0].Message.Content)
	} else {
		http.Error(w, "OpenAI API returned empty choices", http.StatusInternalServerError)
		log.Println("OpenAI API returned empty choices")
	}
}

func handleGetPrice(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	cmd := exec.Command("python3", "scrape_amazon.py", query)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Printf("Error running the python script: %s", err)
		log.Printf("Stderr: %s", stderr.String())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	var product AmazonProduct
	err = json.Unmarshal(stdout.Bytes(), &product)
	if err != nil {
		log.Printf("Error unmarshalling the python script output: %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error encoding product to JSON:", err)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
	}

	// GitHubのシークレットトークンを使用してリクエストの真正性を検証
	signature := r.Header.Get("X-Hub-Signature")
	if !validateSignature(signature, payload, "s3cr3tT0k3nF0rW3bh00k12345") {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
	}

	// デプロイスクリプトを実行
	cmd := exec.Command("/Users/izumi_handa/Documents/01_会社\:事 業/03_事業/03_ツール開発/91_出品くん/01_shupinkun/deploy.sh")
	err = cmd.Run()
	if err != nil {
			http.Error(w, "Deployment failed", http.StatusInternalServerError)
			return
	}

	w.WriteHeader(http.StatusOK)
}


func validateSignature(signature string, payload []byte, secret string) bool {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := "sha1=" + hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}


func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	http.HandleFunc("/generate-description", handleGenerateDescription)

	http.HandleFunc("/get-price", handleGetPrice)

	log.Println("Server is starting at :8080")
	http.ListenAndServe("0.0.0.0:8080", nil)
}
