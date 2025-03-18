# autoresponderbot

GitHubApp example to learn about GitHub Apps

## Overview

The `autoresponderbot` is a GitHub App that automatically responds to newly opened issues in a repository. When an issue is opened, the bot posts a comment thanking the user for opening the issue and informs them that the team will review it soon.

## Features

- Automatically responds to new issues with a predefined comment.
- Uses GitHub App authentication to securely interact with the GitHub API.

## Setup

### Prerequisites

- Go 1.16 or higher
- A GitHub App with the following permissions:
  - Issues: Read & Write
  - Metadata: Read-only
- The private key file for your GitHub App

### Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/lentregu/github-app-go.git
   cd github-app-go
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Set up environment variables:

   ```sh
   export APP_ID=<your-github-app-id>
   export PORT=8080
   ```

4. Place your GitHub App's private key file in the project directory and name it `autoresponderbot.<expiration-date>.private-key.pem`.

### Running the App

Start the server:

```sh
go run main.go
```

The server will start listening for webhook events on the specified port (default is 8080).

### Usage

1. Configure your GitHub App to send webhook events to your server's `/webhook` endpoint.
2. Open a new issue in a repository where your GitHub App is installed.
3. The bot will automatically post a comment on the new issue.

## Contributing

Feel free to submit issues and pull requests for new features, bug fixes, or improvements.

## License

This project is licensed under the MIT License.
