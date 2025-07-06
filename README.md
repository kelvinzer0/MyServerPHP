# PHP Server Manager

PHP Server Manager is a simple web-based tool to manage your PHP development servers. It allows you to start, stop, and configure multiple PHP servers running on different hosts, ports, and directories.

## Features

-   Manage multiple PHP development servers.
-   Start and stop servers with a single click.
-   Configure server details including name, host, port, document root, and custom commands.
-   Cross-platform compatibility (Linux, Windows, macOS).
-   Modern and responsive UI built with Vue 3 and Tailwind CSS.

## Technologies Used

-   **Backend:** Go
-   **Frontend:** Vue 3, Vite, Tailwind CSS
-   **PHP Server:** FrankenPHP (required)

## Prerequisites

To run PHP Server Manager, you need to have **FrankenPHP** installed and accessible in your system's PATH. FrankenPHP is the underlying PHP server that PHP Server Manager controls.

### Installing FrankenPHP

You can find the official FrankenPHP installation instructions at [https://frankenphp.dev/docs/install/](https://frankenphp.dev/docs/install/).

**Quick Install Example (Linux/macOS):**

```bash
curl -L https://frankenphp.dev/install.sh | sh
sudo mv frankenphp /usr/local/bin/
```

Ensure that the `frankenphp` command is executable from your terminal after installation.

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/phpservermanager.git # Replace with your repo URL
cd phpservermanager
```

### 2. Frontend Development Setup

Navigate to the `frontend` directory, install dependencies, and start the development server.

```bash
cd frontend
npm install
npm run dev
```

This will start the frontend development server, usually on `http://localhost:5173`.

### 3. Backend Development Setup

Open a new terminal, navigate to the project root, and run the Go backend.

```bash
go run cmd/server/main.go
```

This will start the Go backend server, usually on `http://localhost:8080`.

### 4. Access the Application

Open your web browser and go to `http://localhost:5173` (or the address where your Vite development server is running). The frontend will proxy API requests to the Go backend.

## Building for Production

To build the frontend for production, navigate to the `frontend` directory and run:

```bash
cd frontend
npm run build
```

This will create an optimized `dist` directory inside `frontend/`.

The Go backend is configured to embed and serve these static files when built. To build the Go binary:

```bash
go build -o phpservermanager cmd/server/main.go
```

## Installation (Linux with systemd)

For Linux systems, you can use the provided `install.sh` script to install `phpservermanager` to `/usr/local/bin` and register it as a `systemd` service.

```bash
chmod +x install.sh
sudo ./install.sh
```

To uninstall:

```bash
chmod +x uninstall.sh
sudo ./uninstall.sh
```

## Configuration Files

Configuration files (`config.yaml` and `servers.json`) are stored in a platform-specific directory:

-   **Linux:** `/etc/phpservermanager/`
-   **macOS:** `~/Library/Application Support/phpservermanager/`
-   **Windows:** `%APPDATA%\phpservermanager\`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
