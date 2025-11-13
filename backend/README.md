# FreezeTag Backend
This is a [Go](https://go.dev) project.

## Getting Started

### Requirements
* Go toolchain, this project is built on version 1.24.9.
* ImageMagick 7 (the imagick Go bindings require ImageMagick version 7+). 
    * Note: some Linux distributions still ship ImageMagick 6 by default, including Ubuntu for WSL2. This may require a build from [source](https://github.com/ImageMagick/ImageMagick).
* A C compiler and build tools.
    * [CGO](https://pkg.go.dev/cmd/cgo) is required for the ImageMagick bindings, see [here](#enabling-cgo) to enable it in this project    
* `pkg-config`.
* Your distro may require libraries like `libraw`, `libtiff`, `libheif`, and `libwebp` as well as `base-devel` for the C toolchain 

# Running  
Below are example install commands for common distributions.

### Arch Linux (Recommended)
```bash
$ sudo pacman -Syu
$ sudo pacman -S --needed go imagemagick pkgconf gcc base-devel

# libraries that ImageMagick needs that may not be installed
$ sudo pacman -S --needed libheif libraw libtiff libwebp 
```

### Ubuntu / Debian
```bash
$ sudo apt update
$ sudo apt install -y golang-go build-essential pkg-config

# Install ImageMagick and common delegate libraries. Note: older Ubuntu releases may ship ImageMagick 6 by default.
$ sudo apt install -y imagemagick libmagickcore-7*
```

If your distribution provides ImageMagick 6 by default, either upgrade to a release that supplies ImageMagick 7, use a third-party package, or build ImageMagick 7 from source.

### Enabling CGO
The `imagick` Go bindings use CGO to call the ImageMagick C API. Ensure CGO is enabled when building or running tests. This is only necessary when not using the makefile's `run` and `test` targets. Example:
```bash
# enable CGO for a single command
$ CGO_ENABLED=1 go build ./...
$ CGO_ENABLED=1 go test ./... -v

# or export for your session (the Makefile also takes care of this)
$ export CGO_ENABLED=1
```
Make sure `gcc` (or an equivalent C compiler) is installed. On MacOS, you might have to set the `CXX` and `CC` variables to the actual `gcc`, sometimes it aliases `clang` and that does not work.

### Building the project
After the dependencies are installed and `CGO_ENABLED=1` is set when needed, you can run the backend from the `backend/` directory:
```bash
# build only the necessary parts
$ make run
# build everything
$ go run ./...
```


### Testing
Many testing configs rely on [`Mockery`](https://github.com/vektra/mockery). Make sure you have version 3.6.0 or higher installed. The [`.mockery.yml`](./.mockery.yml) file is used to generate mocks of relevant interfaces, and those mocked tests can be generated via running `$ go tool mockery` in the `backend` root directory.  

Project tests can be executed with:
```bash
$ make test
# or run a specific package
$ go test ./pkg/images -v
```
- if you want code coverage, install [`gocover-cobertura`](#optional-tools) and make sure you have `$HOME/go/bin` in your PATH
    -  For bash, that can be accomplished by adding `export PATH="$HOME/go/bin:$PATH"` to `~/.bashrc` and reloading the shell
    - `./coverage/coverage.html` provides a human readable code coverage summary
        - Use [live preview](https://marketplace.visualstudio.com/items?itemName=ms-vscode.live-server) extension for viewing HTML files if using VSCode

### Troubleshooting common failures
- ERROR_MODULE / unable to load module '.../coders/dng.la' or similar: means ImageMagick could not load a coder module at runtime because a shared dependency (for example `libraw`) was missing. 
    - Install the missing `lib*` package (e.g. `libraw`) and re-run tests.
- `imagick.NewMagickWand` or other symbols not found at build time
    - Ensure the `gopkg.in/gographics/imagick.v3` Go module is present (`go mod tidy`) and [CGO](https://pkg.go.dev/cmd/cgo) is enabled with a working C compiler.

## Contributing
Read our project's [Contribution Guidelines](https://capstone.cs.utah.edu/groups/freezetag/-/wikis/Contribution-Guidelines)

You'll want to install `golangci-lint` so that you can run the linter with `golangci-lint run`.

If you want to check formatting, run `gofmt -l .` to see if any Go files are incorrectly formatted.

You can run unit tests and get coverage using `go test -cover -tags test`.

Contributing is easiest if you use the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) for Visual Studio Code. It handles most of this stuff automatically (although you still need to install `golangci-lint` separately for it to work).
