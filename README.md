## Code Structure

```
.
├── Dockerfile             # Builds Go backend API
├── Dockerfile.frontend    # Builds Nginx frontend server
├── README.md              # Project documentation
├── compose.yaml           # Docker Compose definition for both services
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── nginx.conf             # Nginx configuration used by frontend container
├── pizza-mix.go           # Go backend source code (API only)
└── static/                # Root for static frontend assets
    ├── index.html         # Main HTML file
    ├── css/
    │   └── style.css    # Frontend CSS styles
    └── js/
        └── script.js    # Frontend JavaScript logic
```

## Running with Docker (Split Frontend/Backend)

You can run the Pizza Recommender application using Docker and Docker Compose, with the frontend and backend split into separate containers.

- The Go backend API runs in one container (`backend`).
- The static frontend files (HTML/CSS/JS) are served by Nginx in another container (`frontend`).

**Requirements:**
* Docker and Docker Compose installed

**Steps:**
1. **Build and start the application:**
    ```bash
    docker compose up --build -d # Use -d to run in detached mode (optional)
    ```
    This command builds both the backend (Go) and frontend (Nginx) images using their respective Dockerfiles (`Dockerfile` and `Dockerfile.frontend`) and starts both services defined in `compose.yaml`.

2. **Access the application:**
    - Open your browser and go to `http://localhost:8000` (Note the port change - it's now served by the frontend Nginx container on port 8000).

**Details:**
- The Docker Compose setup defines two services: `backend` and `frontend`.
- They communicate over a dedicated Docker network (`pizza-net`).
- The `frontend` service depends on the `backend` service.
- The frontend Nginx container exposes port `80` internally, which is mapped to port `8000` on your host machine.
- The backend Go API listens on port `8080` internally within the Docker network, but this port is **not** exposed directly to your host machine by default.
- The frontend JavaScript (`static/js/script.js`) is configured to make API calls to `http://backend:8080`, using Docker's internal DNS to resolve the `backend` service name.
- Both containers run as non-root users where applicable.

**Stopping the application:**
```bash
docker compose down
```

*No special environment variables or configuration are required for this Docker setup.*
