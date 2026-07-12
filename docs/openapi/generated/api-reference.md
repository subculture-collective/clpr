---
title: "API Reference"
summary: "Complete API reference for Clipper platform"
tags: ["api", "reference", "openapi"]
area: "openapi"
status: "stable"
version: "1.0.0"
generated: 2026-07-12T07:39:51.820Z
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
- [Creators](#creators)
- [Broadcasters](#broadcasters)
- [Categories](#categories)
- [Streams](#streams)
- [Games](#games)
- [Discovery](#discovery)
- [Leaderboards](#leaderboards)
- [Feeds](#feeds)
- [Recommendations](#recommendations)
- [Notifications](#notifications)
- [Verification](#verification)
- [Subscriptions](#subscriptions)
- [Webhooks](#webhooks)
- [Contact](#contact)
- [Chat](#chat)
- [Ads](#ads)
- [Documentation](#documentation)
- [Communities](#communities)
- [Playlists](#playlists)
- [Forum](#forum)
- [Queue](#queue)
- [Watch History](#watch-history)
- [Watch Parties](#watch-parties)
- [Service Status](#service-status)
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

Returns 200 if all services are ready (database, redis, etc.)

**Tags:** Health

#### Responses

**200** - Service is ready

**503** - Success

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

**201** - Log submitted successfully

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

**201** - View tracked

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

Submit a content report (rate limited - 10/hour)

**Tags:** Reports

#### Request Body

Content-Type: `application/json`

#### Responses

**201** - Report submitted

**400** - Success

**401** - Success

**429** - Success

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

**201** - User banned on Twitch

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

**204** - User unbanned on Twitch

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

### Export user data

`GET /api/v1/users/me/export`

Exports user data (GDPR compliance, rate limited - 1/hour)

**Tags:** Users

#### Responses

**200** - Data export initiated

**401** - Success

**429** - Success

#### Code Examples

##### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/users/me/export"
```

##### JavaScript

```javascript
// Using fetch API
try {
  const response = await fetch('/api/v1/users/me/export', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('HTTP error ' + response.status);
  }

  const data = await response.json();
  // Process data
} catch (error) {
  console.error('Error:', error);
}
```

##### Python

```python
import requests

try:
    response = requests.get(
        '/api/v1/users/me/export'
    )
    response.raise_for_status()  # Raise error for bad status
    data = response.json()
    # Process data
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

##### Go

```go
package main

import (
    "net/http"
    "io"
)

func main() {
    req, err := http.NewRequest("GET", "/api/v1/users/me/export", nil)
    if err != nil {
        // Handle error
        return
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        // Handle error
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
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

