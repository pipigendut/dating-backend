# Dating App Backend

A high-performance, scalable backend for a dating application built with Go, adhering to the principles of Clean Architecture.

## 🚀 Tech Stack

- **Language**: [Go](https://go.dev/) (1.26+)
- **Web Framework**: [Gin Gonic](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/)
- **Database**: [PostgreSQL](https://www.postgresql.org/) (PostGIS enabled)
- **Cache**: [Redis](https://redis.io/)
- **Documentation**: [Swaggo](https://github.com/swaggo/swag)
- **Utilities**: JWT for auth, Bcrypt for password hashing.

## 🏗 Architecture

This project follows **Clean Architecture** to ensure separation of concerns, scalability, and ease of testing.

```mermaid
graph TD
    A[Delivery Layer / HTTP] --> B[Usecase Layer / Business Logic]
    B --> C[Repository Layer / Data Access]
    C --> D[Domain Layer / Entities]
    B --> D
```

- **`internal/entities`**: Core business models (User, Profile, Photo).
- **`internal/usecases`**: Business logic orchestration.
- **`internal/repository`**: Data persistence interfaces and GORM implementations.
- **`internal/delivery/http`**: HTTP handlers, requests/responses DTOs.
- **`pkg`**: Shared utilities (Auth, JWT, etc.).

## 🔑 Key Features

- **Authentication System**:
  - Google OAuth Login with automatic **Account Linking**.
  - Email & Password Login/Registration.
  - JWT-based session management.
- **Standardized API Responses**: Centralized response envelope for consistency.
- **Scalable HTTP Layer**: Handlers and DTOs grouped by resource.
- **Database Migrations**: Versioned SQL migrations using `golang-migrate`.

## 🛠 Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://go.dev/doc/install) (1.20+)

### Setup

1. **Clone the repository**:
   ```bash
   git clone <repo-url>
   cd dating-backend
   ```

2. **Configure Environment Variables**:
   Copy the example environment file and adjust the values (especially database credentials).
   ```bash
   cp .env.example .env
   ```

3. **Start Infrastructure**:
   ```bash
   docker-compose up -d
   ```

4. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

5. **Database Migrations & Seeding**:
   Migrations and seeding are now manual. Refer to the commands below:
   ```bash
   # Run all pending migrations
   go run cmd/migrate/main.go up

   # Rollback one step (Undo the last migration)
   go run cmd/migrate/main.go step -1

   # Full Rollback (CAUTION: Removes all tables!)
   go run cmd/migrate/main.go down

   # Seed master data
   go run cmd/seed/main.go
   ```

6. **Run the Application**:
   ```bash
   go run cmd/app/main.go
   ```

## 📖 API Documentation

The project uses Swagger for API documentation.

1. **Access Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
2. **Regenerate Documentation** (after adding annotations):
   ```bash
   swag init -g cmd/app/main.go
   ```

## 🌐 Environment & Development

This project supports separate **development** and **production** environments.

### 🏁 Running the App

Using the **Makefile**:
- **Development**: `make dev` (Uses `.env.development` and `swipee://` deep links)
- **Production**: `make prod` (Uses `.env.production` and `https://swipee.app` universal links)

### 🛠 Development with ngrok

Whenever you restart ngrok, you must update the tunnel URLs:

1.  **Backend**: Update `BASE_URL` in `dating-backend/.env.development` and restart with `make dev`.
2.  **Mobile**: Update `EXPO_PUBLIC_API_URL` in `dating-mobile/.env.development`.
3.  **Mobile Switch**: Use `npm run start:dev` to load the development environment.

## 🧪 Testing

```bash
go test ./...
```

## 🛠 Extra Commands

### Advanced Migrations
For more granular control over database migrations:

```bash
# Check current migration version
go run cmd/migrate/main.go version

# Rollback exactly 1 migration step
go run cmd/migrate/main.go step -1

# Run exactly 1 next migration step
go run cmd/migrate/main.go step 1

# Force a specific version (use this if the database is in a 'dirty' state)
go run cmd/migrate/main.go force <version_number>
```
