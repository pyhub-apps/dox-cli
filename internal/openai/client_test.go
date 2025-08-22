package openai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "Valid API key",
			apiKey:  "test-api-key",
			wantErr: false,
		},
		{
			name:    "Empty API key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_GenerateContent(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"message": "Invalid API key",
				},
			})
			return
		}

		// Parse request
		var req ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return mock response
		response := ChatCompletionResponse{
			ID:     "test-id",
			Object: "chat.completion",
			Model:  req.Model,
			Choices: []struct {
				Index   int     `json:"index"`
				Message Message `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Generated content for: " + req.Messages[1].Content,
					},
					FinishReason: "stop",
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server URL
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.apiURL = server.URL

	tests := []struct {
		name    string
		prompt  string
		options GenerateOptions
		wantErr bool
	}{
		{
			name:   "Valid generation",
			prompt: "Test prompt",
			options: GenerateOptions{
				ContentType: "blog",
				Model:       "gpt-3.5-turbo",
				MaxTokens:   100,
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name:   "Custom content type",
			prompt: "Another test",
			options: GenerateOptions{
				ContentType: "custom",
				Model:       "gpt-4",
				MaxTokens:   200,
				Temperature: 0.5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := client.GenerateContent(tt.prompt, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && content == "" {
				t.Error("GenerateContent() returned empty content")
			}
			if !tt.wantErr && content != "Generated content for: "+tt.prompt {
				t.Errorf("GenerateContent() = %v, want %v", content, "Generated content for: "+tt.prompt)
			}
		})
	}
}

func TestBuildSystemMessage(t *testing.T) {
	client, _ := NewClient("test-key")

	tests := []struct {
		contentType string
		wantContains string
	}{
		{"blog", "blog writer"},
		{"report", "business analyst"},
		{"summary", "summarization"},
		{"code", "programmer"},
		{"custom", "helpful assistant"},
		{"unknown", "helpful assistant"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			msg := client.buildSystemMessage(tt.contentType)
			if len(msg) == 0 {
				t.Error("buildSystemMessage() returned empty message")
			}
		})
	}
}