package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://generativelanguage.googleapis.com/v1beta/models"

// Client wraps the Gemini API client
type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient creates a new Gemini client
func NewClient(ctx context.Context, apiKey string, model string) (*Client, error) {
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
	}, nil
}

// GenerateContentRequest represents a request to generate content
type GenerateContentRequest struct {
	SystemPrompt string
	Messages     []Message
	UserMessage  string
}

// Message represents a chat message
type Message struct {
	Role    string
	Content string
}

// GenerateContentResponse represents the response from Gemini
type GenerateContentResponse struct {
	Text         string
	PromptTokens int
	OutputTokens int
	TotalTokens  int
}

// InsightResponse represents the structured insight response
type InsightResponse struct {
	Analysis string `json:"analysis"`
}

// API request/response structures
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata *usageMetadata    `json:"usageMetadata,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GenerateContent sends a request to Gemini and returns the response
func (c *Client) GenerateContent(ctx context.Context, req GenerateContentRequest) (*GenerateContentResponse, error) {
	var contents []geminiContent

	// Add system instruction as the first message
	if req.SystemPrompt != "" {
		contents = append(contents, geminiContent{
			Role: "user",
			Parts: []geminiPart{
				{Text: "SYSTEM INSTRUCTION: " + req.SystemPrompt + "\n\nPlease acknowledge this instruction."},
			},
		})
		contents = append(contents, geminiContent{
			Role: "model",
			Parts: []geminiPart{
				{Text: "I understand and will follow these instructions."},
			},
		})
	}

	// Add conversation history
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: msg.Content},
			},
		})
	}

	// Add current user message
	contents = append(contents, geminiContent{
		Role: "user",
		Parts: []geminiPart{
			{Text: req.UserMessage},
		},
	})

	// Build request
	reqBody := geminiRequest{
		Contents: contents,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", baseURL, c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text

	// Extract token usage
	var promptTokens, outputTokens, totalTokens int
	if geminiResp.UsageMetadata != nil {
		promptTokens = geminiResp.UsageMetadata.PromptTokenCount
		outputTokens = geminiResp.UsageMetadata.CandidatesTokenCount
		totalTokens = geminiResp.UsageMetadata.TotalTokenCount
	}

	return &GenerateContentResponse{
		Text:         text,
		PromptTokens: promptTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
	}, nil
}

// GenerateInsight generates structured insights about a conversation
func (c *Client) GenerateInsight(ctx context.Context, messages []Message) (*InsightResponse, error) {
	// Build prompt for insight generation
	conversationText := "Analyze the following conversation and provide structured insights:\n\n"
	for _, msg := range messages {
		conversationText += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	conversationText += "\n\nProvide insights in the following format:\n"
	conversationText += "1. Main topics discussed\n"
	conversationText += "2. User's primary concerns or questions\n"
	conversationText += "3. Sentiment analysis\n"
	conversationText += "4. Suggested follow-up topics\n"
	conversationText += "\nIMPORTANT: Respond as an analyst, NOT in character. Be objective and analytical."

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: conversationText},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", baseURL, c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	return &InsightResponse{
		Analysis: geminiResp.Candidates[0].Content.Parts[0].Text,
	}, nil
}

// Close closes the Gemini client
func (c *Client) Close() error {
	return nil
}
