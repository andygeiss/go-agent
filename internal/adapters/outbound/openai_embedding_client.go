package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/cloud-native-utils/stability"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/openai"
)

// Default configuration for embedding client (alphabetically sorted).
const (
	defaultEmbeddingModel   = "text-embedding-3-small"
	defaultEmbeddingTimeout = 30 * time.Second
)

// OpenAIEmbeddingClient implements the agent.EmbeddingClient interface.
// It communicates with embedding APIs using the OpenAI-compatible format.
type OpenAIEmbeddingClient struct {
	httpClient    *http.Client
	logger        *slog.Logger
	baseURL       string
	model         string
	timeout       time.Duration
	retryDelay    time.Duration
	breakerThresh int
	retryAttempts int
}

// NewOpenAIEmbeddingClient creates a new OpenAIEmbeddingClient instance.
func NewOpenAIEmbeddingClient(baseURL string) *OpenAIEmbeddingClient {
	return &OpenAIEmbeddingClient{
		httpClient: &http.Client{
			Timeout: defaultEmbeddingTimeout,
		},
		baseURL:       baseURL,
		model:         defaultEmbeddingModel,
		timeout:       defaultEmbeddingTimeout,
		retryDelay:    defaultRetryDelay,
		breakerThresh: defaultBreakerThresh,
		retryAttempts: defaultRetryAttempts,
	}
}

// WithHTTPClient sets a custom HTTP client.
func (c *OpenAIEmbeddingClient) WithHTTPClient(httpClient *http.Client) *OpenAIEmbeddingClient {
	c.httpClient = httpClient
	return c
}

// WithLogger sets the logger for the client.
func (c *OpenAIEmbeddingClient) WithLogger(logger *slog.Logger) *OpenAIEmbeddingClient {
	c.logger = logger
	return c
}

// WithModel sets the embedding model to use.
func (c *OpenAIEmbeddingClient) WithModel(model string) *OpenAIEmbeddingClient {
	c.model = model
	return c
}

// WithRetry configures retry behavior for transient failures.
func (c *OpenAIEmbeddingClient) WithRetry(attempts int, delay time.Duration) *OpenAIEmbeddingClient {
	c.retryAttempts = attempts
	c.retryDelay = delay
	return c
}

// WithTimeout sets the timeout for embedding calls.
func (c *OpenAIEmbeddingClient) WithTimeout(timeout time.Duration) *OpenAIEmbeddingClient {
	c.timeout = timeout
	return c
}

// Embed generates an embedding vector for the given text.
func (c *OpenAIEmbeddingClient) Embed(ctx context.Context, text string) (agent.Embedding, error) {
	// Create the base function that performs the actual embedding call
	baseFn := func(ctx context.Context, input string) (agent.Embedding, error) {
		return c.doEmbed(ctx, input)
	}

	// Wrap with stability patterns (innermost to outermost):
	// 1. Timeout - enforce maximum execution time
	// 2. Retry - handle transient failures
	// 3. Circuit Breaker - prevent cascading failures
	var fn service.Function[string, agent.Embedding] = baseFn
	fn = stability.Timeout(fn, c.timeout)
	fn = stability.Retry(fn, c.retryAttempts, c.retryDelay)
	fn = stability.Breaker(fn, c.breakerThresh)

	return fn(ctx, text)
}

// doEmbed performs the actual embedding API call.
func (c *OpenAIEmbeddingClient) doEmbed(ctx context.Context, text string) (agent.Embedding, error) {
	request := openai.NewEmbeddingRequest(c.model, text)
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/v1/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var response openai.EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	embedding := response.GetFirstEmbedding()
	if embedding == nil {
		return nil, errors.New("empty embedding response")
	}

	c.logSuccess(len(embedding), response.Usage.TotalTokens)

	return embedding, nil
}

// logSuccess logs a successful embedding request.
func (c *OpenAIEmbeddingClient) logSuccess(dimensions, tokens int) {
	if c.logger != nil {
		c.logger.Debug("embedding generated",
			slog.Int("dimensions", dimensions),
			slog.Int("tokens", tokens),
		)
	}
}
