package service

import (
	"bytes"
	"net/http"
)

type Service struct {
	AuthServiceURL string
	HTTPClient     *http.Client
}

type ServiceInterface interface {
	SignUp(payload []byte) (*http.Response, error)
	SignIn(payload []byte) (*http.Response, error)
}

func NewService(authServiceURL string, client *http.Client) ServiceInterface {
	return &Service{
		AuthServiceURL: authServiceURL,
		HTTPClient:     client,
	}
}

func (s *Service) SignUp(payload []byte) (*http.Response, error) {
	url := s.AuthServiceURL + "/signup"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return s.HTTPClient.Do(req)
}

func (s *Service) SignIn(payload []byte) (*http.Response, error) {
	url := s.AuthServiceURL + "/signin"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return s.HTTPClient.Do(req)
}
