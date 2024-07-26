# neon-lights

A flexible load-testing/monitoring tool for Neon databases. Periocally executes specified rules and stores the results in a Postgres database.

Example of the rules:
- `{"act": "do_global_rules", "args": {}}` – load and execute all rules from the `global_rules` postgres table.
- `{"act": "create_project", "args": {"Interval": "10m"}}` – create a database in every region, if there were no projects created for the last 10 minutes
- `{"act": "delete_project", "args": {"ProjectsN": 3}}` – delete a random database in random region, if there are >3 existing databases
- `{"act": "query_project", "args": {"Scenario": "activityV1"}}` - send a SQL query to the random project

The default rule is `{"act": "do_global_rules", "args": {}, "periodic": "random(5,35)"}`, which will fetch and execute all rules from the database every 5-35 seconds.

```bash
# this will download dependencies
go mod download

# this will run existing code
go run main.go

# read .env and run the code
set -a; source .env; set +a
go run main.go | tee -a app.log

# now program should be running without errors, until Ctrl+C is pressed
```

One more example:
```bash
go run main.go '{"act": "query_project", "timeout": "10m", "args": {"Driver": [{"Weight": 1, "Item": "vercel-node=pool09"}], "Scenario": "activityV1", "RawProjectFilter": "projects.suspend_timeout_seconds = 0", "MaxRandomProjects": 1}}' | tee -a app.log
```

## Deploying

1. Get a Neon account. Don't forget to increase a limit for the projects.
2. Get a Postgres database.
3. Put env variables.
4. Run.

<details>
<summary>Development</summary>

Make sure you have:
- Go 1.16, [install](https://golang.org/doc/install)
- GoLand / VSCode / other IDE, [install goland](https://www.jetbrains.com/go/)
- golangci-lint 1.40, [install](https://golangci-lint.run/usage/install/)


### EnvFile plugin

EnvFile plugin for GoLand is useful for applying conf from .env files. Install [here](https://plugins.jetbrains.com/plugin/7861-envfile).

To use it:
- Open [Run configuration]
- Select EnvFile tab
- Add file .env from repo root
  * On macOS press shirt+cmd+. to display hidden files

### Run a linter

```
golangci-lint run --fix
```

</details>
