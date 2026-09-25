---
title: "OpenAPI Specifications"
summary: "OpenAPI 3.0 specifications for the Clipper REST API including clip submission, comments, and core endpoints."
tags: ["openapi", "hub", "index", "api", "swagger"]
area: "openapi"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
aliases: ["openapi hub", "swagger", "api specs"]
---

# OpenAPI Specifications

OpenAPI 3.0 specifications for the Clipper REST API.

## Available Specifications

### Clip Submission API

- **File**: [[clip-submission-api.yaml]]
- **Description**: API for submitting Twitch clips to the platform
- **Version**: 1.0.0
- **Endpoints**:
  - `GET /api/v1/submissions/metadata` - Fetch clip metadata from Twitch
  - `POST /api/v1/submissions` - Submit a clip for review
  - `GET /api/v1/submissions` - List user's submissions
  - `GET /api/v1/submissions/stats` - Get submission statistics

### Comments API

- **File**: [[comments-api.yaml]]
- **Description**: API for managing comments on clips with nested threading support
- **Version**: 1.0.0
- **Endpoints**:
  - `GET /api/v1/clips/{clipId}/comments` - List comments for a clip
  - `POST /api/v1/clips/{clipId}/comments` - Create a comment
  - `GET /api/v1/comments/{commentId}/replies` - Get replies to a comment
- **Features**:
  - Flat list or nested tree structure (up to 10 levels deep)
  - Vote scores and user vote status
  - Reply counts and markdown rendering

## Documentation Index

```dataview
TABLE title, summary, status
FROM "docs/openapi"
WHERE file.name != "index" AND file.name != "README"
SORT title ASC
```

## Viewing the Specifications

### Online Viewers

1. **Swagger Editor**: <https://editor.swagger.io/>
   - Copy and paste the YAML content into the editor

2. **Redoc**: Generate a beautiful HTML documentation
   ```bash
   npx @redocly/cli build-docs docs/openapi/clip-submission-api.yaml
   ```

3. **Swagger UI**: Run a local Swagger UI server
   ```bash
   docker run -p 8081:8080 \
     -e SWAGGER_JSON=/api/clip-submission-api.yaml \
     -v $(pwd)/docs/openapi:/api \
     swaggerapi/swagger-ui
   ```
   Then visit <http://localhost:8081>

### VS Code Extension

Install the [OpenAPI (Swagger) Editor](https://marketplace.visualstudio.com/items?itemName=42Crunch.vscode-openapi) extension for:
- Syntax highlighting and validation
- Auto-completion
- Live preview

## Generating Client SDKs

Generate client libraries using [OpenAPI Generator](https://openapi-generator.tech/):

### TypeScript/JavaScript

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi/clip-submission-api.yaml \
  -g typescript-axios \
  -o generated/typescript-client
```

### Python

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi/clip-submission-api.yaml \
  -g python \
  -o generated/python-client
```

### Go

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi/clip-submission-api.yaml \
  -g go \
  -o generated/go-client
```

## Validating Specifications

```bash
# Using Redocly CLI
npx @redocly/cli lint docs/openapi/clip-submission-api.yaml

# Using Swagger CLI
npx swagger-cli validate docs/openapi/clip-submission-api.yaml
```

## Contributing

When adding or modifying API endpoints:

1. Update the corresponding OpenAPI specification
2. Validate the specification
3. Update the developer guide with examples
4. Test all endpoints
5. Submit a pull request

## Related Documentation

- [[../backend/clip-submission-api-guide|Clip Submission API Guide]] - Complete developer guide
- [[../backend/api|API Reference]] - Main API documentation
- [[../examples/index|Code Examples]] - Sample code and scripts
- [[../backend/authentication|Authentication]] - OAuth and JWT setup

---

**See also:**
[[../backend/api|API Reference]] ·
[[../examples/index|Code Examples]] ·
[[../index|Documentation Home]]
