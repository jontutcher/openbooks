# openbooks

> NOTE: Going forward only the latest release will be supported. If you encounter any issues, be sure you are using the latest version.

[![Docker Pulls](https://img.shields.io/docker/pulls/evanbuss/openbooks.svg)](https://hub.docker.com/r/evanbuss/openbooks/)

Openbooks allows you to download ebooks from irc.irchighway.net quickly and easily.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./.github/home_v3_dark.png">
  <img alt="openbooks screenshot" src="./.github/home_v3.png">
</picture>


## Getting Started

### Binary

1. Download the latest release for your platform from the [releases page](https://github.com/evan-buss/openbooks/releases).
2. Run the binary
   - Linux users may have to run `chmod +x [binary name]` to make it executable
3. `./openbooks --help`
   - This will display all possible configuration values and introduce the two modes; CLI or Server.

### Docker

- Basic config
  - `docker run -p 8080:80 evanbuss/openbooks`
- Config to persist all eBook files to disk
  - `docker run -p 8080:80 -v /home/evan/Downloads/openbooks:/books evanbuss/openbooks --persist`

### Setting the Base Path

OpenBooks server doesn't have to be hosted at the root of your webserver. The basepath value allows you to host it behind a reverse proxy. The base path value must have opening and closing forward slashes (default "/").

- Docker
  - `docker run -p 8080:80 -e BASE_PATH=/openbooks/ evanbuss/openbooks`
- Binary
  - `./openbooks server --basepath /openbooks/`

## Usage

For a complete list of features use the `--help` flags on all subcommands.
For example `openbooks cli --help or openbooks cli download --help`. There are
two modes; Server or CLI. In CLI mode you interact and download books through
a terminal interface. In server mode the application runs as a web application
that you can visit in your browser.

Double clicking the executable will open the UI in your browser. In the future it may use [webviews](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) to provide a "native-like" desktop application. 

## Development

OpenBooks uses [Task](https://taskfile.dev/) as its build system. `task --list-all`
shows everything that is available, and the full guide lives in the
[developer documentation](https://evan-buss.github.io/openbooks/developers/).

### Install the dependencies

- `task dev:init`

### Run against the mock IRC server

The mock server lets you debug requests and responses against simplified IRC / DCC
servers that mimic IRC Highway, rather than hitting the real network. Run each of
these in its own terminal:

```bash
task dev:irc     # a local ergo IRC server in Docker on port 6667
task dev:mock    # the mock search / download bot
task dev:server  # OpenBooks in server mode, pointed at the mock
task dev:client  # optional: the React client with hot reload
```

`task dev:irc-stop` removes the IRC container again. Use `task dev:cli` in place of
`task dev:server` to work on the terminal interface.

### Build the React SPA and compile binaries for multiple platforms

- `task release:build`
- This installs the npm packages, builds the React app, and compiles the executables
  into `build/`.

### Build just the Go binary (if you haven't changed the frontend)

- `cd cmd/openbooks && go build`

### Desktop App
Compile OpenBooks with experimental webview support:

``` shell
cd cmd/openbooks
go build -tags webview
```


## Why / How

- I wrote this as an easier way to search and download books from irchighway.net. It handles all the extraction and data processing for you. You just have to click the book you want. Hopefully you find it much easier than the IRC interface.
- It was also interesting to learn how the [IRC](https://en.wikipedia.org/wiki/Internet_Relay_Chat) and [DCC](https://en.wikipedia.org/wiki/Direct_Client-to-Client) protocols work and write custom implementations.

## Technology

- Backend
  - Golang
  - Chi
  - gorilla/websocket
  - Archiver (extract files from various archive formats)
- Frontend
  - React.js
  - TypeScript
  - Redux / Redux Toolkit
  - Mantine UI / @emotion/react
  - Framer Motion
