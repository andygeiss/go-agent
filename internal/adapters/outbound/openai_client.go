package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/cloud-native-utils/stability"
	"github.com/andygeiss/go-agent/pkg/agent"
	"github.com/andygeiss/go-agent/pkg/openai"
)

// Default configuration for LLM client resilience.
const (
	defaultHTTPTimeout   = 60 * time.Second  // HTTP client timeout
	defaultLLMTimeout    = 120 * time.Second // LLM call timeout (longer for complex prompts)
	defaultRetryAttempts = 3                 // Number of retry attempts
	defaultRetryDelay    = 2 * time.Second   // Delay between retries
	defaultBreakerThresh = 5                 // Circuit breaker failure threshold
)

// The adapter translates between domain types (agent.Message, agent.ToolCall)
// and the OpenAI chat payload that LM Studio expects.

// OpenAIClient implements the agent.LLMClient interface.
// It communicates with LM Studio using the OpenAI-compatible API.
// It wraps LLM calls with resilience patterns (timeout, retry, circuit breaker).
type OpenAIClient struct {
	httpClient    *http.Client
	baseURL       string
	model         string
	llmTimeout    time.Duration
	retryAttempts int
	retryDelay    time.Duration
	breakerThresh int
}

// NewOpenAIClient creates a new OpenAIClient instance with sensible defaults.
// The client is configured with:
// - HTTP timeout: 60s.
// - LLM call timeout: 120s.
// - Retry: 3 attempts with 2s delay.
// - Circuit breaker: opens after 5 consecutive failures.
func NewOpenAIClient(baseURL, model string) *OpenAIClient {
	return &OpenAIClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		llmTimeout:    defaultLLMTimeout,
		retryAttempts: defaultRetryAttempts,
		retryDelay:    defaultRetryDelay,
		breakerThresh: defaultBreakerThresh,
	}
}

// WithHTTPClient sets a custom HTTP client for the OpenAIClient.
func (c *OpenAIClient) WithHTTPClient(httpClient *http.Client) *OpenAIClient {
	c.httpClient = httpClient
	return c
}

// WithLLMTimeout sets the timeout for LLM calls.
func (c *OpenAIClient) WithLLMTimeout(timeout time.Duration) *OpenAIClient {
	c.llmTimeout = timeout
	return c
}

// WithRetry configures retry behavior for transient failures.
func (c *OpenAIClient) WithRetry(attempts int, delay time.Duration) *OpenAIClient {
	c.retryAttempts = attempts
	c.retryDelay = delay
	return c
}

// WithCircuitBreaker configures the circuit breaker threshold.
func (c *OpenAIClient) WithCircuitBreaker(threshold int) *OpenAIClient {
	c.breakerThresh = threshold
	return c
}

// llmInput bundles the inputs for an LLM call.
type llmInput struct {
	messages []agent.Message
	tools    []agent.ToolDefinition
}

// Run sends the conversation messages to LM Studio and returns the response.
// It translates between domain types and the OpenAI-compatible API format.
// The call is wrapped with resilience patterns:
// - Timeout: prevents hanging on slow responses.
// - Retry: handles transient network failures.
// - Circuit Breaker: prevents cascading failures when LLM is down.
func (c *OpenAIClient) Run(ctx context.Context, messages []agent.Message, tools []agent.ToolDefinition) (agent.LLMResponse, error) {
	// Create the base function that performs the actual LLM call
	baseFn := func(ctx context.Context, in llmInput) (agent.LLMResponse, error) {
		apiMessages := c.convertToAPIMessages(in.messages)
		apiTools := c.convertToAPITools(in.tools)

		respPayload, err := c.sendRequest(ctx, apiMessages, apiTools)
		if err != nil {
			return agent.LLMResponse{}, err
		}

		return c.convertToResponse(respPayload)
	}

	// Wrap with stability patterns (innermost to outermost):
	// 1. Timeout - enforce maximum execution time
	// 2. Retry - handle transient failures
	// 3. Circuit Breaker - prevent cascading failures
	var fn service.Function[llmInput, agent.LLMResponse] = baseFn
	fn = stability.Timeout(fn, c.llmTimeout)
	fn = stability.Retry(fn, c.retryAttempts, c.retryDelay)
	fn = stability.Breaker(fn, c.breakerThresh)

	return fn(ctx, llmInput{messages: messages, tools: tools})
}

// convertToAPIMessages converts domain messages to API format.
func (c *OpenAIClient) convertToAPIMessages(messages []agent.Message) []openai.Message {
	apiMessages := make([]openai.Message, len(messages))
	for i, msg := range messages {
		apiMessages[i] = openai.Message{
			Role:       string(msg.Role),
			Content:    msg.Content,
			ToolCallID: string(msg.ToolCallID),
		}
		if len(msg.ToolCalls) > 0 {
			apiMessages[i].ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				apiMessages[i].ToolCalls[j] = openai.NewToolCall(
					string(tc.ID),
					tc.Name,
					tc.Arguments,
				)
			}
		}
	}
	return apiMessages
}

// sendRequest sends the chat completion request to LM Studio.
func (c *OpenAIClient) sendRequest(ctx context.Context, apiMessages []openai.Message, apiTools []openai.Tool) (*openai.ChatCompletionResponse, error) {
	reqPayload := openai.NewChatCompletionRequest(c.model, apiMessages).WithTools(apiTools)

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LM Studio returned status %d: %s", resp.StatusCode, string(body))
	}

	var respPayload openai.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &respPayload, nil
}

// convertToResponse converts the API response to domain types.
func (c *OpenAIClient) convertToResponse(respPayload *openai.ChatCompletionResponse) (agent.LLMResponse, error) {
	choice := respPayload.GetFirstChoice()
	if choice == nil {
		return agent.LLMResponse{}, errors.New("no choices in response")
	}

	domainMessage := agent.NewMessage(agent.Role(choice.Message.Role), choice.Message.Content)

	var domainToolCalls []agent.ToolCall
	if len(choice.Message.ToolCalls) > 0 {
		domainToolCalls = make([]agent.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			domainToolCalls[i] = agent.NewToolCall(
				agent.ToolCallID(tc.ID),
				tc.Function.Name,
				tc.Function.Arguments,
			)
		}
		domainMessage = domainMessage.WithToolCalls(domainToolCalls)
	}

	return agent.NewLLMResponse(domainMessage, choice.FinishReason).WithToolCalls(domainToolCalls), nil
}

// convertToAPITools converts domain tool definitions to API format.
func (c *OpenAIClient) convertToAPITools(tools []agent.ToolDefinition) []openai.Tool {
	if len(tools) == 0 {
		return nil
	}
	apiTools := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		properties := make(map[string]openai.PropertyDefinition)
		required := make([]string, 0, len(tool.Parameters))

		for _, param := range tool.Parameters {
			prop := openai.PropertyDefinition{
				Type:        string(param.Type),
				Description: param.Description,
			}
			if len(param.Enum) > 0 {
				prop.Enum = param.Enum
			}
			properties[param.Name] = prop
			if param.Required {
				required = append(required, param.Name)
			}
		}

		apiTools[i] = openai.Tool{
			Type: "function",
			Function: openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: openai.ParametersDefinition{
					Type:       "object",
					Properties: properties,
					Required:   required,
				},
			},
		}
	}
	return apiTools
}
