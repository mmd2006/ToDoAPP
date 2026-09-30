# ToDoAPP

A RESTful Task Management API built with Go, Echo, MongoDB, MySQL, JWT and bcrypt.

The project is designed as a portfolio-level backend application with a layered structure, authentication, role-based authorization, validation, pagination, login activity logging, MongoDB indexing and automated tests.

## Features

- User signup and login
- Password hashing with bcrypt
- JWT authentication with 24-hour expiration
- Role-based access control (`user` / `admin`)
- Task CRUD operations
- User-specific task access
- Admin task access
- Pagination for tasks and login activities
- MongoDB task storage
- MySQL user storage with GORM
- Login activity logging
- Input validation
- MongoDB compound index for task queries
- Request timeouts for database operations
- Unit tests for controllers, middleware, services and validation
- `go vet` compatible codebase
- OpenAPI 3 documentation

## Tech Stack

| Technology | Purpose |
|---|---|
| Go | Backend language |
| Echo v4 | HTTP framework |
| MySQL | User and authentication data |
| GORM | MySQL ORM |
| MongoDB | Tasks and login activities |
| MongoDB Go Driver | MongoDB access |
| JWT | Authentication |
| bcrypt | Password hashing |
| godotenv | Environment configuration |
| OpenAPI 3 | API documentation |

## Architecture

The project follows a simple layered structure:

```text
ToDoAPP/
├── config/          # Database connections and MongoDB indexes
├── controller/      # HTTP handlers and request/response handling
├── middleware/      # JWT authentication and role authorization
├── model/           # Data models
├── repository/      # Database access layer
├── router/          # Route definitions
├── service/         # Task business logic
├── validation/      # Request validation
├── docs/             # OpenAPI documentation
├── main.go
├── .env.example
├── go.mod
└── README.md
```

### Request flow

```text
HTTP Request
    ↓
Echo Router
    ↓
JWT Middleware / Role Middleware
    ↓
Controller
    ↓
Service
    ↓
Repository
    ↓
MongoDB / MySQL
```

Authentication-related user operations currently use the MySQL repository directly from the controller, while task operations use the full Controller → Service → Repository flow.

## Database Design

### MySQL

The `users` table stores:

- user ID
- username
- hashed password
- role

Passwords are never stored as plain text.

### MongoDB

The `tasks` collection stores:

- task ID
- user ID
- title
- description
- completion status
- creation time

The application creates this compound index:

```text
{ user_id: 1, created_at: -1 }
```

This supports the main user task query while keeping recent tasks first.

The `login_activities` collection stores login information such as username, user ID, IP address, user agent and login time.

## Authentication

Send the JWT returned by `/login` using the `Authorization` header:

```http
Authorization: Bearer <token>
```

Users can manage their own tasks. Admins can access the admin task endpoint and admin-only user/login activity endpoints.

## API Endpoints

### Public

| Method | Endpoint | Description |
|---|---|---|
| GET | `/` | API health message |
| POST | `/signup` | Create a user account |
| POST | `/login` | Authenticate and receive a JWT |

### Authenticated

| Method | Endpoint | Description |
|---|---|---|
| POST | `/tasks` | Create a task |
| GET | `/tasks` | List tasks |
| GET | `/tasks/:id` | Get one task |
| PUT | `/tasks/:id` | Update a task |
| DELETE | `/tasks/:id` | Delete a task |

### Admin

| Method | Endpoint | Description |
|---|---|---|
| GET | `/admin/tasks` | List tasks with admin access |
| GET | `/admin/users` | List users |
| GET | `/admin/login-activities` | List login activities |

## Validation Rules

### User

- Username is required
- Username length: 3–50 characters
- Username may contain letters, numbers and `_`
- Password is required
- Password minimum length: 6 characters

### Task

- Title is required
- Title length: 3–150 characters
- Description maximum length: 500 characters

## Pagination

Tasks support:

```text
GET /tasks?page=1&limit=10
```

- `page` starts at `1`
- default page: `1`
- default limit: `10`
- maximum task limit: `100`

Login activities support:

```text
GET /admin/login-activities?page=1&limit=10
```

- default page: `1`
- default limit: `10`
- maximum login activity limit: `50`

Invalid pagination values return `400 Bad Request`.

## Environment Variables

Create `.env` from `.env.example`:

```env
MONGODB_URI=mongodb://localhost:27017
MONGO_DB=todoapp
JWT_SECRET=change_me_to_a_long_random_secret
PORT=1323

MYSQL_USER=root
MYSQL_PASSWORD=
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DB=todoapp
```

> Never commit `.env` or real database credentials. The repository ignores `.env` through `.gitignore`.

## Local Setup

### Requirements

- Go 1.24+
- MySQL
- MongoDB

### 1. Clone the project

```bash
git clone <your-repository-url>
cd ToDoAPP
```

### 2. Configure environment variables

```bash
copy .env.example .env
```

On Linux/macOS:

```bash
cp .env.example .env
```

Update `.env` with your local database settings and a strong JWT secret.

### 3. Start MySQL and MongoDB

Make sure both database services are running before starting the API.

### 4. Download dependencies

```bash
go mod tidy
```

### 5. Run the application

```bash
go run .
```

The default API address is:

```text
http://localhost:1323
```

## Run with Docker

The project can run locally with Docker Compose using three containers:

- Go API
- MySQL 8
- MongoDB 8

### Prerequisites

- Docker Desktop
- Docker Compose v2

### Start

```bash
docker compose up --build
```

The API will be available at `http://localhost:1323`.

The MySQL and MongoDB data are stored in named Docker volumes, so restarting the containers does not remove the database data.

To stop the stack:

```bash
docker compose down
```

To stop it and remove the database volumes as well:

```bash
docker compose down -v
```

### Environment overrides

For local Docker use, Compose provides development defaults. You can override them with environment variables, for example:

```bash
JWT_SECRET=your_long_random_secret MYSQL_ROOT_PASSWORD=your_mysql_password docker compose up --build
```

The default values are intended only for local development and should not be used in production.

## Testing

Format the code:

```bash
go fmt ./...
```

Run all tests:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

## API Documentation

The OpenAPI 3 specification is available at:

```text
docs/openapi.yaml
```

You can import this file into Swagger UI, Swagger Editor, Postman or another OpenAPI-compatible tool.

## Example Request Flow

### 1. Signup

```http
POST /signup
Content-Type: application/json
```

```json
{
  "username": "reza_test",
  "password": "123456"
}
```

### 2. Login

```http
POST /login
Content-Type: application/json
```

```json
{
  "username": "reza_test",
  "password": "123456"
}
```

Response:

```json
{
  "token": "<jwt-token>"
}
```

### 3. Create a task

```http
POST /tasks
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

```json
{
  "title": "Learn Go interfaces",
  "description": "Review interfaces and error handling"
}
```

### 4. List tasks

```http
GET /tasks?page=1&limit=10
Authorization: Bearer <jwt-token>
```

## Security Notes

- Passwords are hashed with bcrypt.
- JWTs are signed with HS256 and require a configured `JWT_SECRET`.
- Protected routes require a valid Bearer token.
- Admin routes require the `admin` role.
- Users cannot access another user's tasks through normal task endpoints.
- Database error details are not returned directly to API clients.

## Project Status

The current codebase passes:

```text
go fmt ./...
go test ./...
go vet ./...
```

The project is intentionally kept within a practical Junior/Intern backend scope: authentication, authorization, persistence, validation, testing and documentation are included without adding unnecessary enterprise-level complexity.
