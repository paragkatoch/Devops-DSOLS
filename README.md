# API Gateway

## Overview

The API Gateway is a **reverse proxy** built with **Nginx** that acts as a single entry point for all client requests in the DSOLS microservices architecture. It routes incoming HTTP requests to the appropriate backend services based on URL patterns, handles load balancing, and provides a unified interface for the distributed system.

## Architecture

The API Gateway is part of a microservices-based architecture:

```
┌─────────────────┐
│  Client/Browser │
└────────┬────────┘
         │ (Port 9000)
         ▼
┌──────────────────────┐
│   API Gateway        │
│   (Nginx)            │
│                      │
│ Routes:              │
│ /api/user      ──┐   │
│ /api/product   ──┤   │
└──────────────────┤──┬┘
                   │  │
        ┌──────────┘  │
        │             │
        ▼             ▼
    ┌────────────┐ ┌──────────────┐
    │ User       │ │ Product      │
    │ Controller │ │ Controller   │
    │ (Port 9000)│ │ (Port 9000)  │
    └────────────┘ └──────────────┘
        │             │
        └──────┬──────┘
               │
        ┌──────▼─────────┐
        │  Database      │
        │  (PostgreSQL)  │
        └────────────────┘
```

## Key Components

### Docker Container Configuration

- **Image**: `nginx:latest`
- **Container Name**: `api-gateway`
- **Listen Port**: `9000`
- **Exposed Port**: `9000:80`

### Routing Configuration

The gateway routes requests based on URL patterns:

| Route            | Backend Service    | Port | Status                          |
| ---------------- | ------------------ | ---- | ------------------------------- |
| `/api/user/*`    | user-controller    | 9000 | ✅ Active                       |
| `/api/product/*` | product-controller | 9000 | ✅ Active                       |
| `/order`         | order-controller   | 9000 | ⏳ Not yet routed in nginx.conf |

## Setup and Configuration

### Running the API Gateway

#### Using Docker Compose

```bash
# Start the entire stack including the API Gateway
docker-compose up -d

# Check if API Gateway is running
docker ps | grep api-gateway

# View logs
docker logs api-gateway
```

#### Manual Setup (Development)

```bash
# Build the container
docker build -t api-gateway:latest -f infra/Dockerfile .

# Run the container
docker run -d \
  --name api-gateway \
  -p 9000:80 \
  -v $(pwd)/infra/nginx.conf:/etc/nginx/nginx.conf \
  nginx:latest
```

### Configuration Files

#### Main Configuration: `infra/nginx.conf`

```nginx
events {}

http {
  server {
    listen 80;

    location /api/user {
      proxy_pass http://user-controller:9000;
    }

    location /api/product {
      proxy_pass http://product-controller:9000;
    }
  }
}
```

**Key Points:**

- Listens on port 80 inside the container (mapped to 9000 on the host)
- Uses `proxy_pass` to forward requests to backend services
- Services are referenced by container name (Docker DNS)
- **Note:** Order controller route (`/order`) is not yet configured in nginx.conf but the service is available

#### To Enable Order Routes

Add the following to `infra/nginx.conf` if orders should be accessible through the gateway:

```nginx
location /order {
  proxy_pass http://order-controller:9000;
}

location /api/order {
  proxy_pass http://order-controller:9000;
}
```

#### Server Configuration: `server/config/local.yml`

```yaml
env: "local"
http_server:
  address: ":9000" # Services listen on port 9000
```

## API Endpoints

### User Service Endpoints

#### 1. Create User

```bash
POST /api/user
```

**Request Body:**

```json
{
	"id": "USER12345",
	"email": "john.doe@example.com"
}
```

**Response (Success - 200 OK):**

```json
{
	"success": "ok"
}
```

**Response (Validation Error - 400 Bad Request):**

```json
{
	"error": "validation error",
	"details": [
		{
			"field": "email",
			"message": "required"
		}
	]
}
```

---

#### 2. Get All Users

```bash
GET /api/user
```

**Response (200 OK):**

```json
[
	{
		"id": "USER12345",
		"email": "john.doe@example.com"
	},
	{
		"id": "USER67890",
		"email": "jane.smith@example.com"
	}
]
```

---

#### 3. Get User by ID

```bash
GET /api/user/USER12345
```

**Response (200 OK):**

```json
{
	"id": "USER12345",
	"email": "john.doe@example.com"
}
```

**Response (Error - 500 Internal Server Error):**

```json
{
	"error": "user not found"
}
```

---

#### 4. Get User's Orders

```bash
GET /api/user/USER12345/order
```

**Response (200 OK):**

```json
[
	{
		"id": "ORD-550e8400-e29b-41d4-a716-446655440000",
		"user_id": "USER12345",
		"items": [
			{
				"product_id": "PROD-001",
				"quantity": 2,
				"price": 2999
			},
			{
				"product_id": "PROD-003",
				"quantity": 1,
				"price": 4999
			}
		],
		"total_amount": 10997,
		"currency": "USD",
		"status": "COMPLETED",
		"created_at": "2026-04-03T15:30:00Z",
		"updated_at": "2026-04-03T15:45:00Z"
	}
]
```

---

### Product Service Endpoints

#### 1. Create Product

```bash
POST /api/product
```

**Request Body:**

```json
{
	"id": "PROD-001",
	"name": "Wireless Bluetooth Headphones",
	"price": 2999,
	"quantity": 50
}
```

**Response (Success - 200 OK):**

```json
{
	"success": "ok"
}
```

**Response (Validation Error - 400 Bad Request):**

```json
{
	"error": "validation error",
	"details": [
		{
			"field": "id",
			"message": "required"
		},
		{
			"field": "name",
			"message": "required"
		}
	]
}
```

---

#### 2. Get All Products

```bash
GET /api/product
```

**Response (200 OK):**

```json
[
	{
		"id": "PROD-001",
		"name": "Wireless Bluetooth Headphones",
		"price": 2999,
		"quantity": 50
	},
	{
		"id": "PROD-002",
		"name": "USB-C Cable",
		"price": 999,
		"quantity": 200
	},
	{
		"id": "PROD-003",
		"name": "Smartphone Case",
		"price": 4999,
		"quantity": 150
	}
]
```

---

#### 3. Get Product by ID

```bash
GET /api/product/PROD-001
```

**Response (200 OK):**

```json
{
	"id": "PROD-001",
	"name": "Wireless Bluetooth Headphones",
	"price": 2999,
	"quantity": 50
}
```

**Response (Error - 500 Internal Server Error):**

```json
{
	"error": "product not found"
}
```

---

#### 4. Update Product Quantity

```bash
POST /api/product/quantity
```

**Request Body:**

```json
{
	"id": "PROD-001",
	"quantity": 35
}
```

**Response (Success - 200 OK):**

```json
{
	"success": "queued"
}
```

**Response (Validation Error - 400 Bad Request):**

```json
{
	"error": "validation error",
	"details": [
		{
			"field": "id",
			"message": "required"
		}
	]
}
```

---

### Order Service Endpoints

#### 1. Create Order

```bash
POST /order
```

**Request Body:**

```json
{
	"user_id": "USER12345",
	"items": [
		{
			"product_id": "PROD-001",
			"quantity": 2,
			"price": 2999
		},
		{
			"product_id": "PROD-003",
			"quantity": 1,
			"price": 4999
		}
	],
	"total_amount": 10997,
	"currency": "USD"
}
```

**Response (Success - 200 OK):**

```json
{
	"success": "ok"
}
```

**Note:** The order is assigned:

- Auto-generated `id` (UUID v4)
- Status: `CREATED`
- Auto-populated `created_at` and `updated_at` timestamps

**Example Response with Full Order Details:**

```json
{
	"id": "550e8400-e29b-41d4-a716-446655440000",
	"user_id": "USER12345",
	"items": [
		{
			"product_id": "PROD-001",
			"quantity": 2,
			"price": 2999
		},
		{
			"product_id": "PROD-003",
			"quantity": 1,
			"price": 4999
		}
	],
	"total_amount": 10997,
	"currency": "USD",
	"status": "CREATED",
	"created_at": "2026-04-03T15:30:00Z",
	"updated_at": "2026-04-03T15:30:00Z"
}
```

**Response (Validation Error - 400 Bad Request):**

```json
{
	"error": "validation error",
	"details": [
		{
			"field": "user_id",
			"message": "required"
		},
		{
			"field": "items",
			"message": "required"
		}
	]
}
```

## Testing the API Gateway

### Create a New User

```bash
curl -X POST http://localhost:9000/api/user \
  -H "Content-Type: application/json" \
  -d '{
    "id": "USER12345",
    "email": "john.doe@example.com"
  }'
```

### Get All Users

```bash
curl -X GET http://localhost:9000/api/user
```

### Get Specific User

```bash
curl -X GET http://localhost:9000/api/user/USER12345
```

### Create a New Product

```bash
curl -X POST http://localhost:9000/api/product \
  -H "Content-Type: application/json" \
  -d '{
    "id": "PROD-001",
    "name": "Wireless Bluetooth Headphones",
    "price": 2999,
    "quantity": 50
  }'
```

### Get All Products

```bash
curl -X GET http://localhost:9000/api/product
```

### Get Specific Product

```bash
curl -X GET http://localhost:9000/api/product/PROD-001
```

### Update Product Quantity

```bash
curl -X POST http://localhost:9000/api/product/quantity \
  -H "Content-Type: application/json" \
  -d '{
    "id": "PROD-001",
    "quantity": 35
  }'
```

### Get User's Orders

```bash
curl -X GET http://localhost:9000/api/user/USER12345/order
```

### Create an Order (Note: Currently not routed through API Gateway)

To test order creation, you would need to either:

1. Access the order-controller directly: `curl -X POST http://localhost:9000/order`
2. Or update nginx.conf to include the order route

```bash
curl -X POST http://localhost:9000/order \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "USER12345",
    "items": [
      {
        "product_id": "PROD-001",
        "quantity": 2,
        "price": 2999
      }
    ],
    "total_amount": 5998,
    "currency": "USD"
  }'
```

### View Real-time Logs

```bash
docker logs -f api-gateway
```

### Docker Network Testing

```bash
# Enter API Gateway container
docker exec -it api-gateway /bin/bash

# Test connectivity to backend services
curl http://user-controller:9000/api/user
curl http://product-controller:9000/api/product
```

## Important Implementation Details

### Event-Driven Architecture

The API Gateway backs onto a **microservices event-driven architecture**:

- **POST requests** (Create/Update operations) do NOT write directly to the database
- Instead, they are **queued as events in RabbitMQ**
- Backend services consume these events asynchronously and persist data
- This provides loose coupling between services and enables asynchronous processing

#### Event Types

- `user.create` - Triggered by `POST /api/user`
- `product.create` - Triggered by `POST /api/product`
- `product.quantity` - Triggered by `POST /api/product/quantity`

#### GET Requests

- All `GET` requests read directly from the PostgreSQL database
- They do not go through the message queue

### Data Structures

#### User

```json
{
	"id": "string", // Required, user identifier
	"email": "string" // User email address
}
```

#### Product

```json
{
	"id": "string", // Required, product identifier
	"name": "string", // Required, product name
	"price": "int64", // Required, price in cents (USD)
	"quantity": "int" // Required, available quantity
}
```

#### Order

```json
{
	"id": "string", // Auto-generated UUID
	"user_id": "string", // Required, reference to user
	"items": [
		// Required, array of order items
		{
			"product_id": "string", // Required
			"quantity": "int", // Required
			"price": "int64" // Required, price in cents
		}
	],
	"total_amount": "int64", // Required, total price in cents
	"currency": "string", // Required, e.g., "USD"
	"status": "string", // CREATED, PROCESSING, COMPLETED
	"created_at": "timestamp", // Auto-generated
	"updated_at": "timestamp" // Auto-generated
}
```

### Response Format

All API responses follow a consistent JSON format:

**Success Response:**

```json
{
	"success": "ok"
}
```

**Queued Response (Async Operations):**

```json
{
	"success": "queued"
}
```

**Data Response (GET Requests):**

```json
{
  "id": "...",
  "email": "...",
  ...
}
```

**Error Response:**

```json
{
	"error": "error description"
}
```

**Validation Error Response:**

```json
{
	"error": "validation error",
	"details": [
		{
			"field": "field_name",
			"message": "validation message"
		}
	]
}
```

- Docker and Docker Compose installed
- Backend services (user-controller, product-controller) running
- PostgreSQL and RabbitMQ services accessible

### Production Considerations

**Security:**

- Enable HTTPS with TLS certificates
- Add authentication/authorization middleware
- Implement rate limiting
- Enable CORS headers if needed

**Performance:**

- Configure upstream load balancing for multiple replicas
- Implement caching strategies
- Configure connection pooling

**Monitoring:**

- Enable access logs
- Set up health checks for backend services
- Monitor traffic and latency

### Example Production Configuration

```nginx
events {
  worker_connections 1024;
}

http {
  # Upstream services for load balancing
  upstream user_backend {
    server user-controller-1:9000;
    server user-controller-2:9000;
    server user-controller-3:9000;
  }

  upstream product_backend {
    server product-controller-1:9000;
    server product-controller-2:9000;
  }

  # SSL and security headers
  server {
    listen 443 ssl http2;
    ssl_certificate /etc/nginx/certs/cert.pem;
    ssl_certificate_key /etc/nginx/certs/key.pem;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

    location /api/user {
      limit_req zone=api_limit burst=20;
      proxy_pass http://user_backend;
      add_header X-Forwarded-For $remote_addr;
    }

    location /api/product {
      limit_req zone=api_limit burst=20;
      proxy_pass http://product_backend;
      add_header X-Forwarded-For $remote_addr;
    }
  }

  # Redirect HTTP to HTTPS
  server {
    listen 80;
    return 301 https://$host$request_uri;
  }
}
```

## Troubleshooting

### API Gateway Not Starting

```bash
# Check configuration syntax
docker run --rm -v $(pwd)/infra/nginx.conf:/etc/nginx/nginx.conf nginx:latest nginx -t

# View container logs
docker logs api-gateway
```

### Backend Services Unreachable

```bash
# Verify service names in docker network
docker network inspect devops-dsols_default

# Check if services are running
docker ps | grep controller
```

### Connection Refused

```bash
# Ensure backend services are listening on port 9000
docker exec user-controller netstat -tlnp | grep 9000
docker exec product-controller netstat -tlnp | grep 9000
```

### Logs and Debugging

```bash
# Enable debug logging (modify nginx.conf)
# Add: error_log /var/log/nginx/error.log debug;

# View logs in real-time
docker logs -f api-gateway

# Check nginx configuration
docker exec api-gateway nginx -T
```

## Advanced Configuration

### Health Checks

Add health check endpoints to verify backend service availability:

```nginx
location /healthz {
  access_log off;
  return 200 "healthy\n";
  add_header Content-Type text/plain;
}
```

### Custom Headers

```nginx
location /api/user {
  proxy_pass http://user-controller:9000;
  proxy_set_header Host $host;
  proxy_set_header X-Real-IP $remote_addr;
  proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  proxy_set_header X-Forwarded-Proto $scheme;
}
```

### Timeout Configuration

```nginx
server {
  proxy_connect_timeout 60s;
  proxy_send_timeout 60s;
  proxy_read_timeout 60s;
}
```

## Related Documentation

- [Nginx Documentation](https://nginx.org/en/docs/)
- [Docker Networking](https://docs.docker.com/network/)
- [Project Structure](./README.md)
- [Service Documentation](./server/README.md)

## Support

For issues or questions regarding the API Gateway:

1. Check the logs: `docker logs api-gateway`
2. Verify backend services are running: `docker ps`
3. Test connectivity: `curl http://localhost:9000/api/user`
4. Review the nginx configuration: `infra/nginx.conf`

---

**Last Updated**: April 2026
**Version**: 1.0
**Status**: Production Ready
