package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"proxy-go/config"
)

type AzureHandler struct {
	Config *config.Config
	Target *url.URL
}

func NewAzureHandler(cfg *config.Config) *AzureHandler {
	target, err := url.Parse(cfg.AzureOpenAIEndpoint)
	if err != nil {
		log.Fatalf("Invalid Azure Endpoint: %v", err)
	}
	return &AzureHandler{Config: cfg, Target: target}
}

func (h *AzureHandler) Proxy(c *gin.Context) {
	// We expect the path to be something like /azure/...
	// We need to strip /azure and forward the rest to the Azure Endpoint
    // The user might be calling /azure/openai/deployments/...
    // So we just rewrite the path.

	proxy := httputil.NewSingleHostReverseProxy(h.Target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// Set Azure API Key
		req.Header.Set("api-key", h.Config.AzureOpenAIKey)
        // Remove Authorization header from incoming request so it doesn't conflict (we use JWT, Azure uses api-key)
        req.Header.Del("Authorization")

		// Rewrite Path: /azure/X -> /X
		// Example: /azure/openai/deployments/gpt-4/chat/completions -> /openai/deployments/gpt-4/chat/completions
		req.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/azure")
        
        // Ensure Host is set correctly for SSL/TLS
		req.Host = h.Target.Host
	}

    // Error handler
    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        log.Printf("Azure Proxy Error: %v", err)
        w.WriteHeader(http.StatusBadGateway)
    }

	proxy.ServeHTTP(c.Writer, c.Request)
}
