# Home Task - Entain

## Tech Stack

- **Language**: Go 1.24
- **Framework**: Chi Router v5
- **Database**: PostgreSQL with GORM
- **Logging**: Logrus
- **Tracing**: OpenTelemetry
- **Testing**: Testify
- **Containerization**: Docker & Docker Compose

## Project Structure

```
├── cmd/home-task/           # Application entry point
├── internal/
│   ├── api/                 # HTTP server and routing
│   ├── config/              # Configuration management
│   ├── db/                  # Database layer (GORM, migrations)
│   ├── model/api/           # API request/response models
│   ├── service/             # Business logic layer
│   └── util/                # Utility packages (env, response)
├── compose.yaml             # Docker Compose configuration
├── Dockerfile               # Container build instructions
└── Makefile                 # Build and development commands
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development)

### Running with Docker Compose

1. Start the services:

```bash
docker-compose up -d
```

This will start:

- **Web server** on port `8080`
- **PostgreSQL database** on port `5432`

## API Endpoints

### Update User Balance

Updates a user's balance with transaction tracking.

**Endpoint**: `POST /user/{user_id}/balance`

**Headers**:

- `Source-Type`: Required. Must be one of: `game`, `server`, `payment`

**Request Body**:

```json
{
  "state": "win",
  // Required. Can be "win" or "lose"
  "transaction_id": "some generated identification",
  // Required. Unique transaction ID, e.g., UUID
  "amount": "10.50"
  // Required. Amount in string format, e.g., "10.50"
}
```

**Response**:

- `200 OK`: Balance updated successfully
- `400 Bad Request`: Invalid request data or missing/invalid Source-Type header
- `409 Conflict`: Invalid request with conflicting data (e.g., duplicate transaction ID)
- `500 Internal Server Error`: Server error

**Example**:

```bash
curl -X POST http://localhost:8080/user/1/transaction \
  -H "Content-Type: application/json" \
  -H "Source-Type: game" \
  -d '{
  "state": "win",
  "transaction_id": "e48a6dd8-09bc-4cb2-b036-59c8b497b7e2", 
  "amount": "10.50"
  }'
```

## Configuration

The application uses environment variables for configuration:

| Variable      | Description       | Default      |
|---------------|-------------------|--------------|
| `PORT`        | Server port       | `8080`       |
| `DB_HOST`     | Database host     | `localhost`  |
| `DB_PORT`     | Database port     | `5432`       |
| `DB_USER`     | Database username | `myuser`     |
| `DB_PASSWORD` | Database password | `mypassword` |
| `DB_NAME`     | Database name     | `mydb`       |

## Database Schema

The application automatically runs migrations on startup. Key entities:

- **Users**: User account information
- **Transactions**: Transaction history with amounts and source types

## Logging

The application provides logging:

- **Request Logging**: Every HTTP request is logged with method, URI, status code, and duration
- **OpenTelemetry Integration**: Trace and span IDs are included in all logs
- **Body Logging**: Optional request body logging for POST/PUT/PATCH requests

Example log output:

```json
{
  "http_method": "POST",
  "request_uri": "/user/123/balance",
  "span_id": "abc123...",
  "trace_id": "def456...",
  "status_code": 200,
  "duration_ms": 45,
  "level": "info",
  "msg": "HTTP Request Completed",
  "time": "2025-08-18T19:17:29Z"
}
```

## Development

### Running Tests

Run all tests:

```bash
make test
```

### Building

Build the binary:

```bash
make build
```

### Mock Generation

Generate mocks for testing:

```bash
make generate
```

## Docker

### Building the Image

```bash
make docker-build
```

