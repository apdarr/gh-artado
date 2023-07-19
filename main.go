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
	rootCmd := &cobra.Command{
		Use:   "artado <subcommand> [flags]",
		Short: "gh artado",
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
			repoUrl, err := cmd.Flags().GetString("repo")
			if err != nil {
				return err
			}

			connectionID, err := cmd.Flags().GetString("connection")
			if err != nil {
				return fmt.Errorf("error retrieving connection flag: %w", err)
			}

			return runAddRepo(repoUrl, connectionID)
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

			connectionID, err := cmd.Flags().GetString("connection")
			if err != nil {
				return fmt.Errorf("error retrieving connection flag: %w", err)
			}

			return runAddBulkRepos(file, connectionID)
		},
	}

	addRepoCmd.Flags().StringP("repo", "r", "", "Repository URL to add to a given connection")
	addRepoCmd.Flags().StringP("connection", "c", "", "Connection ID to add the repo to")

	addBulkReposCmd.Flags().StringP("file", "f", "", "Text file with a list of repos to add to a given connection")
	addBulkReposCmd.Flags().StringP("connection", "c", "", "Connection ID to add the repos to")

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

	fmt.Println(adoResponse)

	var jsonResponse struct {
		Count int          `json:"count"`
		Value []Connection `json:"value"`
	}

	if err := json.Unmarshal([]byte(adoResponse), &jsonResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Debug ado response
	fmt.Println("---Debug--")
	fmt.Println(jsonResponse)
	fmt.Printf("The type of jsonResponse is %T\n", jsonResponse)
	fmt.Println("-----")

	for i, conn := range jsonResponse.Value {
		connectionUrl := fmt.Sprintf("https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections/%s/repos?api-version=7.1-preview", conn.ID)
		connectionResponse := returnURlBody("GET", connectionUrl)

		fmt.Println(connectionResponse)

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

func runAddRepo(repoUrl string, connectionID string) error {
	// handle error if URL is nil
	if repoUrl == "" {
		return fmt.Errorf("must specify a repo URL")
	}

	// Create the request body
	requestBody := struct {
		GitHubRepositoryUrls []struct {
			GitHubRepositoryUrl string `json:"gitHubRepositoryUrl"`
		} `json:"gitHubRepositoryUrls"`
		OperationType string `json:"operationType"`
	}{
		GitHubRepositoryUrls: []struct {
			GitHubRepositoryUrl string `json:"gitHubRepositoryUrl"`
		}{
			{
				GitHubRepositoryUrl: repoUrl,
			},
		},
		OperationType: "add",
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error encoding request body: %w", err)
	}

	if getToken() == "" {
		// handle error if token is not set
		return fmt.Errorf("must set ADO_TOKEN environment variable")
	}

	// Add the repo to the connection using this endpoint (POST): https://dev.azure.com/{organization}/{project}/_apis/githubconnections/{connectionId}/reposBatch?api-version=7.1-preview
	endpoint := "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections/%s/repos?api-version=7.1-preview"
	endpoint = fmt.Sprintf(endpoint, connectionID)
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBodyBytes))
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

	if resp.StatusCode == 200 {
		fmt.Printf("Added repo %s to connection %s\n", repoUrl, connectionID)

		// create a new table showing successfully adding repo to connection
		tb1 := table.NewWriter()
		tb1.SetOutputMirror(os.Stdout)
		tb1.AppendHeader(table.Row{"Connection ID", "Repo Name (added)"})
		tb1.AppendRow([]interface{}{connectionID, repoUrl})
		tb1.SetStyle(table.StyleColoredDark)
		tb1.Render()
	} else {
		fmt.Println("Error adding repo to connection")
	}

	return nil
}

func runAddBulkRepos(txtFile string, connectionID string) error {
	// Allows user to specify a text file with a list of repos to add to a given connection
	file, err := os.Open(txtFile)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var repos []string

	for scanner.Scan() {
		repoUrl := scanner.Text()
		if repoUrl != "" {
			repos = append(repos, repoUrl)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file: %w", err)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories found in file")
	}

	fmt.Printf("Adding %d repositories to connection %s\n", len(repos), connectionID)

	// Add each repo to the connection
	var addedRepos []string
	var failedRepos []string

	for _, repo := range repos {
		err := runAddRepo(repo, connectionID)
		if err != nil {
			failedRepos = append(failedRepos, repo)
		} else {
			addedRepos = append(addedRepos, repo)
		}
	}

	// Render the table of added repositories
	if len(addedRepos) > 0 {
		renderAddedReposTable(connectionID, addedRepos)
	}

	// Inform the user of any failed repositories
	if len(failedRepos) > 0 {
		fmt.Println("Failed to add the following repositories:")
		for _, repo := range failedRepos {
			fmt.Println(repo)
		}
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repos found in file")
	}

	fmt.Printf("Adding %d repositories to connection %s\n", len(repos), connectionID)

	// Add each repo to the connection

	return nil
}

func renderAddedReposTable(connectionID string, addedRepos []string) {
	// create a new table showing successfully adding repos to connection
	tb1 := table.NewWriter()
	tb1.SetOutputMirror(os.Stdout)
	tb1.AppendHeader(table.Row{"Connection ID", "Repo Name (added)"})
	for _, repo := range addedRepos {
		tb1.AppendRow([]interface{}{connectionID, repo})
	}
	tb1.SetStyle(table.StyleColoredDark)
	tb1.Render()
}
