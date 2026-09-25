---
title: "MFA Admin Guide"
summary: "Multi-Factor Authentication (MFA) is **required** for all admin and moderator accounts to protect against credential compromise and unauthorized access. This document explains how to set up and manage"
tags: ["operations","guide"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Multi-Factor Authentication (MFA) for Admin Accounts

## Overview

Multi-Factor Authentication (MFA) is **required** for all admin and moderator accounts to protect against credential compromise and unauthorized access. This document explains how to set up and manage MFA.

## Why MFA is Required

Admin and moderator accounts have elevated privileges and are high-value targets for attackers. MFA provides an additional layer of security beyond just passwords by requiring:

1. **Something you know** (password)
2. **Something you have** (authenticator app)

This significantly reduces the risk of:
- Account takeover
- Unauthorized access to sensitive data
- System compromise
- Malicious actions performed by attackers

## MFA Enforcement Policy

### Who Needs MFA?

- **Admin accounts**: Required
- **Moderator accounts**: Required
- **Regular users**: Optional (but recommended)

### Grace Period

When you are promoted to admin or moderator:
- You have **7 days** to set up MFA
- During this grace period, you can access admin features with a warning
- After 7 days, admin actions will be **blocked** until MFA is enabled

## Setting Up MFA

### Step 1: Install an Authenticator App

Choose one of these TOTP-compatible authenticator apps:
- **Google Authenticator** (iOS, Android)
- **Authy** (iOS, Android, Desktop)
- **1Password** (iOS, Android, Desktop) - includes password manager
- **Microsoft Authenticator** (iOS, Android)

### Step 2: Initiate MFA Enrollment

1. Log in to your admin account
2. Navigate to **Settings** → **Security**
3. Click **Enable Multi-Factor Authentication**

### Step 3: Scan QR Code

1. The system will display a QR code
2. Open your authenticator app
3. Tap **Add Account** or **Scan QR Code**
4. Scan the QR code with your camera

**Alternative**: If you can't scan the QR code, manually enter the secret key shown below the QR code.

### Step 4: Save Backup Codes

⚠️ **IMPORTANT**: You will be shown 10 backup codes. These are single-use codes that allow you to access your account if you lose your authenticator device.

**Save these codes securely:**
- Print them and store in a safe place
- Save to a password manager
- Store in a secure document

**Never share these codes with anyone.**

### Step 5: Verify Setup

1. Enter the 6-digit code from your authenticator app
2. Click **Verify and Enable MFA**
3. MFA is now active on your account

## Using MFA to Login

1. Enter your username and password as usual
2. You'll be prompted for an MFA code
3. Open your authenticator app
4. Enter the 6-digit code shown for Clipper
5. (Optional) Check "Remember this device" to skip MFA for 30 days on trusted devices

## Managing MFA

### Viewing MFA Status

Check your MFA status at any time:
```
GET /api/v1/auth/mfa/status
```

Response includes:
- `enabled`: Whether MFA is active
- `required`: Whether MFA is mandatory for your role
- `grace_period_end`: When grace period expires (if applicable)
- `in_grace_period`: Whether you're still in grace period
- `backup_codes_remaining`: Number of unused backup codes
- `trusted_devices_count`: Number of remembered devices

### Regenerating Backup Codes

If you've used some backup codes or want to refresh them:

```
POST /api/v1/auth/mfa/regenerate-backup-codes
{
  "code": "123456"
}
```

You'll receive 10 new backup codes. **Old backup codes will be invalidated.**

### Managing Trusted Devices

View your trusted devices:
```
GET /api/v1/auth/mfa/trusted-devices
```

Revoke a trusted device:
```
DELETE /api/v1/auth/mfa/trusted-devices/:id
```

### Disabling MFA

⚠️ **Warning**: Disabling MFA is not recommended for admin/moderator accounts and will only work if it's no longer required by policy.

```
POST /api/v1/auth/mfa/disable
{
  "code": "123456"
}
```

## Using Backup Codes

If you lose access to your authenticator app:

1. On the MFA login screen, click **Use backup code**
2. Enter one of your saved backup codes
3. The code will be consumed and cannot be reused
4. **Important**: Set up a new authenticator app immediately

## Troubleshooting

### "Code is invalid" Error

**Causes**:
- Clock on your device is out of sync
- Wrong code entered
- Code expired (codes refresh every 30 seconds)

**Solutions**:
1. Wait for a new code to appear
2. Check your device's time settings (enable automatic time)
3. Ensure you're using the correct account in your authenticator

### Lost Authenticator Device

**If you have backup codes**:
1. Use a backup code to login
2. Go to Settings → Security
3. Disable and re-enable MFA with a new device

**If you don't have backup codes**:
1. Contact a system administrator
2. Provide proof of identity
3. Admin can temporarily disable MFA for recovery
4. Set up MFA immediately after regaining access

### Grace Period Expired

If your grace period has expired and you haven't set up MFA:
1. You'll be blocked from admin actions
2. Set up MFA immediately using the steps above
3. Access will be restored once MFA is enabled

## API Endpoints Reference

### Enrollment

- `POST /api/v1/auth/mfa/enroll` - Start MFA enrollment
- `POST /api/v1/auth/mfa/verify-enrollment` - Complete enrollment

### Login

- `POST /api/v1/auth/mfa/verify-login` - Verify MFA code during login

### Management

- `GET /api/v1/auth/mfa/status` - Get MFA status
- `POST /api/v1/auth/mfa/regenerate-backup-codes` - Get new backup codes
- `POST /api/v1/auth/mfa/disable` - Disable MFA
- `GET /api/v1/auth/mfa/trusted-devices` - List trusted devices
- `DELETE /api/v1/auth/mfa/trusted-devices/:id` - Revoke trusted device

## Security Best Practices

1. **Never share your MFA codes** with anyone
2. **Don't screenshot your QR code** during setup
3. **Store backup codes securely** offline
4. **Don't use SMS-based 2FA** - use app-based TOTP
5. **Review trusted devices regularly** and revoke unused ones
6. **Set up MFA on a secure device** in a private location
7. **Keep your authenticator app updated**

## Audit Logging

All MFA-related actions are logged for security audit purposes:
- MFA enrollment
- Login attempts (success/failure)
- Backup code usage
- Trusted device management
- MFA disable/enable

These logs are available to administrators for security reviews.

## Support

If you need assistance with MFA setup:
- Contact system administrators via the admin panel
- Create a support ticket in the system
- For urgent security issues, contact the security team

⚠️ **Never share your MFA codes or backup codes in support requests**

## Technical Details

### TOTP Configuration

- **Algorithm**: SHA-1 (RFC 6238 compliant)
- **Period**: 30 seconds
- **Digits**: 6
- **Time skew tolerance**: ±30 seconds

### Backup Codes

- **Count**: 10 codes per generation
- **Length**: 8 alphanumeric characters
- **Storage**: Bcrypt hashed
- **Single-use**: Each code can only be used once

### Trusted Devices

- **Duration**: 30 days
- **Fingerprinting**: Based on user agent + IP address
- **Auto-expiration**: Devices are automatically removed after 30 days
- **Renewal**: Using a trusted device extends its expiration

### Grace Period

- **Duration**: 7 days from admin/moderator promotion
- **Actions**: Admin actions allowed with warnings
- **Expiration**: Admin actions blocked until MFA enabled
