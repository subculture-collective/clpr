import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import apiClient from '@/lib/api';
import { CommentActions } from './CommentActions';

const toast = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn(), info: vi.fn() }));

vi.mock('@/hooks', async () => {
    const actual = await vi.importActual('@/hooks');
    return {
        ...actual,
        useIsAuthenticated: vi.fn(() => true),
        useToast: vi.fn(() => toast),
    };
});

const renderActions = () =>
    render(
        <QueryClientProvider client={new QueryClient({ defaultOptions: { mutations: { retry: false } } })}>
            <CommentActions
                commentId='3f1e2d4c-0000-4000-8000-000000000001'
                clipId='clip-1'
                isAuthor={false}
                isAdmin={false}
                createdAt='2024-01-01T00:00:00Z'
            />
        </QueryClientProvider>,
    );

describe('CommentActions report dialog', () => {
    afterEach(() => {
        vi.restoreAllMocks();
        toast.success.mockClear();
        toast.error.mockClear();
    });

    it('offers only reasons the backend accepts and sends reportable fields', async () => {
        const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: { message: 'ok' } });
        renderActions();

        fireEvent.click(screen.getByRole('button', { name: 'Report' }));
        const options = Array.from(screen.getByRole('combobox').querySelectorAll('option')).map(o => o.value);
        expect(options).toEqual(['spam', 'harassment', 'nsfw', 'violence', 'copyright', 'other']);

        fireEvent.change(screen.getByRole('combobox'), { target: { value: 'copyright' } });
        fireEvent.click(screen.getByRole('button', { name: 'Submit Report' }));

        await waitFor(() => expect(post).toHaveBeenCalledWith('/reports', {
            reportable_type: 'comment',
            reportable_id: '3f1e2d4c-0000-4000-8000-000000000001',
            reason: 'copyright',
            description: undefined,
        }));
        await waitFor(() => expect(toast.success).toHaveBeenCalled());
    });

    it('shows an error toast when the report fails', async () => {
        vi.spyOn(apiClient, 'post').mockRejectedValue(new Error('400'));
        renderActions();

        fireEvent.click(screen.getByRole('button', { name: 'Report' }));
        fireEvent.click(screen.getByRole('button', { name: 'Submit Report' }));

        await waitFor(() => expect(toast.error).toHaveBeenCalledWith('The report could not be sent. Please try again.'));
    });
});
