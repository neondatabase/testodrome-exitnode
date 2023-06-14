# neon-lights

TODO: write a README
```bash
# this will download dependencies
go mod download

# this will run existing code
go run main.go

# read .env and run the code
export $(cat .env | xargs) && go run main.go | tee -a app.log

# now program should be running without errors, until Ctrl+C is pressed
```

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
</details>
