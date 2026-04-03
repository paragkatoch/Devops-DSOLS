# DSOLS - Distributed Systems Observability & Load Simulation

A **microservices-based order management and load simulation system** built with Go, featuring event-driven architecture, distributed components, and containerized deployment using Docker and Nginx.

## 🎯 Project Overview

DSOLS is a modern, scalable backend system for managing users, products, and orders in a distributed environment. It follows microservices principles with independent services communicating through message queues, API gateways, and databases.

### Key Features

- 🏗️ **Microservices Architecture** - Decoupled, independent services
- 📨 **Event-Driven** - Asynchronous processing with RabbitMQ
- 🔄 **Service Communication** - REST APIs + Message Queue integration
- 🐳 **Docker Containerized** - Full Docker & Docker Compose setup
- 📦 **PostgreSQL Database** - Persistent data storage
- 🚀 **Nginx API Gateway** - Single entry point for all client requests

## 📋 Table of Contents

- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Technology Stack](#technology-stack)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Services](#services)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Development](#development)
- [Troubleshooting](#troubleshooting)

## 🏛️ Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Clients/Browser                      │
└────────────────────────┬────────────────────────────────────┘
						 │ (Port 9000)
						 ▼
┌────────────────────────────────────────────────────────────┐
│                    API Gateway (Nginx)                     │
│  Routes:                                                   │
│  /api/user      → user-controller:9000                   │
│  /api/product   → product-controller:9000               │
└──────┬──────────────────────────────────┬──────────────────┘
	   │                                  │
	   ▼                                  ▼
┌──────────────────┐            ┌──────────────────────┐
│ User Controller  │            │ Product Controller   │
│  (HTTP Server)   │            │   (HTTP Server)      │
└─────────┬────────┘            └──────────┬───────────┘
		  │                                │
		  └───────────────┬────────────────┘
						  │
			┌─────────────┴──────────────┐
			│                            │
			▼                            ▼
	┌────────────────┐         ┌──────────────────┐
	│  PostgreSQL    │         │   RabbitMQ       │
	│   Database     │         │  Message Queue   │
	└────────────────┘         └──────────────────┘
			▲                            ▲
			└────────────────┬───────────┘
							 │
					┌────────┴────────┐
					│                 │
					▼                 ▼
			┌──────────────┐  ┌──────────────┐
			│ User Service │  │ Product Svc  │
			│  (Consumers) │  │  (Consumers) │
			└──────────────┘  └──────────────┘
```

### Data Flow

1. **Client Request** → API Gateway (Nginx)
2. **Gateway Routes** → Appropriate Controller
3. **GET requests** → Read directly from PostgreSQL
4. **POST/PUT requests** → Queue event to RabbitMQ
5. **Services Consume** → Process events and persist to database

## 📁 Project Structure

```
.
├── README.md                          # Main project documentation
├── API_GATEWAY_README.md              # Detailed API Gateway docs
├── docker-compose.yml                 # Docker Compose configuration
│
├── infra/
│   ├── Dockerfile                     # Container build configuration
│   └── nginx.conf                     # Nginx routing configuration
│
└── server/
	├── main.go                        # Application entry point
	├── go.mod                         # Go module dependencies
	│
	├── config/
	│   └── local.yml                  # Local environment configuration
	│
	├── cmd/                           # Command-line utilities
	│
	├── controllers/
	│   ├── userController.go          # User service HTTP handler
	│   ├── productController.go       # Product service HTTP handler
	│   └── orderController.go         # Order service HTTP handler
	│
	├── services/
	│   ├── userService.go             # User business logic
	│   ├── productService.go          # Product business logic
	│   └── orderService.go            # Order business logic
	│
	├── internal/
	│   ├── config.go                  # Configuration loader
	│   ├── rabbitmq/
	│   │   └── rabbitmq.go            # RabbitMQ integration
	│   └── storage/
	│       ├── storage.go             # Storage interface
	│       └── postgres/
	│           ├── postgres.go        # PostgreSQL driver
	│           ├── product.go         # Product queries
	│           └── user.go            # User queries
	│
	├── types/
	│   └── types.go                   # Data structures & types
	│
	└── util/
		├── DB/
		│   └── setupDB.go             # Database initialization
		├── errHandler/
		│   └── errHandler.go          # Error handling utilities
		├── response/
		│   └── response.go            # JSON response formatting
		└── serverHandler/
			└── serverHandler.go       # HTTP server utilities
```

## 🛠️ Technology Stack

| Component            | Technology              | Purpose                          |
| -------------------- | ----------------------- | -------------------------------- |
| **Language**         | Go 1.21+                | Backend services                 |
| **API Gateway**      | Nginx                   | Request routing & load balancing |
| **Message Queue**    | RabbitMQ                | Asynchronous event processing    |
| **Database**         | PostgreSQL 15           | Data persistence                 |
| **Containerization** | Docker & Docker Compose | Deployment & orchestration       |
| **Framework**        | Go std lib (net/http)   | HTTP routing                     |
| **Validation**       | go-playground/validator | Data validation                  |

## 📦 Prerequisites

### System Requirements

- **Docker** (version 20.10+)
- **Docker Compose** (version 1.29+)
- **Go** (version 1.21+) - for local development
- **Git** - for version control

### Ports Required

- **9000** - API Gateway (Nginx)
- **5432** - PostgreSQL
- **5672** - RabbitMQ (AMQP)
- **15672** - RabbitMQ Management UI

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd Devops-DSOLS
   ```

2. **Build and start all services**

   ```bash
   docker-compose up -d
   ```

3. **Verify all services are running**

   ```bash
   docker-compose ps
   ```

   Expected output:

   ```
   NAME                    STATUS
   postgres                Up
   rabbitmq                Up (healthy)
   postgres-setup          Exited (0)
   user-controller         Up
   product-controller      Up
   api-gateway             Up
   ```

4. **Test the API**

   ```bash
   # Create a user
   curl -X POST http://localhost:9000/api/user \
    -H "Content-Type: application/json" \
    -d '{"id":"USER1","email":"test@example.com"}'

   # Get all users
   curl http://localhost:9000/api/user
   ```

5. **Access Services**
   - **API Gateway**: http://localhost:9000
   - **RabbitMQ UI**: http://localhost:15672 (guest:guest)
   - **PostgreSQL**: localhost:5432

### Local Development

1. **Install dependencies**

   ```bash
   cd server
   go mod download
   ```

2. **Build the application**

   ```bash
   go build -o app main.go
   ```

3. **Initialize database** (requires running postgres-setup)

   ```bash
   docker-compose up postgres postgres-setup
   ```

4. **Run a specific component**

   ```bash
   # Run user controller
   ./app -type=controller -component=user -config=config/local.yml

   # Run user service
   ./app -type=service -component=user -config=config/local.yml

   # Initialize database
   ./app -type=init -config=config/local.yml
   ```

## 🔧 Services

### User Controller

**Port:** 9000  
**Endpoints:** `/api/user`  
**Responsibilities:**

- Accept user creation requests
- Retrieve user information
- Query user orders

### Product Controller

**Port:** 9000  
**Endpoints:** `/api/product`  
**Responsibilities:**

- Accept product creation requests
- Manage product catalog
- Update product quantities

### User Service (Consumer)

**Queue:** `user`  
**Events:**

- `user.create` - Process new user events

### Product Service (Consumer)

**Queue:** `product`  
**Events:**

- `product.create` - Process new product events
- `product.quantity` - Update product quantities

### Order Controller

**Port:** 9000  
**Endpoints:** `/order`  
**Responsibilities:**

- Accept order creation requests
- Assign order IDs and timestamps

## 📡 API Documentation

For comprehensive API documentation including all endpoints, request/response examples, and sample data, see **[API_GATEWAY_README.md](API_GATEWAY_README.md)**.

### Postman Collection

Import and test all API endpoints using our Postman Workspace:

🔗 **[DSOLS Postman Workspace](https://www.postman.com/preidiot-ream/workspace/spe-major)**

The workspace includes:

- Pre-configured requests for all endpoints
- Environment variables for different deployments
- Example request/response pairs
- API testing scripts

### Quick API Reference

| Method | Endpoint                | Description             |
| ------ | ----------------------- | ----------------------- |
| POST   | `/api/user`             | Create a new user       |
| GET    | `/api/user`             | List all users          |
| GET    | `/api/user/{id}`        | Get user by ID          |
| GET    | `/api/user/{id}/order`  | Get user's orders       |
| POST   | `/api/product`          | Create a new product    |
| GET    | `/api/product`          | List all products       |
| GET    | `/api/product/{id}`     | Get product by ID       |
| POST   | `/api/product/quantity` | Update product quantity |
| POST   | `/order`                | Create an order         |

## ⚙️ Configuration

### Environment Variables

All services use `server/config/local.yml` for configuration:

```yaml
env: "local"
storage_path: "postgres://admin:admin@postgres:5432/retail?sslmode=disable"
queue_path: "amqp://guest:guest@rabbitmq:5672/"
http_server:
  address: ":9000"
```

### Database Schema

Database schema is automatically initialized by the `postgres-setup` service on first run. This includes:

- `users` table
- `products` table
- `orders` table
- `order_items` table

## 🐳 Deployment

### Docker Compose Services

**PostgreSQL** (postgres)

- Image: `postgres:15`
- Port: 5432
- Credentials: admin/admin
- Database: retail

**RabbitMQ** (rabbitmq)

- Image: `rabbitmq:management`
- AMQP Port: 5672
- Management UI: 15672
- Credentials: guest/guest

**Postgres Setup** (postgres-setup)

- Runs initialization on startup
- Depends on: postgres, rabbitmq

**User Controller** (user-controller)

- Listens on port 9000
- Dependencies: postgres, rabbitmq

**Product Controller** (product-controller)

- Listens on port 9000
- Dependencies: postgres, rabbitmq

**API Gateway** (api-gateway)

- Routes requests to controllers
- Port: 9000 (external)
- Dependencies: all controllers

### Production Considerations

For production deployments:

1. Use environment-specific `.yml` files for different configs
2. Enable TLS/SSL certificates for secure communication
3. Set up proper database backups and replication
4. Implement logging and monitoring (ELK, Prometheus)
5. Use container orchestration (Kubernetes) for scaling
6. Implement rate limiting and authentication
7. Set up health checks and auto-restart policies

See [API_GATEWAY_README.md](API_GATEWAY_README.md) for advanced gateway configuration.

## 🛠️ Development

### Running Tests

```bash
cd server
go test ./...
```

### Building Docker Image

```bash
docker build -t devops-dsols:latest -f infra/Dockerfile .
```

### Viewing Logs

```bash
# View all services
docker-compose logs -f

# View specific service
docker-compose logs -f user-controller

# View API Gateway
docker-compose logs -f api-gateway
```

### Connecting to PostgreSQL

```bash
# Using docker exec
docker exec -it postgres psql -U admin -d retail

# Or using local psql (if installed)
psql -h localhost -p 5432 -U admin -d retail
```

### Accessing RabbitMQ

```bash
# Management UI
open http://localhost:15672
# Credentials: guest / guest

# Check queues
docker exec -it rabbitmq rabbitmqctl list_queues
```

## 🐛 Troubleshooting

### Services Won't Start

```bash
# Check logs
docker-compose logs

# Verify all ports are available
lsof -i :9000
lsof -i :5432
lsof -i :5672

# Rebuild images
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Database Connection Error

```bash
# Verify postgres is running and healthy
docker-compose ps postgres

# Check if postgres-setup completed successfully
docker-compose logs postgres-setup

# Verify database exists
docker exec -it postgres psql -U admin -l
```

### RabbitMQ Issues

```bash
# Check RabbitMQ health
docker-compose exec rabbitmq rabbitmq-diagnostics ping

# View queue status
docker-compose exec rabbitmq rabbitmqctl list_queues

# Reset (⚠️ clears queues)
docker-compose exec rabbitmq rabbitmqctl reset
```

### API Gateway Not Routing Requests

```bash
# Verify nginx configuration
docker exec api-gateway nginx -t

# Check controller connectivity from gateway
docker exec api-gateway curl http://user-controller:9000/api/user

# View gateway logs
docker logs -f api-gateway
```

### Port Already in Use

```bash
# Find process using port
lsof -i :9000

# Kill process (use with caution)
kill -9 <PID>

# Or use different port in docker-compose.yml
# Change "9000:80" to "9001:80"
```

## 📚 Additional Documentation

- **[API Gateway Documentation](API_GATEWAY_README.md)** - Detailed API routes, examples, and configuration
- **Go Module**: github.com/paragkatoch/Devops-DSOLS
- **Docker Compose**: See `docker-compose.yml` for all service configurations

## 🔐 Security Notes

⚠️ **Development Only** - Current setup uses:

- Default credentials (admin/admin for PostgreSQL, guest/guest for RabbitMQ)
- No HTTPS/TLS
- No authentication/authorization

For production, implement:

- Strong, unique credentials
- TLS certificates
- API authentication (JWT, OAuth2)
- Rate limiting
- Input validation and sanitization
- CORS policies

## 📝 License

This project is part of the DevOps & Distributed Systems learning initiative.

## 👤 Author

Parag Katoch

## 🙋 Support & Contribution

For issues, questions, or contributions:

1. Check existing issues and documentation
2. Review logs: `docker-compose logs`
3. Verify all services are healthy: `docker-compose ps`
4. Test individual services with curl commands

---

**Last Updated**: April 2026  
**Status**: Active Development
