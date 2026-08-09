---
title: "README"
summary: "This directory contains practical examples for using the Clip Submission API."
tags: ["examples"]
area: "examples"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Clip Submission API Examples

This directory contains practical examples for using the Clip Submission API.

## Available Examples

### 1. Shell Script Example (`clip-submission-test.sh`)

A complete bash script demonstrating the full submission workflow using cURL.

**Prerequisites:**
- `curl` command-line tool
- `jq` for JSON parsing (install with `brew install jq` or `apt-get install jq`)
- Valid JWT token from Clipper authentication

**Usage:**
```bash
# Set your JWT token
export CLIPPER_TOKEN="your_jwt_token_here"

# Run the test script
chmod +x clip-submission-test.sh
./clip-submission-test.sh https://clips.twitch.tv/YourClipID

# Use custom API endpoint (optional)
API_BASE_URL="https://api.clpr.tv/v1" ./clip-submission-test.sh
```

**What it does:**
1. Fetches clip metadata from Twitch
2. Submits the clip to Clipper
3. Retrieves submission statistics

### 2. TypeScript/JavaScript Example (`clip-submission-example.ts`)

A comprehensive TypeScript SDK implementation with type safety and error handling.

**Prerequisites:**
- Node.js 18+ or TypeScript-enabled project
- `axios` package

**Installation:**
```bash
npm install axios
# or
yarn add axios
```

**Usage:**

As a standalone script:
```bash
# Compile TypeScript
npx tsc clip-submission-example.ts

# Run
CLIPPER_TOKEN="your_token" node clip-submission-example.js
```

As a library in your project:
```typescript
import { ClipSubmissionClient } from './clip-submission-example';

const client = new ClipSubmissionClient(
  'http://localhost:8080/api/v1',
  'your_jwt_token'
);

// Submit a clip
const result = await client.submitClipWorkflow(clipUrl, {
  tags: ['epic', 'valorant'],
  customTitle: 'Amazing Play'
});

console.log(result);
```

**Features:**
- Type-safe API client
- Comprehensive error handling
- Complete workflow example
- Retry logic for transient errors
- User-friendly error messages

### 3. React Hook Example

See the [clip submission implementation guide](../CLIP_SUBMISSION_IMPLEMENTATION_SUMMARY.md#frontend-already-complete) for the frontend flow.

## Running the Examples

### Quick Test with cURL

```bash
# 1. Fetch metadata
curl "http://localhost:8080/api/v1/submissions/metadata?url=CLIP_URL" \
  -H "Authorization: Bearer TOKEN"

# 2. Submit clip
curl -X POST "http://localhost:8080/api/v1/submissions" \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"clip_url":"CLIP_URL","tags":["epic"]}'

# 3. Get stats
curl "http://localhost:8080/api/v1/submissions/stats" \
  -H "Authorization: Bearer TOKEN"
```

### Using the TypeScript SDK

```typescript
// Basic submission
const { submission } = await client.submitClip({
  clip_url: 'https://clips.twitch.tv/ClipID'
});

// With all options
const { submission } = await client.submitClip({
  clip_url: 'https://clips.twitch.tv/ClipID',
  custom_title: 'Epic Moment',
  tags: ['clutch', 'epic'],
  is_nsfw: false,
  submission_reason: 'Great gameplay'
});

// Complete workflow
const result = await client.submitClipWorkflow(clipUrl, {
  customTitle: 'My Title',
  tags: ['tag1', 'tag2']
});
```

## Error Handling Examples

### TypeScript Error Handling

```typescript
try {
  await client.submitClip({ clip_url: url });
} catch (error) {
  if (error instanceof Error) {
    if (error.message.includes('duplicate')) {
      console.log('Clip already submitted');
    } else if (error.message.includes('karma')) {
      console.log('Need more karma');
    } else if (error.message.includes('rate limit')) {
      console.log('Too many submissions, try again later');
    } else {
      console.error('Error:', error.message);
    }
  }
}
```

### Shell Script Error Handling

```bash
if ! RESPONSE=$(curl -s -X POST "$API_URL" -d "$DATA"); then
  echo "Request failed"
  exit 1
fi

if ! echo "$RESPONSE" | jq -e '.success' > /dev/null; then
  ERROR=$(echo "$RESPONSE" | jq -r '.error')
  echo "Error: $ERROR"
  exit 1
fi
```

## Testing Tips

### 1. Test with Invalid Inputs

```bash
# Invalid URL
./clip-submission-test.sh "not-a-valid-url"

# Non-existent clip
./clip-submission-test.sh "https://clips.twitch.tv/InvalidClipID123"
```

### 2. Test Rate Limits

```bash
# Submit multiple clips quickly to trigger rate limit
for i in {1..6}; do
  ./clip-submission-test.sh "$CLIP_URL"
  sleep 1
done
```

### 3. Test Different Clip Formats

```bash
# Standard format
./clip-submission-test.sh "https://clips.twitch.tv/ClipID"

# Streamer path format
./clip-submission-test.sh "https://www.twitch.tv/streamer/clip/ClipID"

# Mobile format
./clip-submission-test.sh "https://m.twitch.tv/streamer/clip/ClipID"

# Direct clip ID
./clip-submission-test.sh "ClipID"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CLIPPER_TOKEN` | JWT authentication token | Required |
| `API_BASE_URL` | Base URL for API | `http://localhost:8080/api/v1` |

## Troubleshooting

### "jq: command not found"

Install jq:
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Windows (via Chocolatey)
choco install jq
```

### "Unauthorized" Error

Make sure your JWT token is valid and not expired. Re-authenticate if necessary.

### "Clip not found on Twitch"

Verify the clip exists by visiting it directly on Twitch. Some clips may be deleted or from suspended channels.

### TypeScript Compilation Errors

Ensure you have the correct type definitions:
```bash
npm install --save-dev @types/node axios
```

## Additional Resources

- [Clip submission implementation guide](../CLIP_SUBMISSION_IMPLEMENTATION_SUMMARY.md) - Supported submission behavior and tests
- Quick Reference - Quick reference card
- OpenAPI Specification - Formal API specification
- API Reference - Full API documentation

## Contributing

When adding new examples:

1. Follow the existing code style
2. Include comprehensive error handling
3. Add inline comments explaining key steps
4. Test thoroughly with different scenarios
5. Update this README with usage instructions

## Support

If you encounter issues with these examples:

1. Check the [testing troubleshooting guide](../testing/TESTING.md#troubleshooting)
2. Review the FAQ
3. Open an issue on GitHub with the `api` and `documentation` labels
