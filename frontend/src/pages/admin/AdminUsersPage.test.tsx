import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import apiClient from '@/lib/api';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AdminUsersPage } from './AdminUsersPage';

describe('AdminUsersPage identity taxonomy', () => {
    beforeEach(() => {
        vi.spyOn(apiClient, 'get').mockResolvedValue({
            data: {
                users: [],
                total: 2966,
                page: 1,
                per_page: 25,
                summary: {
                    signed_in_users: 0,
                    unclaimed_creators: 2965,
                    staff: 1,
                    other: 0,
                },
            },
        });
    });

    afterEach(() => vi.restoreAllMocks());

    it('does not label imported creator identities as users', async () => {
        const queryClient = new QueryClient({
            defaultOptions: { queries: { retry: false } },
        });
        render(
            <QueryClientProvider client={queryClient}>
                <MemoryRouter>
                    <AdminUsersPage />
                </MemoryRouter>
            </QueryClientProvider>,
        );

        expect(await screen.findByText('Identity Records')).toBeInTheDocument();
        expect(screen.getByText('Signed-in Users')).toBeInTheDocument();
        expect(screen.getByText('Unclaimed Creators')).toBeInTheDocument();
        expect(screen.getByText('Staff')).toBeInTheDocument();
        expect(screen.getByText('Other Identities')).toBeInTheDocument();
        expect(screen.queryByText('Total Users')).not.toBeInTheDocument();
    });

    it('sends moderation actions through the CSRF-aware API client', async () => {
        vi.spyOn(apiClient, 'get').mockResolvedValue({
            data: {
                users: [{
                    id: 'u1', username: 'target', display_name: '', email: 'target@example.test',
                    avatar_url: '', role: 'user', karma_points: 0, is_banned: false,
                    account_status: 'active', created_at: '2026-01-01T00:00:00Z', last_login_at: '',
                }],
                total: 1,
                page: 1,
                per_page: 25,
                summary: { signed_in_users: 1, unclaimed_creators: 0, staff: 0, other: 0 },
            },
        });
        const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: {} });
        const queryClient = new QueryClient({
            defaultOptions: { queries: { retry: false } },
        });
        render(
            <QueryClientProvider client={queryClient}>
                <MemoryRouter>
                    <AdminUsersPage />
                </MemoryRouter>
            </QueryClientProvider>,
        );

        fireEvent.click(await screen.findByTitle('Ban'));
        fireEvent.change(screen.getByLabelText('Reason (Required)'), { target: { value: 'spam account' } });
        fireEvent.click(screen.getByRole('button', { name: 'Confirm' }));

        await waitFor(() => expect(post).toHaveBeenCalledWith('/admin/users/u1/ban', { reason: 'spam account' }));
    });
});
