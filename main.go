package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

type Connection struct {
	ID                  string `json:"id"`
	Url                 string `json:"url"`
	Repository          string `json:"repository"`
	AccessToken         string `json:"accessToken"`
	AuthorizationHeader string `json:"authorizationHeader"`
	GitHubRepositoryUrl string `json:"gitHubRepositoryUrl"`
	Name                string `json:"name"`
}

type Response struct {
	Value []struct {
		ID                string `json:"id"`
		Name              string `json:"name"`
		GitHubConnections []struct {
			GitHubRepositoryUrl string `json:"gitHubRepositoryUrl"`
		} `json:"gitHubConnections"`
	} `json:"value"`
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

func getUsername() string {
	username := os.Getenv("ADO_USERNAME")
	return username
}

// function for outputting body of HTTP requests, takes an API URL as input
func returnURlBody(operation, url string) string {
	username := getUsername()
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+getToken()))

	// Create the HTTP client
	client := http.Client{}

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

	listConnectionsCmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List GitHub connection IDs for a given Azure DevOps board",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			connections, err := runListConnections()
			if err != nil {
				return err
			}

			// Print the table
			tb1 := table.NewWriter()
			tb1.SetOutputMirror(os.Stdout)
			tb1.AppendHeader(table.Row{"Connection ID", "Connection Name", "Repo Name"})

			for _, connection := range connections {
				tb1.AppendRow([]interface{}{connection.ID, connection.Name, connection.GitHubRepositoryUrl})
			}
			tb1.SetStyle(table.StyleColoredDark)
			tb1.Render()

			return nil
		},
	}

	addRepoCmd := &cobra.Command{
		Use:   "add [flags]",
		Short: "Add a repo to a given connection",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fmt.Printf("Added repo %s to connection\n", repo)
			return runAddRepo(repoOverride)
		},
	}

	addBulkReposCmd := &cobra.Command{
		Use:   "add-bulk [flags]",
		Short: "Specify a text file with a list of repos to add to a given connection",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			file, err := cmd.Flags().GetString("file")
			if err != nil {
				return fmt.Errorf("error retrieving file flag: %w", err)
			}
			if file == "" {
				return fmt.Errorf("file flag is required")
			}
			return runAddBulkRepos(file)
		},
	}

	addBulkReposCmd.Flags().StringP("file", "f", "", "Text file with a list of repos to add to a given connection")

	rootCmd.AddCommand(listConnectionsCmd)
	rootCmd.AddCommand(addRepoCmd)
	rootCmd.AddCommand(addBulkReposCmd)

	return rootCmd.Execute()
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, "X %s", err.Error())
	}
}

func runListConnections() ([]Connection, error) {
	// handle error if token is not set
	if getToken() == "" {
		return nil, fmt.Errorf("must set ADO_TOKEN environment variable")
	}

	adoResponse := returnURlBody("GET", "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections?api-version=7.1-preview")

	var jsonResponse struct {
		Count int          `json:"count"`
		Value []Connection `json:"value"`
	}

	if err := json.Unmarshal([]byte(adoResponse), &jsonResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	for i, conn := range jsonResponse.Value {
		connectionUrl := fmt.Sprintf("https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections/%s/repos?api-version=7.1-preview", conn.ID)
		connectionResponse := returnURlBody("GET", connectionUrl)

		var connection struct {
			Name string `json:"name"`
		}

		if err := json.Unmarshal([]byte(connectionResponse), &connection); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %w", err)
		}

		jsonResponse.Value[i].Name = connection.Name

		// Get the list of respitories connected to the connection
		connectedReposUrl := fmt.Sprintf("https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections/%s/repos?api-version=7.1-preview", conn.ID)
		connectedReposResponse := returnURlBody("GET", connectedReposUrl)

		var connectedRepos struct {
			Value []struct {
				GitHubRepositoryUrl string `json:"gitHubRepositoryUrl"`
			} `json:"value"`
		}

		if err := json.Unmarshal([]byte(connectedReposResponse), &connectedRepos); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %w", err)
		}

		repoUrls := make([]string, 0, len(connectedRepos.Value))

		for _, repo := range connectedRepos.Value {
			repoUrls = append(repoUrls, repo.GitHubRepositoryUrl)
		}

		jsonResponse.Value[i].GitHubRepositoryUrl = strings.Join(repoUrls, "\n")
		jsonResponse.Value[i].Name = conn.Name
	}

	return jsonResponse.Value, nil
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

	println("connectionID for runAddRepo", connectionID)

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
	username := getUsername()
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

func runAddBulkRepos(txtFile string) error {
	// Allows user to specify a text file with a list of repos to add to a given connection
	file, err := os.Open(txtFile)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the file into a byte slice
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repoUrl := strings.TrimSpace(scanner.Text())
		// call the runAddRepo function for each repo in the file

		url := repoUrl
		fmt.Println("Adding repo:", url)
		if err := runAddRepo(&url); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}
