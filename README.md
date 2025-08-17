# Financial Transaction System

A robust, production-ready financial transaction system built with Go, showcasing microservices architecture, clean code practices, and modern development patterns. This project demonstrates comprehensive backend development skills including concurrent processing, database design, API development, and observability.

## Features

### Core Functionality
- **User Management**: Secure user registration, authentication, and role-based authorization
- **Transaction Processing**: Credit, debit, and transfer operations with atomic guarantees
- **Balance Management**: Thread-safe balance updates with historical tracking
- **Scheduled Transactions**: Automated recurring and future-dated transactions
- **Transaction Limits**: Configurable limits and rules for different user types

### Advanced Features
- **Concurrent Processing**: Worker pool architecture for high-throughput transaction processing
- **Event Sourcing**: Audit logging for all system changes with replay capability
- **Caching Layer**: Redis-based caching with intelligent invalidation strategies
- **Batch Processing**: Efficient bulk transaction operations
- **Multi-currency Support**: Extensible currency handling system

### Technical Excellence
- **Clean Architecture**: Separation of concerns with domain-driven design
- **Observability**: OpenTelemetry integration with Prometheus metrics and Grafana dashboards
- **Security**: JWT authentication, input validation, and secure defaults
- **Testing**: Comprehensive test coverage with unit, integration, and performance tests
- **Performance**: Optimized database queries, connection pooling, and caching

### Key Components
- **Handlers**: HTTP request/response handling with validation
- **Services**: Business logic and transaction orchestration
- **Repositories**: Data access abstraction with PostgreSQL implementation
- **Workers**: Concurrent transaction processing with channel-based queues
- **Middleware**: Authentication, logging, metrics, and error handling

## Technology Stack

- **Language**: Go 1.21+ (Go modules, generics, context)
- **Database**: PostgreSQL with migrations
- **Cache**: Redis for session management and caching
- **Message Queue**: Go channels for internal communication
- **Authentication**: JWT with role-based access control
- **Monitoring**: Prometheus, Grafana, OpenTelemetry
- **Containerization**: Docker with multi-stage builds
- **Testing**: Go testing framework with table-driven tests

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 14+ (included in Docker setup)
- Redis 7+ (included in Docker setup)

## Quick Start

### 1. Clone the Repository
```bash
git clone <repository-url>
cd InsiderProject
```

### 2. Build and Run with Docker Compose
```bash
# Build the application
docker compose build app

# Start all services
docker compose up -d
```

The system will be available at:
- **API**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

### 3. Verify Installation
```bash
# Check service status
docker compose ps

# View application logs
docker compose logs app

# Test API health
curl http://localhost:8080/health
```

## Local Development

### 1. Install Dependencies
```bash
go mod download
```

### 2. Set Environment Variables
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Run Database Migrations
```bash
# Using the included migration tool
go run cmd/migrate/main.go
```

### 4. Start the Application
```bash
go run cmd/backend/main.go
```

### 5. Run Tests
```bash
# Unit tests
go test ./...

# Tests with coverage
go test -cover ./...

# Performance benchmarks
go test -bench=. ./...
```

## Testing

### Test Structure
- **Unit Tests**: Individual component testing with mocks
- **Integration Tests**: Database and service integration
- **Performance Tests**: Benchmarking critical paths
- **API Tests**: End-to-end HTTP endpoint testing

### Running Tests
```bash
# All tests
go test ./...

# Specific package
go test ./internal/service

# Tests with verbose output
go test -v ./...

# Tests with race detection
go test -race ./...
```

## Monitoring and Observability

### Metrics (Prometheus)
- Request latency and throughput
- Error rates and response codes
- Database connection pool status
- Worker pool performance metrics
- Business metrics (transaction volume, user activity)

### Logging
- Structured JSON logging
- Request correlation IDs
- Performance tracing
- Error context and stack traces

### Dashboards (Grafana)
- System performance overview
- Business metrics dashboard
- Database performance monitoring
- Custom alerting rules

## 🔧 Configuration

### Environment Variables
```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=financial_system
DB_USER=postgres
DB_PASSWORD=password

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Worker Configuration
WORKER_POOL_SIZE=10
WORKER_QUEUE_SIZE=1000
```

## Docker

### Multi-stage Dockerfile
- **Build Stage**: Go compilation with optimizations
- **Runtime Stage**: Minimal Alpine Linux image
- **Security**: Non-root user, minimal attack surface

### Docker Compose Services
- **app**: Main application service
- **postgres**: PostgreSQL database
- **redis**: Redis cache and session store
- **grafana**: Monitoring dashboards
- **prometheus**: Metrics collection

## Deployment

### Production Considerations
- **Environment Variables**: Secure configuration management
- **Database**: Connection pooling and read replicas
- **Caching**: Redis cluster for high availability
- **Load Balancing**: Reverse proxy with health checks
- **Monitoring**: Centralized logging and alerting

### Scaling
- **Horizontal**: Multiple application instances
- **Vertical**: Resource optimization and tuning
- **Database**: Read replicas and connection pooling
- **Caching**: Distributed cache with consistent hashing

## Project Structure

```
InsiderProject/
├── cmd/                    # Application entrypoints
│   └── backend/           # Main server binary
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── domain/           # Business logic and models
│   ├── handler/          # HTTP request handlers
│   ├── middleware/       # HTTP middleware
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic services
│   └── worker/           # Background workers
├── pkg/                  # Public packages
│   ├── cache/            # Redis caching utilities
│   ├── jwt/              # JWT authentication
│   ├── metrics/          # Prometheus metrics
│   └── tracing/          # OpenTelemetry tracing
├── migrations/           # Database schema migrations
├── configs/              # Configuration files
│   ├── grafana/          # Grafana dashboards
│   └── prometheus/       # Prometheus configuration
└── test/                 # Test utilities and scripts
```

## Contributing

This project demonstrates industry-standard development practices:

### Code Quality
- **Go Modules**: Dependency management
- **Go Fmt**: Code formatting
- **Go Lint**: Static analysis
- **Go Vet**: Code correctness checks

### Testing Strategy
- **Table-driven Tests**: Comprehensive test coverage
- **Mock Interfaces**: Clean dependency injection
- **Integration Tests**: Real database testing
- **Performance Tests**: Benchmarking critical paths

### Architecture Patterns
- **Clean Architecture**: Separation of concerns
- **Domain-Driven Design**: Business logic isolation
- **Interface Segregation**: Small, focused interfaces
- **Dependency Injection**: Testable and maintainable code

## Learning Outcomes

This project demonstrates mastery of:

- **Go Language**: Idiomatic Go code, concurrency patterns, error handling
- **Microservices**: Service decomposition, API design, inter-service communication
- **Database Design**: Schema design, migrations, connection management
- **Security**: Authentication, authorization, input validation
- **Performance**: Caching, connection pooling, worker pools
- **Observability**: Metrics, logging, tracing, monitoring
- **DevOps**: Docker, containerization, service orchestration
- **Testing**: Unit, integration, and performance testing

## License

This project is created for educational and portfolio purposes.

---

**Note**: This is a showcase project demonstrating backend development skills. For production use, additional security, monitoring, and deployment considerations should be implemented.
