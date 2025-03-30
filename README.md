# wsgo

golang webSocket server implementation created for learning purposes

trying to follow [rfc-6455](https://datatracker.ietf.org/doc/html/rfc6455) a maximum

you can find the **autobahn-testsuite** results on [wsgo.drafted.dev](https://wsgo.drafted.dev)

## Build

To build the program you have to run:

```bash
# this will output an binary file "wsgo"
go build -o wsgo ./src
```

## Testing

### Running the Autobahn Test Suite

1. Docker is required to run the test suite
2. Make sure your wsgo server is running on port 8080 (or update the port in `./autobahn/config/fuzzingclient.json`)
3. Run the test suite:
    ```bash
    bash ./autobahn.sh
    ```

### Viewing Test Results

Open the HTML report in your browser:

```bash
# macOS
open ./autobahn/reports/servers/index.html

# Linux (most distros)
xdg-open ./autobahn/reports/servers/index.html
```

<!-- ## Getting Started -->
