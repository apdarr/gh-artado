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

## WIP: Update a connection's repos to match a file: 

There are two options to connect ADO boards to GitHub Enterprise Cloud: either through a GitHub OAuth connection or a GitHub Personal Access Token (PAT). With either option, it's possible for the connection to expire. This is especially common for PAT-based ADO<->GitHub connections when the GitHub PAT expires. Currently there's no method in the ADO UI or a direct method in the API to refresh the PAT _or_ re-create the connection automatically.

This CLI method provides a workaround for these use cases: 

`gh artado graft connections-2023-10-23-16.yaml --from 4b53-a947-fcffb033f7ec --to ce189438-3344`

What this command does: 
- We read in a YAML file that contains a list of connections and their repos for a certain date. Running `gh artado ouput` will generate a timestamped YAML file that captures the state of all connections and their repos at that time. You can run this command on a regular basis, for example on a cron job or in a GitHub Action workflow, to create snapshots of your ADO connections.
- This .yaml file is then parsed as an argument to `gh artado graft`. The CLI will use the connection state described in that .yaml file to rebuild a newly created repo (described below). 
- The `--from` argument is the connection ID of the connection in the .yaml file that you want to use as the source of truth for the new connection. The intention is to reference a historical connection ID (represented in our .yaml file) and transfer its repos to to a new connection. 
- The `--to` argument represents the new connection ID. The repos listed under the `4b53-a947-fcffb033f7ec` for example will be added to the newly built `ce189438-3344` connection.  

