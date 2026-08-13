import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import axios from 'axios';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AdminUsersPage } from './AdminUsersPage';

describe('AdminUsersPage identity taxonomy', () => {
    beforeEach(() => {
        vi.spyOn(axios, 'get').mockResolvedValue({
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
});
