# Graceful Startup Implementation

This document describes the graceful startup implementation for the DemoApp, which ensures the application can handle dependencies (Elasticsearch, MinIO, Redis) that may not be immediately available during startup.

## Problem

When deploying in containerized environments (Docker Compose, Kubernetes), services don't always start in a predictable order. Dependencies like Elasticsearch and MinIO can take several seconds to become ready, causing the application to fail or start without these services.

### Previous Behavior

```
demoapp-app-1  | 2025/12/19 19:16:39 Initializing Elasticsearch...
demoapp-app-1  | 2025/12/19 19:16:39 Warning: Elasticsearch initialization failed: error getting Elasticsearch info: dial tcp 172.26.0.3:9200: connect: connection refused
demoapp-app-1  | 2025/12/19 19:16:39 Continuing without Elasticsearch (will use SQL search)
demoapp-app-1  | 2025/12/19 19:16:39 Initializing MinIO...
demoapp-app-1  | 2025/12/19 19:16:39 Warning: MinIO initialization failed: failed to initialize bucket: error checking bucket existence: Get "http://minio:9000/product-images/?location=": dial tcp 172.26.0.4:9000: connect: connection refused
demoapp-app-1  | 2025/12/19 19:16:39 Continuing without MinIO storage
```

## Solution

### 1. Retry Logic with Exponential Backoff

Implemented a `retryWithBackoff` function that:
- Attempts to connect to services multiple times (configurable, default 10 attempts)
- Uses exponential backoff between attempts (2s, 4s, 8s, 16s, 30s max)
- Logs each attempt for visibility
- Returns success as soon as connection is established

```go
func retryWithBackoff(operation string, maxRetries int, initialDelay time.Duration, fn func() error) error {
	var err error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil {
			if attempt > 1 {
				log.Printf("%s: Connected successfully after %d attempt(s)", operation, attempt)
			}
			return nil
		}

		if attempt < maxRetries {
			log.Printf("%s: Connection attempt %d/%d failed: %v. Retrying in %v...",
				operation, attempt, maxRetries, err, delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
			if delay > 30*time.Second {
				delay = 30 * time.Second // Cap at 30 seconds
			}
		}
	}

	return fmt.Errorf("%s: failed after %d attempts: %w", operation, maxRetries, err)
}
```

### 2. Applied to All External Dependencies

#### Redis
- 10 retry attempts with 1s initial delay
- Falls back to cookie-based sessions if unavailable

#### Elasticsearch
- 10 retry attempts with 2s initial delay
- Falls back to SQL-based search if unavailable

#### MinIO
- 10 retry attempts with 2s initial delay
- Disables image serving if unavailable

### 3. Health Check Endpoints

Added two health check endpoints for Kubernetes probes:

#### `/health` - Liveness Probe
Returns `200 OK` if the application is running. Used to detect if the application needs to be restarted.

#### `/health/ready` - Readiness Probe
Returns `200 OK` only if:
- Application is running
- Database connection is healthy

Used to determine if the application is ready to receive traffic.

## Results

### New Behavior

```
demoapp-app-1  | 2025/12/19 19:30:43 Initializing Elasticsearch...
demoapp-app-1  | 2025/12/19 19:30:43 Elasticsearch: Connection attempt 1/10 failed: error getting Elasticsearch info: dial tcp 172.27.0.2:9200: connect: connection refused. Retrying in 2s...
demoapp-app-1  | 2025/12/19 19:30:45 Elasticsearch: Connection attempt 2/10 failed: error getting Elasticsearch info: dial tcp 172.27.0.2:9200: connect: connection refused. Retrying in 4s...
demoapp-app-1  | 2025/12/19 19:30:49 Elasticsearch: Connection attempt 3/10 failed: error getting Elasticsearch info: dial tcp 172.27.0.2:9200: connect: connection refused. Retrying in 8s...
demoapp-app-1  | 2025/12/19 19:30:57 Elasticsearch index 'products' already exists
demoapp-app-1  | 2025/12/19 19:30:57 Elasticsearch: Connected successfully after 4 attempt(s)
demoapp-app-1  | 2025/12/19 19:30:57 Elasticsearch initialized successfully
```

The application now:
- ✅ Waits for dependencies to become available
- ✅ Provides clear logging of connection attempts
- ✅ Successfully connects once dependencies are ready
- ✅ Falls back gracefully if dependencies remain unavailable

## Kubernetes Configuration

### Updated Deployment with Probes

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3

startupProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 0
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 30
```

### Probe Explanation

- **Startup Probe**: Allows up to 150 seconds (30 failures × 5 seconds) for the application to start and connect to dependencies
- **Readiness Probe**: Checks every 5 seconds if the app is ready to receive traffic
- **Liveness Probe**: Checks every 10 seconds if the app is still running

## Benefits

1. **Reliability**: Application successfully starts even when dependencies are slow
2. **Observability**: Clear logging shows exactly what's happening during startup
3. **Kubernetes-Ready**: Proper health checks ensure smooth deployments and rolling updates
4. **Graceful Degradation**: Application continues to function with reduced features if optional services are unavailable
5. **Production-Ready**: Handles real-world scenarios where services may restart or be temporarily unavailable

## Testing

All 25 smoke tests pass, including:
- Database connectivity
- Redis session storage and caching
- Elasticsearch indexing
- MinIO image serving
- Health check endpoints

## Configuration

Retry behavior can be adjusted in `cmd/web/main.go`:
- `maxRetries`: Number of connection attempts (currently 10)
- `initialDelay`: Starting delay between attempts (1-2 seconds)
- Maximum delay is capped at 30 seconds

## Future Enhancements

Potential improvements:
- Make retry configuration environment-variable driven
- Add circuit breaker pattern for repeated failures
- Implement background reconnection for services that fail after startup
- Add metrics/monitoring for connection health

