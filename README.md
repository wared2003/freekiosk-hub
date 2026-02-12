# FreeKiosk Hub

FreeKiosk Hub is a central server designed to manage and monitor a fleet of kiosk devices (tablets). It leverages Tailscale for secure, zero-configuration networking, allowing you to manage your devices from anywhere.

## Features

- **Web Dashboard:** A clean and simple web interface to view and manage all your kiosks.
- **Device Management:** Track the status, configuration, and health of each connected device.
- **Group Management:** Organize your kiosks into logical groups for easier management.
- **Secure Networking:** Uses Tailscale's secure network layer for all communications.
- **Real-time Monitoring:** Employs Server-Sent Events (SSE) for live status updates.
- **Extensible:** Built with a modular structure in Go for easy extension.

## Architecture

This project is the **server-side** component of the FreeKiosk ecosystem. It requires a client application to be running on the managed tablets.

The official and compatible client is **[FreeKiosk](https://github.com/RushB-fr/freekiosk/)**, which should be installed on each tablet you wish to manage.

## Prerequisites

Before you begin, ensure you have the following installed:
- **Go:** The programming language used for the project.
- **Make:** To use the provided Makefile for common tasks.
- **[Templ](https://templ.guide/):** For generating Go code from templates.

Install Templ with the following command:
```sh
go install github.com/a-h/templ/cmd/templ@latest
```

## Getting Started

### 1. Clone the Repository

```sh
git clone https://github.com/your-username/freekiosk-hub.git
cd freekiosk-hub
```

### 2. Install Dependencies

Run the following command to download and tidy the Go modules:

```sh
make deps
```

### 3. Configuration

Configuration is managed via environment variables or a `.env` file in the project root.

Create a `.env` file by copying the example below:

```dotenv
# .env.example

# -- Server Configuration --
SERVER_PORT=8081
LOG_LEVEL=INFO # DEBUG, INFO, WARN, ERROR

# -- Database --
DB_PATH=freekiosk.db
RETENTION_DAYS=31 # How long to keep historical data

# -- Kiosk Communication --
KIOSK_PORT=8080
KIOSK_API_KEY=your-secret-api-key # A shared secret between the hub and kiosks

# -- Tailscale Integration --
# Required for fetching device information from your Tailnet.
# Create an API key from the Tailscale admin console: https://login.tailscale.com/admin/settings/keys
TS_AUTHKEY=tskey-auth-your-key-here...

# -- Performance --
POLL_INTERVAL=30s
MAX_WORKERS=5
```

### Environment Variables

| Variable         | Description                                                 | Required | Default        |
|------------------|-------------------------------------------------------------|----------|----------------|
| `SERVER_PORT`    | The port for the FreeKiosk Hub web interface.               | No       | `8081`         |
| `DB_PATH`        | Path to the SQLite database file.                           | No       | `freekiosk.db` |
| `TS_AUTHKEY`     | Your Tailscale API authentication key.                      | **Yes**  | -              |
| `LOG_LEVEL`      | The application log level (`DEBUG`, `INFO`, `WARN`, `ERROR`). | No       | `INFO`         |
| `KIOSK_PORT`     | The port on which the kiosk client API runs.                | No       | `8080`         |
| `KIOSK_API_KEY`  | A shared API key to authenticate requests from kiosks.      | No       | -              |
| `POLL_INTERVAL`  | The interval for polling device statuses.                   | No       | `30s`          |
| `RETENTION_DAYS` | How many days of historical report data to retain.          | No       | `31`           |
| `MAX_WORKERS`    | Number of concurrent workers for polling device statuses.   | No       | `5`            |


## Usage

### Generate UI Components

The UI is built using `templ`. You must generate the Go code from the template files before building or running the application.

```sh
make generate
```

### Build the Application

Compile the application into a single binary located in the `bin/` directory.

```sh
make build
```

### Run the Application

Start the server directly for development purposes. The server will automatically reload if you make changes to the Go files.

```sh
make run
```

Once running, you can access the web dashboard at **http://localhost:8081**.

## Project Structure

The project is organized into several key directories:

- `cmd/server/`: The main entry point for the web server.
- `internal/`: Contains the core application logic, separated by domain.
  - `api/`: HTTP handlers and router setup.
  - `config/`: Environment variable loading and application configuration.
  - `database/`: Database connection and schema management.
  - `models/`: Core data structures.
  - `repositories/`: Data access layer for interacting with the database.
  - `services/`: Business logic and coordination.
  - `network/`: Clients for external services like Tailscale.
- `static/`: Static assets (CSS, JS, images) served by the web server.
- `ui/`: `.templ` files for the web interface components.
- `Makefile`: Contains helper commands for building, running, and cleaning the project.

## License

This project is licensed under the GNU AGPLv3. See the `LICENSE` file for details.
