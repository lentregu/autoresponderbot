package main

import (
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func TestGenerateJWT(t *testing.T) {
	// Load a test private key (or use a mock key)
	privateKey, err := loadPrivateKey("test-private-key.pem")
	if err != nil {
		t.Fatalf("Failed to load private key: %v", err)
	}

	// Define the app ID for testing
	appID := int64(12345)

	// Generate the JWT
	tokenString, err := generateJWT(appID, privateKey)
	if err != nil {
		t.Fatalf("generateJWT failed: %v", err)
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			t.Fatalf("Unexpected signing method: %v", token.Header["alg"])
		}
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v", err)
	}

	// Validate the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if iss, ok := claims["iss"].(float64); !ok || int64(iss) != appID {
			t.Errorf("Invalid issuer claim: got %v, want %v", iss, appID)
		}
		if iat, ok := claims["iat"].(float64); !ok {
			t.Errorf("Invalid iat claim: %v", claims["iat"])
		} else {
			now := time.Now().Unix()
			if int64(iat) < now-10 || int64(iat) > now+10 {
				t.Errorf("iat claim is out of range: got %v, expected near %v", iat, now)
			}
		}
		if exp, ok := claims["exp"].(float64); !ok {
			t.Errorf("Invalid exp claim: %v", claims["exp"])
		} else {
			expectedExp := time.Now().Add(time.Minute * 10).Unix()
			if int64(exp) < expectedExp-10 || int64(exp) > expectedExp+10 {
				t.Errorf("exp claim is out of range: got %v, expected near %v", exp, expectedExp)
			}
		}
	} else {
		t.Fatalf("Invalid token or claims")
	}
}

func TestHandleWebhook(t *testing.T) {
	// Create a mock private key (in practice, you would use a real or test key)
	mockPrivateKey := &rsa.PrivateKey{}

	// Create a request with a sample payload
	payload := `{"action": "opened", "issue": {"number": 1}}`
	req, err := http.NewRequest("POST", "/webhook", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a router and handle the request
	r := mux.NewRouter()
	r.HandleFunc("/webhook", handleWebhook(mockPrivateKey)).Methods("POST")
	r.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Add more assertions as needed
}

// func TestHandleWebhook(t *testing.T) {
// 	// Create a mock IssuesEvent payload
// 	payload := `{
// 		"action": "opened",
// 		"issue": {
// 			"number": 1
// 		},
// 		"repository": {
// 			"owner": {
// 				"login": "owner"
// 			},
// 			"name": "repo"
// 		},
// 		"installation": {
// 			"id": 67890
// 		}
// 	}`

// 	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
// 	w := httptest.NewRecorder()

// 	handleWebhook(w, req)

// 	resp := w.Result()
// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// }

// func TestGenerateJWT(t *testing.T) {
// 	privateKey := &rsa.PrivateKey{}
// 	appID := int64(12345)

// 	token, err := generateJWT(appID)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, token)
// }
