package handlers

import (
	service "apigateway/internal/services"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type Handler struct {
	service service.ServiceInterface
}

type HandlerInterface interface {
	SignUpHandler(ctx *gin.Context)
	SignInHandler(ctx *gin.Context)
	ProxyRequestHandler(ctx *gin.Context)
}

type Route struct {
	Prefix  string `yaml:"prefix"`
	Service string `yaml:"service"`
	Port    int    `yaml:"port"`
}

type Config struct {
	Routes []Route `yaml:"routes"`
}

var routeConfig Config

func (h *Handler) SignUpHandler(ctx *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding request body: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		fmt.Printf("Invalid request: email and password are required\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	payload, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
		return
	}

	resp, err := h.service.SignUp(payload)
	if err != nil {
		fmt.Printf("Error calling auth service: %+v\n", err)
		ctx.JSON(http.StatusBadGateway, gin.H{"error": err, "message": "auth service error"})
		return
	}
	defer resp.Body.Close()

	ctx.Status(resp.StatusCode)
	ctx.Writer.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	if _, copyErr := io.Copy(ctx.Writer, resp.Body); copyErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read auth service response"})
	}

	// Configure rate limiter for the new user
	var signUpResp struct {
        Data    interface{} `json:"data"`
        Success bool        `json:"success"`
        Message string      `json:"message"`
        Err     interface{} `json:"err"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&signUpResp); err != nil {
        ctx.JSON(http.StatusBadGateway, gin.H{"error": err, "message": "invalid response from auth service"})
        return
    }

	// create payload for rate limiter
	// TODO: handle constants
	rateLimiterPayload := map[string]interface{}{
		"userId": signUpResp.Data,
		"capacity": "10",
		"refillRate": "1000",
	}
	payloadBytes, err := json.Marshal(rateLimiterPayload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err,"message": "failed to create rate limiter payload"})
		return
	}

	_, err = h.service.ConfigRateLimiter(payloadBytes)
	if err != nil {
		fmt.Printf("Error calling rate limiter service: %+v\n", err)
		ctx.JSON(http.StatusBadGateway, gin.H{"error": err, "message": "rate limiter service error"})
		return
	}
}

func (h *Handler) SignInHandler(ctx *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding request body: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		fmt.Printf("Invalid request: email and password are required\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	payload, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
		return
	}

	resp, err := h.service.SignIn(payload)
	if err != nil {
		fmt.Printf("Error calling auth service: %+v\n", err)
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "auth service error"})
		return
	}
	defer resp.Body.Close()

	ctx.Status(resp.StatusCode)
	ctx.Writer.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	if _, copyErr := io.Copy(ctx.Writer, resp.Body); copyErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read auth service response"})
	}
}

func loadRoutes() {
	data, err := ioutil.ReadFile("/config/routes.yaml")
	if err != nil {
		fmt.Printf("failed to read routes config: %v", err)
		return
	}
	if err := yaml.Unmarshal(data, &routeConfig); err != nil {
		fmt.Printf("failed to parse routes.yaml: %v", err)
	}
}

func (h *Handler) ProxyRequestHandler(ctx *gin.Context) {
    loadRoutes()

    path := ctx.Request.RequestURI
	for _, rt := range routeConfig.Routes {
		if strings.HasPrefix(path, "/api"+rt.Prefix) {
			proxyToService(ctx, rt)
			return
		}
	} 
}

func proxyToService(ctx *gin.Context, rt Route) {
	targetURL := fmt.Sprintf("http://%s.default.svc.cluster.local:%d", rt.Service, rt.Port)
	target, _ := url.Parse(targetURL)

	proxy := httputil.NewSingleHostReverseProxy(target)
	
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.Host = target.Host

        originalPath := ctx.Request.URL.Path
        req.URL.Path = strings.TrimPrefix(originalPath, "/api"+rt.Prefix)
        if req.URL.Path == "" {
            req.URL.Path = "/"
        }

        req.URL.RawQuery = ctx.Request.URL.RawQuery
        req.Header.Set("X-Forwarded-Host", ctx.Request.Host)
        req.Header.Set("X-Real-IP", ctx.ClientIP())
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        fmt.Printf("proxy error: %v", err)
        ctx.JSON(http.StatusBadGateway, gin.H{"error": "upstream service unavailable"})
    }

	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}

func NewHandler(service service.ServiceInterface) HandlerInterface {
	return &Handler{
		service: service,
	}
}
