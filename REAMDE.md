# `gh-artado`

A [GitHub CLI](https://cli.github.com/) extension for view Azure DevOps (ADO) connections to GitHub repositories.

What does the name mean? **a**dd-**r**epo-**t**o-**ado**: `artado`. Rhymes with cortado ☕️.

# Install 

This extension requires the [GitHub CLI](https://cli.github.com/) to be installed.

(TBD as this repo is still in private, since the APIs themselves are in private preview)


# Required env vars

`ADO_USER` = your ADO username
`ADO_PAT` = your ADO personal access token

Must also configure ADO: 
- Install the GitHub ADO app to your GitHub (or even to a single repo)
- Set up a connection with GitHub in the ADO UI 
- Add at least one GitHub repo to your 

# Usage 

## `gh artado list`

List all ADO connections to GitHub repos.

## `gh artado add --repo REPO -c CONNECTION_ID`

Add a single GitHub repo to an ADO connection. For the `REPO` argument, you must provide the full repo URL. You can find the connection ID by running `gh artado list`.