# mcstatusserver
A simple minecraft status server written in Golang.

# Installation
## Download executable binary
You can download the binary for your system at the [releases page](https://github.com/michivip/mcstatusserver/releases).
## Build your own version
You can get the source by using the built in `go get` command:
```
go get -t github.com/michivip/mcstatusserver
```
To build the binary just run the `go build` command:
```
go build main.go
```

# Configuration
- **address**: Address, the server will bind to (IPv4).
- **connection_timeout**: Timeout until an idle connection gets automatically closed.
- **log_file**: Path to the access log file.
- **motd**:
  - **version**:
    - **name**: Version name which will be displayed for clients with wrong versions.
    - **protocol**: The protocol version of the server.
  - **players**:
    - **max**: Maximum amount of players.
    - **online**: Amount of players online.
    - **sample**: Array of samples which will be displayed when hovering over players info/version name.
      - **name**: Name which will be displayed.
      - **id**: A unique id v4.
  - **description**:
    - **text**: MOTD text (use of color codes or \n allowed).
  - **favicon-path**: Path to the png favicon file.
- **login_attempt**: Text which is displayed on a login attempt.
  - **DisconnectText**:
    - **text**: Text which is displayed
    - **bold**: Determines whether text is bold ("true"/"false")
    - **italic**: Determines whether text is italic ("true"/"false")
    - **underlined**: Determines whether text is underlined ("true"/"false")
    - **obfuscated**: Determines whether text is obfuscated ("true"/"false")
    - **color**: The color for the text ("true"/"false")
    - **insertion**: Determines whether text is inserted ("true"/"false")
    - **extra**: Array of values with the same values like in DisconnectText

# Contributing
If you want to contribute, just open an issue. Then your issue will be discussed.

# Used libraries
Until now no external Golang library is used.