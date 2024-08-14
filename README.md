# Jerusalem CLI Client

Jerusalem CLI Client is a cross-platform client for the Jerusalem Tunnel project written in Go. It allows clients to
create a secure tunnel from their localhost to a public server, making it accessible from anywhere.

## Features

- **Cross-Platform**: Works on Windows, macOS, and Linux.
- **Tunnel Creation**: Create a secure tunnel to expose localhost to the public.
- **Secure Handshake**: Complete a handshake using a secret key and unique clientID with the server.

## Installation

1. Ensure you have [Go 1.22](https://golang.org/dl/) or later installed.
2. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/jerusalem-cli-client.git
    cd jerusalem-cli-client
    ```

3. Build the project:

    ```bash
    go build -o jerusalem-cli-client -v ./...
    ```

## Usage

Start the client to create a tunnel:

    ./jerusalem-cli-client config.yaml

## Configuration

The client requires a configuration file in YAML format to run. Example `client.yaml`:

```yaml
local-host: "127.0.0.0"
local-port: "9090"
server: "0.0.0.0"
server-port: "8901"
client-id: "TEST"
secret-key: "2y6sUp8cBSfNDk7Jq5uLm0xHAIOb9ZGqE4hR1WVXtCwKjP3dYzvTn2QiFXe8rMb6"
```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.