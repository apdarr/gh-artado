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
	"time"

	"log"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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

type ConnectionFile struct {
	ID                  string   `yaml:"id"`
	GitHubRepositoryUrl []string `yaml:"githubrepositoryurl"`
	Name                string   `yaml:"name"`
}

// Used to store the data from the YAML file
type ConnectionFileData struct {
	ID                  string `yaml:"id"`
	GitHubRepositoryUrl string `yaml:"githubrepositoryurl"`
	Name                string `yaml:"name"`
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

			m, err := runAddRepo(repoUrl, connectionID)
			if err != nil {
				fmt.Println(err)
				return err
			}

			tb1 := table.NewWriter()
			tb1.SetOutputMirror(os.Stdout)
			tb1.AppendHeader(table.Row{"Connection ID", "Repo Name (added)"})

			for k, v := range m {
				// Append the rows even though there's only one
				tb1.AppendRow([]interface{}{k, v})
			}

			tb1.SetStyle(table.StyleColoredDark)
			tb1.Render()
			// Return nil as we're outputting the table
			return nil

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

			reposSlice, _, err := runAddBulkRepos(file, connectionID)
			if err != nil {
				fmt.Println(err)
				return err
			}

			tb1 := table.NewWriter()
			tb1.SetOutputMirror(os.Stdout)
			tb1.AppendHeader(table.Row{"Connection ID", "Repo Name (added)"})

			// Loop through the slice of maps and then the key value pairs of each map
			for _, m := range reposSlice {
				for k, v := range m {
					// Append the rows even though there's only one
					tb1.AppendRow([]interface{}{k, v})
				}
			}

			tb1.SetStyle(table.StyleColoredDark)
			tb1.Render()
			// Return nil as we're outputting the table
			return nil
		},
	}

	outputConnectionFileCmd := &cobra.Command{
		Use:   "output",
		Short: "Output a YAML file with the list of connections and connected repos",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := outputConnectionFile()
			if err != nil {
				return err
			}
			return nil
		},
	}

	// To-do, add flags for file, target + source connections
	graftConnectionCmd := &cobra.Command{
		Use:   "graft",
		Short: "Graft repositories from an expired connection to a newer connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := cmd.Flags().GetString("file")

			if err != nil {
				return fmt.Errorf("error retrieving file flag: %w", err)
			}
			if file == "" {
				return fmt.Errorf("file flag is required")
			}

			urls, err := graftConnection(file)

			fmt.Printf("URLs: %v", urls)
			fmt.Printf("URLS length: %v", len(urls))

			if err != nil {
				return err
			}
			return nil
		},
	}

	addRepoCmd.Flags().StringP("repo", "r", "", "Repository URL to add to a given connection")
	addRepoCmd.Flags().StringP("connection", "c", "", "Connection ID to add the repo to")

	addBulkReposCmd.Flags().StringP("file", "f", "", "Text file with a list of repos to add to a given connection")
	addBulkReposCmd.Flags().StringP("connection", "c", "", "Connection ID to add the repos to")

	graftConnectionCmd.Flags().StringP("file", "f", "", "YAML file with the list of connections and connected repos")

	rootCmd.AddCommand(listConnectionsCmd)
	rootCmd.AddCommand(addRepoCmd)
	rootCmd.AddCommand(addBulkReposCmd)
	rootCmd.AddCommand(outputConnectionFileCmd)
	rootCmd.AddCommand(graftConnectionCmd)

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

func runAddRepo(repoUrl string, connectionID string) (map[string]string, error) {
	// handle error if URL is nil
	if repoUrl == "" {
		return nil, fmt.Errorf("must specify a repo URL")
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
		return nil, fmt.Errorf("error encoding request body: %w", err)
	}

	if getToken() == "" {
		// handle error if token is not set
		return nil, fmt.Errorf("must set ADO_TOKEN environment variable")
	}

	// Add the repo to the connection using this endpoint (POST): https://dev.azure.com/{organization}/{project}/_apis/githubconnections/{connectionId}/reposBatch?api-version=7.1-preview
	endpoint := "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections/%s/repos?api-version=7.1-preview"
	endpoint = fmt.Sprintf(endpoint, connectionID)
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBodyBytes))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	defer resp.Body.Close()

	// Inform the user of the result

	if resp.StatusCode == 200 {

		// return a map of the connection ID and the repo name that was added
		m := make(map[string]string)
		m[connectionID] = repoUrl

		fmt.Printf("Successfully added repo %s to connection %s\n", repoUrl, connectionID)

		return m, nil

	} else {
		fmt.Println("Error adding repo to connection")
		err := fmt.Errorf("failed to add repo %s to connection %s", repoUrl, connectionID)
		return nil, err
	}

}

func runAddBulkRepos(txtFile string, connectionID string) ([]map[string]string, []string, error) {
	// Allows user to specify a text file with a list of repos to add to a given connection
	file, err := os.Open(txtFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening file: %w", err)
	}

	// Get the file information
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting file info: %w", err)
	}

	// Make sure the file is not empty
	if fileInfo.Size() == 0 {
		return nil, nil, fmt.Errorf("file is empty")
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
		return nil, nil, fmt.Errorf("error scanning file: %w", err)
	}

	if len(repos) == 0 {
		return nil, nil, fmt.Errorf("no repositories found in file")
	}

	fmt.Printf("Adding %d repositories to connection %s\n", len(repos), connectionID)

	// Create a slice of maps of successfully added repos
	var addedRepos []map[string]string
	var failedRepos []string

	for _, repo := range repos {
		m, err := runAddRepo(repo, connectionID)
		if err != nil {
			failedRepos = append(failedRepos, repo)
		} else {
			addedRepos = append(addedRepos, m)
		}
	}

	// Render the table of added repositories
	if len(failedRepos) > 0 {
		return nil, failedRepos, fmt.Errorf("failed to add the following repos: %v", failedRepos)
	}

	return addedRepos, nil, nil
}

func outputConnectionFile() (string, error) {
	connections, err := runListConnections()
	if err != nil {
		log.Fatal(err)
	}

	// Filter the connections, meaning, make it marshable to YAML
	var filteredConnections []ConnectionFile
	for _, c := range connections {
		urls := strings.Split(c.GitHubRepositoryUrl, "\n")
		filteredConnections = append(filteredConnections, ConnectionFile{
			ID:                  c.ID,
			GitHubRepositoryUrl: urls,
			Name:                c.Name,
		})
	}

	// Marshal the connections to YAML
	data, err := yaml.Marshal(&filteredConnections)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = os.MkdirAll("connections", 0755)
	if err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("connections/connections-%s.yml", time.Now().Format("2006-01-02-15-04-05"))

	// Write the YAML to a file.
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return string(data), nil
}

// // Consume a .yml file and add the repos to a new connection in ADO using the connection name as the key
// func graftConnection(connFile string, connSource string, connTarget string) error

func graftConnection(fileName string) ([]string, error) {
	// Read the yaml file
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	// Unmarshal the yaml file
	var connFile []ConnectionFileData
	err = yaml.Unmarshal(data, &connFile)
	if err != nil {
		return nil, err
	}

	connSource := "3aa9d254-413a-4b53-a947-fcffb033f7ec"

	connTarget := "3aa9d254-413a-4b53-a947-fcffb033f7ec"

	var urls []string
	for _, c := range connFile {
		if c.ID == connSource {
			urls = strings.Split(c.GitHubRepositoryUrl, "\n")
			break
		}
	}

	for _, url := range urls {
		_, err := runAddRepo(url, connTarget)
		if err != nil {
			return nil, err
		}
	}
	return urls, nil
}
