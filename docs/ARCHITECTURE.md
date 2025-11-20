# Gordon Watcher - Architecture

## Overview

Gordon Watcher is designed with production-grade patterns to handle high-volume file processing without overwhelming system resources.

## Components

### 1. File Watcher
- Uses `fsnotify` for efficient file system monitoring
- Recursive directory watching
- Automatic cleanup of empty directories

### 2. Worker Pool
- Fixed number of workers (prevents memory overflow)
- Buffered queue with backpressure
- Graceful shutdown

### 3. Rate Limiter
- Token bucket algorithm
- Protects downstream services
- Configurable rate per second

### 4. Circuit Breaker
- Protects against cascading failures
- Three states: Closed, Open, Half-Open
- Automatic recovery attempts

### 5. Storage Layer
- Redis for distributed deployments
- In-memory for single instance
- Idempotency guarantees

## Data Flow
```
File Detected → Worker Pool → Rate Limiter → Circuit Breaker → Queue/Storage
```

## Scaling

- Horizontal: Multiple watcher instances with Redis
- Vertical: Increase worker pool size
- Queue: RabbitMQ handles load distribution
