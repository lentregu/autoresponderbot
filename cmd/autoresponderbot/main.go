package main

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v56/github"
	"github.com/gorilla/mux"
	"github.com/lentregu/github-app-go/internal/app"
)

// loadPrivateKey loads the RSA private key from a file.
func loadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	log.Println("Loading private key...")
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, err
	}
	log.Println("Private key loaded successfully.")
	return privateKey, nil
}

// generateJWT generates a JWT token for the GitHub App.
func generateJWT(appID int64, privateKey *rsa.PrivateKey) (string, error) {
	log.Println("Generating JWT token...")
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix()
	claims["iss"] = appID

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		log.Printf("Error signing JWT token: %v", err)
		return "", err
	}
	log.Println("JWT token generated successfully.")
	return signedToken, nil
}

// handleWebhook handles incoming webhook requests from GitHub.
// We inmplement this closure for the handleWebhook to retain access to
// privateKey even after handleWebhook returns.
func handleWebhook(privateKey *rsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received webhook request.")

		var issue github.IssuesEvent
		err := json.NewDecoder(r.Body).Decode(&issue)
		if err != nil {
			log.Printf("Error decoding webhook payload: %v", err)
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		if issue.Action != nil && *issue.Action == "opened" {
			log.Printf("New issue opened: %d", issue.GetIssue().GetNumber())
			app.PostComment(issue, r, privateKey)
		} else {
			log.Println("Webhook event is not an issue opened event.")
		}
	}
}

// setupRouter initializes the router and sets up the routes.
func setupRouter(privateKey *rsa.PrivateKey) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/webhook", handleWebhook(privateKey)).Methods("POST")
	return r
}

func main() {
	// Get the server port from an environment variable, default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Load the private key
	privateKey, err := loadPrivateKey("autoresponderbot.2025-03-13.private-key.pem")
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Set up the router
	r := setupRouter(privateKey)

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
