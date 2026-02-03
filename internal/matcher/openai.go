package matcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type openAIClient struct {
	apiKey          string
	baseURL         string
	httpClient      *http.Client
	embedModel      string
	llmModel        string
	explainTopN     int
	embedBatchSize  int
	embedChunkWords int
	explainMaxChars int
	temperature     float64
}

func newOpenAIClientFromEnv() (*openAIClient, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return nil, ErrMissingOpenAIKey
	}

	baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	embedModel := strings.TrimSpace(os.Getenv("RESUMEGPT_EMBED_MODEL"))
	if embedModel == "" {
		embedModel = "text-embedding-3-large"
	}

	llmModel := strings.TrimSpace(os.Getenv("RESUMEGPT_LLM_MODEL"))
	if llmModel == "" {
		llmModel = "gpt-4o"
	}

	explainTopN := envInt("RESUMEGPT_EXPLAIN_TOPN", 20)
	embedBatch := envInt("RESUMEGPT_EMBED_BATCH", 96)
	embedChunkWords := envInt("RESUMEGPT_EMBED_CHUNK_WORDS", 2000)
	explainMaxChars := envInt("RESUMEGPT_EXPLAIN_MAX_CHARS", 12000)
	temperature := envFloat("RESUMEGPT_LLM_TEMPERATURE", 0.2)

	return &openAIClient{
		apiKey:          apiKey,
		baseURL:         strings.TrimRight(baseURL, "/"),
		httpClient:      &http.Client{Timeout: 90 * time.Second},
		embedModel:      embedModel,
		llmModel:        llmModel,
		explainTopN:     explainTopN,
		embedBatchSize:  max(1, embedBatch),
		embedChunkWords: max(500, embedChunkWords),
		explainMaxChars: max(2000, explainMaxChars),
		temperature:     clamp(temperature, 0, 1),
	}, nil
}

func envInt(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return def
	}
	return val
}

func envFloat(key string, def float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return def
	}
	return val
}

func envBool(key string, def bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (c *openAIClient) chatCompletionJSON(ctx context.Context, schemaName string, schema map[string]any, system, user string, out any) error {
	req := chatCompletionRequest{
		Model: c.llmModel,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		ResponseFormat: &responseFormat{
			Type: "json_schema",
			JSONSchema: &jsonSchema{
				Name:        schemaName,
				Description: "Return only valid JSON for the schema.",
				Schema:      schema,
				Strict:      true,
			},
		},
		Temperature: c.temperature,
	}

	var resp chatCompletionResponse
	if err := c.doJSON(ctx, "/chat/completions", req, &resp); err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return errors.New("openai: empty response")
	}
	msg := resp.Choices[0].Message
	if strings.TrimSpace(msg.Refusal) != "" {
		return fmt.Errorf("openai refusal: %s", msg.Refusal)
	}
	content := strings.TrimSpace(msg.Content)
	if content == "" {
		return errors.New("openai: empty content")
	}
	if err := json.Unmarshal([]byte(content), out); err != nil {
		return fmt.Errorf("openai: invalid json: %w", err)
	}
	return nil
}

func (c *openAIClient) embedTexts(ctx context.Context, texts []string) ([][]float64, error) {
	cleaned := make([]string, 0, len(texts))
	for _, t := range texts {
		trimmed := strings.TrimSpace(t)
		if trimmed == "" {
			trimmed = "empty"
		}
		cleaned = append(cleaned, trimmed)
	}

	out := make([][]float64, len(cleaned))
	for i := 0; i < len(cleaned); i += c.embedBatchSize {
		end := i + c.embedBatchSize
		if end > len(cleaned) {
			end = len(cleaned)
		}
		req := embeddingsRequest{
			Model: c.embedModel,
			Input: cleaned[i:end],
		}
		var resp embeddingsResponse
		if err := c.doJSON(ctx, "/embeddings", req, &resp); err != nil {
			return nil, err
		}
		for _, item := range resp.Data {
			idx := i + item.Index
			if idx >= 0 && idx < len(out) {
				out[idx] = item.Embedding
			}
		}
	}
	return out, nil
}

func (c *openAIClient) doJSON(ctx context.Context, path string, reqBody any, respBody any) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		var apiErr openAIErrorResponse
		if json.Unmarshal(raw, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("openai: %s (%s)", apiErr.Error.Message, apiErr.Error.Type)
		}
		return fmt.Errorf("openai: http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	return json.NewDecoder(resp.Body).Decode(respBody)
}

type chatCompletionRequest struct {
	Model          string           `json:"model"`
	Messages       []chatMessage     `json:"messages"`
	ResponseFormat *responseFormat   `json:"response_format,omitempty"`
	Temperature    float64          `json:"temperature,omitempty"`
	MaxTokens      int              `json:"max_tokens,omitempty"`
	Seed           int              `json:"seed,omitempty"`
	Stop           []string         `json:"stop,omitempty"`
	TopP           float64          `json:"top_p,omitempty"`
	Stream         bool             `json:"stream,omitempty"`
	Tools          []map[string]any `json:"tools,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Refusal string `json:"refusal,omitempty"`
}

type responseFormat struct {
	Type       string      `json:"type"`
	JSONSchema *jsonSchema `json:"json_schema,omitempty"`
}

type jsonSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema"`
	Strict      bool           `json:"strict,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

type embeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingsResponse struct {
	Data []embeddingData `json:"data"`
}

type embeddingData struct {
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
