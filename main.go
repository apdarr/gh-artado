package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Connection struct {
	ID string `json:"id"`
}

type Response struct {
	Value []Connection `json:"value"`
}

func main() {
	fmt.Println("hi world, this is the gh-artado extension!")

	// Set up the basic authentication credentials
	username := "alexdarr@gmail.com"
	token := "AZURE PAT"
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+token))

	// Create the HTTP client
	client := http.Client{}

	// Create the HTTP request
	request, err := http.NewRequest("GET", "https://dev.azure.com/ursa-minus/ursa/_apis/githubconnections?api-version=7.1-preview", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Add the authentication header to the request
	request.Header.Set("Authorization", authHeader)

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer response.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	// Print response body
	fmt.Println(string(body))

	// Parse JSON response
	var jsonResponse Response
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Access the ID value
	if len(jsonResponse.Value) > 0 {
		connectionID := jsonResponse.Value[0].ID
		fmt.Println("Connection ID:", connectionID)
	} else {
		fmt.Println("No connections found")
	}

	// access the value of the first connection

}
