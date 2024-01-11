# `gh-artado`

A [GitHub CLI](https://cli.github.com/) extension for view Azure DevOps (ADO) connections to GitHub repositories.

What does the name mean? **a**dd-**r**epo-**t**o-**ado**: `artado`. Rhymes with cortado ☕️.

# Install 

This extension requires the [GitHub CLI](https://cli.github.com/) to be installed.

(TBD as this repo is still in private, since the APIs themselves are in private preview)


# Required env vars

- `ADO_USER` = your ADO username
- `ADO_PAT` = your ADO personal access token
- `ADO_PROJECT` = the ADO project name in which your boards are located. For example, `fabrikam/fabric`

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

## Snapshot the state of connections and their repos to a YAML file:

`gh artado output`

This command will output a YAML file that contains the state of all connections and their repos. This is useful for creating a snapshot of the state of your connections and repos at a certain time. You can use this file as an argument to `gh artado graft` to rebuild a connection and its repos. Here's an example of the output: 

```yaml
- id: 3aa9d254-413ac
  url: ""
  repository: ""
  accesstoken: ""
  authorizationheader: ""
  githubrepositoryurl: |-
    https://github.com/ursa-minus/ab
    https://github.com/ursa-minus/za
    https://github.com/ursa-minus/foobar
  name: apdarr
- id: 6f6969a7-26b0
  repository: ""
  accesstoken: ""
  authorizationheader: ""
  githubrepositoryurl: |-
    https://github.com/ursa-minus/try-foo
    https://github.com/ursa-minus/try-bar
  name: apdarr
  name: apdarr_fabrikam
```

By using the `cron_workflow.yml` file in this repo, you can run this command on a regular basis in a GitHub Actions workflow, which will thereby create a record of connections and repos at a certain time. Checkout the below section for why this is important. 

## Update a connection's repos to match a past snapshot: 

When managing the connection between ADO boards and GitHub repos, it's possible for the connection to expire. This is especially common for PAT-based ADO<->GitHub connections when the GitHub PAT expires. Currently there's no method in the ADO UI or a direct method in the API to refresh the PAT _or_ re-create the connection automatically. The alternative is to manually re-create the connection in the ADO UI, using the UI to click through the steps to re-create the connection. 

This CLI method provides a workaround for these use cases. By referencing a past snapshot (thanks to the `gh artado output` command) of the connection and its repos, we can re-create the connection and its repos.: 

`gh artado graft connections-2023-10-23-16.yml --from 3aa9d254-413ac --to ce189438-3344`

Let's dig into what this command is doing: 
- We read in a YAML file that contains a list of connections and their repos for a certain date. Running `gh artado ouput` will generate a timestamped YAML file that captures the state of all connections and their repos at that time. You can run this command on a regular basis, for example on a cron job or in a GitHub Action workflow, to create snapshots of your ADO connections.
  - At the root of this repo, you'll finde a `cron_workflow.yml` file that you can use to run `artado` on a regular basis. The workflow will run `gh artado output` and then commit the resulting YAML file to the repo from which the workflow runs.
- This .yml file is then parsed as an argument to `gh artado graft`. The CLI will use the connection state described in that .yml file to rebuild a newly created repo (described below). 
- The `--from` argument references a chosen connection ID in the .yml file. As seen above, each connection ID lists under a connection ID key. So this argument instructs the CLI to build _from_ this snapshot state _to_ a new, active connection. 
- The `--to` argument represents the new connection ID. The repos listed under the `4b53-a947-fcffb033f7ec` for example will be added to the newly built `ce189438-3344` connection.  

