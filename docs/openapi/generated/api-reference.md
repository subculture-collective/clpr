---
title: "API Reference"
summary: "Complete API reference for Clipper platform"
tags: ["api", "reference", "openapi"]
area: "openapi"
status: "stable"
version: "1.0.0"
generated: 2026-07-12T17:41:02.029Z
---

# Clipper API

Complete API documentation for the Clipper platform - a social platform for Twitch clip curation and sharing.

## Authentication
Most endpoints require authentication via JWT Bearer token:
```
Authorization: Bearer <your_jwt_token>
```

## Rate Limiting
API endpoints are rate-limited to prevent abuse. Rate limits vary by endpoint and are documented in each operation.
Common limits:
- Public endpoints: 60-100 requests/minute
- Authenticated actions: 10-30 requests/minute
- Submission operations: 5-10 requests/hour

## Pagination
List endpoints support pagination via `page` and `limit` query parameters.
Default limit is 20, maximum is 100.

## Errors
The API uses standard HTTP status codes:
- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 429: Too Many Requests
- 500: Internal Server Error


**Version:** 1.0.0

## Base URLs

- **Local development server:** `http://localhost:8080`
- **Staging environment:** `https://staging.clpr.tv`
- **Production API server:** `https://api.clpr.tv`

## Table of Contents

- [Health](#health)
- [Authentication](#authentication)
- [MFA](#mfa)
- [Clips](#clips)
- [Comments](#comments)
- [Tags](#tags)
- [Search](#search)
- [Submissions](#submissions)
- [Reports](#reports)
- [Moderation](#moderation)
- [Users](#users)
- [Broadcasters](#broadcasters)
- [Webhooks](#webhooks)
- [Documentation](#documentation)
- [Watch History](#watch-history)
- [Admin](#admin)

## Health

Health check and monitoring endpoints

### Get sitemap XML

`GET /sitemap.xml`

Returns sitemap for search engine crawlers

**Tags:** Health

#### Responses

**200** - Sitemap XML

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/sitemap.xml"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/sitemap.xml', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/sitemap.xml'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/sitemap.xml", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get robots.txt

`GET /robots.txt`

Returns robots.txt for search engine crawlers

**Tags:** Health

#### Responses

**200** - robots.txt content

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/robots.txt"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/robots.txt', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/robots.txt'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/robots.txt", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Readiness check

`GET /health/ready`

Returns 200 when required services are ready; optional dependency failures are listed as degradation

**Tags:** Health

#### Responses

**200** - Service is ready

**503** - A required dependency is unavailable

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/health/ready"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/health/ready', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/health/ready'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/health/ready", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Liveness check

`GET /health/live`

Returns 200 if service is alive

**Tags:** Health

#### Responses

**200** - Service is alive

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/health/live"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/health/live', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/health/live'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/health/live", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Health statistics

`GET /health/stats`

Returns database connection stats and other health metrics

**Tags:** Health

#### Responses

**200** - Health statistics

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/health/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/health/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/health/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/health/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Cache statistics

`GET /health/cache`

Returns Redis cache statistics

**Tags:** Health

#### Responses

**200** - Cache statistics

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/health/cache"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/health/cache', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/health/cache'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/health/cache", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Cache health check

`GET /health/cache/check`

Verifies Redis cache connectivity

**Tags:** Health

#### Responses

**200** - Cache is healthy

**503** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/health/cache/check"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/health/cache/check', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/health/cache/check'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/health/cache/check", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Webhook retry statistics

`GET /health/webhooks`

Returns webhook delivery retry statistics

**Tags:** Health

#### Responses

**200** - Webhook statistics

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/health/webhooks"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/health/webhooks', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/health/webhooks'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/health/webhooks", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Prometheus metrics

`GET /debug/metrics`

Returns Prometheus metrics (debug mode only)

**Tags:** Health

#### Responses

**200** - Prometheus metrics

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/metrics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/metrics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/metrics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/metrics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### API health check

`GET /api/v1/health`

Basic health check for v1 API

**Tags:** Health

#### Responses

**200** - API is healthy

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/health"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/health', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/health'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/health", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Ping endpoint

`GET /api/v1/ping`

Simple ping to check API responsiveness

**Tags:** Health

#### Responses

**200** - Pong response

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/ping"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/ping', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/ping'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/ping", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get public configuration

`GET /api/v1/config`

Returns public configuration values (features, limits, etc.)

**Tags:** Health

#### Responses

**200** - Public configuration

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/config"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/config', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/config'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/config", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Submit application logs

`POST /api/v1/logs`

Submit client-side application logs (rate limited - 60/minute)

**Tags:** Health

#### Request Body

Content-Type: `application/json`

#### Responses

**204** - Log submitted successfully

**400** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/logs" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/logs', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/logs',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/logs", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get log statistics

`GET /api/v1/logs/stats`

Returns aggregated log statistics (rate limited - 30/minute)

**Tags:** Health

#### Responses

**200** - Log statistics

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/logs/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/logs/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/logs/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/logs/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Authentication

User authentication and OAuth

### Initiate Twitch OAuth

`GET /api/v1/auth/twitch`

Redirects to Twitch OAuth authorization (rate limited - 30/minute)

**Tags:** Authentication

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| redirect_uri | query | string |  |  |

#### Responses

**302** - Redirect to Twitch OAuth

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/auth/twitch"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/twitch', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/auth/twitch'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/auth/twitch", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Twitch OAuth callback

`GET /api/v1/auth/twitch/callback`

Handles OAuth callback from Twitch (rate limited - 50/minute)

**Tags:** Authentication

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| code | query | string | ✓ |  |
| state | query | string | ✓ |  |

#### Responses

**200** - Authentication successful

**400** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/auth/twitch/callback"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/twitch/callback', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/auth/twitch/callback'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/auth/twitch/callback", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PKCE OAuth callback

`POST /api/v1/auth/twitch/callback`

Handles PKCE OAuth callback (rate limited - 50/minute)

**Tags:** Authentication

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Authentication successful

**400** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/twitch/callback" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/twitch/callback', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/twitch/callback',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/auth/twitch/callback", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Test login (development only)

`POST /api/v1/auth/test-login`

Creates test user session for development (rate limited - 30/minute)

**Tags:** Authentication

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Test login successful

**400** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/test-login" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/test-login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/test-login',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/auth/test-login", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Refresh access token

`POST /api/v1/auth/refresh`

Refresh JWT access token using refresh token (rate limited - 50/minute)

**Tags:** Authentication

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Token refreshed successfully

**400** - Success

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/refresh', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/refresh',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Logout

`POST /api/v1/auth/logout`

Invalidates current session and refresh tokens

**Tags:** Authentication

#### Responses

**200** - Logout successful

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/logout"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/logout', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/logout'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/auth/logout", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get current user

`GET /api/v1/auth/me`

Returns currently authenticated user

**Tags:** Authentication

#### Responses

**200** - Current user

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/auth/me"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/me', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/auth/me'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/auth/me", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Reauthorize Twitch

`POST /api/v1/auth/twitch/reauthorize`

Re-initiates Twitch OAuth flow for additional scopes (rate limited - 3/hour)

**Tags:** Authentication

#### Responses

**200** - Reauthorization initiated

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/twitch/reauthorize"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/twitch/reauthorize', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/twitch/reauthorize'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/auth/twitch/reauthorize", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## MFA

Multi-factor authentication

### Start MFA enrollment

`POST /api/v1/auth/mfa/enroll`

Initiates MFA enrollment process (rate limited - 3/hour)

**Tags:** MFA

#### Responses

**200** - MFA enrollment started

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/mfa/enroll"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/mfa/enroll', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/mfa/enroll'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/auth/mfa/enroll", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Verify MFA enrollment

`POST /api/v1/auth/mfa/verify-enrollment`

Completes MFA enrollment by verifying TOTP code (rate limited - 10/minute)

**Tags:** MFA

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - MFA enrollment verified

**400** - Success

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/mfa/verify-enrollment" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/mfa/verify-enrollment', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/mfa/verify-enrollment',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/auth/mfa/verify-enrollment", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get MFA status

`GET /api/v1/auth/mfa/status`

Returns MFA enrollment status for current user

**Tags:** MFA

#### Responses

**200** - MFA status

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/auth/mfa/status"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/mfa/status', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/auth/mfa/status'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/auth/mfa/status", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Regenerate backup codes

`POST /api/v1/auth/mfa/regenerate-backup-codes`

Generates new backup codes (rate limited - 5/hour)

**Tags:** MFA

#### Responses

**200** - New backup codes generated

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/mfa/regenerate-backup-codes"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/mfa/regenerate-backup-codes', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/mfa/regenerate-backup-codes'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/auth/mfa/regenerate-backup-codes", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Disable MFA

`POST /api/v1/auth/mfa/disable`

Disables MFA for current user (rate limited - 3/hour)

**Tags:** MFA

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - MFA disabled

**400** - Success

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/mfa/disable" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/mfa/disable', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/mfa/disable',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/auth/mfa/disable", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get trusted devices

`GET /api/v1/auth/mfa/trusted-devices`

Returns list of trusted devices for MFA

**Tags:** MFA

#### Responses

**200** - List of trusted devices

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/auth/mfa/trusted-devices"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/mfa/trusted-devices', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/auth/mfa/trusted-devices'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/auth/mfa/trusted-devices", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Revoke trusted device

`DELETE /api/v1/auth/mfa/trusted-devices/{id}`

Removes a device from trusted devices list

**Tags:** MFA

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - Trusted device revoked

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/auth/mfa/trusted-devices/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/mfa/trusted-devices/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/auth/mfa/trusted-devices/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/auth/mfa/trusted-devices/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Verify MFA login

`POST /api/v1/auth/mfa/verify-login`

Verifies MFA code during login (rate limited - 10/minute)

**Tags:** MFA

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - MFA verification successful

**400** - Success

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/auth/mfa/verify-login" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/auth/mfa/verify-login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/auth/mfa/verify-login',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/auth/mfa/verify-login", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Clips

Clip management and discovery

### List clips

`GET /api/v1/clips`

Returns paginated list of clips with filtering and sorting options

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
| sort | query | string |  |  |
| game | query | string |  | Filter by game name |
| broadcaster | query | string |  | Filter by broadcaster name |
| tag | query | string |  | Filter by tag slug |
| language | query | string |  | Filter by language code |
| time_range | query | string |  | Time range for 'top' sort |

#### Responses

**200** - List of clips

**400** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get clip by ID

`GET /api/v1/clips/{id}`

Returns detailed information about a specific clip

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Clip details

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update clip (Admin)

`PUT /api/v1/clips/{id}`

Updates clip metadata (admin/moderator only)

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Clip updated

**400** - Success

**401** - Success

**403** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/clips/{id}" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/clips/{id}',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/clips/{id}", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Delete clip (Admin)

`DELETE /api/v1/clips/{id}`

Permanently deletes a clip (admin only)

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - Clip deleted

**401** - Success

**403** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/clips/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/clips/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/clips/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get related clips

`GET /api/v1/clips/{id}/related`

Returns clips related to the specified clip

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Related clips

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}/related"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/related', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}/related'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}/related", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Batch get clip media URLs

`POST /api/v1/clips/batch-media`

Returns media URLs for multiple clips (rate limited - 60/minute)

**Tags:** Clips

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Clip media URLs

**400** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/clips/batch-media" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/batch-media', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/clips/batch-media',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/clips/batch-media", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get clip analytics

`GET /api/v1/clips/{id}/analytics`

Returns analytics data for a specific clip

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Clip analytics

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}/analytics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/analytics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}/analytics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}/analytics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Track clip view

`POST /api/v1/clips/{id}/track-view`

Records a view event for analytics

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - View tracked

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/clips/{id}/track-view" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/track-view', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/clips/{id}/track-view',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/clips/{id}/track-view", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get clip engagement score

`GET /api/v1/clips/{id}/engagement`

Returns engagement metrics and score for a clip

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Engagement score

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}/engagement"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/engagement', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}/engagement'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}/engagement", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Vote on clip

`POST /api/v1/clips/{id}/vote`

Upvote or downvote a clip (rate limited - 20/minute)

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Vote recorded

**400** - Success

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/clips/{id}/vote" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/vote', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/clips/{id}/vote',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/clips/{id}/vote", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Add clip to favorites

`POST /api/v1/clips/{id}/favorite`

Adds a clip to user's favorites

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**201** - Clip favorited

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/clips/{id}/favorite"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/favorite', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/clips/{id}/favorite'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/clips/{id}/favorite", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Remove clip from favorites

`DELETE /api/v1/clips/{id}/favorite`

Removes a clip from user's favorites

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - Favorite removed

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/clips/{id}/favorite"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/favorite', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/clips/{id}/favorite'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/clips/{id}/favorite", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update clip metadata

`PUT /api/v1/clips/{id}/metadata`

Updates clip metadata (creator/submitter only, rate limited - 10/minute)

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Metadata updated

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/clips/{id}/metadata" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/metadata', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/clips/{id}/metadata',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/clips/{id}/metadata", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update clip visibility

`PUT /api/v1/clips/{id}/visibility`

Updates clip visibility (creator/submitter only, rate limited - 10/minute)

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Visibility updated

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/clips/{id}/visibility" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/visibility', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/clips/{id}/visibility',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/clips/{id}/visibility", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Request clip sync

`POST /api/v1/clips/request`

Request a specific clip to be synced from Twitch (rate limited - 10/hour)

**Tags:** Clips

#### Request Body

Content-Type: `application/json`

#### Responses

**202** - Clip sync requested

**400** - Success

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/clips/request" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/request', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/clips/request',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/clips/request", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List scraped clips

`GET /api/v1/scraped-clips`

Returns clips that haven't been claimed by users yet

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - List of scraped clips

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/scraped-clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/scraped-clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/scraped-clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/scraped-clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List user favorites

`GET /api/v1/favorites`

Returns current user's favorite clips

**Tags:** Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - List of favorite clips

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/favorites"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/favorites', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/favorites'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/favorites", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Run a clip synchronization

`POST /api/v1/admin/sync/clips`

Runs a bounded synchronous Twitch clip import. Requires administrator MFA; cookie-authenticated requests must also send the CSRF token. Item-level failures produce 207 rather than full success.

**Tags:** Admin, Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Sync completed without item errors

**207** - Sync completed with one or more item errors

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/sync/clips" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/sync/clips', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/sync/clips',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/admin/sync/clips", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get persisted clip synchronization status

`GET /api/v1/admin/sync/status`

Returns the most recent persisted clip import time rather than a synthetic service-ready placeholder. Requires administrator MFA.

**Tags:** Admin, Clips

#### Responses

**200** - Last persisted synchronization evidence

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/sync/status"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/sync/status', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/sync/status'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/sync/status", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Tags

Tag management and discovery

### Get clip tags

`GET /api/v1/clips/{id}/tags`

Returns all tags associated with a clip

**Tags:** Tags

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Clip tags

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}/tags"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/tags', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}/tags'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}/tags", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Add tags to clip

`POST /api/v1/clips/{id}/tags`

Adds tags to a clip (rate limited - 10/minute)

**Tags:** Tags

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Tags added

**400** - Success

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/clips/{id}/tags" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/tags', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/clips/{id}/tags',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/clips/{id}/tags", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Remove tag from clip

`DELETE /api/v1/clips/{id}/tags/{slug}`

Removes a tag from a clip

**Tags:** Tags

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**204** - Tag removed

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/clips/{id}/tags/{slug}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/tags/{slug}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/clips/{id}/tags/{slug}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/clips/{id}/tags/{slug}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List tags

`GET /api/v1/tags`

Returns list of all tags

**Tags:** Tags

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - List of tags

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/tags"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/tags', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/tags'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/tags", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Search tags

`GET /api/v1/tags/search`

Search tags by name (rate limited - 60/minute)

**Tags:** Tags

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| q | query | string | ✓ | Search query |
|  |  | string |  |  |

#### Responses

**200** - Search results

**400** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/tags/search"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/tags/search', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/tags/search'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/tags/search", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get tag by slug

`GET /api/v1/tags/{slug}`

Returns detailed information about a tag

**Tags:** Tags

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Tag details

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/tags/{slug}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/tags/{slug}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/tags/{slug}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/tags/{slug}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get clips by tag

`GET /api/v1/tags/{slug}/clips`

Returns clips with the specified tag

**Tags:** Tags

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |
| sort | query | string |  |  |

#### Responses

**200** - List of clips

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/tags/{slug}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/tags/{slug}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/tags/{slug}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/tags/{slug}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Watch History

Watch history and progress

### Get resume position

`GET /api/v1/clips/{id}/progress`

Returns saved watch progress for clip (optional auth)

**Tags:** Watch History

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Resume position

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}/progress"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/progress', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}/progress'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}/progress", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Comments

Comment and reply operations

### List clip comments

`GET /api/v1/clips/{id}/comments`

Returns paginated list of comments for a clip

**Tags:** Comments

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |
| sort | query | string |  |  |

#### Responses

**200** - List of comments

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}/comments"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/comments', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}/comments'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}/comments", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Create comment

`POST /api/v1/clips/{id}/comments`

Creates a new comment on a clip (rate limited - 10/minute)

**Tags:** Comments

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Comment created

**400** - Success

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/clips/{id}/comments" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/comments', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/clips/{id}/comments',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/clips/{id}/comments", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get comment replies

`GET /api/v1/comments/{id}/replies`

Returns replies to a specific comment

**Tags:** Comments

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - List of replies

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/comments/{id}/replies"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/comments/{id}/replies', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/comments/{id}/replies'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/comments/{id}/replies", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update comment

`PUT /api/v1/comments/{id}`

Updates a comment (author only)

**Tags:** Comments

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Comment updated

**400** - Success

**401** - Success

**403** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/comments/{id}" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/comments/{id}', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/comments/{id}',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/comments/{id}", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Delete comment

`DELETE /api/v1/comments/{id}`

Deletes a comment (author/admin only)

**Tags:** Comments

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - Comment deleted

**401** - Success

**403** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/comments/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/comments/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/comments/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/comments/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Vote on comment

`POST /api/v1/comments/{id}/vote`

Upvote or downvote a comment (rate limited - 20/minute)

**Tags:** Comments

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Vote recorded

**400** - Success

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/comments/{id}/vote" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/comments/{id}/vote', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/comments/{id}/vote',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/comments/{id}/vote", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Search

Search functionality

### Search clips

`GET /api/v1/search`

Full-text search for clips (rate limited - 60/minute)

**Tags:** Search

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| q | query | string | ✓ | Search query |
|  |  | string |  |  |
|  |  | string |  |  |
| type | query | string |  |  |

#### Responses

**200** - Search results

**400** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/search"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/search', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/search'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/search", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get search suggestions

`GET /api/v1/search/suggestions`

Returns autocomplete suggestions (rate limited - 60/minute)

**Tags:** Search

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| q | query | string | ✓ |  |

#### Responses

**200** - Suggestions

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/search/suggestions"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/search/suggestions', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/search/suggestions'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/search/suggestions", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Search with scores

`GET /api/v1/search/scores`

Hybrid search with similarity scores (rate limited - 60/minute)

**Tags:** Search

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| q | query | string | ✓ |  |
|  |  | string |  |  |

#### Responses

**200** - Search results with scores

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/search/scores"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/search/scores', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/search/scores'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/search/scores", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get trending searches

`GET /api/v1/search/trending`

Returns popular search queries (rate limited - 30/minute)

**Tags:** Search

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Trending searches

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/search/trending"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/search/trending', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/search/trending'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/search/trending", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get search history

`GET /api/v1/search/history`

Returns user's search history

**Tags:** Search

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Search history

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/search/history"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/search/history', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/search/history'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/search/history", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get failed searches (Admin)

`GET /api/v1/search/failed`

Returns searches that returned no results (admin only)

**Tags:** Search

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Failed searches

**401** - Success

**403** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/search/failed"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/search/failed', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/search/failed'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/search/failed", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get search analytics (Admin)

`GET /api/v1/search/analytics`

Returns search analytics summary (admin only)

**Tags:** Search

#### Responses

**200** - Search analytics

**401** - Success

**403** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/search/analytics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/search/analytics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/search/analytics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/search/analytics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Submissions

User clip submissions

### Get user submissions

`GET /api/v1/submissions`

Returns current user's clip submissions

**Tags:** Submissions

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
| status | query | string |  |  |

#### Responses

**200** - List of submissions

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/submissions"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/submissions', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/submissions'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/submissions", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Submit clip

`POST /api/v1/submissions`

Submit a Twitch clip for moderation (rate limited - 10/hour)

**Tags:** Submissions

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Clip submitted

**400** - Success

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/submissions" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/submissions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/submissions',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/submissions", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get submission statistics

`GET /api/v1/submissions/stats`

Returns submission statistics for current user

**Tags:** Submissions

#### Responses

**200** - Submission statistics

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/submissions/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/submissions/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/submissions/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/submissions/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get clip metadata

`GET /api/v1/submissions/metadata`

Fetches metadata for a Twitch clip URL (rate limited - 100/hour)

**Tags:** Submissions

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| clip_url | query | string | ✓ |  |

#### Responses

**200** - Clip metadata

**400** - Success

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/submissions/metadata"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/submissions/metadata', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/submissions/metadata'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/submissions/metadata", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Check clip status

`GET /api/v1/submissions/check/{clip_id}`

Checks if a clip can be claimed/submitted (rate limited - 100/hour)

**Tags:** Submissions

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| clip_id | path | string | ✓ |  |

#### Responses

**200** - Clip status

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/submissions/check/{clip_id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/submissions/check/{clip_id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/submissions/check/{clip_id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/submissions/check/{clip_id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Reports

Content reporting

### Submit report

`POST /api/v1/reports`

Submit a content report (rate limited to 10 per hour per user). Cookie-authenticated requests must also send the CSRF token.

**Tags:** Reports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Report submitted

**400** - Success

**401** - Success

**404** - Success

**409** - The user already reported this item

**429** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/reports" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/reports', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/reports',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/reports", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List content reports

`GET /api/v1/admin/reports`

Requires an administrator session with MFA.

**Tags:** Admin, Reports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| status | query | string |  |  |
| type | query | string |  |  |
| page | query | integer |  |  |
| limit | query | integer |  |  |

#### Responses

**200** - Paginated reports

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/reports"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/reports', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/reports'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/reports", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a content report

`GET /api/v1/admin/reports/{id}`

Requires an administrator session with MFA. Includes reporter information and reports for the same target.

**Tags:** Admin, Reports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Report details

**400** - Success

**401** - Success

**403** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/reports/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/reports/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/reports/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/reports/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Resolve a content report

`PUT /api/v1/admin/reports/{id}`

Requires an administrator session with MFA. Unsupported or incompatible moderation actions fail closed. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin, Reports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Report updated after any moderation action succeeded

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/reports/{id}" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/reports/{id}', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/reports/{id}',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/admin/reports/{id}", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Moderation

Moderation and appeals

### Get user appeals

`GET /api/v1/moderation/appeals`

Returns current user's appeals

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - List of appeals

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/appeals"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/appeals', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/moderation/appeals'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/appeals", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Create appeal

`POST /api/v1/moderation/appeals`

Submit an appeal for moderation action (rate limited - 5/hour)

**Tags:** Moderation

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Appeal created

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/moderation/appeals" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/appeals', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/moderation/appeals',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/moderation/appeals", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Sync bans from Twitch

`POST /api/v1/moderation/sync-bans`

Synchronizes ban status from Twitch for a specific channel. This endpoint triggers an asynchronous
background job that syncs ban data from Twitch. The operation has a 5-minute timeout.

**Required Permissions**: Channel owner, moderator, or admin

**Rate Limit**: 5 requests per hour

**Requirements**:
- Valid Twitch moderator OAuth scopes
- Moderator permissions on the target channel


**Tags:** Moderation

🔒 **Authentication Required**

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Ban sync started successfully

**400** - Bad request - Invalid or missing channel_id

**401** - Success

**429** - Success

**503** - Service unavailable - Twitch ban sync service is unavailable

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/moderation/sync-bans" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/sync-bans', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.post(
        '/api/v1/moderation/sync-bans',
        headers=headers,
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/moderation/sync-bans", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List channel bans

`GET /api/v1/moderation/bans`

Retrieves a paginated list of bans for a specific channel.

**Required Permissions**: Channel owner, moderator, or admin

**Rate Limit**: 60 requests per minute


**Tags:** Moderation

🔒 **Authentication Required**

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channelId | query | string | ✓ | Channel ID to list bans for |
| limit | query | integer |  | Number of results per page |
| offset | query | integer |  | Number of results to skip (must be multiple of limit) |

#### Responses

**200** - List of bans retrieved successfully

**400** - Bad request - Invalid parameters

**401** - Success

**403** - Forbidden - Insufficient permissions

**404** - Not found - Channel does not exist

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/bans" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/bans', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
    }
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.get(
        '/api/v1/moderation/bans',
        headers=headers
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/bans", nil)
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Create ban

`POST /api/v1/moderation/bans`

Creates a new ban for a user in a specific channel. Permanent bans are created by default.

**Required Permissions**: Channel owner, admin, or moderator

**Rate Limit**: 10 requests per hour

**Notes**:
- Cannot ban the channel owner
- Creates an audit log entry automatically
- Broadcasts ban event via WebSocket


**Tags:** Moderation

🔒 **Authentication Required**

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Ban created successfully

**400** - Bad request - Invalid parameters or cannot ban owner

**401** - Success

**403** - Forbidden - Insufficient permissions

**404** - Not found - User or channel not found

**409** - Conflict - User already banned

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/moderation/bans" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/bans', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.post(
        '/api/v1/moderation/bans',
        headers=headers,
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/moderation/bans", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get ban details

`GET /api/v1/moderation/ban/{id}`

Returns details about a specific ban (rate limited - 60/minute)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Ban details

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/ban/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/moderation/ban/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/ban/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Revoke ban

`DELETE /api/v1/moderation/ban/{id}`

Revokes a ban (rate limited - 10/hour)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - Ban revoked

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/moderation/ban/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/moderation/ban/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/moderation/ban/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Ban user on Twitch

`POST /api/v1/moderation/twitch/ban`

Bans a user on Twitch (requires Twitch moderator scopes, rate limited - 10/hour)

**Tags:** Moderation

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - User banned on Twitch

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/moderation/twitch/ban" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/twitch/ban', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/moderation/twitch/ban',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/moderation/twitch/ban", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Unban user on Twitch

`DELETE /api/v1/moderation/twitch/ban`

Unbans a user on Twitch (requires Twitch moderator scopes, rate limited - 10/hour)

**Tags:** Moderation

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - User unbanned on Twitch

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/moderation/twitch/ban" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/twitch/ban', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/moderation/twitch/ban',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("DELETE", "/api/v1/moderation/twitch/ban", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List moderators

`GET /api/v1/moderation/moderators`

Retrieves a paginated list of moderators for a specific channel.

**Required Permissions**: Authenticated users can view moderators

**Rate Limit**: 60 requests per minute


**Tags:** Moderation

🔒 **Authentication Required**

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channelId | query | string | ✓ | Channel ID to list moderators for |
| limit | query | integer |  | Number of results per page |
| offset | query | integer |  | Number of results to skip |

#### Responses

**200** - List of moderators retrieved successfully

**400** - Bad request - Missing or invalid channelId

**401** - Success

**403** - Forbidden - Insufficient permissions

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/moderators" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/moderators', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
    }
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.get(
        '/api/v1/moderation/moderators',
        headers=headers
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/moderators", nil)
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Add moderator

`POST /api/v1/moderation/moderators`

Adds a new moderator to a specific channel.

**Required Permissions**: Channel owner or admin only

**Rate Limit**: 10 requests per hour

**Notes**:
- If user is already a member, upgrades them to moderator
- If user is not a member, adds them as a new moderator
- Validates moderator scope for community-specific permissions
- Cannot add users who are already admins


**Tags:** Moderation

🔒 **Authentication Required**

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Moderator added successfully

**400** - Bad request - Invalid UUID or user already admin

**401** - Success

**403** - Forbidden - Not channel owner/admin

**404** - Not found - User or channel not found

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/moderation/moderators" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/moderators', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.post(
        '/api/v1/moderation/moderators',
        headers=headers,
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/moderation/moderators", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update moderator permissions

`PATCH /api/v1/moderation/moderators/{id}`

Updates moderator permissions (rate limited - 10/hour)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Permissions updated

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/moderation/moderators/{id}" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/moderators/{id}', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/moderation/moderators/{id}',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PATCH", "/api/v1/moderation/moderators/{id}", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Remove moderator

`DELETE /api/v1/moderation/moderators/{id}`

Removes a moderator (rate limited - 10/hour)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - Moderator removed

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/moderation/moderators/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/moderators/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/moderation/moderators/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/moderation/moderators/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List moderation audit logs

`GET /api/v1/moderation/audit-logs`

Returns moderation audit logs (moderator/admin only, rate limited - 60/minute)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Audit logs

**401** - Success

**403** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/audit-logs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/audit-logs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/moderation/audit-logs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/audit-logs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Export moderation audit logs

`GET /api/v1/moderation/audit-logs/export`

Exports audit logs to CSV (moderator/admin only, rate limited - 10/hour)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| start_date | query | string |  |  |
| end_date | query | string |  |  |

#### Responses

**200** - CSV file

**401** - Success

**403** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/audit-logs/export"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/audit-logs/export', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/moderation/audit-logs/export'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/audit-logs/export", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get audit log

`GET /api/v1/moderation/audit-logs/{id}`

Returns specific audit log entry (moderator/admin only, rate limited - 60/minute)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Audit log entry

**401** - Success

**403** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/audit-logs/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/audit-logs/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/moderation/audit-logs/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/audit-logs/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List ban reason templates

`GET /api/v1/moderation/ban-templates`

Returns ban reason templates (rate limited - 60/minute)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - List of templates

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/ban-templates"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban-templates', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/moderation/ban-templates'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/ban-templates", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Create ban template

`POST /api/v1/moderation/ban-templates`

Creates a new ban reason template (rate limited - 20/hour)

**Tags:** Moderation

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Template created

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/moderation/ban-templates" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban-templates', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/moderation/ban-templates',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/moderation/ban-templates", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get template usage statistics

`GET /api/v1/moderation/ban-templates/stats`

Returns usage statistics for ban templates (rate limited - 60/minute)

**Tags:** Moderation

#### Responses

**200** - Template statistics

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/ban-templates/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban-templates/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/moderation/ban-templates/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/ban-templates/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get ban template

`GET /api/v1/moderation/ban-templates/{id}`

Returns specific ban template (rate limited - 60/minute)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Template details

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/moderation/ban-templates/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban-templates/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/moderation/ban-templates/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/moderation/ban-templates/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update ban template

`PATCH /api/v1/moderation/ban-templates/{id}`

Updates a ban template (rate limited - 20/hour)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Template updated

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/moderation/ban-templates/{id}" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban-templates/{id}', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/moderation/ban-templates/{id}',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PATCH", "/api/v1/moderation/ban-templates/{id}", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Delete ban template

`DELETE /api/v1/moderation/ban-templates/{id}`

Deletes a ban template (rate limited - 20/hour)

**Tags:** Moderation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - Template deleted

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/moderation/ban-templates/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban-templates/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/moderation/ban-templates/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/moderation/ban-templates/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Users

User profiles and social features

### Get user by username

`GET /api/v1/users/by-username/{username}`

Returns user profile by username

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| username | path | string | ✓ |  |

#### Responses

**200** - User profile

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/by-username/{username}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/by-username/{username}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/by-username/{username}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/by-username/{username}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### User autocomplete

`GET /api/v1/users/autocomplete`

Search users for mentions/suggestions (rate limited - 100/hour)

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| q | query | string | ✓ |  |
|  |  | string |  |  |

#### Responses

**200** - User suggestions

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/autocomplete"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/autocomplete', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/autocomplete'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/autocomplete", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user profile

`GET /api/v1/users/{id}`

Returns detailed user profile (optional auth for follow status)

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - User profile

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Claim account

`POST /api/v1/users/claim-account`

Claims an unclaimed broadcaster/creator account

**Tags:** Users

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Account claimed

**400** - Success

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/claim-account" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/claim-account', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/claim-account',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/users/claim-account", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user reputation

`GET /api/v1/users/{id}/reputation`

Returns reputation score and breakdown

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Reputation data

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/reputation"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/reputation', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/reputation'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/reputation", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user karma

`GET /api/v1/users/{id}/karma`

Returns karma points and history

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Karma data

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/karma"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/karma', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/karma'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/karma", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user badges

`GET /api/v1/users/{id}/badges`

Returns earned badges for user

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - User badges

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/badges"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/badges', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/badges'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/badges", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user comments

`GET /api/v1/users/{id}/comments`

Returns user's comment history

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - List of comments

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/comments"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/comments', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/comments'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/comments", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user clips

`GET /api/v1/users/{id}/clips`

Returns clips submitted by user (optional auth for hidden clips)

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - List of clips

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user activity

`GET /api/v1/users/{id}/activity`

Returns recent user activity (optional auth)

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Activity feed

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/activity"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/activity', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/activity'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/activity", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user upvoted clips

`GET /api/v1/users/{id}/upvoted`

Returns clips user has upvoted

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Upvoted clips

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/upvoted"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/upvoted', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/upvoted'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/upvoted", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user downvoted clips

`GET /api/v1/users/{id}/downvoted`

Returns clips user has downvoted

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Downvoted clips

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/downvoted"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/downvoted', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/downvoted'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/downvoted", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user followers

`GET /api/v1/users/{id}/followers`

Returns list of user's followers (optional auth)

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Followers list

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/followers"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/followers', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/followers'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/followers", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get following

`GET /api/v1/users/{id}/following`

Returns users being followed (optional auth)

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Following list

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/following"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/following', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/following'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/following", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get followed broadcasters

`GET /api/v1/users/{id}/following/broadcasters`

Returns broadcasters being followed (optional auth)

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Followed broadcasters

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/following/broadcasters"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/following/broadcasters', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/following/broadcasters'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/following/broadcasters", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Follow user

`POST /api/v1/users/{id}/follow`

Follows a user

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**201** - User followed

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/{id}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/follow', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/{id}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/users/{id}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Unfollow user

`DELETE /api/v1/users/{id}/follow`

Unfollows a user

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - User unfollowed

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/users/{id}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/follow', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/users/{id}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/users/{id}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Block user

`POST /api/v1/users/{id}/block`

Blocks a user (rate limited - 20/minute)

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**201** - User blocked

**401** - Success

**404** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/{id}/block"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/block', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/{id}/block'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/users/{id}/block", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Unblock user

`DELETE /api/v1/users/{id}/block`

Unblocks a user

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**204** - User unblocked

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/users/{id}/block"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/block', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/users/{id}/block'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/users/{id}/block", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get blocked users

`GET /api/v1/users/me/blocked`

Returns list of blocked users

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Blocked users

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/me/blocked"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/blocked', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/me/blocked'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/me/blocked", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get personal statistics

`GET /api/v1/users/me/stats`

Returns statistics for current user

**Tags:** Users

#### Responses

**200** - User statistics

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/me/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/me/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/me/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user engagement score

`GET /api/v1/users/{id}/engagement`

Returns engagement metrics for user

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Engagement score

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/engagement"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/engagement', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/engagement'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/engagement", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update profile

`PUT /api/v1/users/me/profile`

Updates current user's profile

**Tags:** Users

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Profile updated

**400** - Success

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/users/me/profile" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/profile', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/users/me/profile',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/users/me/profile", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update social links

`PUT /api/v1/users/me/social-links`

Updates social media links

**Tags:** Users

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Social links updated

**400** - Success

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/users/me/social-links" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/social-links', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/users/me/social-links',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/users/me/social-links", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get user settings

`GET /api/v1/users/me/settings`

Returns current user's settings

**Tags:** Users

#### Responses

**200** - User settings

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/me/settings"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/settings', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/me/settings'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/me/settings", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update settings

`PUT /api/v1/users/me/settings`

Updates user settings

**Tags:** Users

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Settings updated

**400** - Success

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/users/me/settings" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/settings', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/users/me/settings',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/users/me/settings", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get cookie consent

`GET /api/v1/users/me/consent`

Returns current user's cookie consent preferences

**Tags:** Users

#### Responses

**200** - Consent preferences

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/me/consent"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/consent', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/me/consent'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/me/consent", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Save cookie consent

`POST /api/v1/users/me/consent`

Saves cookie consent preferences (rate limited - 30/minute)

**Tags:** Users

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Consent saved

**400** - Success

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/me/consent" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/consent', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/me/consent',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/users/me/consent", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get email logs

`GET /api/v1/users/me/email-logs`

Returns email logs for current user

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
|  |  | string |  |  |

#### Responses

**200** - Email logs

**401** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/me/email-logs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/email-logs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/me/email-logs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/me/email-logs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Convert the current account to broadcaster

`POST /api/v1/users/me/convert-to-broadcaster`

Converts the authenticated member account to broadcaster. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Account converted to broadcaster

**400** - Success

**401** - Success

**429** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/me/convert-to-broadcaster" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/convert-to-broadcaster', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/me/convert-to-broadcaster',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/users/me/convert-to-broadcaster", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a user's account type

`GET /api/v1/users/{id}/account-type`

Returns the account type, permissions, and any included conversion history for the requested user.

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Account-type information

**400** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/account-type"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/account-type', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/account-type'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/account-type", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List a user's account-type conversion history

`GET /api/v1/users/{id}/account-type/history`

Returns account-type conversions for the requested user, ordered from newest to oldest.

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| limit | query | integer |  |  |
| offset | query | integer |  |  |

#### Responses

**200** - Conversion history

**400** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/account-type/history"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/account-type/history', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/account-type/history'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/account-type/history", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get the current user's watch history

`GET /api/v1/watch-history`

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| filter | query | string |  |  |
| limit | query | integer |  |  |

#### Responses

**200** - Watch-history entries

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/watch-history"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/watch-history', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/watch-history'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/watch-history", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Record clip watch progress

`POST /api/v1/watch-history`

Records a bounded progress position when watch-history tracking is enabled. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Progress recorded or tracking disabled

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/watch-history" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/watch-history', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/watch-history',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/watch-history", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Clear the current user's watch history

`DELETE /api/v1/watch-history`

Permanently deletes all watch-history entries for the authenticated user. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Users

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Watch history cleared

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/watch-history"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/watch-history', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/watch-history'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/watch-history", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Admin

Administrative operations

### List recent account-type conversions

`GET /api/v1/admin/account-types/conversions`

Requires an administrator session with MFA. Results are ordered from newest to oldest.

**Tags:** Admin

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| limit | query | integer |  |  |
| offset | query | integer |  |  |

#### Responses

**200** - Recent conversions

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/account-types/conversions"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/account-types/conversions', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/account-types/conversions'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/account-types/conversions", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get account-type totals

`GET /api/v1/admin/account-types/stats`

Requires an administrator session with MFA.

**Tags:** Admin

#### Responses

**200** - Counts for every supported account type

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/account-types/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/account-types/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/account-types/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/account-types/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Convert a user to a moderator account

`POST /api/v1/admin/account-types/users/{id}/convert-to-moderator`

Requires an administrator session with MFA. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - User converted to moderator

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/account-types/users/{id}/convert-to-moderator" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/account-types/users/{id}/convert-to-moderator', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/account-types/users/{id}/convert-to-moderator',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/admin/account-types/users/{id}/convert-to-moderator", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List contact messages

`GET /api/v1/admin/contact`

Requires an administrator session with MFA. Responses contain private submitter and abuse-prevention data.

**Tags:** Admin, Support

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| page | query | integer |  |  |
| limit | query | integer |  |  |
| category | query | string |  |  |
| status | query | string |  |  |

#### Responses

**200** - Paginated contact messages

**400** - Invalid filter or pagination value

**401** - Success

**403** - Success

**500** - Contact messages could not be retrieved

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/contact"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/contact', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/contact'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/contact", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update a contact message status

`PUT /api/v1/admin/contact/{id}/status`

Requires an administrator session with MFA. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin, Support

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Status updated

**400** - Invalid message ID or status

**401** - Success

**403** - Success

**404** - Contact message not found

**500** - Status could not be updated

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/contact/{id}/status" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/contact/{id}/status', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/contact/{id}/status',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/admin/contact/{id}/status", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List content reports

`GET /api/v1/admin/reports`

Requires an administrator session with MFA.

**Tags:** Admin, Reports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| status | query | string |  |  |
| type | query | string |  |  |
| page | query | integer |  |  |
| limit | query | integer |  |  |

#### Responses

**200** - Paginated reports

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/reports"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/reports', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/reports'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/reports", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a content report

`GET /api/v1/admin/reports/{id}`

Requires an administrator session with MFA. Includes reporter information and reports for the same target.

**Tags:** Admin, Reports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Report details

**400** - Success

**401** - Success

**403** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/reports/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/reports/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/reports/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/reports/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Resolve a content report

`PUT /api/v1/admin/reports/{id}`

Requires an administrator session with MFA. Unsupported or incompatible moderation actions fail closed. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin, Reports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Report updated after any moderation action succeeded

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/reports/{id}" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/reports/{id}', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/reports/{id}',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/admin/reports/{id}", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get aggregate feed event metrics

`GET /api/v1/feeds/analytics`

Requires an administrator session.

**Tags:** Admin, Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| hours | query | integer |  |  |

#### Responses

**200** - Aggregate event metrics

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/feeds/analytics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/feeds/analytics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/feeds/analytics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/feeds/analytics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get hourly metrics for an event type

`GET /api/v1/feeds/analytics/hourly`

Requires an administrator session.

**Tags:** Admin, Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| event_type | query | string | ✓ |  |
| hours | query | integer |  |  |

#### Responses

**200** - Hourly event metrics

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/feeds/analytics/hourly"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/feeds/analytics/hourly', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/feeds/analytics/hourly'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/feeds/analytics/hourly", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List outbound webhook dead-letter items

`GET /api/v1/admin/webhooks/dlq`

Requires an administrator session with MFA.

**Tags:** Admin, Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| page | query | integer |  |  |
| limit | query | integer |  |  |

#### Responses

**200** - Paginated dead-letter items

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/webhooks/dlq"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/webhooks/dlq', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/webhooks/dlq'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/webhooks/dlq", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Permanently delete a dead-letter item

`DELETE /api/v1/admin/webhooks/dlq/{id}`

Requires an administrator session with MFA. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin, Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Responses

**200** - Item deleted

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/webhooks/dlq/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/webhooks/dlq/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/webhooks/dlq/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/webhooks/dlq/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Replay a dead-lettered webhook

`POST /api/v1/admin/webhooks/dlq/{id}/replay`

Atomically claims and redelivers an item once. Concurrent or already-successful replay attempts fail with 409. Requires administrator MFA; cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin, Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Responses

**200** - Replay delivered successfully

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**409** - Item is currently replaying or was already replayed successfully

**502** - Downstream replay failed

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/webhooks/dlq/{id}/replay"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/webhooks/dlq/{id}/replay', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/webhooks/dlq/{id}/replay'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/webhooks/dlq/{id}/replay", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get subscription revenue metrics

`GET /api/v1/admin/revenue`

Requires an administrator session with MFA. Monetary values are integer-compatible cents; recognized revenue comes from paid invoice amounts while MRR uses monthly recurring equivalents. The endpoint fails closed when any metric query is incomplete.

**Tags:** Admin, Payments

#### Responses

**200** - Complete revenue dashboard metrics

**401** - Success

**403** - Success

**500** - One or more revenue metrics could not be calculated

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/revenue"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/revenue', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/revenue'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/revenue", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Run a clip synchronization

`POST /api/v1/admin/sync/clips`

Runs a bounded synchronous Twitch clip import. Requires administrator MFA; cookie-authenticated requests must also send the CSRF token. Item-level failures produce 207 rather than full success.

**Tags:** Admin, Clips

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Sync completed without item errors

**207** - Sync completed with one or more item errors

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/sync/clips" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/sync/clips', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/sync/clips',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/admin/sync/clips", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get persisted clip synchronization status

`GET /api/v1/admin/sync/status`

Returns the most recent persisted clip import time rather than a synthetic service-ready placeholder. Requires administrator MFA.

**Tags:** Admin, Clips

#### Responses

**200** - Last persisted synchronization evidence

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/sync/status"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/sync/status', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/sync/status'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/sync/status", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List immutable moderation audit events

`GET /api/v1/admin/audit-logs`

Requires an administrator session with MFA. Results may include IP address, user-agent, and event metadata.

**Tags:** Admin, Audit

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| page | query | integer |  |  |
| limit | query | integer |  |  |
| moderator_id | query | string |  |  |
| action | query | string |  |  |
| entity_type | query | string |  |  |
| entity_id | query | string |  |  |
| channel_id | query | string |  |  |
| start_date | query | string |  |  |
| end_date | query | string |  |  |
| search | query | string |  |  |

#### Responses

**200** - Paginated audit events

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/audit-logs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/audit-logs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/audit-logs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/audit-logs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Export filtered audit events as CSV

`GET /api/v1/admin/audit-logs/export`

Requires administrator MFA. Uses the same filters as the list route and refuses results over 10,000 rows; narrow the filters and retry.

**Tags:** Admin, Audit

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| moderator_id | query | string |  |  |
| action | query | string |  |  |
| entity_type | query | string |  |  |
| entity_id | query | string |  |  |
| channel_id | query | string |  |  |
| start_date | query | string |  |  |
| end_date | query | string |  |  |
| search | query | string |  |  |

#### Responses

**200** - Complete CSV export

**400** - Success

**401** - Success

**403** - Success

**413** - Export exceeds 10,000 rows

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/audit-logs/export"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/audit-logs/export', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/audit-logs/export'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/audit-logs/export", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Webhooks

Webhook management

### List webhook subscriptions

`GET /api/v1/webhooks`

Lists outbound webhook subscriptions owned by the authenticated user.

**Tags:** Webhooks

#### Responses

**200** - Webhook subscriptions

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/webhooks"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/webhooks'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/webhooks", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Create a webhook subscription

`POST /api/v1/webhooks`

Creates an outbound webhook subscription. The signing secret is returned only by this response. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Subscription created and signing secret issued

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/webhooks" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/webhooks',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List supported webhook events

`GET /api/v1/webhooks/events`

**Tags:** Webhooks

#### Responses

**200** - Supported event identifiers

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/webhooks/events"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks/events', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/webhooks/events'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/webhooks/events", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a webhook subscription

`GET /api/v1/webhooks/{id}`

**Tags:** Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Webhook subscription

**400** - Success

**401** - Success

**404** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/webhooks/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/webhooks/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/webhooks/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update a webhook subscription

`PATCH /api/v1/webhooks/{id}`

Updates mutable subscription settings. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Subscription updated

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/webhooks/{id}" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks/{id}', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/webhooks/{id}',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PATCH", "/api/v1/webhooks/{id}", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Delete a webhook subscription

`DELETE /api/v1/webhooks/{id}`

Permanently deletes the subscription. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Responses

**200** - Subscription deleted

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/webhooks/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/webhooks/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/webhooks/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List webhook delivery attempts

`GET /api/v1/webhooks/{id}/deliveries`

**Tags:** Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| page | query | integer |  |  |
| limit | query | integer |  |  |

#### Responses

**200** - Paginated delivery attempts

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/webhooks/{id}/deliveries"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks/{id}/deliveries', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/webhooks/{id}/deliveries'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/webhooks/{id}/deliveries", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Regenerate a webhook signing secret

`POST /api/v1/webhooks/{id}/regenerate-secret`

Invalidates the previous secret and returns its replacement once. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Responses

**200** - Replacement signing secret

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/webhooks/{id}/regenerate-secret"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks/{id}/regenerate-secret', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/webhooks/{id}/regenerate-secret'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/webhooks/{id}/regenerate-secret", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List outbound webhook dead-letter items

`GET /api/v1/admin/webhooks/dlq`

Requires an administrator session with MFA.

**Tags:** Admin, Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| page | query | integer |  |  |
| limit | query | integer |  |  |

#### Responses

**200** - Paginated dead-letter items

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/webhooks/dlq"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/webhooks/dlq', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/webhooks/dlq'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/webhooks/dlq", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Permanently delete a dead-letter item

`DELETE /api/v1/admin/webhooks/dlq/{id}`

Requires an administrator session with MFA. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin, Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Responses

**200** - Item deleted

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/webhooks/dlq/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/webhooks/dlq/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/webhooks/dlq/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/webhooks/dlq/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Replay a dead-lettered webhook

`POST /api/v1/admin/webhooks/dlq/{id}/replay`

Atomically claims and redelivers an item once. Concurrent or already-successful replay attempts fail with 409. Requires administrator MFA; cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin, Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Responses

**200** - Replay delivered successfully

**400** - Success

**401** - Success

**403** - Success

**404** - Success

**409** - Item is currently replaying or was already replayed successfully

**502** - Downstream replay failed

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/webhooks/dlq/{id}/replay"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/webhooks/dlq/{id}/replay', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/webhooks/dlq/{id}/replay'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/webhooks/dlq/{id}/replay", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get webhook delivery and retry health

`GET /internal/operations/webhooks`

Requires the dedicated operational bearer token. Returns 503 rather than reporting healthy when delivery statistics are incomplete.

**Tags:** Operations, Webhooks

🔒 **Authentication Required**

#### Responses

**200** - Complete webhook operational statistics

**401** - Success

**500** - Retry queue statistics unavailable

**503** - Delivery statistics unavailable; partial retry statistics returned

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/internal/operations/webhooks" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/internal/operations/webhooks', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
    }
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.get(
        '/internal/operations/webhooks',
        headers=headers
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/internal/operations/webhooks", nil)
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Ingest signed SendGrid event notifications

`POST /api/v1/webhooks/sendgrid`

Verifies SendGrid's ECDSA signature and five-minute timestamp window before processing 1–1000 events. Payloads are limited to 1 MiB. A 500 response requests provider retry when any event could not be persisted.

**Tags:** Webhooks

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| X-Twilio-Email-Event-Webhook-Signature | header | string | ✓ |  |
| X-Twilio-Email-Event-Webhook-Timestamp | header | string | ✓ |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Every event processed

**400** - Invalid JSON, batch bounds, or required event fields

**401** - Missing, stale, future, or invalid SendGrid signature

**413** - Payload exceeds 1 MiB

**500** - One or more verified events could not be persisted

**503** - Signature verification key is unavailable

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/webhooks/sendgrid" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks/sendgrid', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/webhooks/sendgrid',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/webhooks/sendgrid", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Support

### Submit a contact message

`POST /api/v1/contact`

Submits a rate-limited support request. Authentication is optional; authenticated requests associate the message with the current user.

**Tags:** Support

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Contact message accepted

**400** - Invalid contact message

**429** - Success

**500** - Contact message could not be stored

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/contact" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/contact', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/contact',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/contact", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List contact messages

`GET /api/v1/admin/contact`

Requires an administrator session with MFA. Responses contain private submitter and abuse-prevention data.

**Tags:** Admin, Support

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| page | query | integer |  |  |
| limit | query | integer |  |  |
| category | query | string |  |  |
| status | query | string |  |  |

#### Responses

**200** - Paginated contact messages

**400** - Invalid filter or pagination value

**401** - Success

**403** - Success

**500** - Contact messages could not be retrieved

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/contact"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/contact', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/contact'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/contact", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Update a contact message status

`PUT /api/v1/admin/contact/{id}/status`

Requires an administrator session with MFA. Cookie-authenticated requests must also send the CSRF token.

**Tags:** Admin, Support

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**200** - Status updated

**400** - Invalid message ID or status

**401** - Success

**403** - Success

**404** - Contact message not found

**500** - Status could not be updated

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/contact/{id}/status" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/contact/{id}/status', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/contact/{id}/status',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("PUT", "/api/v1/admin/contact/{id}/status", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Analytics

### Queue one or more analytics events

`POST /api/v1/events`

Accepts a single event or a batch of up to 100 events. A response confirms queue acceptance, not database persistence. Requests are limited to 64 KiB and 100 requests per minute.

**Tags:** Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| X-Session-ID | header | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**202** - Event or complete batch accepted into the queue

**207** - Only part of the batch was accepted

**400** - Invalid event, batch, or session identifier

**413** - Request exceeds 64 KiB

**429** - Success

**500** - Event could not be queued

**503** - Event queue is saturated; retry later

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/events" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/events', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/events',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get aggregate feed event metrics

`GET /api/v1/feeds/analytics`

Requires an administrator session.

**Tags:** Admin, Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| hours | query | integer |  |  |

#### Responses

**200** - Aggregate event metrics

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/feeds/analytics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/feeds/analytics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/feeds/analytics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/feeds/analytics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get hourly metrics for an event type

`GET /api/v1/feeds/analytics/hourly`

Requires an administrator session.

**Tags:** Admin, Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| event_type | query | string | ✓ |  |
| hours | query | integer |  |  |

#### Responses

**200** - Hourly event metrics

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/feeds/analytics/hourly"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/feeds/analytics/hourly', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/feeds/analytics/hourly'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/feeds/analytics/hourly", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Operations

### Get webhook delivery and retry health

`GET /internal/operations/webhooks`

Requires the dedicated operational bearer token. Returns 503 rather than reporting healthy when delivery statistics are incomplete.

**Tags:** Operations, Webhooks

🔒 **Authentication Required**

#### Responses

**200** - Complete webhook operational statistics

**401** - Success

**500** - Retry queue statistics unavailable

**503** - Delivery statistics unavailable; partial retry statistics returned

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/internal/operations/webhooks" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/internal/operations/webhooks', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
    }
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.get(
        '/internal/operations/webhooks',
        headers=headers
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/internal/operations/webhooks", nil)
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get typed Redis operational statistics

`GET /internal/operations/cache`

Requires the dedicated operational bearer token. Missing or malformed required Redis counters return degraded 503 instead of a healthy response.

**Tags:** Operations

🔒 **Authentication Required**

#### Responses

**200** - Complete typed cache statistics

**401** - Success

**503** - Redis unavailable or statistics incomplete

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/internal/operations/cache" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/internal/operations/cache', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
    }
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.get(
        '/internal/operations/cache',
        headers=headers
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/internal/operations/cache", nil)
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Check Redis connectivity

`GET /internal/operations/cache/check`

Requires the dedicated operational bearer token.

**Tags:** Operations

🔒 **Authentication Required**

#### Responses

**200** - Redis is accessible

**401** - Success

**503** - Redis is unavailable

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/internal/operations/cache/check" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/internal/operations/cache/check', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
    }
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
try:
    response = requests.get(
        '/internal/operations/cache/check',
        headers=headers
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/internal/operations/cache/check", nil)
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Payments

### Get subscription revenue metrics

`GET /api/v1/admin/revenue`

Requires an administrator session with MFA. Monetary values are integer-compatible cents; recognized revenue comes from paid invoice amounts while MRR uses monthly recurring equivalents. The endpoint fails closed when any metric query is incomplete.

**Tags:** Admin, Payments

#### Responses

**200** - Complete revenue dashboard metrics

**401** - Success

**403** - Success

**500** - One or more revenue metrics could not be calculated

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/revenue"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/revenue', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/revenue'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/revenue", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Broadcasters

Broadcaster profiles and content

### List currently live broadcasters

`GET /api/v1/broadcasters/live`

Returns only statuses checked within the last two minutes, ordered by viewer count.

**Tags:** Broadcasters

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| page | query | integer |  |  |
| limit | query | integer |  |  |

#### Responses

**200** - Fresh live broadcaster statuses

**400** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/broadcasters/live"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/broadcasters/live', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/broadcasters/live'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/broadcasters/live", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a broadcaster's live status

`GET /api/v1/broadcasters/{id}/live-status`

A stored live record older than two minutes is returned as offline with is_stale set to true.

**Tags:** Broadcasters

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Live status, or a default offline status when none has been observed

**400** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/broadcasters/{id}/live-status"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/broadcasters/{id}/live-status', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/broadcasters/{id}/live-status'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/broadcasters/{id}/live-status", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Audit

### List immutable moderation audit events

`GET /api/v1/admin/audit-logs`

Requires an administrator session with MFA. Results may include IP address, user-agent, and event metadata.

**Tags:** Admin, Audit

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| page | query | integer |  |  |
| limit | query | integer |  |  |
| moderator_id | query | string |  |  |
| action | query | string |  |  |
| entity_type | query | string |  |  |
| entity_id | query | string |  |  |
| channel_id | query | string |  |  |
| start_date | query | string |  |  |
| end_date | query | string |  |  |
| search | query | string |  |  |

#### Responses

**200** - Paginated audit events

**400** - Success

**401** - Success

**403** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/audit-logs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/audit-logs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/audit-logs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/audit-logs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Export filtered audit events as CSV

`GET /api/v1/admin/audit-logs/export`

Requires administrator MFA. Uses the same filters as the list route and refuses results over 10,000 rows; narrow the filters and retry.

**Tags:** Admin, Audit

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| moderator_id | query | string |  |  |
| action | query | string |  |  |
| entity_type | query | string |  |  |
| entity_id | query | string |  |  |
| channel_id | query | string |  |  |
| start_date | query | string |  |  |
| end_date | query | string |  |  |
| search | query | string |  |  |

#### Responses

**200** - Complete CSV export

**400** - Success

**401** - Success

**403** - Success

**413** - Export exceeds 10,000 rows

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/audit-logs/export"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/audit-logs/export', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/audit-logs/export'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/audit-logs/export", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Pages

### Render a game clip collection page

`GET /clips/game/{gameSlug}`

Server-rendered SEO page. Missing games return 404; incomplete repository data returns a non-cacheable 500 rather than an empty indexable page.

**Tags:** Pages

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| gameSlug | path | string | ✓ |  |

#### Responses

**200** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/clips/game/{gameSlug}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/clips/game/{gameSlug}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/clips/game/{gameSlug}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/clips/game/{gameSlug}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Render a broadcaster clip profile page

`GET /clips/streamer/{broadcasterName}`

Server-rendered SEO page. Missing broadcasters return 404; incomplete repository data returns a non-cacheable 500.

**Tags:** Pages

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| broadcasterName | path | string | ✓ |  |

#### Responses

**200** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/clips/streamer/{broadcasterName}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/clips/streamer/{broadcasterName}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/clips/streamer/{broadcasterName}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/clips/streamer/{broadcasterName}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Render a broadcaster-by-game clip page

`GET /clips/streamer/{broadcasterName}/{gameSlug}`

Server-rendered SEO page. Missing entities return 404; incomplete repository data returns a non-cacheable 500.

**Tags:** Pages

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| broadcasterName | path | string | ✓ |  |
| gameSlug | path | string | ✓ |  |

#### Responses

**200** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/clips/streamer/{broadcasterName}/{gameSlug}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/clips/streamer/{broadcasterName}/{gameSlug}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/clips/streamer/{broadcasterName}/{gameSlug}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/clips/streamer/{broadcasterName}/{gameSlug}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Documentation

API documentation

### List available Markdown documentation

`GET /api/v1/docs`

Returns a tree of regular Markdown files, excluding hidden, archive, vault, and symlink entries.

**Tags:** Documentation

#### Responses

**200** - Documentation tree

**406** - Client explicitly does not accept application/json

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/docs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/docs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/docs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/docs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Search documentation text

`GET /api/v1/docs/search`

Searches bounded regular Markdown files and returns at most 100 results ordered by relevance. Rate limited to 60 requests per minute.

**Tags:** Documentation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| q | query | string | ✓ |  |

#### Responses

**200** - Ranked search results

**400** - Success

**429** - Success

**503** - Documentation source unavailable

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/docs/search"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/docs/search', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/docs/search'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/docs/search", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a top-level Markdown document

`GET /api/v1/docs/{path}`

**Tags:** Documentation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| path | path | string | ✓ |  |

#### Responses

**200** - Document content and optional edit URL

**400** - Success

**403** - Success

**404** - Success

**413** - Document exceeds 1 MiB

**500** - Success

**503** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/docs/{path}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/docs/{path}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/docs/{path}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/docs/{path}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a nested Markdown document

`GET /api/v1/docs/content/{path}`

Resolves the real path and rejects symlinks that escape the configured documentation root.

**Tags:** Documentation

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| path | path | string | ✓ |  |

#### Responses

**200** - Nested document content and optional edit URL

**400** - Success

**403** - Success

**404** - Success

**413** - Document exceeds 1 MiB

**500** - Success

**503** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/docs/content/{path}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/docs/content/{path}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/docs/content/{path}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/docs/content/{path}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Creator Exports

### Request a creator data export

`POST /api/v1/creators/me/export/request`

Queues an authenticated creator's clip export. Limited to three requests per day.

**Tags:** Creator Exports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Export request queued

**400** - Success

**401** - Success

**429** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/creators/me/export/request" \
  -H "Content-Type: application/json" \
  -d '{"example": "data"}'
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/me/export/request', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      // Your request data
    })
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/creators/me/export/request',
        json={}  # Your request data
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
    "bytes"
    "encoding/json"
)

func main() {
    // Create request body
    data := map[string]interface{}{
        // Your request data
    }
    jsonBody, err := json.Marshal(data)
    if err != nil {
        // Handle error
        return
    }

    req, err := http.NewRequest("POST", "/api/v1/creators/me/export/request", bytes.NewBuffer(jsonBody))
    if err != nil {
        // Handle error
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a creator export's status

`GET /api/v1/creators/me/export/status/{id}`

Returns only an export owned by the authenticated user. Completed, unexpired exports include a download URL.

**Tags:** Creator Exports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Export status

**400** - Success

**401** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/creators/me/export/status/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/me/export/status/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/creators/me/export/status/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/creators/me/export/status/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Download a completed creator export

`GET /api/v1/creators/me/export/download/{id}`

Downloads an owned, completed, unexpired export artifact.

**Tags:** Creator Exports

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Export file

**400** - Invalid ID or export is not ready

**401** - Success

**404** - Success

**410** - Export artifact has expired

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/creators/me/export/download/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/me/export/download/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/creators/me/export/download/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/creators/me/export/download/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List the current creator's export requests

`GET /api/v1/creators/me/exports`

**Tags:** Creator Exports

#### Responses

**200** - Export request history

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/creators/me/exports"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/me/exports', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/creators/me/exports'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/creators/me/exports", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Creator Analytics

### Get public creator analytics totals

`GET /api/v1/creators/{creatorName}/analytics/overview`

**Tags:** Creator Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |

#### Responses

**200** - Creator analytics totals

**400** - Success

**404** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/creators/{creatorName}/analytics/overview"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/{creatorName}/analytics/overview', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/creators/{creatorName}/analytics/overview'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/creators/{creatorName}/analytics/overview", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### List a creator's top-performing clips

`GET /api/v1/creators/{creatorName}/analytics/clips`

**Tags:** Creator Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
| sort | query | string |  |  |
| limit | query | integer |  |  |

#### Responses

**200** - Ranked clips

**400** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/creators/{creatorName}/analytics/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/{creatorName}/analytics/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/creators/{creatorName}/analytics/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/creators/{creatorName}/analytics/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get a creator metric time series

`GET /api/v1/creators/{creatorName}/analytics/trends`

**Tags:** Creator Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
| metric | query | string |  |  |
| days | query | integer |  |  |

#### Responses

**200** - Daily metric values

**400** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/creators/{creatorName}/analytics/trends"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/{creatorName}/analytics/trends', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/creators/{creatorName}/analytics/trends'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/creators/{creatorName}/analytics/trends", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### Get aggregate creator audience device insights

`GET /api/v1/creators/{creatorName}/analytics/audience`

Returns device categories for views in the last 90 days. Geographic results remain empty until a production GeoIP provider and accuracy policy are configured.

**Tags:** Creator Analytics

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
|  |  | string |  |  |
| limit | query | integer |  | Reserved maximum number of geographic rows once geography is supported. |

#### Responses

**200** - Aggregate audience insights

**400** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/creators/{creatorName}/analytics/audience"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/{creatorName}/analytics/audience', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/creators/{creatorName}/analytics/audience'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/creators/{creatorName}/analytics/audience", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

## Generated Route Contracts

### GET /api/v1/admin/ads/campaigns

`GET /api/v1/admin/ads/campaigns`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/ads/campaigns"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/campaigns', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/ads/campaigns'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/ads/campaigns", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/ads/campaigns

`POST /api/v1/admin/ads/campaigns`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/ads/campaigns"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/campaigns', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/ads/campaigns'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/ads/campaigns", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/ads/campaigns/{id}

`GET /api/v1/admin/ads/campaigns/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/ads/campaigns/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/campaigns/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/ads/campaigns/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/ads/campaigns/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/admin/ads/campaigns/{id}

`PUT /api/v1/admin/ads/campaigns/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/ads/campaigns/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/campaigns/{id}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/ads/campaigns/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/admin/ads/campaigns/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/admin/ads/campaigns/{id}

`DELETE /api/v1/admin/ads/campaigns/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/ads/campaigns/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/campaigns/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/ads/campaigns/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/ads/campaigns/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/ads/experiments

`GET /api/v1/admin/ads/experiments`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/ads/experiments"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/experiments', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/ads/experiments'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/ads/experiments", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/ads/experiments/{id}/report

`GET /api/v1/admin/ads/experiments/{id}/report`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/ads/experiments/{id}/report"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/experiments/{id}/report', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/ads/experiments/{id}/report'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/ads/experiments/{id}/report", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/ads/reports/by-campaign

`GET /api/v1/admin/ads/reports/by-campaign`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/ads/reports/by-campaign"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/reports/by-campaign', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/ads/reports/by-campaign'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/ads/reports/by-campaign", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/ads/reports/by-date

`GET /api/v1/admin/ads/reports/by-date`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/ads/reports/by-date"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/reports/by-date', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/ads/reports/by-date'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/ads/reports/by-date", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/ads/reports/by-placement

`GET /api/v1/admin/ads/reports/by-placement`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/ads/reports/by-placement"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/reports/by-placement', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/ads/reports/by-placement'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/ads/reports/by-placement", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/ads/reports/by-slot

`GET /api/v1/admin/ads/reports/by-slot`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/ads/reports/by-slot"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/reports/by-slot', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/ads/reports/by-slot'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/ads/reports/by-slot", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/ads/validate-creative

`POST /api/v1/admin/ads/validate-creative`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/ads/validate-creative"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/ads/validate-creative', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/ads/validate-creative'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/ads/validate-creative", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/analytics/alerts

`GET /api/v1/admin/analytics/alerts`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/alerts"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/analytics/alerts', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/analytics/alerts'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/analytics/alerts", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/analytics/content

`GET /api/v1/admin/analytics/content`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/content"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/analytics/content', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/analytics/content'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/analytics/content", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/analytics/export

`GET /api/v1/admin/analytics/export`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/export"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/analytics/export', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/analytics/export'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/analytics/export", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/analytics/health

`GET /api/v1/admin/analytics/health`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/health"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/analytics/health', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/analytics/health'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/analytics/health", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/analytics/overview

`GET /api/v1/admin/analytics/overview`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/overview"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/analytics/overview', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/analytics/overview'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/analytics/overview", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/analytics/trending

`GET /api/v1/admin/analytics/trending`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/trending"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/analytics/trending', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/analytics/trending'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/analytics/trending", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/analytics/trends

`GET /api/v1/admin/analytics/trends`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/analytics/trends"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/analytics/trends', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/analytics/trends'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/analytics/trends", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/broadcasters/refresh-rankings

`POST /api/v1/admin/broadcasters/refresh-rankings`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/broadcasters/refresh-rankings"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/broadcasters/refresh-rankings', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/broadcasters/refresh-rankings'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/broadcasters/refresh-rankings", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/discovery-lists

`GET /api/v1/admin/discovery-lists`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/discovery-lists"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/discovery-lists'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/discovery-lists", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/discovery-lists

`POST /api/v1/admin/discovery-lists`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/discovery-lists"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/discovery-lists'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/discovery-lists", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/discovery-lists/{id}

`GET /api/v1/admin/discovery-lists/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/discovery-lists/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/discovery-lists/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/discovery-lists/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/admin/discovery-lists/{id}

`PUT /api/v1/admin/discovery-lists/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/discovery-lists/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists/{id}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/discovery-lists/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/admin/discovery-lists/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/admin/discovery-lists/{id}

`DELETE /api/v1/admin/discovery-lists/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/discovery-lists/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/discovery-lists/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/discovery-lists/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/discovery-lists/{id}/clips

`GET /api/v1/admin/discovery-lists/{id}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/discovery-lists/{id}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists/{id}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/discovery-lists/{id}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/discovery-lists/{id}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/discovery-lists/{id}/clips

`POST /api/v1/admin/discovery-lists/{id}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/discovery-lists/{id}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists/{id}/clips', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/discovery-lists/{id}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/discovery-lists/{id}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/admin/discovery-lists/{id}/clips/reorder

`PUT /api/v1/admin/discovery-lists/{id}/clips/reorder`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/discovery-lists/{id}/clips/reorder"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists/{id}/clips/reorder', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/discovery-lists/{id}/clips/reorder'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/admin/discovery-lists/{id}/clips/reorder", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/admin/discovery-lists/{id}/clips/{clipId}

`DELETE /api/v1/admin/discovery-lists/{id}/clips/{clipId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| clipId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/discovery-lists/{id}/clips/{clipId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/discovery-lists/{id}/clips/{clipId}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/discovery-lists/{id}/clips/{clipId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/discovery-lists/{id}/clips/{clipId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/email/alerts

`GET /api/v1/admin/email/alerts`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/email/alerts"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/email/alerts', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/email/alerts'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/email/alerts", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/email/alerts/{id}/acknowledge

`POST /api/v1/admin/email/alerts/{id}/acknowledge`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/email/alerts/{id}/acknowledge"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/email/alerts/{id}/acknowledge', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/email/alerts/{id}/acknowledge'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/email/alerts/{id}/acknowledge", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/email/alerts/{id}/resolve

`POST /api/v1/admin/email/alerts/{id}/resolve`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/email/alerts/{id}/resolve"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/email/alerts/{id}/resolve', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/email/alerts/{id}/resolve'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/email/alerts/{id}/resolve", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/email/logs

`GET /api/v1/admin/email/logs`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/email/logs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/email/logs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/email/logs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/email/logs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/email/metrics

`GET /api/v1/admin/email/metrics`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/email/metrics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/email/metrics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/email/metrics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/email/metrics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/email/metrics/dashboard

`GET /api/v1/admin/email/metrics/dashboard`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/email/metrics/dashboard"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/email/metrics/dashboard', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/email/metrics/dashboard'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/email/metrics/dashboard", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/email/metrics/templates

`GET /api/v1/admin/email/metrics/templates`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/email/metrics/templates"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/email/metrics/templates', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/email/metrics/templates'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/email/metrics/templates", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/forum/bans

`GET /api/v1/admin/forum/bans`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/forum/bans"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/forum/bans', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/forum/bans'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/forum/bans", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/forum/flagged

`GET /api/v1/admin/forum/flagged`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/forum/flagged"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/forum/flagged', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/forum/flagged'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/forum/flagged", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/forum/moderation-log

`GET /api/v1/admin/forum/moderation-log`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/forum/moderation-log"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/forum/moderation-log', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/forum/moderation-log'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/forum/moderation-log", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/forum/threads/{id}/delete

`POST /api/v1/admin/forum/threads/{id}/delete`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/forum/threads/{id}/delete"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/forum/threads/{id}/delete', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/forum/threads/{id}/delete'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/forum/threads/{id}/delete", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/forum/threads/{id}/lock

`POST /api/v1/admin/forum/threads/{id}/lock`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/forum/threads/{id}/lock"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/forum/threads/{id}/lock', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/forum/threads/{id}/lock'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/forum/threads/{id}/lock", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/forum/threads/{id}/pin

`POST /api/v1/admin/forum/threads/{id}/pin`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/forum/threads/{id}/pin"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/forum/threads/{id}/pin', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/forum/threads/{id}/pin'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/forum/threads/{id}/pin", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/forum/users/{id}/ban

`POST /api/v1/admin/forum/users/{id}/ban`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/forum/users/{id}/ban"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/forum/users/{id}/ban', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/forum/users/{id}/ban'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/forum/users/{id}/ban", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/abuse/{userId}

`GET /api/v1/admin/moderation/abuse/{userId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| userId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/abuse/{userId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/abuse/{userId}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/abuse/{userId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/abuse/{userId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/analytics

`GET /api/v1/admin/moderation/analytics`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/analytics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/analytics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/analytics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/analytics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/appeals

`GET /api/v1/admin/moderation/appeals`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/appeals"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/appeals', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/appeals'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/appeals", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/moderation/appeals/{id}/resolve

`POST /api/v1/admin/moderation/appeals/{id}/resolve`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/moderation/appeals/{id}/resolve"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/appeals/{id}/resolve', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/moderation/appeals/{id}/resolve'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/moderation/appeals/{id}/resolve", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/audit

`GET /api/v1/admin/moderation/audit`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/audit"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/audit', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/audit'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/audit", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/moderation/bulk

`POST /api/v1/admin/moderation/bulk`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/moderation/bulk"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/bulk', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/moderation/bulk'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/moderation/bulk", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/events

`GET /api/v1/admin/moderation/events`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/events"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/events', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/events'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/events", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/moderation/events/{id}/process

`POST /api/v1/admin/moderation/events/{id}/process`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/moderation/events/{id}/process"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/events/{id}/process', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/moderation/events/{id}/process'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/moderation/events/{id}/process", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/moderation/events/{id}/review

`POST /api/v1/admin/moderation/events/{id}/review`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/moderation/events/{id}/review"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/events/{id}/review', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/moderation/events/{id}/review'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/moderation/events/{id}/review", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/events/{type}

`GET /api/v1/admin/moderation/events/{type}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| type | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/events/{type}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/events/{type}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/events/{type}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/events/{type}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/queue

`GET /api/v1/admin/moderation/queue`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/queue"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/queue', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/queue'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/queue", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/queue/stats

`GET /api/v1/admin/moderation/queue/stats`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/queue/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/queue/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/queue/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/queue/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/stats

`GET /api/v1/admin/moderation/stats`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/moderation/toxicity/metrics

`GET /api/v1/admin/moderation/toxicity/metrics`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/moderation/toxicity/metrics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/toxicity/metrics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/moderation/toxicity/metrics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/moderation/toxicity/metrics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/moderation/{id}/approve

`POST /api/v1/admin/moderation/{id}/approve`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/moderation/{id}/approve"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/{id}/approve', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/moderation/{id}/approve'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/moderation/{id}/approve", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/moderation/{id}/reject

`POST /api/v1/admin/moderation/{id}/reject`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/moderation/{id}/reject"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/moderation/{id}/reject', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/moderation/{id}/reject'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/moderation/{id}/reject", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/nsfw/batch-detect

`POST /api/v1/admin/nsfw/batch-detect`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/nsfw/batch-detect"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/nsfw/batch-detect', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/nsfw/batch-detect'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/nsfw/batch-detect", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/nsfw/config

`GET /api/v1/admin/nsfw/config`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/nsfw/config"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/nsfw/config', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/nsfw/config'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/nsfw/config", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/nsfw/detect

`POST /api/v1/admin/nsfw/detect`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/nsfw/detect"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/nsfw/detect', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/nsfw/detect'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/nsfw/detect", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/nsfw/health

`GET /api/v1/admin/nsfw/health`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/nsfw/health"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/nsfw/health', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/nsfw/health'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/nsfw/health", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/nsfw/metrics

`GET /api/v1/admin/nsfw/metrics`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/nsfw/metrics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/nsfw/metrics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/nsfw/metrics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/nsfw/metrics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/nsfw/scan-clips

`POST /api/v1/admin/nsfw/scan-clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/nsfw/scan-clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/nsfw/scan-clips', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/nsfw/scan-clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/nsfw/scan-clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/playlist-scripts

`GET /api/v1/admin/playlist-scripts`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/playlist-scripts"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/playlist-scripts', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/playlist-scripts'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/playlist-scripts", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/playlist-scripts

`POST /api/v1/admin/playlist-scripts`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/playlist-scripts"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/playlist-scripts', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/playlist-scripts'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/playlist-scripts", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/admin/playlist-scripts/{id}

`PUT /api/v1/admin/playlist-scripts/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/playlist-scripts/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/playlist-scripts/{id}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/playlist-scripts/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/admin/playlist-scripts/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/admin/playlist-scripts/{id}

`DELETE /api/v1/admin/playlist-scripts/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/playlist-scripts/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/playlist-scripts/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/playlist-scripts/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/playlist-scripts/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/playlist-scripts/{id}/generate

`POST /api/v1/admin/playlist-scripts/{id}/generate`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/playlist-scripts/{id}/generate"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/playlist-scripts/{id}/generate', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/playlist-scripts/{id}/generate'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/playlist-scripts/{id}/generate", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/submissions

`GET /api/v1/admin/submissions`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/submissions"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/submissions', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/submissions'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/submissions", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/submissions/bulk-approve

`POST /api/v1/admin/submissions/bulk-approve`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/submissions/bulk-approve"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/submissions/bulk-approve', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/submissions/bulk-approve'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/submissions/bulk-approve", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/submissions/bulk-reject

`POST /api/v1/admin/submissions/bulk-reject`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/submissions/bulk-reject"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/submissions/bulk-reject', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/submissions/bulk-reject'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/submissions/bulk-reject", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/submissions/rejection-reasons

`GET /api/v1/admin/submissions/rejection-reasons`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/submissions/rejection-reasons"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/submissions/rejection-reasons', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/submissions/rejection-reasons'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/submissions/rejection-reasons", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/submissions/{id}/approve

`POST /api/v1/admin/submissions/{id}/approve`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/submissions/{id}/approve"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/submissions/{id}/approve', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/submissions/{id}/approve'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/submissions/{id}/approve", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/submissions/{id}/reject

`POST /api/v1/admin/submissions/{id}/reject`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/submissions/{id}/reject"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/submissions/{id}/reject', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/submissions/{id}/reject'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/submissions/{id}/reject", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/tags

`POST /api/v1/admin/tags`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/tags"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/tags', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/tags'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/tags", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/tags/blacklist

`GET /api/v1/admin/tags/blacklist`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/tags/blacklist"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/tags/blacklist', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/tags/blacklist'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/tags/blacklist", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/tags/blacklist

`POST /api/v1/admin/tags/blacklist`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/tags/blacklist"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/tags/blacklist', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/tags/blacklist'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/tags/blacklist", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/admin/tags/blacklist/{id}

`DELETE /api/v1/admin/tags/blacklist/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/tags/blacklist/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/tags/blacklist/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/tags/blacklist/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/tags/blacklist/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/admin/tags/{id}

`PUT /api/v1/admin/tags/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/tags/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/tags/{id}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/admin/tags/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/admin/tags/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/admin/tags/{id}

`DELETE /api/v1/admin/tags/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/tags/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/tags/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/tags/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/tags/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/users

`GET /api/v1/admin/users`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/users"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/users'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/users", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/users/{id}/badges

`POST /api/v1/admin/users/{id}/badges`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/users/{id}/badges"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/badges', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/users/{id}/badges'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/users/{id}/badges", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/admin/users/{id}/badges/{badgeId}

`DELETE /api/v1/admin/users/{id}/badges/{badgeId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| badgeId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/users/{id}/badges/{badgeId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/badges/{badgeId}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/admin/users/{id}/badges/{badgeId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/admin/users/{id}/badges/{badgeId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/users/{id}/ban

`POST /api/v1/admin/users/{id}/ban`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/users/{id}/ban"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/ban', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/users/{id}/ban'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/users/{id}/ban", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/users/{id}/comment-suspension-history

`GET /api/v1/admin/users/{id}/comment-suspension-history`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/users/{id}/comment-suspension-history"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/comment-suspension-history', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/users/{id}/comment-suspension-history'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/users/{id}/comment-suspension-history", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PATCH /api/v1/admin/users/{id}/karma

`PATCH /api/v1/admin/users/{id}/karma`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/admin/users/{id}/karma"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/karma', {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/admin/users/{id}/karma'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PATCH", "/api/v1/admin/users/{id}/karma", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/users/{id}/lift-comment-suspension

`POST /api/v1/admin/users/{id}/lift-comment-suspension`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/users/{id}/lift-comment-suspension"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/lift-comment-suspension', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/users/{id}/lift-comment-suspension'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/users/{id}/lift-comment-suspension", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PATCH /api/v1/admin/users/{id}/role

`PATCH /api/v1/admin/users/{id}/role`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/admin/users/{id}/role"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/role', {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/admin/users/{id}/role'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PATCH", "/api/v1/admin/users/{id}/role", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/users/{id}/suspend-comments

`POST /api/v1/admin/users/{id}/suspend-comments`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/users/{id}/suspend-comments"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/suspend-comments', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/users/{id}/suspend-comments'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/users/{id}/suspend-comments", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/users/{id}/toggle-comment-review

`POST /api/v1/admin/users/{id}/toggle-comment-review`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/users/{id}/toggle-comment-review"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/toggle-comment-review', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/users/{id}/toggle-comment-review'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/users/{id}/toggle-comment-review", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/users/{id}/unban

`POST /api/v1/admin/users/{id}/unban`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/users/{id}/unban"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/users/{id}/unban', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/users/{id}/unban'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/users/{id}/unban", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/verification/applications

`GET /api/v1/admin/verification/applications`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/verification/applications"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/verification/applications', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/verification/applications'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/verification/applications", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/verification/applications/{id}

`GET /api/v1/admin/verification/applications/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/verification/applications/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/verification/applications/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/verification/applications/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/verification/applications/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/admin/verification/applications/{id}/review

`POST /api/v1/admin/verification/applications/{id}/review`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/admin/verification/applications/{id}/review"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/verification/applications/{id}/review', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/admin/verification/applications/{id}/review'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/admin/verification/applications/{id}/review", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/verification/audit-logs

`GET /api/v1/admin/verification/audit-logs`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/verification/audit-logs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/verification/audit-logs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/verification/audit-logs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/verification/audit-logs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/verification/stats

`GET /api/v1/admin/verification/stats`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/verification/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/verification/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/verification/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/verification/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/admin/verification/users/{user_id}/audit-logs

`GET /api/v1/admin/verification/users/{user_id}/audit-logs`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| user_id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/admin/verification/users/{user_id}/audit-logs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/admin/verification/users/{user_id}/audit-logs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/admin/verification/users/{user_id}/audit-logs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/admin/verification/users/{user_id}/audit-logs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/ads/select

`GET /api/v1/ads/select`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/ads/select"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/ads/select', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/ads/select'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/ads/select", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/ads/track/{id}

`POST /api/v1/ads/track/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/ads/track/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/ads/track/{id}', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/ads/track/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/ads/track/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/ads/{id}

`GET /api/v1/ads/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/ads/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/ads/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/ads/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/ads/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/badges

`GET /api/v1/badges`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/badges"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/badges', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/badges'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/badges", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/broadcasters/popular

`GET /api/v1/broadcasters/popular`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/broadcasters/popular"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/broadcasters/popular', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/broadcasters/popular'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/broadcasters/popular", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/broadcasters/rankings

`GET /api/v1/broadcasters/rankings`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/broadcasters/rankings"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/broadcasters/rankings', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/broadcasters/rankings'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/broadcasters/rankings", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/broadcasters/{id}

`GET /api/v1/broadcasters/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/broadcasters/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/broadcasters/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/broadcasters/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/broadcasters/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/broadcasters/{id}/clips

`GET /api/v1/broadcasters/{id}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/broadcasters/{id}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/broadcasters/{id}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/broadcasters/{id}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/broadcasters/{id}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/broadcasters/{id}/follow

`POST /api/v1/broadcasters/{id}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/broadcasters/{id}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/broadcasters/{id}/follow', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/broadcasters/{id}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/broadcasters/{id}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/broadcasters/{id}/follow

`DELETE /api/v1/broadcasters/{id}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/broadcasters/{id}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/broadcasters/{id}/follow', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/broadcasters/{id}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/broadcasters/{id}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/categories

`GET /api/v1/categories`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/categories"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/categories', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/categories'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/categories", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/categories/{slug}

`GET /api/v1/categories/{slug}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| slug | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/categories/{slug}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/categories/{slug}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/categories/{slug}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/categories/{slug}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/categories/{slug}/clips

`GET /api/v1/categories/{slug}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| slug | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/categories/{slug}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/categories/{slug}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/categories/{slug}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/categories/{slug}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/categories/{slug}/games

`GET /api/v1/categories/{slug}/games`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| slug | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/categories/{slug}/games"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/categories/{slug}/games', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/categories/{slug}/games'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/categories/{slug}/games", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/channels

`GET /api/v1/chat/channels`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/channels"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/channels'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/channels", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/chat/channels

`POST /api/v1/chat/channels`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/chat/channels"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/chat/channels'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/chat/channels", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/channels/{id}

`GET /api/v1/chat/channels/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/channels/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/channels/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/channels/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PATCH /api/v1/chat/channels/{id}

`PATCH /api/v1/chat/channels/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/chat/channels/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}', {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/chat/channels/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PATCH", "/api/v1/chat/channels/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/chat/channels/{id}

`DELETE /api/v1/chat/channels/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/chat/channels/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/chat/channels/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/chat/channels/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/chat/channels/{id}/ban

`POST /api/v1/chat/channels/{id}/ban`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/chat/channels/{id}/ban"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/ban', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/chat/channels/{id}/ban'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/chat/channels/{id}/ban", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/chat/channels/{id}/ban/{user_id}

`DELETE /api/v1/chat/channels/{id}/ban/{user_id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| user_id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/chat/channels/{id}/ban/{user_id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/ban/{user_id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/chat/channels/{id}/ban/{user_id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/chat/channels/{id}/ban/{user_id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/channels/{id}/check-ban

`GET /api/v1/chat/channels/{id}/check-ban`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/channels/{id}/check-ban"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/check-ban', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/channels/{id}/check-ban'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/channels/{id}/check-ban", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/channels/{id}/members

`GET /api/v1/chat/channels/{id}/members`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/channels/{id}/members"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/members', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/channels/{id}/members'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/channels/{id}/members", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/chat/channels/{id}/members

`POST /api/v1/chat/channels/{id}/members`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/chat/channels/{id}/members"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/members', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/chat/channels/{id}/members'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/chat/channels/{id}/members", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PATCH /api/v1/chat/channels/{id}/members/{user_id}

`PATCH /api/v1/chat/channels/{id}/members/{user_id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| user_id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/chat/channels/{id}/members/{user_id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/members/{user_id}', {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/chat/channels/{id}/members/{user_id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PATCH", "/api/v1/chat/channels/{id}/members/{user_id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/chat/channels/{id}/members/{user_id}

`DELETE /api/v1/chat/channels/{id}/members/{user_id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| user_id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/chat/channels/{id}/members/{user_id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/members/{user_id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/chat/channels/{id}/members/{user_id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/chat/channels/{id}/members/{user_id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/channels/{id}/messages

`GET /api/v1/chat/channels/{id}/messages`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/channels/{id}/messages"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/messages', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/channels/{id}/messages'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/channels/{id}/messages", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/channels/{id}/moderation-log

`GET /api/v1/chat/channels/{id}/moderation-log`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/channels/{id}/moderation-log"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/moderation-log', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/channels/{id}/moderation-log'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/channels/{id}/moderation-log", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/chat/channels/{id}/mute

`POST /api/v1/chat/channels/{id}/mute`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/chat/channels/{id}/mute"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/mute', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/chat/channels/{id}/mute'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/chat/channels/{id}/mute", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/channels/{id}/role

`GET /api/v1/chat/channels/{id}/role`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/channels/{id}/role"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/role', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/channels/{id}/role'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/channels/{id}/role", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/chat/channels/{id}/timeout

`POST /api/v1/chat/channels/{id}/timeout`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/chat/channels/{id}/timeout"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/timeout', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/chat/channels/{id}/timeout'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/chat/channels/{id}/timeout", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/channels/{id}/ws

`GET /api/v1/chat/channels/{id}/ws`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/channels/{id}/ws"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/channels/{id}/ws', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/channels/{id}/ws'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/channels/{id}/ws", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/health

`GET /api/v1/chat/health`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/health"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/health', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/health'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/health", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/chat/messages/{id}

`DELETE /api/v1/chat/messages/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/chat/messages/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/messages/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/chat/messages/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/chat/messages/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/chat/stats

`GET /api/v1/chat/stats`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/chat/stats"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/chat/stats', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/chat/stats'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/chat/stats", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/clips/{id}/backfill

`POST /api/v1/clips/{id}/backfill`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/clips/{id}/backfill"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/backfill', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/clips/{id}/backfill'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/clips/{id}/backfill", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/clips/{id}/media

`GET /api/v1/clips/{id}/media`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}/media"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/media', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}/media'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}/media", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/clips/{id}/processing-status

`GET /api/v1/clips/{id}/processing-status`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/clips/{id}/processing-status"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/clips/{id}/processing-status', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/clips/{id}/processing-status'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/clips/{id}/processing-status", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/communities

`GET /api/v1/communities`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/communities"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/communities'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/communities", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/communities

`POST /api/v1/communities`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/communities"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/communities'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/communities", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/communities/search

`GET /api/v1/communities/search`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/communities/search"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/search', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/communities/search'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/communities/search", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/communities/{id}

`GET /api/v1/communities/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/communities/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/communities/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/communities/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/communities/{id}

`PUT /api/v1/communities/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/communities/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/communities/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/communities/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/communities/{id}

`DELETE /api/v1/communities/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/communities/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/communities/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/communities/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/communities/{id}/ban

`POST /api/v1/communities/{id}/ban`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/communities/{id}/ban"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/ban', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/communities/{id}/ban'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/communities/{id}/ban", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/communities/{id}/ban/{userId}

`DELETE /api/v1/communities/{id}/ban/{userId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| userId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/communities/{id}/ban/{userId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/ban/{userId}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/communities/{id}/ban/{userId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/communities/{id}/ban/{userId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/communities/{id}/bans

`GET /api/v1/communities/{id}/bans`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/communities/{id}/bans"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/bans', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/communities/{id}/bans'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/communities/{id}/bans", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/communities/{id}/clips

`POST /api/v1/communities/{id}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/communities/{id}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/clips', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/communities/{id}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/communities/{id}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/communities/{id}/clips/{clipId}

`DELETE /api/v1/communities/{id}/clips/{clipId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| clipId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/communities/{id}/clips/{clipId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/clips/{clipId}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/communities/{id}/clips/{clipId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/communities/{id}/clips/{clipId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/communities/{id}/discussions

`GET /api/v1/communities/{id}/discussions`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/communities/{id}/discussions"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/discussions', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/communities/{id}/discussions'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/communities/{id}/discussions", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/communities/{id}/discussions

`POST /api/v1/communities/{id}/discussions`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/communities/{id}/discussions"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/discussions', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/communities/{id}/discussions'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/communities/{id}/discussions", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/communities/{id}/discussions/{discussionId}

`GET /api/v1/communities/{id}/discussions/{discussionId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| discussionId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/communities/{id}/discussions/{discussionId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/discussions/{discussionId}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/communities/{id}/discussions/{discussionId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/communities/{id}/discussions/{discussionId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/communities/{id}/discussions/{discussionId}

`PUT /api/v1/communities/{id}/discussions/{discussionId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| discussionId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/communities/{id}/discussions/{discussionId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/discussions/{discussionId}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/communities/{id}/discussions/{discussionId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/communities/{id}/discussions/{discussionId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/communities/{id}/discussions/{discussionId}

`DELETE /api/v1/communities/{id}/discussions/{discussionId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| discussionId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/communities/{id}/discussions/{discussionId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/discussions/{discussionId}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/communities/{id}/discussions/{discussionId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/communities/{id}/discussions/{discussionId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/communities/{id}/feed

`GET /api/v1/communities/{id}/feed`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/communities/{id}/feed"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/feed', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/communities/{id}/feed'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/communities/{id}/feed", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/communities/{id}/join

`POST /api/v1/communities/{id}/join`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/communities/{id}/join"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/join', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/communities/{id}/join'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/communities/{id}/join", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/communities/{id}/leave

`POST /api/v1/communities/{id}/leave`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/communities/{id}/leave"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/leave', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/communities/{id}/leave'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/communities/{id}/leave", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/communities/{id}/members

`GET /api/v1/communities/{id}/members`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/communities/{id}/members"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/members', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/communities/{id}/members'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/communities/{id}/members", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/communities/{id}/members/{userId}/role

`PUT /api/v1/communities/{id}/members/{userId}/role`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| userId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/communities/{id}/members/{userId}/role"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/communities/{id}/members/{userId}/role', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/communities/{id}/members/{userId}/role'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/communities/{id}/members/{userId}/role", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/creators/{creatorName}/clips

`GET /api/v1/creators/{creatorName}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| creatorName | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/creators/{creatorName}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/creators/{creatorName}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/creators/{creatorName}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/creators/{creatorName}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/discovery-lists

`GET /api/v1/discovery-lists`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/discovery-lists"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/discovery-lists', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/discovery-lists'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/discovery-lists", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/discovery-lists/{id}

`GET /api/v1/discovery-lists/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/discovery-lists/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/discovery-lists/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/discovery-lists/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/discovery-lists/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/discovery-lists/{id}/bookmark

`POST /api/v1/discovery-lists/{id}/bookmark`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/discovery-lists/{id}/bookmark"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/discovery-lists/{id}/bookmark', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/discovery-lists/{id}/bookmark'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/discovery-lists/{id}/bookmark", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/discovery-lists/{id}/bookmark

`DELETE /api/v1/discovery-lists/{id}/bookmark`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/discovery-lists/{id}/bookmark"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/discovery-lists/{id}/bookmark', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/discovery-lists/{id}/bookmark'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/discovery-lists/{id}/bookmark", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/discovery-lists/{id}/clips

`GET /api/v1/discovery-lists/{id}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/discovery-lists/{id}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/discovery-lists/{id}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/discovery-lists/{id}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/discovery-lists/{id}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/discovery-lists/{id}/follow

`POST /api/v1/discovery-lists/{id}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/discovery-lists/{id}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/discovery-lists/{id}/follow', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/discovery-lists/{id}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/discovery-lists/{id}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/discovery-lists/{id}/follow

`DELETE /api/v1/discovery-lists/{id}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/discovery-lists/{id}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/discovery-lists/{id}/follow', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/discovery-lists/{id}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/discovery-lists/{id}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/feeds/clips

`GET /api/v1/feeds/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/feeds/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/feeds/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/feeds/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/feeds/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/feeds/discover

`GET /api/v1/feeds/discover`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/feeds/discover"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/feeds/discover', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/feeds/discover'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/feeds/discover", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/feeds/following

`GET /api/v1/feeds/following`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/feeds/following"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/feeds/following', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/feeds/following'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/feeds/following", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/feeds/search

`GET /api/v1/feeds/search`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/feeds/search"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/feeds/search', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/feeds/search'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/feeds/search", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/forum/analytics

`GET /api/v1/forum/analytics`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/forum/analytics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/analytics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/forum/analytics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/forum/analytics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/forum/flag

`POST /api/v1/forum/flag`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/forum/flag"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/flag', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/forum/flag'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/forum/flag", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/forum/helpful-replies

`GET /api/v1/forum/helpful-replies`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/forum/helpful-replies"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/helpful-replies', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/forum/helpful-replies'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/forum/helpful-replies", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/forum/popular

`GET /api/v1/forum/popular`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/forum/popular"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/popular', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/forum/popular'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/forum/popular", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PATCH /api/v1/forum/replies/{id}

`PATCH /api/v1/forum/replies/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/forum/replies/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/replies/{id}', {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/forum/replies/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PATCH", "/api/v1/forum/replies/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/forum/replies/{id}

`DELETE /api/v1/forum/replies/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/forum/replies/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/replies/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/forum/replies/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/forum/replies/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/forum/replies/{id}/vote

`POST /api/v1/forum/replies/{id}/vote`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/forum/replies/{id}/vote"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/replies/{id}/vote', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/forum/replies/{id}/vote'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/forum/replies/{id}/vote", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/forum/replies/{id}/votes

`GET /api/v1/forum/replies/{id}/votes`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/forum/replies/{id}/votes"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/replies/{id}/votes', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/forum/replies/{id}/votes'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/forum/replies/{id}/votes", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/forum/search

`GET /api/v1/forum/search`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/forum/search"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/search', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/forum/search'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/forum/search", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/forum/threads

`GET /api/v1/forum/threads`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/forum/threads"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/threads', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/forum/threads'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/forum/threads", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/forum/threads

`POST /api/v1/forum/threads`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/forum/threads"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/threads', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/forum/threads'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/forum/threads", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/forum/threads/{id}

`GET /api/v1/forum/threads/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/forum/threads/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/threads/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/forum/threads/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/forum/threads/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/forum/threads/{id}/replies

`POST /api/v1/forum/threads/{id}/replies`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/forum/threads/{id}/replies"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/threads/{id}/replies', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/forum/threads/{id}/replies'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/forum/threads/{id}/replies", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/forum/users/{id}/reputation

`GET /api/v1/forum/users/{id}/reputation`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/forum/users/{id}/reputation"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/forum/users/{id}/reputation', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/forum/users/{id}/reputation'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/forum/users/{id}/reputation", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/games/trending

`GET /api/v1/games/trending`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/games/trending"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/games/trending', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/games/trending'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/games/trending", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/games/{gameId}

`GET /api/v1/games/{gameId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| gameId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/games/{gameId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/games/{gameId}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/games/{gameId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/games/{gameId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/games/{gameId}/clips

`GET /api/v1/games/{gameId}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| gameId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/games/{gameId}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/games/{gameId}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/games/{gameId}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/games/{gameId}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/games/{gameId}/follow

`POST /api/v1/games/{gameId}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| gameId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/games/{gameId}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/games/{gameId}/follow', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/games/{gameId}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/games/{gameId}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/games/{gameId}/follow

`DELETE /api/v1/games/{gameId}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| gameId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/games/{gameId}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/games/{gameId}/follow', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/games/{gameId}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/games/{gameId}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/leaderboards/{type}

`GET /api/v1/leaderboards/{type}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| type | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/leaderboards/{type}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/leaderboards/{type}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/leaderboards/{type}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/leaderboards/{type}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/moderation/ban

`POST /api/v1/moderation/ban`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/moderation/ban"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/moderation/ban', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/moderation/ban'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/moderation/ban", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/notifications

`GET /api/v1/notifications`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/notifications"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/notifications'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/notifications", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/notifications/count

`GET /api/v1/notifications/count`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/notifications/count"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/count', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/notifications/count'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/notifications/count", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/notifications/preferences

`GET /api/v1/notifications/preferences`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/notifications/preferences"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/preferences', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/notifications/preferences'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/notifications/preferences", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/notifications/preferences

`PUT /api/v1/notifications/preferences`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/notifications/preferences"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/preferences', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/notifications/preferences'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/notifications/preferences", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/notifications/preferences/reset

`POST /api/v1/notifications/preferences/reset`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/notifications/preferences/reset"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/preferences/reset', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/notifications/preferences/reset'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/notifications/preferences/reset", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/notifications/read-all

`PUT /api/v1/notifications/read-all`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/notifications/read-all"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/read-all', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/notifications/read-all'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/notifications/read-all", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/notifications/register

`POST /api/v1/notifications/register`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/notifications/register"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/register', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/notifications/register'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/notifications/register", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/notifications/unregister

`DELETE /api/v1/notifications/unregister`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/notifications/unregister"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/unregister', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/notifications/unregister'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/notifications/unregister", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/notifications/unsubscribe

`GET /api/v1/notifications/unsubscribe`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/notifications/unsubscribe"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/unsubscribe', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/notifications/unsubscribe'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/notifications/unsubscribe", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/notifications/{id}

`DELETE /api/v1/notifications/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/notifications/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/notifications/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/notifications/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/notifications/{id}/read

`PUT /api/v1/notifications/{id}/read`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/notifications/{id}/read"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/notifications/{id}/read', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/notifications/{id}/read'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/notifications/{id}/read", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlist-scripts

`GET /api/v1/playlist-scripts`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlist-scripts"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlist-scripts', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlist-scripts'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlist-scripts", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlist-scripts

`POST /api/v1/playlist-scripts`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlist-scripts"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlist-scripts', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlist-scripts'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlist-scripts", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/playlist-scripts/{id}

`PUT /api/v1/playlist-scripts/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/playlist-scripts/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlist-scripts/{id}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/playlist-scripts/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/playlist-scripts/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/playlist-scripts/{id}

`DELETE /api/v1/playlist-scripts/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/playlist-scripts/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlist-scripts/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/playlist-scripts/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/playlist-scripts/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlist-scripts/{id}/generate

`POST /api/v1/playlist-scripts/{id}/generate`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlist-scripts/{id}/generate"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlist-scripts/{id}/generate', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlist-scripts/{id}/generate'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlist-scripts/{id}/generate", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists

`GET /api/v1/playlists`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlists

`POST /api/v1/playlists`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlists"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlists'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlists", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists/bookmarks

`GET /api/v1/playlists/bookmarks`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists/bookmarks"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/bookmarks', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists/bookmarks'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists/bookmarks", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists/featured

`GET /api/v1/playlists/featured`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists/featured"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/featured', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists/featured'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists/featured", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists/public

`GET /api/v1/playlists/public`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists/public"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/public', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists/public'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists/public", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists/share/{token}

`GET /api/v1/playlists/share/{token}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| token | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists/share/{token}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/share/{token}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists/share/{token}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists/share/{token}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists/today

`GET /api/v1/playlists/today`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists/today"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/today', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists/today'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists/today", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists/{id}

`GET /api/v1/playlists/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PATCH /api/v1/playlists/{id}

`PATCH /api/v1/playlists/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/playlists/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}', {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/playlists/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PATCH", "/api/v1/playlists/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/playlists/{id}

`DELETE /api/v1/playlists/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/playlists/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/playlists/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/playlists/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlists/{id}/bookmark

`POST /api/v1/playlists/{id}/bookmark`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlists/{id}/bookmark"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/bookmark', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlists/{id}/bookmark'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlists/{id}/bookmark", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/playlists/{id}/bookmark

`DELETE /api/v1/playlists/{id}/bookmark`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/playlists/{id}/bookmark"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/bookmark', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/playlists/{id}/bookmark'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/playlists/{id}/bookmark", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlists/{id}/clips

`POST /api/v1/playlists/{id}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlists/{id}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/clips', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlists/{id}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlists/{id}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/playlists/{id}/clips/order

`PUT /api/v1/playlists/{id}/clips/order`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/playlists/{id}/clips/order"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/clips/order', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/playlists/{id}/clips/order'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/playlists/{id}/clips/order", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/playlists/{id}/clips/{clip_id}

`DELETE /api/v1/playlists/{id}/clips/{clip_id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| clip_id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/playlists/{id}/clips/{clip_id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/clips/{clip_id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/playlists/{id}/clips/{clip_id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/playlists/{id}/clips/{clip_id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists/{id}/collaborators

`GET /api/v1/playlists/{id}/collaborators`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists/{id}/collaborators"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/collaborators', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists/{id}/collaborators'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists/{id}/collaborators", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlists/{id}/collaborators

`POST /api/v1/playlists/{id}/collaborators`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlists/{id}/collaborators"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/collaborators', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlists/{id}/collaborators'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlists/{id}/collaborators", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PATCH /api/v1/playlists/{id}/collaborators/{user_id}

`PATCH /api/v1/playlists/{id}/collaborators/{user_id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| user_id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/playlists/{id}/collaborators/{user_id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/collaborators/{user_id}', {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/playlists/{id}/collaborators/{user_id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PATCH", "/api/v1/playlists/{id}/collaborators/{user_id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/playlists/{id}/collaborators/{user_id}

`DELETE /api/v1/playlists/{id}/collaborators/{user_id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| user_id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/playlists/{id}/collaborators/{user_id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/collaborators/{user_id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/playlists/{id}/collaborators/{user_id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/playlists/{id}/collaborators/{user_id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlists/{id}/copy

`POST /api/v1/playlists/{id}/copy`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlists/{id}/copy"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/copy', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlists/{id}/copy'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlists/{id}/copy", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlists/{id}/like

`POST /api/v1/playlists/{id}/like`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlists/{id}/like"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/like', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlists/{id}/like'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlists/{id}/like", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/playlists/{id}/like

`DELETE /api/v1/playlists/{id}/like`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/playlists/{id}/like"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/like', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/playlists/{id}/like'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/playlists/{id}/like", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/playlists/{id}/share-link

`GET /api/v1/playlists/{id}/share-link`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/playlists/{id}/share-link"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/share-link', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/playlists/{id}/share-link'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/playlists/{id}/share-link", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/playlists/{id}/track-share

`POST /api/v1/playlists/{id}/track-share`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/playlists/{id}/track-share"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/playlists/{id}/track-share', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/playlists/{id}/track-share'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/playlists/{id}/track-share", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/queue

`GET /api/v1/queue`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/queue"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/queue', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/queue'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/queue", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/queue

`POST /api/v1/queue`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/queue"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/queue', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/queue'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/queue", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/queue

`DELETE /api/v1/queue`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/queue"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/queue', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/queue'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/queue", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/queue/convert-to-playlist

`POST /api/v1/queue/convert-to-playlist`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/queue/convert-to-playlist"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/queue/convert-to-playlist', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/queue/convert-to-playlist'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/queue/convert-to-playlist", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/queue/count

`GET /api/v1/queue/count`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/queue/count"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/queue/count', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/queue/count'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/queue/count", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PATCH /api/v1/queue/reorder

`PATCH /api/v1/queue/reorder`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PATCH "http://localhost:8080/api/v1/queue/reorder"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/queue/reorder', {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.patch(
        '/api/v1/queue/reorder'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PATCH", "/api/v1/queue/reorder", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/queue/{id}

`DELETE /api/v1/queue/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/queue/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/queue/{id}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/queue/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/queue/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/queue/{id}/played

`POST /api/v1/queue/{id}/played`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/queue/{id}/played"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/queue/{id}/played', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/queue/{id}/played'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/queue/{id}/played", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/recommendations/clips

`GET /api/v1/recommendations/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/recommendations/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/recommendations/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/recommendations/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/recommendations/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/recommendations/feedback

`POST /api/v1/recommendations/feedback`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/recommendations/feedback"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/recommendations/feedback', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/recommendations/feedback'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/recommendations/feedback", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/recommendations/onboarding

`POST /api/v1/recommendations/onboarding`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/recommendations/onboarding"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/recommendations/onboarding', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/recommendations/onboarding'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/recommendations/onboarding", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/recommendations/preferences

`GET /api/v1/recommendations/preferences`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/recommendations/preferences"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/recommendations/preferences', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/recommendations/preferences'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/recommendations/preferences", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/recommendations/preferences

`PUT /api/v1/recommendations/preferences`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/recommendations/preferences"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/recommendations/preferences', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/recommendations/preferences'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/recommendations/preferences", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/recommendations/track-view/{id}

`POST /api/v1/recommendations/track-view/{id}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/recommendations/track-view/{id}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/recommendations/track-view/{id}', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/recommendations/track-view/{id}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/recommendations/track-view/{id}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/streamer-clip-rooms/{channel}

`GET /api/v1/streamer-clip-rooms/{channel}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channel | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/streamer-clip-rooms/{channel}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streamer-clip-rooms/{channel}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/streamer-clip-rooms/{channel}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/streamer-clip-rooms/{channel}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/streamer-clip-rooms/{channel}/items

`GET /api/v1/streamer-clip-rooms/{channel}/items`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channel | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/streamer-clip-rooms/{channel}/items"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streamer-clip-rooms/{channel}/items', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/streamer-clip-rooms/{channel}/items'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/streamer-clip-rooms/{channel}/items", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/streamer-clip-rooms/{channel}/items/order

`PUT /api/v1/streamer-clip-rooms/{channel}/items/order`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channel | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/streamer-clip-rooms/{channel}/items/order"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streamer-clip-rooms/{channel}/items/order', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/streamer-clip-rooms/{channel}/items/order'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/streamer-clip-rooms/{channel}/items/order", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/streamer-clip-rooms/{channel}/items/{itemId}/approve

`POST /api/v1/streamer-clip-rooms/{channel}/items/{itemId}/approve`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channel | path | string | ✓ |  |
| itemId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/streamer-clip-rooms/{channel}/items/{itemId}/approve"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streamer-clip-rooms/{channel}/items/{itemId}/approve', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/streamer-clip-rooms/{channel}/items/{itemId}/approve'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/streamer-clip-rooms/{channel}/items/{itemId}/approve", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/streamer-clip-rooms/{channel}/items/{itemId}/reject

`POST /api/v1/streamer-clip-rooms/{channel}/items/{itemId}/reject`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channel | path | string | ✓ |  |
| itemId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/streamer-clip-rooms/{channel}/items/{itemId}/reject"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streamer-clip-rooms/{channel}/items/{itemId}/reject', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/streamer-clip-rooms/{channel}/items/{itemId}/reject'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/streamer-clip-rooms/{channel}/items/{itemId}/reject", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/streamer-clip-rooms/{channel}/start

`POST /api/v1/streamer-clip-rooms/{channel}/start`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channel | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/streamer-clip-rooms/{channel}/start"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streamer-clip-rooms/{channel}/start', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/streamer-clip-rooms/{channel}/start'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/streamer-clip-rooms/{channel}/start", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/streamer-clip-rooms/{channel}/stop

`POST /api/v1/streamer-clip-rooms/{channel}/stop`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channel | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/streamer-clip-rooms/{channel}/stop"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streamer-clip-rooms/{channel}/stop', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/streamer-clip-rooms/{channel}/stop'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/streamer-clip-rooms/{channel}/stop", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/streamer-clip-rooms/{channel}/ws

`GET /api/v1/streamer-clip-rooms/{channel}/ws`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| channel | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/streamer-clip-rooms/{channel}/ws"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streamer-clip-rooms/{channel}/ws', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/streamer-clip-rooms/{channel}/ws'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/streamer-clip-rooms/{channel}/ws", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/streams/following

`GET /api/v1/streams/following`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/streams/following"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streams/following', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/streams/following'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/streams/following", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/streams/{streamer}

`GET /api/v1/streams/{streamer}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| streamer | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/streams/{streamer}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streams/{streamer}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/streams/{streamer}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/streams/{streamer}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/streams/{streamer}/follow

`POST /api/v1/streams/{streamer}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| streamer | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/streams/{streamer}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streams/{streamer}/follow', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/streams/{streamer}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/streams/{streamer}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/streams/{streamer}/follow

`DELETE /api/v1/streams/{streamer}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| streamer | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/streams/{streamer}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streams/{streamer}/follow', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/streams/{streamer}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/streams/{streamer}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/streams/{streamer}/follow-status

`GET /api/v1/streams/{streamer}/follow-status`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| streamer | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/streams/{streamer}/follow-status"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/streams/{streamer}/follow-status', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/streams/{streamer}/follow-status'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/streams/{streamer}/follow-status", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/subscriptions/cancel

`POST /api/v1/subscriptions/cancel`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/subscriptions/cancel"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/subscriptions/cancel', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/subscriptions/cancel'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/subscriptions/cancel", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/subscriptions/change-plan

`POST /api/v1/subscriptions/change-plan`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/subscriptions/change-plan"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/subscriptions/change-plan', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/subscriptions/change-plan'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/subscriptions/change-plan", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/subscriptions/checkout

`POST /api/v1/subscriptions/checkout`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/subscriptions/checkout"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/subscriptions/checkout', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/subscriptions/checkout'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/subscriptions/checkout", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/subscriptions/invoices

`GET /api/v1/subscriptions/invoices`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/subscriptions/invoices"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/subscriptions/invoices', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/subscriptions/invoices'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/subscriptions/invoices", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/subscriptions/me

`GET /api/v1/subscriptions/me`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/subscriptions/me"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/subscriptions/me', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/subscriptions/me'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/subscriptions/me", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/subscriptions/portal

`POST /api/v1/subscriptions/portal`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/subscriptions/portal"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/subscriptions/portal', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/subscriptions/portal'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/subscriptions/portal", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/subscriptions/reactivate

`POST /api/v1/subscriptions/reactivate`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/subscriptions/reactivate"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/subscriptions/reactivate', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/subscriptions/reactivate'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/subscriptions/reactivate", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/twitch/auth

`DELETE /api/v1/twitch/auth`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/twitch/auth"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/twitch/auth', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/twitch/auth'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/twitch/auth", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/twitch/auth/status

`GET /api/v1/twitch/auth/status`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/twitch/auth/status"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/twitch/auth/status', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/twitch/auth/status'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/twitch/auth/status", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/twitch/oauth/authorize

`GET /api/v1/twitch/oauth/authorize`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/twitch/oauth/authorize"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/twitch/oauth/authorize', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/twitch/oauth/authorize'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/twitch/oauth/authorize", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/twitch/oauth/callback

`GET /api/v1/twitch/oauth/callback`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/twitch/oauth/callback"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/twitch/oauth/callback', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/twitch/oauth/callback'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/twitch/oauth/callback", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/users/me/discovery-list-follows

`GET /api/v1/users/me/discovery-list-follows`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/me/discovery-list-follows"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/discovery-list-follows', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/me/discovery-list-follows'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/me/discovery-list-follows", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/users/{id}/feeds

`GET /api/v1/users/{id}/feeds`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/feeds"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/feeds'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/feeds", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/users/{id}/feeds

`POST /api/v1/users/{id}/feeds`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/{id}/feeds"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/{id}/feeds'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/users/{id}/feeds", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/users/{id}/feeds/{feedId}

`GET /api/v1/users/{id}/feeds/{feedId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/feeds/{feedId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/feeds/{feedId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/users/{id}/feeds/{feedId}

`PUT /api/v1/users/{id}/feeds/{feedId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/users/{id}/feeds/{feedId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/users/{id}/feeds/{feedId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/users/{id}/feeds/{feedId}

`DELETE /api/v1/users/{id}/feeds/{feedId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/users/{id}/feeds/{feedId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/users/{id}/feeds/{feedId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/users/{id}/feeds/{feedId}/clips

`GET /api/v1/users/{id}/feeds/{feedId}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}/clips', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/feeds/{feedId}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/feeds/{feedId}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/users/{id}/feeds/{feedId}/clips

`POST /api/v1/users/{id}/feeds/{feedId}/clips`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}/clips"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}/clips', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/{id}/feeds/{feedId}/clips'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/users/{id}/feeds/{feedId}/clips", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/users/{id}/feeds/{feedId}/clips/reorder

`PUT /api/v1/users/{id}/feeds/{feedId}/clips/reorder`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}/clips/reorder"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}/clips/reorder', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/users/{id}/feeds/{feedId}/clips/reorder'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/users/{id}/feeds/{feedId}/clips/reorder", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/users/{id}/feeds/{feedId}/clips/{clipId}

`DELETE /api/v1/users/{id}/feeds/{feedId}/clips/{clipId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |
| clipId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}/clips/{clipId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}/clips/{clipId}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/users/{id}/feeds/{feedId}/clips/{clipId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/users/{id}/feeds/{feedId}/clips/{clipId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/users/{id}/feeds/{feedId}/follow

`POST /api/v1/users/{id}/feeds/{feedId}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}/follow', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/{id}/feeds/{feedId}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/users/{id}/feeds/{feedId}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/users/{id}/feeds/{feedId}/follow

`DELETE /api/v1/users/{id}/feeds/{feedId}/follow`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| feedId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/users/{id}/feeds/{feedId}/follow"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/feeds/{feedId}/follow', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/users/{id}/feeds/{feedId}/follow'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/users/{id}/feeds/{feedId}/follow", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/users/{id}/filter-presets

`GET /api/v1/users/{id}/filter-presets`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/filter-presets"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/filter-presets', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/filter-presets'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/filter-presets", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/users/{id}/filter-presets

`POST /api/v1/users/{id}/filter-presets`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/users/{id}/filter-presets"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/filter-presets', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/users/{id}/filter-presets'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/users/{id}/filter-presets", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/users/{id}/filter-presets/{presetId}

`GET /api/v1/users/{id}/filter-presets/{presetId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| presetId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/filter-presets/{presetId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/filter-presets/{presetId}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/filter-presets/{presetId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/filter-presets/{presetId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### PUT /api/v1/users/{id}/filter-presets/{presetId}

`PUT /api/v1/users/{id}/filter-presets/{presetId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| presetId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X PUT "http://localhost:8080/api/v1/users/{id}/filter-presets/{presetId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/filter-presets/{presetId}', {
    method: 'PUT',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.put(
        '/api/v1/users/{id}/filter-presets/{presetId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("PUT", "/api/v1/users/{id}/filter-presets/{presetId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### DELETE /api/v1/users/{id}/filter-presets/{presetId}

`DELETE /api/v1/users/{id}/filter-presets/{presetId}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |
| presetId | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X DELETE "http://localhost:8080/api/v1/users/{id}/filter-presets/{presetId}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/filter-presets/{presetId}', {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.delete(
        '/api/v1/users/{id}/filter-presets/{presetId}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("DELETE", "/api/v1/users/{id}/filter-presets/{presetId}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/users/{id}/games/following

`GET /api/v1/users/{id}/games/following`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| id | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/{id}/games/following"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/{id}/games/following', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/{id}/games/following'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/{id}/games/following", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/verification/applications

`POST /api/v1/verification/applications`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/verification/applications"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/verification/applications', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/verification/applications'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/verification/applications", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /api/v1/verification/applications/me

`GET /api/v1/verification/applications/me`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/verification/applications/me"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/verification/applications/me', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/verification/applications/me'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/verification/applications/me", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### POST /api/v1/webhooks/stripe

`POST /api/v1/webhooks/stripe`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X POST "http://localhost:8080/api/v1/webhooks/stripe"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/webhooks/stripe', {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.post(
        '/api/v1/webhooks/stripe'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("POST", "/api/v1/webhooks/stripe", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /clips/best/{path}

`GET /clips/best/{path}`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Parameters

| Name | In | Type | Required | Description |
|------|-------|------|----------|-------------|
| path | path | string | ✓ |  |

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/clips/best/{path}"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/clips/best/{path}', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/clips/best/{path}'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/clips/best/{path}", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof

`GET /debug/pprof`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/allocs

`GET /debug/pprof/allocs`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/allocs"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/allocs', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/allocs'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/allocs", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/block

`GET /debug/pprof/block`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/block"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/block', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/block'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/block", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/cmdline

`GET /debug/pprof/cmdline`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/cmdline"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/cmdline', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/cmdline'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/cmdline", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/goroutine

`GET /debug/pprof/goroutine`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/goroutine"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/goroutine', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/goroutine'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/goroutine", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/heap

`GET /debug/pprof/heap`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/heap"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/heap', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/heap'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/heap", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/mutex

`GET /debug/pprof/mutex`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/mutex"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/mutex', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/mutex'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/mutex", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/profile

`GET /debug/pprof/profile`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/profile"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/profile', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/profile'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/profile", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/symbol

`GET /debug/pprof/symbol`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/symbol"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/symbol', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/symbol'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/symbol", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/threadcreate

`GET /debug/pprof/threadcreate`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/threadcreate"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/threadcreate', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/threadcreate'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/threadcreate", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /debug/pprof/trace

`GET /debug/pprof/trace`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/debug/pprof/trace"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/debug/pprof/trace', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/debug/pprof/trace'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/debug/pprof/trace", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /health

`GET /health`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/health"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/health', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/health'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/health", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /internal/operations/database

`GET /internal/operations/database`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/internal/operations/database"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/internal/operations/database', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/internal/operations/database'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/internal/operations/database", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

### GET /internal/operations/metrics

`GET /internal/operations/metrics`

Router-derived operation pending a route-specific response schema.

**Tags:** Generated Route Contracts

#### Responses

**200** - Success

**400** - Success

**401** - Success

**500** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/internal/operations/metrics"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/internal/operations/metrics', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/internal/operations/metrics'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/internal/operations/metrics", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        // Handle error
        return
    }
    // Process body
    _ = body
}
```

---

