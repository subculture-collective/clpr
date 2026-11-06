import { afterEach, describe, expect, it, vi } from 'vitest';
import apiClient from './api';
import { submitContactMessage } from './contact-api';
import { submitReport, updateReport } from './report-api';

describe('state-changing API helpers use the CSRF-aware client', () => {
    afterEach(() => vi.restoreAllMocks());

    it('updates reports through apiClient', async () => {
        const put = vi.spyOn(apiClient, 'put').mockResolvedValue({ data: { message: 'ok' } });
        await updateReport('r1', { status: 'dismissed' });
        expect(put).toHaveBeenCalledWith('/admin/reports/r1', { status: 'dismissed' });
    });

    it('submits reports through apiClient', async () => {
        const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: {} });
        await submitReport({ reportable_type: 'clip', reportable_id: 'c1', reason: 'spam' });
        expect(post).toHaveBeenCalledWith('/reports', { reportable_type: 'clip', reportable_id: 'c1', reason: 'spam' });
    });

    it('submits contact messages through apiClient', async () => {
        const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: { message: 'ok', status: 'sent' } });
        const body = { email: 'a@example.test', category: 'feedback' as const, subject: 'Hi', message: 'Hello there' };
        await submitContactMessage(body);
        expect(post).toHaveBeenCalledWith('/contact', body);
    });
});
