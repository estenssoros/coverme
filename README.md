# coverme
runs `go test ./... -coverprofile=c.out && go tool cover -html=c.out -o=index.html` and refreshes the coverage output live in your browser


requirements

- live-server
    - `npm install -g live-server`

run `coverme` in your project directory

- creates a `test-coverage` folder
- watches your `.go` files and re-runs tests coverage as you make edits to your code
- launches live-server to watch and display changes to your test coverage