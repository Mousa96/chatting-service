# Chat Service

[![CI](https://github.com/Mousa96/chatting-service/actions/workflows/ci.yml/badge.svg)](https://github.com/Mousa96/chatting-service/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Mousa96/chatting-service)](https://goreportcard.com/report/github.com/Mousa96/chatting-service)

A real-time chat service built with Go that supports direct messaging, broadcasting, media sharing, and WebSocket connections.

## Architecture Overview

### System Components

┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ Frontend │ │ Backend │ │ Database │
│ (HTML/JS) │◄──►│ (Go) │◄──►│ (PostgreSQL) │
└─────────────────┘ └─────────────────┘ └─────────────────┘
│
▼
┌─────────────────┐
│ File Storage │
│ (Local/S3) │
└─────────────────┘

### Backend Architecture

The backend follows a clean architecture pattern with clear separation of concerns:
cmd/
├── server/ # Application entry point
internal/
├── auth/ # Authentication module
│ ├── handler/ # HTTP handlers
│ ├── service/ # Business logic
│ ├── repository/ # Data access
│ └── models/ # Data structures
├── message/ # Messaging module
├── user/ # User management module
├── websocket/ # Real-time communication
├── middleware/ # HTTP middleware (auth, rate limiting)
├── router/ # Route configuration
├── db/ # Database utilities
└── storage/ # File storage abstraction

### Key Design Principles

- **SOLID Principles**: Each module has a single responsibility with clear interfaces
- **Dependency Injection**: Services are injected through interfaces for testability
- **Clean Architecture**: Business logic is separated from infrastructure concerns
- **Interface-Driven Design**: All major components implement interfaces for loose coupling

## Tech Stack Used

### Backend

- **Language**: Go 1.21+
- **Web Framework**: Standard library `net/http` with custom routing
- **Database**: PostgreSQL 16
- **Authentication**: JWT tokens
- **Real-time Communication**: WebSockets (gorilla/websocket)
- **File Storage**: Local filesystem (extensible to AWS S3)
- **Migrations**: golang-migrate
- **Documentation**: Swagger/OpenAPI

### Frontend

- **Languages**: HTML5, CSS3, JavaScript (ES6+)
- **Styling**: Custom CSS with responsive design
- **Real-time**: WebSocket client

### Infrastructure

- **Containerization**: Docker & Docker Compose
- **Database**: PostgreSQL with persistent volumes
- **File Storage**: Docker volumes for uploads

### Development & Testing

- **Testing**: Go standard testing + testify
- **Integration Tests**: Full HTTP testing with test database
- **Rate Limiting**: Custom in-memory rate limiter
- **CORS**: Custom middleware for cross-origin requests

## Setup Instructions

### Prerequisites

- Docker and Docker Compose
- Git

### Quick Start

1. **Clone the repository**

```bash
git clone https://github.com/Mousa96/chatting-service.git
cd chatting-service
```

2. **Start the services**

```bash
docker-compose up --build
```

3. **Access the application**

- Web Interface: http://localhost:8080
- API Documentation: http://localhost:8080/swagger/
- Health Check: http://localhost:8080/health

## API Documentation

Complete API documentation is available via Swagger UI:

- **URL**: http://localhost:8080/swagger/

## Known Limitations

### Current Limitations

1. **File Storage**: Currently uses local filesystem. For production, consider:

   - AWS S3 integration (infrastructure ready)
   - CDN for media delivery
   - File size and type restrictions

2. **Rate Limiting**: In-memory rate limiter that doesn't persist across restarts

   - Consider Redis-based rate limiting for production
   - Current limits: 10 requests/minute for most endpoints, 3/minute for broadcasts

3. **WebSocket Scaling**: Single-instance WebSocket connections

   - For horizontal scaling, implement Redis pub/sub
   - Consider WebSocket load balancing

4. **Database Connection Pooling**: Basic connection management

   - Implement connection pooling for high-load scenarios
   - Add database health checks

5. **Security Considerations**:

   - JWT secret key is hardcoded (use environment variables in production)
   - No password complexity requirements
   - No account lockout mechanisms

6. **Message Delivery**: Basic delivery status tracking

   - No offline message queuing
   - No push notifications for mobile devices

7. **Search Functionality**: No message search capabilities
   - Consider implementing full-text search
   - Message indexing for large datasets

### Production Readiness Improvements

- **Environment Configuration**: Use environment variables for all configuration
- **Logging**: Implement structured logging with log levels
- **Monitoring**: Add metrics and health monitoring
- **SSL/TLS**: HTTPS termination and secure WebSocket connections
- **Database Optimization**: Query optimization and indexing
- **Caching**: Redis caching for frequently accessed data
- **Backup Strategy**: Database backup and recovery procedures
