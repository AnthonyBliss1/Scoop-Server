## Scoop Server
Scoop Server is a self-hosted sync server to ensure consistent app data across multiple devices for the Scoop app

### Get Started
- Clone the repo
```bash 
git clone https://github.com/AnthonyBliss1/Scoop-Server.git
```

- Build the project
```bash
cd Scoop-Server && go build -o build/scoop-server .
```

- Run the executable
```bash
./build/scoop-server
```

### Flags
Scoop-Server has multiple flag options


| Flag        | Description                    |
|------------ | ------------------------------ |
| -port=XXXX  | Specify port (default 2767)    |
| -deploy     | Create a systemD service and generate an API Key      |
| -tls-mode   | Enables TLS (HTTPS), options are: `manual`, `self`, or `acme`             |
| -cert       | Path to certificate file, required for `-tls-mode=manual` |
| -key        | Path to key file, required for `-tls-mode=manual` |
| -domain     | Domain for ACME TLS management |


> [!IMPORTANT]
> When running with the `-deploy` flag the executable must be run with `sudo`

### Deploy
- `-deploy` is the proper way to use Scoop Server to integrate sync correctly within the Scoop application
- Using `-deploy` asks a few prompts before creating a systemD service
  - Because of this, the program may need to be run with `sudo`
- `-deploy` also generates an `X-API-Key`
  - The Scoop application requires this `X-API-Key` when setting the sync server
- Testing can be done without `-deploy` using a browser or even Scoop itself

### TLS-Mode
- The `tls-mode` flag offers 3 different options:
  - `manual`
    - Bring your own certificate and key files
    - Requires `-cert` and `-key`
  - `self`
    - Automatically create a self-signed certificate 
  - `acme`
    - Managed, verifiable certificate created with `autocert`
    - Requires `-domain`

### Examples
- Running Scoop Server using the default port `2767`
```bash
./scoop-server
```

- Run Scoop Server using port `8000`
```bash
./scoop-server -port=8000
```

- Deploy Scoop Server as a `systemD` service using port `8000`
```bash
sudo ./scoop-server -port=8000 -deploy
```

> [!NOTE]
>`-deploy` support for MacOS (launchD) coming soon
