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
- Add at least one GitHub repo to your your ADO board, done so manually

# Usage 

## List connections to ADO Boards and connected repos: 

`gh artado list`

List all ADO connections to GitHub repos.

## Add a single repository to an ADO connection:

`gh artado add --repo REPO -c CONNECTION_ID`

Add a single GitHub repo to an ADO connection. For the `REPO` argument, you must provide the full repo URL. You can find the connection ID by running `gh artado list`.

## Bulk-add a selection of repositories to an ADO board connection:

 `gh artado add-bulk -f repos.txt -c 3aa9d254-413a-4b53-a947-fcffb033f7ec`

In a file specify the repos to be added to the connection, one per line. You can find the connection ID by running `gh artado list`.The repo names will the be the full URL. Be sure that the repos are on newlines and not comma separated.