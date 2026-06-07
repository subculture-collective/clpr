---
title: "README"
summary: "Complete OpenAPI 3.1 specification for all Clipper API endpoints"
tags: ["openapi", "api-docs", "swagger"]
area: "openapi"
status: "stable"
owner: "team-core"
version: "1.0.0"
last_reviewed: 2026-01-30
---

# Clipper API - OpenAPI Specification

Complete OpenAPI 3.1 specification documenting all 474+ API endpoints for the Clipper platform.

## 📄 Main Specification

### Complete API Specification

- **File**: [`openapi.yaml`](./openapi.yaml) ⭐ **Primary spec - edit this file**
- **Version**: 1.0.0 (OpenAPI 3.1)
- **Description**: Complete API documentation covering all endpoints
- **Size**: 4,691 lines
- **Endpoints**: 474+ documented routes
- **Status**: ✅ Production-ready and validated

> **Note on file organization:**
> - `openapi.yaml` - **Primary source of truth**. Edit this file when making changes.
> - `openapi-bundled.yaml` - Auto-generated bundled version (created by `npm run openapi:bundle`). Do not edit manually.
> - `openapi-main.yaml` - Alternative simplified spec (currently unused). For most cases, use `openapi.yaml`.

**Major API Groups:**
- Health & Monitoring (10 endpoints)
- Authentication & MFA (15 endpoints)
- Clips & Content Management (40+ endpoints)
- Comments & Engagement
- Users & Profiles (40+ fully documented)
- Search & Discovery
- Submissions & Moderation (30+ endpoints)
- Tags & Categories
- Reports & Appeals
- Analytics & Engagement
- Premium/Subscriptions
- Notifications
- Feeds & Recommendations
- Communities & Forums
- Playlists & Watch Parties
- Broadcasters & Live Status
- Admin Operations
- Webhooks & Events

### Legacy Specifications (Deprecated)

> ⚠️ **Note**: These are maintained for backward compatibility but `openapi.yaml` is now the single source of truth.

- [`clip-submission-api.yaml`](./clip-submission-api.yaml) - Clip submission workflow
- [`comments-api.yaml`](./comments-api.yaml) - Comment threading system

## 🚀 Quick Start

### Viewing the API Documentation

#### Option 1: Swagger UI (Recommended)

Open the interactive API documentation in your browser:

```bash
# Using Make
make openapi-serve

# Or using npm
npm run openapi:serve
```

Then open http://localhost:8081 in your browser.

#### Option 2: Redocly Preview (Live Reload)

For documentation development with live reload:

```bash
# Using npm
npm run openapi:preview

# Or directly
npx @redocly/cli preview-docs docs/openapi/openapi.yaml
```

#### Option 3: Online Viewers

Upload or paste the spec into:
- [Swagger Editor](https://editor.swagger.io/)
- [Redoc](https://redocly.github.io/redoc/)
- [Stoplight Studio](https://stoplight.io/studio/)

### Validating the Specification

```bash
# Using Make
make openapi-validate

# Or using npm
npm run openapi:validate

# Or directly
npx @redocly/cli lint docs/openapi/openapi.yaml
```

### Building Static Documentation

Generate a standalone HTML file:

```bash
# Using Make
make openapi-build

# Or using npm
npm run openapi:build

# Output: docs/openapi/api-docs.html
```

### Getting Spec Statistics

```bash
# Using Make
make openapi-stats

# Or using npm
npm run openapi:stats
```

## 🛠️ Development Workflow

### Making Changes

1. **Edit the specification**: `docs/openapi/openapi.yaml`
2. **Validate your changes**: `make openapi-validate`
3. **Preview locally**: `make openapi-preview`
4. **Commit and push**: Changes are validated automatically in CI

### CI/CD Validation

OpenAPI specifications are automatically validated on:
- Pull requests
- Pushes to main/develop branches
- Changes to OpenAPI files

See [`.github/workflows/openapi.yml`](../../.github/workflows/openapi.yml) for details.

## 📦 Generating Client SDKs

Use OpenAPI Generator to create type-safe clients for any language:

### TypeScript/JavaScript (Axios)

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi/openapi.yaml \
  -g typescript-axios \
  -o generated/typescript-client \
  --additional-properties=npmName=@clpr/api-client,npmVersion=1.0.0
```

### TypeScript/JavaScript (Fetch)

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi/openapi.yaml \
  -g typescript-fetch \
  -o generated/typescript-fetch-client
```

### Python

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi/openapi.yaml \
  -g python \
  -o generated/python-client \
  --additional-properties=packageName=clpr_api,packageVersion=1.0.0
```

### Go

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi/openapi.yaml \
  -g go \
  -o generated/go-client \
  --additional-properties=packageName=clpr
```

### Other Languages

OpenAPI Generator supports 50+ languages. See [available generators](https://openapi-generator.tech/docs/generators).

## 🧪 Testing

### Import into API Testing Tools

#### Postman

1. Open Postman
2. Click "Import" → "Upload Files"
3. Select `docs/openapi/openapi.yaml`
4. Postman creates a collection with all endpoints

#### Insomnia

1. Open Insomnia
2. Click "Create" → "Import From" → "File"
3. Select `docs/openapi/openapi.yaml`

#### REST Client (VS Code)

Use the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension with the generated examples.

## 🔒 Authentication

Most endpoints require JWT Bearer authentication:

```bash
# Example authenticated request
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.clpr.tv/api/v1/clips
```

**Getting a Token:**
1. Authenticate via Twitch OAuth: `GET /api/v1/auth/twitch`
2. Use the returned JWT token in subsequent requests
3. Refresh tokens when expired: `POST /api/v1/auth/refresh`

## 📊 Rate Limiting

Rate limits vary by subscription tier:

- **Free**: 300 requests/minute
- **Premium**: 1,000 requests/minute
- **Enterprise**: Custom limits

Rate limit headers in responses:
- `X-RateLimit-Limit`: Max requests allowed
- `X-RateLimit-Remaining`: Requests remaining
- `X-RateLimit-Reset`: Reset timestamp

## 🌍 API Environments

### Development
- Base URL: `http://localhost:8080`
- Use for local development and testing

### Staging
- Base URL: `https://staging.clpr.tv`
- Pre-production environment for integration testing

### Production
- Base URL: `https://api.clpr.tv`
- Production API with full SLA and monitoring

## 📝 Specification Structure

```yaml
openapi: 3.1.0
info:                       # API metadata
servers:                    # Environment URLs
tags:                       # Endpoint categories
paths:                      # All API endpoints
  /api/v1/clips:           # Clip operations
  /api/v1/users:           # User operations
  /api/v1/auth:            # Authentication
  # ... 474+ more endpoints
components:
  securitySchemes:         # JWT Bearer auth
  schemas:                 # Reusable data models
    Clip:                  # Clip model
    User:                  # User model
    Comment:               # Comment model
    # ... 30+ more models
  parameters:              # Reusable parameters
  responses:               # Reusable responses
```

## 📚 API Documentation Generator

### Generating Markdown Documentation

Convert the OpenAPI spec to formatted Markdown with code samples:

```bash
# Using npm
npm run openapi:generate-docs

# Output: docs/openapi/generated/api-reference.md
```

The generator creates:
- Complete API reference with all endpoints
- Multi-language code samples (cURL, JavaScript, Python, Go)
- Organized by tags/categories
- Table of contents with navigation links
- Request/response documentation

### Generating Changelog

Compare OpenAPI spec versions to track API changes:

```bash
# Generate baseline (first run)
npm run openapi:changelog

# Compare two versions
node scripts/generate-api-changelog.js [old-spec.yaml] [new-spec.yaml]

# Output: docs/openapi/generated/api-changelog.md
```

The changelog includes:
- Added endpoints
- Modified endpoints
- Removed endpoints
- Deprecated endpoints
- Migration guide

### Viewing in Admin Dashboard

Access the generated API documentation through the admin dashboard:

1. Navigate to `/admin/api-docs`
2. Select version (Current, Baseline, Changelog)
3. Search endpoints by keyword
4. Filter by category/tag
5. View code samples in multiple languages

## 🔗 Related Documentation

- [Main API Documentation](../backend/api.md)
- [Clip Submission Guide](../CLIP_SUBMISSION_API_GUIDE.md)
- [Authentication Guide](../backend/authentication.md)
- [Rate Limiting](../backend/rate-limiting.md)
- [WebSocket API](../backend/websocket-api.md)

## 🐛 Troubleshooting

### Validation Errors

If you encounter validation errors:

```bash
# Check detailed error output
npx @redocly/cli lint docs/openapi/openapi.yaml --format=stylish

# Use VS Code extension for inline validation
# Install: OpenAPI (Swagger) Editor by 42Crunch
```

### Missing Endpoints

The spec documents all 474+ endpoints. If an endpoint is missing or incorrect:

1. Check `backend/cmd/api/main.go` for the source definition
2. Update `docs/openapi/openapi.yaml`
3. Validate with `make openapi-validate`
4. Submit a pull request

## 📞 Support

- **Issues**: [GitHub Issues](https://git.subcult.tv/subculture-collective/clpr/issues)
- **Discussions**: [GitHub Discussions](https://git.subcult.tv/subculture-collective/clpr/discussions)
- **Email**: support@clpr.tv

## 📜 License

This API specification is part of the Clipper project and is licensed under the [MIT License](../../LICENSE).

---

**Last Updated**: 2026-01-30  
**Version**: 1.0.0  
**Maintainer**: team-core
