package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

type Connection struct {
	ID             string `json:"id"`
	ConnectedRepos string `json:"githubRepositoryUrl"`
	Name           string `json:"name"`
}

type Response struct {
	Value []Connection `json:"value"`
}

type GitHubRepository struct {
	GitHubRepositoryUrl string `json:"gitHubRepositoryUrl"`
}

type ReposBatchRequest struct {
	GitHubRepositoryUrls []GitHubRepository `json:"gitHubRepositoryUrls"`
	OperationType        string             `json:"operationType"`
}

func getToken() string {
	token := os.Getenv("ADO_TOKEN")
	return token
}

// function for outputting body of HTTP requests, takes an API URL as input
func returnURlBody(operation, url string) string {
	username := "alexdarr@gmail.com"
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+getToken()))

	// Create the HTTP client
	client := http.Client{}

	fmt.Println(operation, url)

	// Create the HTTP request
	request, err := http.NewRequest(operation, url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	// Add the authentication header to the request
	request.Header.Set("Authorization", authHeader)

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}
	defer response.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
	}

	return string(body)
}

func _main() error {
	var repo repository.Repository

	rootCmd := &cobra.Command{
		Use:   "artado <subcommand> [flags]",
		Short: "gh artado",
	}

	repoOverride := rootCmd.PersistentFlags().StringP("repo", "r", "", "Repository to use in OWNER/REPO format")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		if *repoOverride != "" {
			repo, err = repository.Parse(*repoOverride)
		} else {
			repo, err = repository.Current()
		}
		return
	}

	//var tokenValue string
	listConnectionsCmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List GitHub connection IDs for a given Azure DevOps board",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			connectionID, err := runListConnections()
			if err != nil {
				return err
			}
			fmt.Printf("Connection ID: %s\n", connectionID)
			return nil
		},
	}

	// listConnectionsCmd.Flags().StringVarP(&tokenValue, "token", "t", "", "Azure DevOps Personal Access Token")

	// Get the token value for later retrieval: tokenValue, _ := tokenCmd.Flags().GetString("token")

	addRepoCmd := &cobra.Command{
		Use:   "add [flags]",
		Short: "Add a repo to a given connection",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fmt.Printf("Added repo %s to connection\n", repo)
			return runAddRepo(repoOverride)
		},
	}

	rootCmd.AddCommand(listConnectionsCmd)
	rootCmd.AddCommand(addRepoCmd)

	return rootCmd.Execute()
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, "X %s", err.Error())
	}
}

func runListConnections() (string, error) {
	// handle error if token is not set
	if getToken() == "" {
		fmt.Errorf("must set ADO_TOKEN environment variable")
	}

	adoResponse := returnURlBody("GET", "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections?api-version=7.1-preview")

	// Parse JSON response

	var jsonResponse Response

	if err := json.Unmarshal([]byte(adoResponse), &jsonResponse); err != nil {
		return "", fmt.Errorf("error parsing JSON: %w", err)
	}

	if len(jsonResponse.Value) > 0 {
		connectionID := jsonResponse.Value[0].ID
		repoName := jsonResponse.Value[0].Name
		fmt.Println("Connection ID:", connectionID)

		// Get the list of repositories for the connection
		connectedRepos := "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections/%s/repos?api-version=7.1-preview"
		connectedReposUrl := fmt.Sprintf(connectedRepos, connectionID)

		adoResponse := returnURlBody("GET", connectedReposUrl)

		fmt.Println("adoResponse", adoResponse)

		//repoUrl := jsonResponse.Value[0].ConnectedRepos

		fmt.Println("Connected repos:", repoName)

		return connectionID, nil

	} else {
		return "", fmt.Errorf("no connections found")
	}
}

func runAddRepo(repoOverride *string) error {
	if repoOverride == nil || *repoOverride == "" {
		return fmt.Errorf("must set --repo flag")
	}

	repo, err := repository.Parse(*repoOverride)
	if err != nil {
		return err
	}

	// Get the connection ID

	connectionID, err := runListConnections()
	if err != nil {
		// Handle the error
		fmt.Println("Error:", err)
		return err
	}

	requestBody := ReposBatchRequest{
		GitHubRepositoryUrls: []GitHubRepository{
			{GitHubRepositoryUrl: *repoOverride},
		},
		OperationType: "add",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	// Add the repo to the connection using this endpoint (POST): https://dev.azure.com/{organization}/{project}/_apis/githubconnections/{connectionId}/reposBatch?api-version=7.1-preview
	endpoint := "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections/%s/repos?api-version=7.1-preview"
	endpoint = fmt.Sprintf(endpoint, connectionID)
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	// Set the content type header, as well as the authorization header
	username := "alexdarr@gmail.com"
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+getToken()))

	request.Header.Set("Authorization", authHeader)
	request.Header.Set("Content-Type", "application/json")

	// Create the HTTP client
	client := http.Client{}
	resp, err := client.Do(request)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Inform the user of the result

	success := "Added %s to connection"
	if resp.StatusCode == 200 {
		fmt.Printf(success, repo)
	} else {
		fmt.Println("Error adding repo to connection")
	}

	return nil
}
