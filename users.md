

- **Grafana Admin**
  - **Username:** `admin`
  - **Password:** `adminpass`

  "username":"testuser1","email":"test1@example.com","password":"password123","role":"user"
  "username":"testuser2","email":"test2@example.com","password":"password123","role":"user"



## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User authentication
- `POST /api/v1/auth/refresh` - Token refresh

### Users
- `GET /api/v1/users` - List users (admin only)
- `GET /api/v1/users/{id}` - Get user details
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user (admin only)

### Transactions
- `POST /api/v1/transactions/credit` - Credit account
- `POST /api/v1/transactions/debit` - Debit account
- `POST /api/v1/transactions/transfer` - Transfer between accounts
- `GET /api/v1/transactions/history` - Transaction history
- `GET /api/v1/transactions/{id}` - Get transaction details

### Balances
- `GET /api/v1/balances/current` - Current balance
- `GET /api/v1/balances/historical` - Balance history
- `GET /api/v1/balances/at-time` - Balance at specific time

### Scheduled Transactions
- `POST /api/v1/scheduled-transactions` - Create scheduled transaction
- `GET /api/v1/scheduled-transactions` - List scheduled transactions
- `PUT /api/v1/scheduled-transactions/{id}` - Update scheduled transaction
- `DELETE /api/v1/scheduled-transactions/{id}` - Cancel scheduled transaction


## 🏗️ Architecture

This system follows Clean Architecture principles with clear separation of concerns:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP/gRPC     │    │   Business      │    │   Data Access   │
│   Handlers      │◄──►│   Services      │◄──►│   Repositories  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Middleware    │    │   Domain        │    │   Database      │
│   (Auth, Log)   │    │   Models        │    │   (PostgreSQL)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```