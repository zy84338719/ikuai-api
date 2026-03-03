package ikuaisdk_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
)

func TestNewClient(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestClientLoginV4(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Action/login" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"code": 0, "message": "success"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := ikuaisdk.NewClient(server.URL, "admin", "password")
	err := client.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if client.GetVersion() != ikuaisdk.VersionV4 {
		t.Errorf("version = %v, want %v", client.GetVersion(), ikuaisdk.VersionV4)
	}
	if !client.IsLoggedIn() {
		t.Error("IsLoggedIn() should be true after successful login")
	}
}

func TestClientLoginV3(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Action/login" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"Result": 10000, "ErrMsg": ""}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := ikuaisdk.NewClient(server.URL, "admin", "password")
	err := client.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if client.GetVersion() != ikuaisdk.VersionV3 {
		t.Errorf("version = %v, want %v", client.GetVersion(), ikuaisdk.VersionV3)
	}
}

func TestClientLoginFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Action/login" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"code": 10014, "message": "password error"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := ikuaisdk.NewClient(server.URL, "admin", "wrong")
	err := client.Login(context.Background())
	if err == nil {
		t.Fatal("Login() should return error on failed login")
	}
}

func TestNewClientWithLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Action/login" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"code": 0, "message": "success"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := ikuaisdk.NewClientWithLogin(server.URL, "admin", "password", ikuaisdk.WithTimeout(10*time.Second))
	if err != nil {
		t.Fatalf("NewClientWithLogin() error = %v", err)
	}
	defer client.Close()

	if !client.IsLoggedIn() {
		t.Error("client should be logged in")
	}
}
