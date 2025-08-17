# CONTEXT.md

## 📌 Project Title: Backend Path

A learning-focused backend project built in **Go**, simulating a minimal financial system with user registration, transactions, and balance tracking. Designed for internship-level exploration of Go fundamentals, concurrency, clean architecture, and containerization.

---

## 🎯 Goal

Develop a modular, testable backend service in Go that handles:

* User authentication & authorization (JWT-based)
* Credit, debit, and transfer operations
* Safe concurrent balance updates
* RESTful API with role-based access
* Dockerized, observable microservice design

---

## 🧱 Architecture Overview

* **Language**: Go
* **Database**: PostgreSQL
* **Cache (Optional)**: Redis
* **Router**: Chi or Gorilla Mux
* **Observability**: Prometheus, Grafana (Optional)
* **Containerization**: Docker + docker-compose
* **Auth**: JWT-based access tokens
* **Style**: Clean architecture (domain, service, handler)

---

## 📂 Folder Structure (Planned)

```bash
.
├── cmd/                # Main application entrypoint
├── internal/
│   ├── config/         # Env/config loader
│   ├── domain/         # Models and interfaces
│   ├── service/        # Business logic
│   ├── handler/        # HTTP handlers (controllers)
│   ├── repository/     # DB interaction layer
│   ├── middleware/     # Middleware (auth, logging)
│   └── worker/         # Background processing
├── migrations/         # DB migration files
├── api/                # Swagger/openapi or postman collection
├── Dockerfile
├── docker-compose.yml
├── go.mod / go.sum
└── README.md
```

---

## 🔐 Core Features

* **Auth**:

  * Register/Login
  * JWT Token issuance
  * Middleware for route protection

* **User**:

  * CRUD endpoints
  * Role-based access (admin vs user)

* **Transaction**:

  * Credit/Debit operations
  * Transfer between users
  * Transaction rollback (if needed)

* **Balance**:

  * Get current balance
  * Get historical balance

* **Concurrency**:

  * Use `sync.RWMutex` or atomic for safe balance updates
  * Simple worker pool for batch tasks

---

## 🔀 API Endpoints (v1)

```http
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/users
GET    /api/v1/users/{id}
POST   /api/v1/transactions/credit
POST   /api/v1/transactions/debit
POST   /api/v1/transactions/transfer
GET    /api/v1/balances/current
```

---

## 🐳 Docker Services

* `app`: Golang service
* `db`: PostgreSQL
* `redis`: Optional cache
* `monitoring`: Prometheus + Grafana (optional)

---

## 📚 Learning Resources

* [Go by Example](https://gobyexample.com/)
* [Effective Go](https://go.dev/doc/effective_go)
* [Go Doc](https://go.dev/doc/)
* [Go Full Course (freeCodeCamp)](https://www.youtube.com/watch?v=YzLrWHZa-Kc)

---

## 🚀 Notes for Cursor

This is a junior-level backend learning project. Use this context to:

* Enable IntelliSense for domain-service separation
* Identify `main.go` entrypoint under `cmd/`
* Autocomplete API handler and interface logic
* Understand service responsibilities by folder
* Detect and recommend Dockerfile best practices
