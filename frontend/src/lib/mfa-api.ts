import { isAxiosError } from 'axios';
import { apiClient } from './api';
import type { User } from './auth-api';

export interface MFAStatus {
  enabled: boolean;
  enrolled_at?: string;
  backup_codes_remaining: number;
  trusted_devices_count: number;
  required: boolean;
  required_at?: string;
  grace_period_end?: string;
  in_grace_period: boolean;
}

export interface MFAEnrollment {
  secret: string;
  qr_code_url: string;
  backup_codes: string[];
}

export interface TrustedDevice {
  id: number;
  device_name?: string;
  ip_address?: string;
  trusted_at: string;
  expires_at: string;
  last_used_at: string;
}

/**
 * Error codes returned by POST /auth/mfa/verify-login.
 */
export type MFALoginErrorCode =
  | 'MFA_CODE_INVALID'
  | 'MFA_CHALLENGE_MISSING'
  | 'MFA_CHALLENGE_EXPIRED'
  | 'MFA_CHALLENGE_EXHAUSTED'
  | 'MFA_LOCKED';

export class MFALoginError extends Error {
  readonly code?: MFALoginErrorCode;

  constructor(message: string, code?: MFALoginErrorCode) {
    super(message);
    this.name = 'MFALoginError';
    this.code = code;
  }

  /** The sign-in can no longer be completed; the user must start over. */
  get requiresNewSignIn(): boolean {
    return (
      this.code === 'MFA_CHALLENGE_MISSING' ||
      this.code === 'MFA_CHALLENGE_EXPIRED' ||
      this.code === 'MFA_CHALLENGE_EXHAUSTED'
    );
  }
}

/** Returns the server's error message, or fallback. */
export function mfaErrorMessage(error: unknown, fallback: string): string {
  if (isAxiosError(error)) {
    const message = (error.response?.data as { error?: unknown } | undefined)?.error;
    if (typeof message === 'string' && message) return message;
  }
  return fallback;
}

/**
 * Completes a sign-in that stopped for an MFA code. The pending sign-in is
 * identified by an HttpOnly cookie, so only the code is sent.
 */
export async function verifyMFALogin(code: string, trustDevice: boolean): Promise<User> {
  try {
    const response = await apiClient.post<{ user: User }>(
      '/auth/mfa/verify-login',
      { code: code.trim(), trust_device: trustDevice },
      // A 401 here means the pending sign-in expired, not that a session
      // needs refreshing.
      { skipAuthRefresh: true },
    );
    return response.data.user;
  } catch (error) {
    if (isAxiosError(error) && error.response) {
      const data = error.response.data as { error?: string; code?: MFALoginErrorCode } | undefined;
      throw new MFALoginError(
        data?.error || 'Verification failed. Please try again.',
        data?.code,
      );
    }
    throw new MFALoginError('Could not reach the server. Check your connection and try again.');
  }
}

export async function getMFAStatus(): Promise<MFAStatus> {
  const response = await apiClient.get<MFAStatus>('/auth/mfa/status');
  return response.data;
}

export async function startMFAEnrollment(): Promise<MFAEnrollment> {
  const response = await apiClient.post<MFAEnrollment>('/auth/mfa/enroll');
  return response.data;
}

export async function verifyMFAEnrollment(code: string): Promise<void> {
  await apiClient.post('/auth/mfa/verify-enrollment', { code: code.replace(/\s/g, '') });
}

export async function regenerateBackupCodes(code: string): Promise<string[]> {
  const response = await apiClient.post<{ backup_codes: string[] }>(
    '/auth/mfa/regenerate-backup-codes',
    { code: code.replace(/\s/g, '') },
  );
  return response.data.backup_codes;
}

export async function disableMFA(code: string): Promise<void> {
  await apiClient.post('/auth/mfa/disable', { code: code.trim() });
}

export async function getTrustedDevices(): Promise<TrustedDevice[]> {
  const response = await apiClient.get<{ devices: TrustedDevice[] | null }>('/auth/mfa/trusted-devices');
  return response.data.devices ?? [];
}

export async function revokeTrustedDevice(id: number): Promise<void> {
  await apiClient.delete(`/auth/mfa/trusted-devices/${id}`);
}
