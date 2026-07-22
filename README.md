# Go To-Do List Backend API

A high-performance, production-grade RESTful API for managing tasks (To-Do List), built with **Go**, **Gin Web Framework**, **PostgreSQL**, and **Redis**.

---

## 🚀 Features

- **Full Task Management (CRUD)**: Create, list (with pagination & search), view, update, and delete tasks.
- **JWT Authentication & Authorization**: Secure endpoint protection using JSON Web Tokens (`/auth/register`, `/auth/login`, Bearer Middleware).
- **Multi-Tenant User Isolation**: Tasks are linked to individual users, enforcing strict data access control.
- **Concurrent Query Execution**: Parallel PostgreSQL query execution using `golang.org/x/sync/errgroup` for paginated task retrieval and total count calculation.
- **Smart Redis Caching**: Automatic query caching (`GET /tasks`, `GET /tasks/:id`) with immediate cache invalidation on task creation, update, or deletion.
- **Raw SQL Performance**: Low-overhead raw SQL queries using `jackc/pgx/v5` connection pooling.
- **Automated Migrations**: Automatic database DDL schema migrations executed on server startup.
- **Structured JSON Logging**: Observability powered by Go's native `log/slog` structured logger.
- **Comprehensive Unit Testing**: End-to-end handler test suite utilizing `testing`, `httptest`, and `testify`.
- **Containerized Deployment**: Ready to run with Docker and `docker-compose`.

---

## 🛠️ Tech Stack

- **Language**: Go 1.22+
- **HTTP Framework**: [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL 15 (`jackc/pgx/v5`)
- **Cache Store**: Redis 7 (`redis/go-redis/v9`)
- **Authentication**: JWT (`golang-jwt/jwt/v5`) & Bcrypt (`golang.org/x/crypto/bcrypt`)
- **Concurrency**: `golang.org/x/sync/errgroup`
- **Testing**: `net/http/httptest` & `github.com/stretchr/testify`

---

## 🏗️ Architecture & Project Structure

The project follows a **Layered Clean Architecture** pattern with clear separation of concerns:

```mermaid
graph TD
    Client[Client / Postman / App] -->|HTTP Requests| GinRouter[Gin Router / CORS]
    
    subgraph Middleware Layer
        GinRouter --> Logger[slog JSON Logger]
        GinRouter --> AuthMiddleware[JWT Auth Middleware]
    end
    
    subgraph Handler Layer
        AuthMiddleware --> TaskHandler[Task Handler]
        AuthMiddleware --> AuthHandler[Auth Handler]
    end
    
    subgraph Service Layer
        TaskHandler --> TaskService[Task Service]
        AuthHandler --> AuthService[Auth Service]
    end
    
    subgraph Repository Layer
        TaskService -->|errgroup Concurrent Queries| TaskRepo[Task Repository]
        TaskService -->|Cache & Invalidation| CacheRepo[Redis Cache Repo]
        AuthService --> UserRepo[User Repository]
    end
    
    subgraph Storage Layer
        TaskRepo -->|pgx/v5 Connection Pool| Postgres[(PostgreSQL 15)]
        UserRepo -->|pgx/v5 Connection Pool| Postgres
        CacheRepo -->|go-redis| Redis[(Redis 7)]
```

### Directory Layout

```text
.
├── cmd/
│   └── api/
│       └── main.go              # Application entrypoint & dependency injection
├── internal/
│   ├── config/                  # Environment configuration loader
│   ├── domain/                  # Core entities, interfaces & models
│   ├── handler/                 # HTTP handlers & request validators
│   ├── middleware/              # Auth & logger middleware
│   ├── repository/
│   │   ├── postgres/            # PostgreSQL data persistence (raw SQL)
│   │   └── redis/               # Redis caching repository
│   └── service/                 # Core business logic
├── pkg/
│   ├── database/                # Postgres pool & migration runner
│   ├── jwt/                     # JWT token generation & verification
│   ├── logger/                  # slog logger configuration
│   └── redis/                   # Redis client setup
├── migrations/                  # DDL SQL migration scripts
├── docs/                        # Technical architecture specifications
├── Dockerfile                   # Multi-stage Docker build configuration
├── docker-compose.yml           # Multi-container orchestration (App, Postgres, Redis)
├── .env.example                 # Example environment variables
└── go.mod                       # Go module dependencies
```

---

## 🗄️ Database Schema

### Users Table (`users`)
```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Tasks Table (`tasks`)
```sql
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'completed')),
    due_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

## ⚙️ Environment Variables

Copy `.env.example` to `.env` to configure your environment variables:

| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `PORT` | Application server port | `8080` |
| `APP_ENV` | Environment mode (`development` / `production`) | `development` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@db:5432/todo_db?sslmode=disable` |
| `REDIS_URL` | Redis connection URL | `redis://redis:6379` |
| `JWT_SECRET` | Secret key for signing JWT tokens | `super-secret-jwt-key-change-in-production` |
| `JWT_EXPIRATION_HOURS` | Token validity duration in hours | `24` |

---

## 📋 API Specification

### Health Check

#### `GET /health`
Returns the status of the server.
- **Response (200 OK):**
```json
{
  "status": "ok",
  "timestamp": "2026-07-22T06:30:00Z"
}
```

---

### Authentication Endpoints

#### 1. Register User
- **Endpoint**: `POST /auth/register`
- **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```
- **Response (201 Created)**:
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "c7a8f9e0-1234-5678-9abc-def012345678",
    "email": "user@example.com"
  }
}
```

#### 2. Login User
- **Endpoint**: `POST /auth/login`
- **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```
- **Response (200 OK)**:
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

### Protected Task Endpoints *(Header: `Authorization: Bearer <token>`)*

#### 3. Create Task
- **Endpoint**: `POST /tasks`
- **Request Body**:
```json
{
  "title": "Build README documentation",
  "description": "Create detailed markdown documentation for the project",
  "status": "pending",
  "due_date": "2026-08-01"
}
```
- **Response (201 Created)**:
```json
{
  "message": "Task created successfully",
  "task": {
    "id": "e9b1c2d3-4567-890a-bcde-f0123456789a",
    "user_id": "c7a8f9e0-1234-5678-9abc-def012345678",
    "title": "Build README documentation",
    "description": "Create detailed markdown documentation for the project",
    "status": "pending",
    "due_date": "2026-08-01",
    "created_at": "2026-07-22T06:30:00Z",
    "updated_at": "2026-07-22T06:30:00Z"
  }
}
```

#### 4. Get All Tasks (Paginated & Filterable)
- **Endpoint**: `GET /tasks`
- **Query Parameters**:
  - `status` *(optional)*: `pending` | `completed`
  - `page` *(optional, default: `1`)*: Page number
  - `limit` *(optional, default: `10`)*: Tasks per page
  - `search` *(optional)*: Keyword search in title/description
- **Response (200 OK)**:
```json
{
  "tasks": [
    {
      "id": "e9b1c2d3-4567-890a-bcde-f0123456789a",
      "user_id": "c7a8f9e0-1234-5678-9abc-def012345678",
      "title": "Build README documentation",
      "description": "Create detailed markdown documentation for the project",
      "status": "pending",
      "due_date": "2026-08-01",
      "created_at": "2026-07-22T06:30:00Z",
      "updated_at": "2026-07-22T06:30:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 1,
    "total_tasks": 1
  }
}
```

#### 5. Get Task by ID
- **Endpoint**: `GET /tasks/:id`
- **Response (200 OK)**:
```json
{
  "id": "e9b1c2d3-4567-890a-bcde-f0123456789a",
  "user_id": "c7a8f9e0-1234-5678-9abc-def012345678",
  "title": "Build README documentation",
  "description": "Create detailed markdown documentation for the project",
  "status": "pending",
  "due_date": "2026-08-01",
  "created_at": "2026-07-22T06:30:00Z",
  "updated_at": "2026-07-22T06:30:00Z"
}
```

#### 6. Update Task
- **Endpoint**: `PUT /tasks/:id`
- **Request Body**:
```json
{
  "title": "Build README documentation",
  "description": "Updated description with clear diagrams",
  "status": "completed",
  "due_date": "2026-08-01"
}
```
- **Response (200 OK)**:
```json
{
  "message": "Task updated successfully",
  "task": {
    "id": "e9b1c2d3-4567-890a-bcde-f0123456789a",
    "user_id": "c7a8f9e0-1234-5678-9abc-def012345678",
    "title": "Build README documentation",
    "description": "Updated description with clear diagrams",
    "status": "completed",
    "due_date": "2026-08-01",
    "created_at": "2026-07-22T06:30:00Z",
    "updated_at": "2026-07-22T06:35:00Z"
  }
}
```

#### 7. Delete Task
- **Endpoint**: `DELETE /tasks/:id`
- **Response (200 OK)**:
```json
{
  "message": "Task deleted successfully"
}
```

---

## 🏃 Getting Started

### Prerequisites
- [Go 1.22+](https://go.dev/doc/install)
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)

---

### Option A: Running via Docker Compose (Recommended)

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```

2. **Start all services** (Backend API, PostgreSQL 15, Redis 7):
   ```bash
   docker compose up --build
   ```

3. The server will start and automatically run migrations. API will be available at:
   `http://localhost:8080`

---

### Option B: Running Locally

1. **Start PostgreSQL and Redis** on standard ports (`5432` and `6379`).

2. **Configure environment variables**:
   ```bash
   cp .env.example .env
   ```
   *Make sure `DATABASE_URL` uses `localhost:5432` and `REDIS_URL` uses `localhost:6379` if running outside containers.*

3. **Install Go dependencies**:
   ```bash
   go mod download
   ```

4. **Run the API server**:
   ```bash
   go run cmd/api/main.go
   ```

---

## 🧪 Running Unit Tests

Run the full test suite with verbose logging:

```bash
go test -v ./...
```

Run tests with code coverage analysis:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

---

## 💡 Key Design Highlights

1. **Concurrent Database Querying**:
   In [`internal/repository/postgres/task_repo.go`](file:///app/internal/repository/postgres/task_repo.go), the `GetTasks` function uses `golang.org/x/sync/errgroup` to execute the data query (paginated records) and total count query in parallel across distinct goroutines, cutting execution latency in half.

2. **Smart Redis Invalidation**:
   In [`internal/service/task_service.go`](file:///app/internal/service/task_service.go), cache keys are user-scoped (`user:<user_id>:...`). Any write/mutation operation (`CreateTask`, `UpdateTask`, `DeleteTask`) automatically purges matching cache keys (`user:<user_id>:*`) to guarantee data consistency.

3. **Graceful Fallback**:
   If Redis becomes unavailable, the application logs a warning and gracefully continues operations directly with PostgreSQL without crashing.
