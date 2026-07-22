# Technical Architecture Specification: Go To-Do List API

> [!NOTE]
> This document summarizes the requirements extracted from the Larksuite technical test document and the architectural design decisions resolved during the alignment interview.

---

## 1. Executive Summary & Context

The objective is to build a robust, production-grade Go Backend REST API for managing a To-Do list. The solution adheres strictly to the specifications outlined in the technical assessment document and implements all optional enhancements (JWT Auth, Redis Caching, Structured Error Logging, Unit Tests, and Concurrent Execution).

---

## 2. Requirements & Agreed Technical Decisions

| Category | Requirement / Decision | Details & Implementation Strategy |
| :--- | :--- | :--- |
| **Language & Runtime** | Go 1.22+ | Native performance, strict typing, and concurrency features |
| **Web Framework** | Gin (`github.com/gin-gonic/gin`) | High-performance routing, middleware support, and clean JSON binding |
| **Database & Driver** | PostgreSQL 15 + `jackc/pgx/v5` | Direct raw SQL queries with connection pooling for maximum control |
| **Database Migrations** | SQL Migration scripts | Standard DDL migrations (`000001_create_users_table.up.sql`) |
| **Architecture Layout** | Layered / Clean Architecture | Strict separation: `Handler` $\rightarrow$ `Service` $\rightarrow$ `Repository` $\rightarrow$ `Database/Redis` |
| **Authentication** | JWT (JSON Web Tokens) | User Registration (`/auth/register`), Login (`/auth/login`), and Bearer auth middleware |
| **Data Scope** | Multi-tenant User Isolation | Each user manages their own tasks (foreign key `user_id`) |
| **Caching Strategy** | Redis 7 | Cache `GET /tasks` & `GET /tasks/:id`. Invalidate cache on `POST`, `PUT`, `DELETE` |
| **Concurrency** | `golang.org/x/sync/errgroup` | Parallel query execution in `GET /tasks` (fetching tasks + total count concurrently) |
| **Logging** | Go `log/slog` | Structured JSON log output with trace context and error details |
| **Testing** | `go test` + `httptest` + `testify` | Unit tests for handlers, services, and repositories |

---

## 3. System Architecture Diagram

```mermaid
graph TD
    Client[Client / Postman / Frontend] -->|HTTP Requests| GinRouter[Gin Router / Middleware]
    
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
    
    subgraph Concurrency & Repository Layer
        TaskService -->|errgroup Concurrent Query| TaskRepo[Task Repository]
        TaskService -->|Cache Lookups & Invalidation| CacheRepo[Redis Cache Repository]
        AuthService --> UserRepo[User Repository]
    end
    
    subgraph Storage Layer
        TaskRepo -->|pgx/v5 Raw SQL| Postgres[(PostgreSQL 15)]
        UserRepo -->|pgx/v5 Raw SQL| Postgres
        CacheRepo -->|go-redis| Redis[(Redis 7)]
```

---

## 4. Database Schema Design

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

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
```

---

## 5. API Endpoints & Request/Response Contracts

### Authentication Endpoints

#### 1. `POST /auth/register`
* **Request:**
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword123"
  }
  ```
* **Response (201 Created):**
  ```json
  {
    "message": "User registered successfully",
    "user": {
      "id": "c7a8f9e0-1234-5678-9abc-def012345678",
      "email": "user@example.com"
    }
  }
  ```

#### 2. `POST /auth/login`
* **Request:**
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword123"
  }
  ```
* **Response (200 OK):**
  ```json
  {
    "message": "Login successful",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```

---

### Task Management Endpoints *(Requires Header `Authorization: Bearer <token>`)*

#### 3. `POST /tasks`
* **Request Body:**
  ```json
  {
    "title": "Task Title",
    "description": "Task Description",
    "status": "pending",
    "due_date": "2026-08-01"
  }
  ```
* **Response (201 Created):**
  ```json
  {
    "message": "Task created successfully",
    "task": {
      "id": "e9b1c2d3-4567-890a-bcde-f0123456789a",
      "user_id": "c7a8f9e0-1234-5678-9abc-def012345678",
      "title": "Task Title",
      "description": "Task Description",
      "status": "pending",
      "due_date": "2026-08-01",
      "created_at": "2026-07-22T05:50:00Z",
      "updated_at": "2026-07-22T05:50:00Z"
    }
  }
  ```

#### 4. `GET /tasks`
* **Query Parameters:**
  * `status` *(optional)*: `pending` | `completed`
  * `page` *(optional, default 1)*: Page number
  * `limit` *(optional, default 10)*: Items per page
  * `search` *(optional)*: Keyword filter for title/description
* **Response (200 OK):**
  ```json
  {
    "tasks": [
      {
        "id": "e9b1c2d3-4567-890a-bcde-f0123456789a",
        "title": "Task Title",
        "description": "Task Description",
        "status": "pending",
        "due_date": "2026-08-01"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 1,
      "total_tasks": 1
    }
  }
  ```

#### 5. `GET /tasks/:id`
* **Response (200 OK):**
  ```json
  {
    "id": "e9b1c2d3-4567-890a-bcde-f0123456789a",
    "title": "Task Title",
    "description": "Task Description",
    "status": "pending",
    "due_date": "2026-08-01"
  }
  ```

#### 6. `PUT /tasks/:id`
* **Request Body:**
  ```json
  {
    "title": "Updated Task Title",
    "description": "Updated Description",
    "status": "completed",
    "due_date": "2026-08-05"
  }
  ```
* **Response (200 OK):**
  ```json
  {
    "message": "Task updated successfully",
    "task": {
      "id": "e9b1c2d3-4567-890a-bcde-f0123456789a",
      "title": "Updated Task Title",
      "description": "Updated Description",
      "status": "completed",
      "due_date": "2026-08-05"
    }
  }
  ```

#### 7. `DELETE /tasks/:id`
* **Response (200 OK):**
  ```json
  {
    "message": "Task deleted successfully"
  }
  ```
