import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { platformModeratorApi } from '../../lib/platform-moderator-api';
import { render, screen, waitFor } from '../../test/test-utils';
import { AdminModeratorsPage } from './AdminModeratorsPage';

vi.mock('../../lib/platform-moderator-api', () => ({
    platformModeratorApi: {
        list: vi.fn(),
        add: vi.fn(),
        update: vi.fn(),
        revoke: vi.fn(),
    },
}));

const pageOf = (page: number, totalPages: number, names: string[]) => ({
    items: names.map((name) => ({ id: `id-${name}`, username: name, role: 'moderator' as const })),
    total: names.length,
    page,
    limit: 25,
    total_pages: totalPages,
});

describe('AdminModeratorsPage', () => {
    const api = vi.mocked(platformModeratorApi);

    beforeEach(() => {
        vi.clearAllMocks();
        api.list.mockResolvedValue(pageOf(1, 1, ['mod_one']));
    });

    it('asks for confirmation before revoking access', async () => {
        api.revoke.mockResolvedValue();
        const user = userEvent.setup();
        render(<AdminModeratorsPage />);

        await user.click(await screen.findByRole('button', { name: 'Revoke' }));

        expect(await screen.findByText('Revoke moderator access?')).toBeInTheDocument();
        expect(api.revoke).not.toHaveBeenCalled();

        await user.click(screen.getByRole('button', { name: 'Cancel' }));
        expect(api.revoke).not.toHaveBeenCalled();

        await user.click(screen.getByRole('button', { name: 'Revoke' }));
        const buttons = await screen.findAllByRole('button', { name: 'Revoke' });
        await user.click(buttons[buttons.length - 1]);

        await waitFor(() =>
            expect(api.revoke).toHaveBeenCalledWith('id-mod_one', 'Platform moderator access revoked'),
        );
    });

    it('shows the message the API returns', async () => {
        api.add.mockRejectedValue({
            response: { status: 400, data: { error: 'A valid user_id is required' } },
        });
        const user = userEvent.setup();
        render(<AdminModeratorsPage />);

        await user.type(await screen.findByLabelText('User ID'), 'not-a-uuid');
        await user.click(screen.getByRole('button', { name: 'Add moderator' }));

        expect(await screen.findByText('A valid user_id is required')).toBeInTheDocument();
    });

    it('pages through moderators using total_pages', async () => {
        api.list.mockImplementation(async (page = 1) =>
            page === 1 ? pageOf(1, 2, ['mod_one']) : pageOf(2, 2, ['mod_two']),
        );
        const user = userEvent.setup();
        render(<AdminModeratorsPage />);

        expect(await screen.findByText('Page 1 of 2')).toBeInTheDocument();
        await user.click(screen.getByRole('button', { name: 'Next' }));

        expect(await screen.findByText('mod_two')).toBeInTheDocument();
        expect(api.list).toHaveBeenLastCalledWith(2);
        expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled();
    });
});
