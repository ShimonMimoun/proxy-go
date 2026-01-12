package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/gin-gonic/gin"
	"proxy-go/config"
)

type BedrockHandler struct {
	Config   *config.Config
	Signer   *v4.Signer
	AWSCfg   aws.Config
}

func NewBedrockHandler(cfg *config.Config) *BedrockHandler {
	// Load AWS Config
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		log.Fatalf("Unable to load AWS config: %v", err)
	}

	return &BedrockHandler{
		Config: cfg,
		Signer: v4.NewSigner(),
		AWSCfg: awsCfg,
	}
}

func (h *BedrockHandler) Proxy(c *gin.Context) {
	// Expected path: /bedrock/:action/:modelId
	// Actions: invoke, invoke-stream, converse, converse-stream
	// Real Bedrock Runtime URLs:
	// Invoke: POST /model/{modelId}/invoke
	// InvokeStream: POST /model/{modelId}/invoke-with-response-stream
	// Converse: POST /model/{modelId}/converse
	// ConverseStream: POST /model/{modelId}/converse-stream
	
	action := c.Param("action")
	modelId := c.Param("modelId")
    
    // Map simplified actions to real Bedrock API paths
    var apiPath string
    switch action {
    case "invoke":
        apiPath = fmt.Sprintf("/model/%s/invoke", modelId)
    case "invoke-with-stream":
        apiPath = fmt.Sprintf("/model/%s/invoke-with-response-stream", modelId)
    case "converse":
        apiPath = fmt.Sprintf("/model/%s/converse", modelId)
    case "converse-stream":
        apiPath = fmt.Sprintf("/model/%s/converse-stream", modelId)
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Use invoke, invoke-with-stream, converse, or converse-stream"})
        return
    }

	// Construct target URL
	host := fmt.Sprintf("bedrock-runtime.%s.amazonaws.com", h.Config.AWSRegion)
	targetURL := fmt.Sprintf("https://%s%s", host, apiPath)

	// Read Body
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore for other middlewares if needed (though we already consumed it)

	// Create Request
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

    // Hash Body
	payloadHash := sha256.Sum256(bodyBytes)
	hexPayload := hex.EncodeToString(payloadHash[:])

	// Get Credentials
	creds, err := h.AWSCfg.Credentials.Retrieve(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve AWS credentials"})
		return
	}

	// Sign Request
	err = h.Signer.SignHTTP(c.Request.Context(), creds, req, hexPayload, "bedrock", h.Config.AWSRegion, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign request"})
		return
	}

    // Set other headers that SDK/CLI usually sends (optional but good practice)
    req.Header.Set("Content-Type", "application/json")
    
	// Execute Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to call Bedrock", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Stream response back to client
	c.Status(resp.StatusCode)
	for key, values := range resp.Header {
        for _, val := range values {
		    c.Header(key, val)
        }
	}

	io.Copy(c.Writer, resp.Body)
}
