import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const apiGet = vi.hoisted(() => vi.fn());
const userApi = vi.hoisted(() => ({
    fetchUserClips: vi.fn(),
    fetchUserActivity: vi.fn(),
    fetchUserFollowers: vi.fn(),
    fetchUserFollowing: vi.fn(),
    followUser: vi.fn(),
    unfollowUser: vi.fn(),
}));

vi.mock('../lib/api', () => ({ default: { get: apiGet } }));
vi.mock('../lib/user-api', () => userApi);
vi.mock('../context/AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('../components', async () => {
    const ui = await vi.importActual<typeof import('../components/ui')>('../components/ui');
    return {
        ...ui,
        Container: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
        SEO: () => null,
    };
});
vi.mock('../components/clip', () => ({ ClipGridCard: () => null }));

import { UserProfilePage } from './UserProfilePage';

// Production shape for /users/by-username/dani_denos?full=true, captured 2026-09-24.
const profile = {
    id: 'f767aadf-1704-4e4a-a3e4-13b10f1ae6da',
    username: 'dani_denos',
    display_name: 'Dani_Denos',
    karma_points: 0,
    role: 'user',
    is_banned: false,
    follower_count: 0,
    following_count: 0,
    is_verified: false,
    created_at: '2026-09-23T04:23:30.204018Z',
    stats: { clips_submitted: 0, total_upvotes: 0, total_comments: 0, clips_featured: 0, broadcasters_followed: 0 },
    is_following: false,
    is_followed_by: false,
};

function renderProfile(username = 'dani_denos') {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter initialEntries={[`/user/${username}`]}>
                <Routes>
                    <Route path='/user/:username' element={<UserProfilePage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>,
    );
}

describe('UserProfilePage', () => {
    beforeEach(() => {
        apiGet.mockReset();
        Object.values(userApi).forEach(mock => mock.mockReset());
    });

    it('requests the profile relative to the API base and survives an empty clip list', async () => {
        apiGet.mockResolvedValue({ data: { success: true, data: profile } });
        userApi.fetchUserClips.mockResolvedValue({ clips: null, total: 0, page: 1, limit: 20, has_more: false });

        renderProfile();

        expect(await screen.findByRole('heading', { level: 1, name: 'Dani_Denos' })).toBeInTheDocument();
        expect(apiGet).toHaveBeenCalledWith('/users/by-username/dani_denos?full=true');
        expect(screen.getByTestId('avatar-fallback')).toHaveTextContent('D');
    });

    it('shows a friendly not-found state for an unknown username', async () => {
        apiGet.mockRejectedValue(Object.assign(new Error('Request failed with status code 404'), {
            response: { status: 404, data: { error: 'user not found' } },
        }));

        renderProfile('zz_no_such_user_zz');

        expect(await screen.findByRole('heading', { level: 1, name: "This profile isn't here" })).toBeInTheDocument();
        expect(screen.queryByText(/status code/i)).not.toBeInTheDocument();
    });
});
