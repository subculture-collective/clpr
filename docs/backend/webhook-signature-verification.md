---
title: "Webhook Signature Verification"
summary: "This guide explains how to verify webhook signatures sent by Clipper's webhook system. Signature verification is crucial for ensuring that webhook requests are authentic and haven't been tampered with"
tags: ["backend"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Webhook Signature Verification Guide

## Overview

This guide explains how to verify webhook signatures sent by Clipper's webhook system. Signature verification is crucial for ensuring that webhook requests are authentic and haven't been tampered with.

## How It Works

Clipper signs all webhook requests using HMAC-SHA256 (Hash-based Message Authentication Code with SHA-256). The signature is included in the `X-Webhook-Signature` header of each webhook request.

### Signature Generation Process

1. **Payload**: The entire JSON payload (request body) is used as the message
2. **Secret**: Your webhook subscription secret is used as the key
3. **Algorithm**: HMAC-SHA256 is applied to generate the signature
4. **Encoding**: The signature is hex-encoded and sent in the `X-Webhook-Signature` header

### Webhook Headers

Every webhook request includes these headers:

- `X-Webhook-Signature`: The HMAC-SHA256 signature (hex-encoded)
- `X-Webhook-Event`: The event type (e.g., "clip.submitted")
- `X-Webhook-Delivery-ID`: A unique UUID for this delivery attempt
- `Content-Type`: Always `application/json`
- `User-Agent`: `Clipper-Webhooks/1.0`

Optional headers:
- `X-Webhook-Replay`: `true` (only present when a webhook is being replayed from the DLQ)

## Signature Verification Examples

### Node.js / JavaScript

```javascript
const crypto = require('crypto');

/**
 * Verify the webhook signature
 * @param {string} payload - The raw request body as a string
 * @param {string} signature - The X-Webhook-Signature header value
 * @param {string} secret - Your webhook secret
 * @returns {boolean} - True if signature is valid
 */
function verifyWebhookSignature(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  hmac.update(payload);
  const expectedSignature = hmac.digest('hex');
  
  // Use timing-safe comparison to prevent timing attacks
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}

// Express.js example
const express = require('express');
const app = express();

// IMPORTANT: Use express.text() or express.raw() to get the raw body
// Do NOT use express.json() as it modifies the body
app.post('/webhook', express.text({ type: 'application/json' }), (req, res) => {
  const signature = req.headers['x-webhook-signature'];
  const event = req.headers['x-webhook-event'];
  const deliveryId = req.headers['x-webhook-delivery-id'];
  const secret = process.env.WEBHOOK_SECRET;
  
  // Verify the signature
  if (!verifyWebhookSignature(req.body, signature, secret)) {
    console.error('Invalid webhook signature');
    return res.status(401).send('Invalid signature');
  }
  
  // Parse the payload after verification
  const payload = JSON.parse(req.body);
  
  console.log(`Received webhook: ${event} (${deliveryId})`);
  console.log('Payload:', payload);
  
  // Process the webhook event
  // ... your business logic here ...
  
  res.status(200).send('OK');
});

app.listen(3000, () => {
  console.log('Webhook server listening on port 3000');
});
```

### Python

```python
import os
import hmac
import hashlib
import json
from flask import Flask, request, jsonify

app = Flask(__name__)

def verify_webhook_signature(payload: bytes, signature: str, secret: str) -> bool:
    """
    Verify the webhook signature using HMAC-SHA256
    
    Args:
        payload: The raw request body as bytes
        signature: The X-Webhook-Signature header value
        secret: Your webhook secret
        
    Returns:
        True if signature is valid, False otherwise
    """
    expected_signature = hmac.new(
        secret.encode('utf-8'),
        payload,
        hashlib.sha256
    ).hexdigest()
    
    # Use compare_digest for timing-safe comparison
    return hmac.compare_digest(signature, expected_signature)

@app.route('/webhook', methods=['POST'])
def webhook():
    # Get the raw body for signature verification
    payload = request.get_data()
    signature = request.headers.get('X-Webhook-Signature')
    event = request.headers.get('X-Webhook-Event')
    delivery_id = request.headers.get('X-Webhook-Delivery-ID')
    
    # Your webhook secret (store securely, e.g., in environment variables)
    secret = os.environ.get('WEBHOOK_SECRET')
    
    # Verify the signature
    if not verify_webhook_signature(payload, signature, secret):
        app.logger.error('Invalid webhook signature')
        return jsonify({'error': 'Invalid signature'}), 401
    
    # Parse the JSON payload after verification
    data = json.loads(payload)
    
    app.logger.info(f'Received webhook: {event} ({delivery_id})')
    app.logger.info(f'Payload: {data}')
    
    # Process the webhook event
    # ... your business logic here ...
    
    return jsonify({'status': 'success'}), 200

if __name__ == '__main__':
    app.run(port=3000)
```

### Go

```go
package main

import (
 "crypto/hmac"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "io"
 "log"
 "net/http"
 "os"
)

// verifyWebhookSignature verifies the HMAC-SHA256 signature
func verifyWebhookSignature(payload []byte, signature, secret string) bool {
 h := hmac.New(sha256.New, []byte(secret))
 h.Write(payload)
 expectedSignature := hex.EncodeToString(h.Sum(nil))
 
 // Use constant-time comparison to prevent timing attacks
 return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// WebhookPayload represents the structure of incoming webhook data
type WebhookPayload struct {
 Event     string                 `json:"event"`
 Timestamp string                 `json:"timestamp"`
 Data      map[string]interface{} `json:"data"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
 if r.Method != http.MethodPost {
  http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  return
 }
 
 // Read the raw body
 body, err := io.ReadAll(r.Body)
 if err != nil {
  log.Printf("Error reading body: %v", err)
  http.Error(w, "Error reading request body", http.StatusBadRequest)
  return
 }
 defer r.Body.Close()
 
 // Get headers
 signature := r.Header.Get("X-Webhook-Signature")
 event := r.Header.Get("X-Webhook-Event")
 deliveryID := r.Header.Get("X-Webhook-Delivery-ID")
 
 // Get secret from environment
 secret := os.Getenv("WEBHOOK_SECRET")
 
 // Verify signature
 if !verifyWebhookSignature(body, signature, secret) {
  log.Println("Invalid webhook signature")
  http.Error(w, "Invalid signature", http.StatusUnauthorized)
  return
 }
 
 // Parse payload after verification
 var payload WebhookPayload
 if err := json.Unmarshal(body, &payload); err != nil {
  log.Printf("Error parsing payload: %v", err)
  http.Error(w, "Invalid payload", http.StatusBadRequest)
  return
 }
 
 log.Printf("Received webhook: %s (%s)", event, deliveryID)
 log.Printf("Payload: %+v", payload)
 
 // Process the webhook event
 // ... your business logic here ...
 
 w.WriteHeader(http.StatusOK)
 w.Write([]byte("OK"))
}

func main() {
 http.HandleFunc("/webhook", webhookHandler)
 
 log.Println("Webhook server listening on :3000")
 if err := http.ListenAndServe(":3000", nil); err != nil {
  log.Fatal(err)
 }
}
```

### Ruby

```ruby
require 'sinatra'
require 'json'
require 'openssl'

# Verify webhook signature
def verify_webhook_signature(payload, signature, secret)
  expected_signature = OpenSSL::HMAC.hexdigest('SHA256', secret, payload)
  
  # Use secure comparison to prevent timing attacks
  Rack::Utils.secure_compare(signature, expected_signature)
end

# Webhook endpoint
post '/webhook' do
  # Read the raw body
  payload = request.body.read
  
  # Get headers
  signature = request.env['HTTP_X_WEBHOOK_SIGNATURE']
  event = request.env['HTTP_X_WEBHOOK_EVENT']
  delivery_id = request.env['HTTP_X_WEBHOOK_DELIVERY_ID']
  
  # Get secret from environment
  secret = ENV['WEBHOOK_SECRET']
  
  # Verify signature
  unless verify_webhook_signature(payload, signature, secret)
    logger.error 'Invalid webhook signature'
    halt 401, { error: 'Invalid signature' }.to_json
  end
  
  # Parse payload after verification
  data = JSON.parse(payload)
  
  logger.info "Received webhook: #{event} (#{delivery_id})"
  logger.info "Payload: #{data}"
  
  # Process the webhook event
  # ... your business logic here ...
  
  status 200
  body 'OK'
end

# Run the server
set :port, 3000
```

### PHP

```php
<?php

/**
 * Verify webhook signature using HMAC-SHA256
 * 
 * @param string $payload The raw request body
 * @param string $signature The X-Webhook-Signature header value
 * @param string $secret Your webhook secret
 * @return bool True if signature is valid
 */
function verifyWebhookSignature($payload, $signature, $secret) {
    $expectedSignature = hash_hmac('sha256', $payload, $secret);
    
    // Use timing-safe comparison
    return hash_equals($signature, $expectedSignature);
}

// Get the raw POST body
$payload = file_get_contents('php://input');

// Get headers
$signature = $_SERVER['HTTP_X_WEBHOOK_SIGNATURE'] ?? '';
$event = $_SERVER['HTTP_X_WEBHOOK_EVENT'] ?? '';
$deliveryId = $_SERVER['HTTP_X_WEBHOOK_DELIVERY_ID'] ?? '';

// Get secret from environment
$secret = getenv('WEBHOOK_SECRET');

// Verify signature
if (!verifyWebhookSignature($payload, $signature, $secret)) {
    error_log('Invalid webhook signature');
    http_response_code(401);
    echo json_encode(['error' => 'Invalid signature']);
    exit;
}

// Parse payload after verification
$data = json_decode($payload, true);

error_log("Received webhook: $event ($deliveryId)");
error_log("Payload: " . print_r($data, true));

// Process the webhook event
// ... your business logic here ...

http_response_code(200);
echo 'OK';
```

### Java

```java
import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.security.InvalidKeyException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.nio.charset.StandardCharsets;

public class WebhookVerifier {
    
    /**
     * Verify webhook signature using HMAC-SHA256
     * 
     * @param payload The raw request body
     * @param signature The X-Webhook-Signature header value
     * @param secret Your webhook secret
     * @return true if signature is valid
     */
    public static boolean verifyWebhookSignature(String payload, String signature, String secret) {
        try {
            Mac hmac = Mac.getInstance("HmacSHA256");
            SecretKeySpec secretKey = new SecretKeySpec(secret.getBytes(StandardCharsets.UTF_8), "HmacSHA256");
            hmac.init(secretKey);
            
            byte[] hash = hmac.doFinal(payload.getBytes(StandardCharsets.UTF_8));
            String expectedSignature = bytesToHex(hash);
            
            // Use constant-time comparison
            return MessageDigest.isEqual(
                signature.getBytes(StandardCharsets.UTF_8),
                expectedSignature.getBytes(StandardCharsets.UTF_8)
            );
        } catch (NoSuchAlgorithmException | InvalidKeyException e) {
            throw new RuntimeException("Error verifying webhook signature", e);
        }
    }
    
    /**
     * Convert byte array to hex string
     */
    private static String bytesToHex(byte[] bytes) {
        StringBuilder result = new StringBuilder();
        for (byte b : bytes) {
            result.append(String.format("%02x", b));
        }
        return result.toString();
    }
}

// Spring Boot example
import org.springframework.web.bind.annotation.*;
import org.springframework.http.ResponseEntity;
import org.springframework.http.HttpStatus;

@RestController
public class WebhookController {
    
    @PostMapping("/webhook")
    public ResponseEntity<String> handleWebhook(
        @RequestBody String payload,
        @RequestHeader("X-Webhook-Signature") String signature,
        @RequestHeader("X-Webhook-Event") String event,
        @RequestHeader("X-Webhook-Delivery-ID") String deliveryId
    ) {
        String secret = System.getenv("WEBHOOK_SECRET");
        
        // Verify signature
        if (!WebhookVerifier.verifyWebhookSignature(payload, signature, secret)) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body("Invalid signature");
        }
        
        // Parse and process the webhook
        // ... your business logic here ...
        
        return ResponseEntity.ok("OK");
    }
}
```

### C# / .NET

```csharp
using System;
using System.Security.Cryptography;
using System.Text;
using Microsoft.AspNetCore.Mvc;

public class WebhookVerifier
{
    /// <summary>
    /// Verify webhook signature using HMAC-SHA256
    /// </summary>
    /// <param name="payload">The raw request body</param>
    /// <param name="signature">The X-Webhook-Signature header value</param>
    /// <param name="secret">Your webhook secret</param>
    /// <returns>True if signature is valid</returns>
    public static bool VerifyWebhookSignature(string payload, string signature, string secret)
    {
        using (var hmac = new HMACSHA256(Encoding.UTF8.GetBytes(secret)))
        {
            var hash = hmac.ComputeHash(Encoding.UTF8.GetBytes(payload));
            var expectedSignature = BitConverter.ToString(hash).Replace("-", "").ToLower();
            
            // Use constant-time comparison
            return CryptographicOperations.FixedTimeEquals(
                Encoding.UTF8.GetBytes(signature),
                Encoding.UTF8.GetBytes(expectedSignature)
            );
        }
    }
}

// ASP.NET Core example
[ApiController]
[Route("webhook")]
public class WebhookController : ControllerBase
{
    [HttpPost]
    public async Task<IActionResult> HandleWebhook()
    {
        // Read the raw body
        using (var reader = new StreamReader(Request.Body, Encoding.UTF8))
        {
            var payload = await reader.ReadToEndAsync();
            
            // Get headers
            var signature = Request.Headers["X-Webhook-Signature"].ToString();
            var eventType = Request.Headers["X-Webhook-Event"].ToString();
            var deliveryId = Request.Headers["X-Webhook-Delivery-ID"].ToString();
            
            // Get secret from configuration
            var secret = Environment.GetEnvironmentVariable("WEBHOOK_SECRET");
            
            // Verify signature
            if (!WebhookVerifier.VerifyWebhookSignature(payload, signature, secret))
            {
                return Unauthorized(new { error = "Invalid signature" });
            }
            
            // Parse and process the webhook
            // ... your business logic here ...
            
            return Ok("OK");
        }
    }
}
```

## Sample Webhook Payloads

### clip.submitted Event

```json
{
  "event": "clip.submitted",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "submission_id": "123e4567-e89b-12d3-a456-426614174000",
    "clip_id": "987fcdeb-51a2-43e7-9876-123456789abc",
    "user_id": "456e7890-e12f-34g5-h678-901234567def",
    "title": "Amazing Play",
    "description": "Check out this incredible moment!",
    "game": "Valorant",
    "submitted_at": "2024-01-15T10:30:00Z"
  }
}
```

### clip.approved Event

```json
{
  "event": "clip.approved",
  "timestamp": "2024-01-15T11:00:00Z",
  "data": {
    "clip_id": "987fcdeb-51a2-43e7-9876-123456789abc",
    "user_id": "456e7890-e12f-34g5-h678-901234567def",
    "approved_by": "moderator-id",
    "approved_at": "2024-01-15T11:00:00Z"
  }
}
```

### clip.rejected Event

```json
{
  "event": "clip.rejected",
  "timestamp": "2024-01-15T11:00:00Z",
  "data": {
    "submission_id": "123e4567-e89b-12d3-a456-426614174000",
    "clip_id": "987fcdeb-51a2-43e7-9876-123456789abc",
    "user_id": "456e7890-e12f-34g5-h678-901234567def",
    "rejected_by": "moderator-id",
    "rejected_at": "2024-01-15T11:00:00Z",
    "reason": "Does not meet content guidelines"
  }
}
```

## Testing Your Webhook Integration

### Generating Test Signatures

You can generate test signatures using the following approach:

```bash
# Using OpenSSL command line
echo -n '{"event":"clip.submitted","timestamp":"2024-01-15T10:30:00Z","data":{}}' | \
  openssl dgst -sha256 -hmac "your-webhook-secret" | \
  awk '{print $2}'
```

### Example Test Data

**Secret:** `test-secret-key-12345`

**Payload:**
```json
{"event":"clip.submitted","timestamp":"2024-01-15T10:30:00Z","data":{"submission_id":"123e4567-e89b-12d3-a456-426614174000"}}
```

**Expected Signature:** `eb09d13b20c12e7e8e12f24eb9bc4803e3eb6faadd641796ca5503f25cb32a69`

You can use this test data to verify your signature verification implementation is working correctly.

## Security Best Practices

### 1. Always Verify Signatures

Never process webhook requests without verifying the signature first. This ensures the request is authentic.

```javascript
// ❌ BAD: Processing before verification
const payload = JSON.parse(req.body);
processWebhook(payload);
if (!verifySignature(...)) { /* too late */ }

// ✅ GOOD: Verify first, then process
if (!verifySignature(req.body, signature, secret)) {
  return res.status(401).send('Invalid signature');
}
const payload = JSON.parse(req.body);
processWebhook(payload);
```

### 2. Use the Raw Request Body

The signature is computed on the raw request body. If you parse the body before verification, you may introduce changes that invalidate the signature.

```javascript
// ❌ BAD: Using express.json() middleware
app.use(express.json());
app.post('/webhook', (req, res) => {
  // req.body is already parsed - signature will fail!
});

// ✅ GOOD: Use raw body middleware
app.post('/webhook', express.text({ type: 'application/json' }), (req, res) => {
  // req.body is the raw string
});
```

### 3. Use Timing-Safe Comparison

Always use constant-time comparison functions to prevent timing attacks:

- Node.js: `crypto.timingSafeEqual()`
- Python: `hmac.compare_digest()`
- Go: `hmac.Equal()`
- Ruby: `Rack::Utils.secure_compare()`
- PHP: `hash_equals()`
- Java: `MessageDigest.isEqual()`
- C#: `CryptographicOperations.FixedTimeEquals()`

### 4. Store Secrets Securely

Never hardcode webhook secrets in your code. Use environment variables or secure secret management systems:

```javascript
// ❌ BAD
const secret = 'my-webhook-secret';

// ✅ GOOD
const secret = process.env.WEBHOOK_SECRET;
```

### 5. Use HTTPS

Always use HTTPS endpoints for your webhooks to prevent man-in-the-middle attacks:

```javascript
// ❌ BAD
const webhookUrl = 'http://example.com/webhook';

// ✅ GOOD
const webhookUrl = 'https://example.com/webhook';
```

### 6. Implement Idempotency

Use the `X-Webhook-Delivery-ID` header to track processed webhooks and prevent duplicate processing:

```javascript
const deliveryId = req.headers['x-webhook-delivery-id'];

// Check if this delivery has already been processed
if (await isProcessed(deliveryId)) {
  return res.status(200).send('Already processed');
}

// Process the webhook
await processWebhook(payload);

// Mark as processed
await markAsProcessed(deliveryId);
```

### 7. Respond Quickly

Respond with a 2xx status code within 10 seconds. Process complex logic asynchronously:

```javascript
app.post('/webhook', async (req, res) => {
  // Verify signature
  if (!verifySignature(...)) {
    return res.status(401).send('Invalid signature');
  }
  
  // Respond immediately
  res.status(200).send('OK');
  
  // Process asynchronously (don't await)
  processWebhookAsync(JSON.parse(req.body)).catch(err => {
    console.error('Error processing webhook:', err);
  });
});
```

### 8. Handle Errors Gracefully

Log errors but don't expose sensitive information in error responses:

```javascript
try {
  processWebhook(payload);
} catch (error) {
  // Log detailed error internally
  console.error('Webhook processing error:', error);
  
  // Return generic error to client
  res.status(500).send('Internal server error');
}
```

### 9. Monitor and Alert

Set up monitoring for:
- Failed signature verifications (possible attack or misconfiguration)
- Webhook processing errors
- Slow response times
- High error rates

### 10. Rotate Secrets Regularly

Periodically rotate your webhook secrets for better security. The API provides a secret rotation endpoint:

```bash
curl -X POST https://api.clpr.example/api/v1/webhooks/{id}/regenerate-secret \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Troubleshooting

### Signature Verification Fails

**Issue:** Your signature verification always fails.

**Solutions:**

1. **Check you're using the raw body:**
   - Don't parse the body before verification
   - Ensure no middleware modifies the body
   - Use the exact bytes received

2. **Verify the secret:**
   - Ensure you're using the correct secret
   - Check for whitespace or encoding issues
   - Verify the secret hasn't been rotated

3. **Check the signature format:**
   - The signature should be hex-encoded
   - It should be 64 characters long (SHA-256)
   - It should be lowercase

4. **Test with known values:**
   - Use the test data provided above
   - Generate a signature manually and compare

### Common Mistakes

```javascript
// ❌ MISTAKE 1: Parsing body before verification
app.use(express.json());
app.post('/webhook', (req, res) => {
  const signature = verifySignature(JSON.stringify(req.body), ...);
  // This will fail because JSON.stringify may not match the original
});

// ❌ MISTAKE 2: Using wrong secret
const signature = verifySignature(payload, signature, 'wrong-secret');

// ❌ MISTAKE 3: Comparing strings with ===
if (signature === expectedSignature) { // Vulnerable to timing attacks

// ❌ MISTAKE 4: Not reading the full body
const payload = req.body.slice(0, 100); // Don't truncate!

// ✅ CORRECT: Use raw body and proper comparison
app.post('/webhook', express.text({ type: 'application/json' }), (req, res) => {
  if (!crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  )) {
    return res.status(401).send('Invalid signature');
  }
});
```

## Additional Resources

- [Main Webhook Documentation](./WEBHOOK_SUBSCRIPTION_MANAGEMENT.md)
- [HMAC-SHA256 Specification (RFC 2104)](https://tools.ietf.org/html/rfc2104)
- [Webhook Security Best Practices](https://webhooks.fyi/security/overview)

## Support

If you encounter issues with webhook signature verification:

1. Check the [Troubleshooting](#troubleshooting) section above
2. Review the delivery history in your webhook subscription settings
3. Check the webhook logs for error messages
4. Contact support with the `X-Webhook-Delivery-ID` for specific failed deliveries
