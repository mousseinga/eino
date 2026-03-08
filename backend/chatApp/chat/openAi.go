package chat

import (
	"ai-eino-interview-agent/internal/errors"
	usermodel "ai-eino-interview-agent/internal/model"
	"ai-eino-interview-agent/internal/service/common"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func CreatOpenAiChatModel(ctx context.Context, userId uint) (model.ToolCallingChatModel, error) {
	result, err := usermodel.UserModelDao.GetDefaultUserModel(int64(userId))
	if err != nil {

		return nil, errors.NewDBError("Failed to get user model", err)
	}
	apiKey, err := common.DecryptAPIKey(result.APIKeyEncrypted)
	if err != nil {
		return nil, errors.NewInternalError("Failed to decrypt API key", err)
	}
	key := apiKey

	// Validate API Key format (basic check)
	key = strings.TrimSpace(key)
	if len(key) < 10 || key == "123456" {
		return nil, errors.NewInvalidParamError("Invalid API Key detected (too short or default value). Please update your API Key in settings.")
	}

	//模型名称
	modelName := strings.TrimSpace(result.ModelKey)
	//api url
	rawURL := result.BaseURL
	// Filter out non-printable and non-ASCII characters to prevent "invisible char" issues
	url := strings.Map(func(r rune) rune {
		if r > 126 || r < 33 { // Keep only printable ASCII (33-126)
			return -1
		}
		return r
	}, rawURL)

	fmt.Printf("[OpenAI Debug] Raw BaseURL bytes: %v\n", []byte(rawURL))
	fmt.Printf("[OpenAI Debug] Cleaned URL: %q\n", url)
	// Remove /chat/completions if present (to avoid duplication when SDK adds it)
	url = strings.TrimSuffix(url, "/chat/completions")
	url = strings.TrimSuffix(url, "/") // Trim again in case it was .../chat/completions/

	// Log the configuration (masking key)
	// fmt.Printf("Creating OpenAI Chat Model: BaseURL=%s, Model=%s, Key=...%s\n", url, modelName, key[len(key)-4:])

	// Check if using Volcengine but model ID doesn't look like an endpoint ID
	if strings.Contains(url, "volces.com") && !strings.HasPrefix(modelName, "ep-") {
		// Just a warning log, or maybe we can hint the user in error message later
		// fmt.Printf("Warning: Volcengine usually requires Endpoint ID (starting with ep-) as Model, but got: %s\n", modelName)
	}

	// Create a custom HTTP client with logging and robust settings
	httpClient := &http.Client{
		Timeout: 0, // No timeout, use context deadline
		Transport: &loggingTransport{
			Transport: &http.Transport{
				DisableKeepAlives:     true,              // Disable KeepAlives to avoid EOF on idle connections
				TLSHandshakeTimeout:   10 * time.Second,  // Timeout for TLS handshake
				ResponseHeaderTimeout: 120 * time.Second, // Timeout for waiting for response headers
				ForceAttemptHTTP2:     false,             // Force HTTP/1.1 to avoid HTTP/2 framing errors
			},
		},
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:     key,
		Model:      modelName,
		BaseURL:    url,
		HTTPClient: httpClient,
	})
	if err != nil {
		// Check for specific error types and wrap them appropriately
		errMsg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMsg, "insufficient_quota") ||
			strings.Contains(errMsg, "billing_not_active") ||
			strings.Contains(errMsg, "quota_exceeded") ||
			strings.Contains(errMsg, "insufficient tokens"):
			return nil, errors.NewInsufficientTokensError("Model API: Insufficient tokens or quota exceeded. Please check your account balance.", err)

		case strings.Contains(errMsg, "rate_limit_exceeded") ||
			strings.Contains(errMsg, "too_many_requests") ||
			strings.Contains(errMsg, "rate limit"):
			return nil, errors.NewRateLimitExceededError("Model API: Rate limit exceeded. Please try again later.", err)

		case strings.Contains(errMsg, "context_length_exceeded") ||
			strings.Contains(errMsg, "maximum context length") ||
			strings.Contains(errMsg, "token limit"):
			return nil, errors.NewContextLengthExceededError("Model API: Context length exceeded. Please try with shorter input.", err)

		default:
			return nil, errors.NewOpenAIError("Failed to create OpenAI chat model", err)
		}
	}

	return chatModel, nil
}

// loggingTransport logs the actual request URL for debugging
type loggingTransport struct {
	Transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("[OpenAI Debug] Requesting: %s %s\n", req.Method, req.URL.String())

	// Peek at body if present
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			fmt.Printf("[OpenAI Debug] Body Size: %d bytes\n", len(bodyBytes))
			if len(bodyBytes) > 1000 {
				fmt.Printf("[OpenAI Debug] Body Preview: %s...\n", string(bodyBytes[:200]))
			} else {
				fmt.Printf("[OpenAI Debug] Body: %s\n", string(bodyBytes))
			}
			// Restore body
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		fmt.Printf("[OpenAI Debug] Request failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("[OpenAI Debug] Response Status: %s\n", resp.Status)
	return resp, nil
}
