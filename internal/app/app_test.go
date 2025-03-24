package app

import (
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-github/v56/github"
	"github.com/stretchr/testify/assert"
)

func TestPostComment(t *testing.T) {
	// Set up environment variables
	os.Setenv("APP_ID", "12345")
	defer os.Unsetenv("APP_ID")

	// Create a mock IssuesEvent
	event := github.IssuesEvent{
		Issue: &github.Issue{
			Number: github.Int(1),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: github.String("owner"),
			},
			Name: github.String("repo"),
		},
		Installation: &github.Installation{
			ID: github.Int64(67890),
		},
	}

	// Create a mock request
	req := httptest.NewRequest("POST", "/webhook", nil)
	w := httptest.NewRecorder()

	// Call the function
	PostComment(event, req, &rsa.PrivateKey{})

	// Check the response
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGenerateJWT(t *testing.T) {
	privateKey := &rsa.PrivateKey{}
	appID := int64(12345)

	token, err := generateJWT(appID, privateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
