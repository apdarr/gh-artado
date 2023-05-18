package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Connection struct {
	ID             string `json:"id"`
	ConnectedRepos string `json:"githubRepositoryUrl"`
}

type Response struct {
	Value []Connection `json:"value"`
}

// function for outputting body of HTTP requests, takes an API URL as input

func returnURlBody(operation, url string) string {
	username := "alexdarr@gmail.com"
	token := "r5zeftbdioarxxli2vztyv2pzmn2kmz2ouj5jtd2vvhsap4k4ioq"
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+token))

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

func main() {
	fmt.Println("hi world, this is the gh-artado extension!")

	adoResponse := returnURlBody("GET", "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections?api-version=7.1-preview")

	fmt.Println(adoResponse)

	fmt.Println("test ^^ test")

	// Parse JSON response

	var jsonResponse Response

	if err := json.Unmarshal([]byte(adoResponse), &jsonResponse); err != nil {
		fmt.Println("Error parsing JSON:", err)
	}

	if len(jsonResponse.Value) > 0 {
		connectionID := jsonResponse.Value[0].ID
		fmt.Println("Connection ID:", connectionID)

		// Get the list of repositories for the connection
		connectedRepos := "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections/%s/repos?api-version=7.1-preview"
		connectedReposUrl := fmt.Sprintf(connectedRepos, connectionID)

		adoResponse := returnURlBody("GET", connectedReposUrl)

		fmt.Println("adoResponse", adoResponse)

		if err := json.Unmarshal([]byte(adoResponse), &jsonResponse); err != nil {
			fmt.Println("Error parsing JSON:", err)
		}

		repoUrl := jsonResponse.Value[0].ConnectedRepos

		fmt.Println("Connected repos:", repoUrl)

	} else {
		fmt.Println("No connections found")
	}
}
