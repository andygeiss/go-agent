package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andygeiss/go-agent/internal/domain/agent/aggregates"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
	"github.com/andygeiss/go-agent/pkg/openai"
)

// The adapter translates between domain types (agent.Message, agent.ToolCall)
// and the OpenAI chat payload that LM Studio expects.

// OpenAIClient implements the agent.LLMClient interface.
// It communicates with LM Studio using the OpenAI-compatible API.
type OpenAIClient struct {
	httpClient *http.Client
	baseURL    string
	model      string
}

// NewOpenAIClient creates a new LMStudioClient instance.
func NewOpenAIClient(baseURL, model string) *OpenAIClient {
	return &OpenAIClient{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{},
	}
}

// WithHTTPClient sets a custom HTTP client for the LMStudioClient.
func (c *OpenAIClient) WithHTTPClient(httpClient *http.Client) *OpenAIClient {
	c.httpClient = httpClient
	return c
}

// Run sends the conversation messages to LM Studio and returns the response.
// It translates between domain types and the OpenAI-compatible API format.
func (c *OpenAIClient) Run(ctx context.Context, messages []entities.Message, tools []immutable.ToolDefinition) (aggregates.LLMResponse, error) {
	apiMessages := c.convertToAPIMessages(messages)
	apiTools := c.convertToAPITools(tools)

	respPayload, err := c.sendRequest(ctx, apiMessages, apiTools)
	if err != nil {
		return aggregates.LLMResponse{}, err
	}

	return c.convertToResponse(respPayload)
}

// convertToAPIMessages converts domain messages to API format.
func (c *OpenAIClient) convertToAPIMessages(messages []entities.Message) []openai.Message {
	apiMessages := make([]openai.Message, len(messages))
	for i, msg := range messages {
		apiMessages[i] = openai.Message{
			Role:       string(msg.Role),
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
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
func (c *OpenAIClient) convertToResponse(respPayload *openai.ChatCompletionResponse) (aggregates.LLMResponse, error) {
	choice := respPayload.GetFirstChoice()
	if choice == nil {
		return aggregates.LLMResponse{}, errors.New("no choices in response")
	}

	domainMessage := entities.NewMessage(immutable.Role(choice.Message.Role), choice.Message.Content)

	var domainToolCalls []entities.ToolCall
	if len(choice.Message.ToolCalls) > 0 {
		domainToolCalls = make([]entities.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			domainToolCalls[i] = entities.NewToolCall(
				immutable.ToolCallID(tc.ID),
				tc.Function.Name,
				tc.Function.Arguments,
			)
		}
		domainMessage = domainMessage.WithToolCalls(domainToolCalls)
	}

	return aggregates.NewLLMResponse(domainMessage, choice.FinishReason).WithToolCalls(domainToolCalls), nil
}

// convertToAPITools converts domain tool definitions to API format.
func (c *OpenAIClient) convertToAPITools(tools []immutable.ToolDefinition) []openai.Tool {
	if len(tools) == 0 {
		return nil
	}
	apiTools := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		properties := make(map[string]openai.PropertyDefinition)
		for paramName, paramDesc := range tool.Parameters {
			properties[paramName] = openai.PropertyDefinition{
				Type:        "string",
				Description: paramDesc,
			}
		}
		// Build required fields (all parameters are required by default, except "limit")
		required := make([]string, 0, len(tool.Parameters))
		for paramName := range tool.Parameters {
			if paramName != "limit" {
				required = append(required, paramName)
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
