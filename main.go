package main

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v56/github"
	"github.com/gorilla/mux"
)

var privateKey *rsa.PrivateKey

func init() {
	log.Println("Loading private key...")
	keyData, err := os.ReadFile("autoresponderbot.2025-03-13.private-key.pem")
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}
	// parses a PEM encoded PKCS1 or PKCS8 private key
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	log.Println("Private key loaded successfully.")
}

func generateJWT(appID int64) (string, error) {
	log.Println("Generating JWT token...")
	// Creates a new token using the R256 signing method that is the one used by GitHub Apps.
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix()
	claims["iss"] = appID

	// The token is signed with the GitHub App private key.
	// This is essential for GitHub to verify the token authenticity
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		log.Printf("Error signing JWT token: %v", err)
		return "", err
	}
	log.Println("JWT token generated successfully.")
	return signedToken, nil
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("Received webhook request.")

	// As we are monitoring issues, we look for an IssuesEvent
	var issue github.IssuesEvent
	err := json.NewDecoder(r.Body).Decode(&issue)
	if err != nil {
		log.Printf("Error decoding webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if issue.Action != nil && *issue.Action == "opened" {
		log.Printf("New issue opened: %d", issue.GetIssue().GetNumber())
		postComment(issue, r)
	} else {
		log.Println("Webhook event is not an issue opened event.")
	}
}

func postComment(event github.IssuesEvent, r *http.Request) {
	// Retrieve the App ID from an environment variable
	appIDStr := os.Getenv("APP_ID")
	if appIDStr == "" {
		log.Println("APP_ID environment variable is not set.")
		return
	}

	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		log.Printf("Error parsing APP_ID: %v", err)
		return
	}

	installation := event.GetInstallation()
	if installation == nil {
		log.Println("No installation data found in event.")
		return
	}
	installationID := installation.GetID()

	log.Printf("Generating JWT for App ID: %d", appID)

	// Generate JWT using the private key and the App ID
	jwtToken, err := generateJWT(appID)
	if err != nil {
		log.Println("Failed to generate JWT token.")
		return
	}

	client := github.NewClient(nil).WithAuthToken(jwtToken)
	ctx := r.Context()

	log.Printf("Requesting installation token for ID: %d", installationID)
	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, &github.InstallationTokenOptions{})
	if err != nil {
		log.Printf("Error getting installation token: %v", err)
		return
	}
	log.Println("Installation token acquired.")

	client = github.NewClient(nil).WithAuthToken(token.GetToken())
	comment := &github.IssueComment{
		Body: github.String("¡Gracias por abrir este issue! Nuestro equipo lo revisará pronto. 🚀"),
	}

	log.Printf("Posting comment on issue #%d in %s/%s", event.GetIssue().GetNumber(), event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName())
	_, _, err = client.Issues.CreateComment(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), event.GetIssue().GetNumber(), comment)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
	} else {
		log.Println("Comment posted successfully.")
	}
}

func main() {
	// Get the server port from an environment variable, default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize the router and define the webhook route
	r := mux.NewRouter()
	r.HandleFunc("/webhook", handleWebhook).Methods("POST")

	log.Printf("Server listening on port %s", port)
	// Start the server
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
