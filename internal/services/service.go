package services

import (
	"bytes"
	"fmt"
	"net/http"
)

type Service struct {
	HTTPClient     *http.Client
}

type ServiceInterface interface {
	SignUp(payload []byte) (*http.Response, error)
	SignIn(payload []byte) (*http.Response, error)
	ConfigRateLimiter(payload []byte) (*http.Response, error)
}

func NewService(client *http.Client) ServiceInterface {
	return &Service{
		HTTPClient:     client,
	}
}

var rateLimiterServiceURL = "http://localhost:8080"
var authServiceURL = "http://localhost:3001"

func (s *Service) SignUp(payload []byte) (*http.Response, error) {
	url := authServiceURL + "/signup"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("req: %+v\n", req)

	return s.HTTPClient.Do(req)
}

func (s *Service) SignIn(payload []byte) (*http.Response, error) {
	url := authServiceURL + "/signin"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return s.HTTPClient.Do(req)
}

func (s *Service) ConfigRateLimiter(payload []byte) (*http.Response, error) {
	url := rateLimiterServiceURL + "/config"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return s.HTTPClient.Do(req)
}
