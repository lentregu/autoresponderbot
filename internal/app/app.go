package app

import (
	"crypto/rsa"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v56/github"
)

func PostComment(event github.IssuesEvent, r *http.Request, privateKey *rsa.PrivateKey) {
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
		log.Printf("No installation data found in event for issue #%d in repository %s/%s.",
			event.GetIssue().GetNumber(), event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName())
		return
	}
	installationID := installation.GetID()

	log.Printf("Generating JWT for App ID: %d", appID)

	// Generate JWT using the private key and the App ID
	jwtToken, err := generateJWT(appID, privateKey)
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

func generateJWT(appID int64, privateKey *rsa.PrivateKey) (string, error) {
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
