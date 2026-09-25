---
title: "API Documentation Generator"
summary: "Automated generator for converting OpenAPI specs to Markdown with admin integration"
tags: ["api", "documentation", "openapi", "generator"]
area: "openapi"
status: "stable"
owner: "team-core"
version: "1.0.0"
last_reviewed: 2026-01-30
---

# API Documentation Generator & Admin Integration

## Overview

This system provides automated generation of API documentation from OpenAPI specifications, with integration into the admin dashboard for easy access and search.

## Features

- ✅ **OpenAPI to Markdown conversion** with multi-language code samples
- ✅ **Multi-language code examples** (cURL, JavaScript, Python, Go)
- ✅ **Changelog generation** from spec diffs
- ✅ **Admin dashboard integration** with search and filtering
- ✅ **Version management** for API documentation
- ✅ **Try-it-out documentation** for testing endpoints

## Architecture

### Components

1. **Generator Script** (`scripts/generate-api-docs.js`)
   - Parses OpenAPI YAML specification
   - Generates formatted Markdown documentation
   - Creates code samples in multiple languages
   - Organizes content by tags/categories

2. **Changelog Generator** (`scripts/generate-api-changelog.js`)
   - Compares two OpenAPI spec versions
   - Identifies added, modified, removed, and deprecated endpoints
   - Generates migration guide
   - Creates baseline documentation

3. **Admin Dashboard Component** (`frontend/src/pages/admin/AdminAPIDocsPage.tsx`)
   - Displays generated API documentation
   - Provides version selection
   - Implements search and filtering
   - Renders markdown with syntax highlighting

### Data Flow

```
OpenAPI Spec (openapi.yaml)
        ↓
Generator Script
        ↓
Markdown Files (docs/openapi/generated/)
        ↓
Backend API (/api/v1/docs/)
        ↓
Admin Dashboard UI
```

## Usage

### Generating API Documentation

```bash
# Generate complete API reference
npm run openapi:generate-docs

# Output location
# docs/openapi/generated/api-reference.md
```

The generator will:
1. Load `docs/openapi/openapi.yaml`
2. Parse all endpoints, schemas, and tags
3. Generate formatted Markdown with:
   - API metadata (version, description, servers)
   - Table of contents by category
   - Detailed endpoint documentation
   - Request/response examples
   - Multi-language code samples

### Generating Changelog

```bash
# Generate baseline (first run)
npm run openapi:changelog

# Compare two versions
node scripts/generate-api-changelog.js old-spec.yaml new-spec.yaml
```

The changelog includes:
- **Summary table** with change counts
- **Added endpoints** grouped by tag
- **Modified endpoints** with before/after comparison
- **Deprecated endpoints** with warnings
- **Removed endpoints** grouped by tag
- **Migration guide** for breaking changes

### Accessing in Admin Dashboard

1. **Navigate to Admin Dashboard**
   ```
   https://yourdomain.com/admin/dashboard
   ```

2. **Click "API Reference"** in Quick Documentation section

3. **Select Version**
   - v1.0.0 (Current) - Latest API reference
   - Baseline - Initial API snapshot
   - Changelog - API changes and diffs

4. **Search and Filter**
   - Search by endpoint, method, or description
   - Filter by category/tag
   - View code samples

## Code Sample Generation

The generator creates code samples in multiple languages for each endpoint:

### cURL

```bash
curl -X GET "https://api.clpr.tv/api/v1/clips" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript/TypeScript

```javascript
const response = await fetch('/api/v1/clips', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN'
  }
});
const data = await response.json();
```

### Python

```python
import requests

headers = {'Authorization': 'Bearer YOUR_TOKEN'}
response = requests.get('/api/v1/clips', headers=headers)
data = response.json()
```

### Go

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

## Customization

### Adding New Languages

Edit `scripts/generate-api-docs.js` and add a new function:

```javascript
function generateRubySample(path, method, operation) {
    // Ruby code generation logic
    return rubyCode;
}
```

Then call it in `generateEndpointMarkdown()`:

```javascript
// Ruby
md += `##### Ruby\n\n`;
md += `\`\`\`ruby\n${generateRubySample(path, method, operation)}\n\`\`\`\n\n`;
```

### Customizing Markdown Output

The generator uses frontmatter for metadata:

```yaml
---
title: "API Reference"
summary: "Complete API reference for Clipper platform"
tags: ["api", "reference", "openapi"]
area: "openapi"
status: "stable"
version: "1.0.0"
generated: 2026-01-30T03:01:39.309Z
---
```

Modify `generateAPIReference()` function to change the structure.

## Admin Dashboard Features

### Version Selection

- **v1.0.0 (Current)**: Latest API reference
- **Baseline**: Initial snapshot for comparison
- **Changelog**: Diff between versions

### Search Functionality

- Real-time search across all endpoints
- Searches in:
  - Endpoint paths
  - Method names
  - Descriptions
  - Parameter names
  - Response descriptions

### Category Filtering

- Filter endpoints by tag/category
- Quick navigation to specific sections
- Categories include:
  - Health, Authentication, Clips, Users
  - Comments, Tags, Search, Submissions
  - And 28+ more categories

### Try-It-Out Information

The page includes documentation for testing endpoints:
- Swagger UI local server
- Postman collection import
- Insomnia collection import
- Direct cURL usage

## Maintenance

### Updating API Documentation

When OpenAPI spec changes:

1. Update `docs/openapi/openapi.yaml`
2. Validate the spec:
   ```bash
   npm run openapi:validate
   ```
3. Regenerate documentation:
   ```bash
   npm run openapi:generate-docs
   ```
4. Generate changelog:
   ```bash
   npm run openapi:changelog
   ```

### Version Management

To create a new API version:

1. Copy current spec:
   ```bash
   cp docs/openapi/openapi.yaml docs/openapi/openapi-v1.0.0.yaml
   ```

2. Generate changelog:
   ```bash
   node scripts/generate-api-changelog.js \
     docs/openapi/openapi-v1.0.0.yaml \
     docs/openapi/openapi.yaml
   ```

3. Update version array in `AdminAPIDocsPage.tsx`:
   ```typescript
   const API_VERSIONS: APIVersion[] = [
       { version: '1.1.0', label: 'v1.1.0 (Current)', file: 'api-reference.md' },
       { version: '1.0.0', label: 'v1.0.0', file: 'api-reference-v1.0.0.md' },
       // ...
   ];
   ```

## CI/CD Integration

Add to your CI pipeline:

```yaml
# .github/workflows/api-docs.yml
name: API Documentation

on:
  push:
    branches: [main, develop]
    paths:
      - 'docs/openapi/openapi.yaml'

jobs:
  generate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: npm install
      - run: npm run openapi:validate
      - run: npm run openapi:generate-docs
      - uses: stefanzweifel/git-auto-commit-action@v4
        with:
          commit_message: 'chore: regenerate API documentation'
          file_pattern: 'docs/openapi/generated/*.md'
```

## Troubleshooting

### Generator Errors

**Error: Cannot find module 'js-yaml'**
```bash
npm install js-yaml --save-dev
```

**Error: Failed to load OpenAPI spec**
- Verify `docs/openapi/openapi.yaml` exists
- Check YAML syntax with `npm run openapi:validate`

### Admin Dashboard Issues

**Error: Failed to load API documentation**
- Run generator first: `npm run openapi:generate-docs`
- Check that files exist in `docs/openapi/generated/`
- Verify backend serves static files from docs directory

**Search not working**
- Ensure markdown is processed correctly
- Check console for JavaScript errors
- Verify `parseMarkdown` utility is imported

### Performance Issues

**Large documentation file**
- Consider splitting by category
- Implement lazy loading in UI
- Use pagination for endpoints

## Best Practices

1. **Keep OpenAPI spec up-to-date**
   - Document all endpoints
   - Add examples and descriptions
   - Use meaningful operation IDs

2. **Regenerate after spec changes**
   - Always run generator after updates
   - Generate changelog for version tracking
   - Review generated output

3. **Maintain version history**
   - Keep archived spec versions
   - Generate changelogs between versions
   - Document breaking changes

4. **Test code samples**
   - Verify generated code is valid
   - Test with different auth scenarios
   - Include error handling examples

## Related Links

- [OpenAPI Specification](https://spec.openapis.org/oas/latest.html)
- [Clipper OpenAPI Spec](./openapi.yaml)
- [Admin Dashboard](../../frontend/src/pages/admin/AdminAPIDocsPage.tsx)
- [Generator Script](../../scripts/generate-api-docs.js)
- [Changelog Script](../../scripts/generate-api-changelog.js)

## Support

For issues or questions:
- Create an issue on GitHub
- Contact the team-core team
- Review the [OpenAPI README](./README.md)

---

**Last Updated**: 2026-01-30  
**Version**: 1.0.0  
**Maintainer**: team-core
