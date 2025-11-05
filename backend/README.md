# FreezeTag Backend
This is a [Go](https://go.dev) project.
## Getting Started
Make sure you've installed the Go toolchain. This project is running Go 1.24.9, so any version at or above that should work.

Running the project should be as simple as `go run .`

## Contributing & Testing
Read our project's [Contribution Guidelines](https://capstone.cs.utah.edu/groups/freezetag/-/wikis/Contribution-Guidelines)

You'll want to install `golangci-lint` so that you can run the linter with `golangci-lint run`

If you want to check formatting, run `gofmt -l .` to see if any Go files are incorrectly formatted

You can run unit tests and get coverage using `go test -cover -tags test`.

Contributing is easiest if you use the Go extension for Visual Studio Code. It handles most of this stuff automatically (although you still need to install `golangci-lint` separately for it to work)