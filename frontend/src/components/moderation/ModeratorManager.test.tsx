import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ModeratorManager } from './ModeratorManager';
import * as moderationApi from '../../lib/moderation-api';
import * as userApi from '../../lib/user-api';

// Mock the API modules
vi.mock('../../lib/moderation-api');
vi.mock('../../lib/user-api');

describe('ModeratorManager', () => {
    const mockChannelId = 'channel-123';
    const mockModerators = [
        {
            id: 'mod-1',
            user_id: 'user-1',
            channel_id: mockChannelId,
            role: 'moderator' as const,
            assigned_at: '2024-01-01T00:00:00Z',
            username: 'john_doe',
            display_name: 'John Doe',
            avatar_url: 'https://example.com/avatar1.jpg',
        },
        {
            id: 'mod-2',
            user_id: 'user-2',
            channel_id: mockChannelId,
            role: 'admin' as const,
            assigned_at: '2024-01-02T00:00:00Z',
            username: 'jane_admin',
            display_name: 'Jane Admin',
            avatar_url: 'https://example.com/avatar2.jpg',
        },
        {
            id: 'mod-3',
            user_id: 'user-3',
            channel_id: mockChannelId,
            role: 'owner' as const,
            assigned_at: '2024-01-03T00:00:00Z',
            username: 'owner_user',
            display_name: 'Owner User',
        },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        // Default mock implementations
        vi.mocked(moderationApi.listChannelModerators).mockResolvedValue({
            success: true,
            data: mockModerators,
            meta: { total: 3, limit: 20, offset: 0 },
        });
    });

    describe('Rendering', () => {
        it('renders the component with title and description', async () => {
            render(<ModeratorManager channelId={mockChannelId} />);

            expect(screen.getByText('Moderators')).toBeInTheDocument();
            expect(screen.getByText('Manage moderators for this channel')).toBeInTheDocument();
            await screen.findByText('John Doe');
        });

        it('displays loading state initially', () => {
            vi.mocked(moderationApi.listChannelModerators).mockReturnValue(new Promise(() => {}));
            render(<ModeratorManager channelId={mockChannelId} />);

            expect(screen.getByLabelText('Loading moderators')).toBeInTheDocument();
        });

        it('displays moderators list after loading', async () => {
            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            expect(screen.getByText('Jane Admin')).toBeInTheDocument();
            expect(screen.getByText('Owner User')).toBeInTheDocument();
        });

        it('displays role badges correctly', async () => {
            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            // Check for role badges - they're spans, not status elements
            expect(screen.getByText('moderator')).toBeInTheDocument();
            expect(screen.getByText('admin')).toBeInTheDocument();
            expect(screen.getByText('owner')).toBeInTheDocument();
        });

        it('does not show Add Moderator button when canManage is false', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={false} />);

            expect(screen.queryByLabelText('Add new moderator')).not.toBeInTheDocument();
            await screen.findByText('John Doe');
        });

        it('shows Add Moderator button when canManage is true', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            expect(screen.getByLabelText('Add new moderator')).toBeInTheDocument();
            await screen.findByText('John Doe');
        });

        it('does not show action buttons when canManage is false', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={false} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            expect(screen.queryByLabelText(/^Edit.*permissions$/)).not.toBeInTheDocument();
            expect(screen.queryByLabelText(/^Remove/)).not.toBeInTheDocument();
        });

        it('shows action buttons for non-owner moderators when canManage is true', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            // Should have edit/remove buttons for moderator and admin, but not owner
            const editButtons = screen.getAllByText('Edit');
            const removeButtons = screen.getAllByText('Remove');
            
            // 2 moderators (not owner) should have edit/remove buttons
            expect(editButtons.length).toBe(2);
            expect(removeButtons.length).toBe(2);
        });
    });

    describe('Search functionality', () => {
        it('filters moderators by username', async () => {
            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            const searchInput = screen.getByLabelText('Search moderators');
            await userEvent.type(searchInput, 'john');

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
                expect(screen.queryByText('Jane Admin')).not.toBeInTheDocument();
            });
        });

        it('filters moderators by display name', async () => {
            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('Jane Admin')).toBeInTheDocument();
            });

            const searchInput = screen.getByLabelText('Search moderators');
            await userEvent.type(searchInput, 'Admin');

            await waitFor(() => {
                expect(screen.getByText('Jane Admin')).toBeInTheDocument();
                expect(screen.queryByText('John Doe')).not.toBeInTheDocument();
            });
        });

        it('shows no results message when search returns no matches', async () => {
            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            const searchInput = screen.getByLabelText('Search moderators');
            await userEvent.type(searchInput, 'nonexistent');

            await waitFor(() => {
                expect(screen.getByText('No moderators found matching your search')).toBeInTheDocument();
            });
        });
    });

    describe('Add Moderator', () => {
        it('opens add moderator modal when button is clicked', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            const addButton = screen.getByLabelText('Add new moderator');
            await userEvent.click(addButton);

            expect(screen.getByRole('dialog')).toBeInTheDocument();
            expect(screen.getByLabelText('Search for user to add as moderator')).toBeInTheDocument();
        });

        it('searches for users when typing in search field', async () => {
            const mockUserSuggestions = [
                {
                    id: 'user-new',
                    username: 'new_user',
                    display_name: 'New User',
                    avatar_url: 'https://example.com/avatar-new.jpg',
                    is_verified: false,
                },
            ];

            vi.mocked(userApi.searchUsersAutocomplete).mockResolvedValue(mockUserSuggestions);

            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            const addButton = screen.getByLabelText('Add new moderator');
            await userEvent.click(addButton);

            const searchInput = screen.getByLabelText('Search for user to add as moderator');
            await userEvent.type(searchInput, 'new');

            await waitFor(() => {
                expect(userApi.searchUsersAutocomplete).toHaveBeenCalledWith('new', 10);
            });
        });

        it('closes modal when cancel is clicked', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            const addButton = screen.getByLabelText('Add new moderator');
            await userEvent.click(addButton);

            expect(screen.getByRole('dialog')).toBeInTheDocument();

            const cancelButton = screen.getAllByText('Cancel')[0];
            await userEvent.click(cancelButton);

            await waitFor(() => {
                expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            });
        });

        it('successfully adds a moderator', async () => {
            const mockNewModerator = {
                id: 'mod-4',
                user_id: 'user-4',
                channel_id: mockChannelId,
                role: 'moderator' as const,
                assigned_at: '2024-01-04T00:00:00Z',
                username: 'new_moderator',
                display_name: 'New Moderator',
            };

            const mockUserSuggestions = [
                {
                    id: 'user-4',
                    username: 'new_moderator',
                    display_name: 'New Moderator',
                    avatar_url: 'https://example.com/avatar-new.jpg',
                    is_verified: false,
                },
            ];

            vi.mocked(userApi.searchUsersAutocomplete).mockResolvedValue(mockUserSuggestions);
            vi.mocked(moderationApi.addChannelModerator).mockResolvedValue({
                success: true,
                message: 'Moderator added successfully',
                moderator: mockNewModerator,
            });

            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            // Open add modal
            const addButton = screen.getByLabelText('Add new moderator');
            await userEvent.click(addButton);

            // Search for user
            const searchInput = screen.getByLabelText('Search for user to add as moderator');
            await userEvent.type(searchInput, 'new');

            // Wait for suggestions and select user
            await waitFor(() => {
                expect(screen.getByText('New Moderator')).toBeInTheDocument();
            });

            const userOption = screen.getByText('New Moderator');
            await userEvent.click(userOption);

            // Submit
            const submitButton = screen.getByRole('button', { name: 'Add Moderator' });
            await userEvent.click(submitButton);

            // Verify success message and modal closed
            await waitFor(() => {
                expect(screen.getByText(/Successfully added new_moderator as a moderator/)).toBeInTheDocument();
            });

            await waitFor(() => {
                expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            });

            // Verify list was reloaded
            expect(moderationApi.listChannelModerators).toHaveBeenCalledTimes(2); // Initial load + reload after add
        });
    });

    describe('Remove Moderator', () => {
        it('opens remove confirmation modal when remove is clicked', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            const removeButtons = screen.getAllByText('Remove');
            await userEvent.click(removeButtons[0]);

            expect(screen.getByText('Remove Moderator')).toBeInTheDocument();
            expect(screen.getByText(/Are you sure you want to remove/)).toBeInTheDocument();
        });

        it('closes remove modal when cancel is clicked', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            const removeButtons = screen.getAllByText('Remove');
            await userEvent.click(removeButtons[0]);

            const dialog = screen.getByRole('dialog');
            expect(dialog).toBeInTheDocument();

            const cancelButtons = screen.getAllByText('Cancel');
            const cancelButton = cancelButtons[cancelButtons.length - 1]; // Get last cancel button (in the modal)
            await userEvent.click(cancelButton);

            await waitFor(() => {
                expect(screen.queryByText('Remove Moderator')).not.toBeInTheDocument();
            }, { timeout: 2000 });
        });

        it('successfully removes a moderator', async () => {
            vi.mocked(moderationApi.removeChannelModerator).mockResolvedValue({
                success: true,
                message: 'Moderator removed successfully',
            });

            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            // Click remove button
            const removeButtons = screen.getAllByText('Remove');
            await userEvent.click(removeButtons[0]);

            // Confirm removal
            const confirmButton = screen.getByRole('button', { name: 'Remove' });
            await userEvent.click(confirmButton);

            // Verify success message and modal closed
            await waitFor(() => {
                expect(screen.getByText(/Successfully removed john_doe from moderators/)).toBeInTheDocument();
            });

            await waitFor(() => {
                expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            });

            // Verify list was reloaded
            expect(moderationApi.listChannelModerators).toHaveBeenCalledTimes(2); // Initial load + reload after remove
        });
    });

    describe('Edit Permissions', () => {
        it('opens edit permissions modal when edit is clicked', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            const editButtons = screen.getAllByText('Edit');
            await userEvent.click(editButtons[0]);

            expect(screen.getByText('Edit Moderator Permissions')).toBeInTheDocument();
            expect(screen.getByLabelText('Select moderator role')).toBeInTheDocument();
        });

        it('displays current role in edit modal', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            const editButtons = screen.getAllByText('Edit');
            await userEvent.click(editButtons[0]);

            const select = screen.getByLabelText('Select moderator role') as HTMLSelectElement;
            expect(select.value).toBe('moderator');
        });

        it('successfully updates moderator permissions', async () => {
            const mockUpdatedModerator = {
                id: 'mod-1',
                user_id: 'user-1',
                channel_id: mockChannelId,
                role: 'admin' as const,
                assigned_at: '2024-01-01T00:00:00Z',
                username: 'john_doe',
                display_name: 'John Doe',
            };

            vi.mocked(moderationApi.updateModeratorPermissions).mockResolvedValue({
                success: true,
                message: 'Permissions updated successfully',
                moderator: mockUpdatedModerator,
            });

            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            // Click edit button
            const editButtons = screen.getAllByText('Edit');
            await userEvent.click(editButtons[0]);

            // Change role to admin
            const select = screen.getByLabelText('Select moderator role') as HTMLSelectElement;
            await userEvent.selectOptions(select, 'admin');

            // Submit
            const updateButton = screen.getByRole('button', { name: 'Update' });
            await userEvent.click(updateButton);

            // Verify success message and modal closed
            await waitFor(() => {
                expect(screen.getByText(/Successfully updated john_doe's permissions to admin/)).toBeInTheDocument();
            });

            await waitFor(() => {
                expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            });

            // Verify list was reloaded
            expect(moderationApi.listChannelModerators).toHaveBeenCalledTimes(2); // Initial load + reload after update
        });
    });

    describe('Pagination', () => {
        it('does not show pagination when total pages is 1', async () => {
            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            expect(screen.queryByLabelText('Pagination')).not.toBeInTheDocument();
        });

        it('shows pagination when there are multiple pages', async () => {
            vi.mocked(moderationApi.listChannelModerators).mockResolvedValue({
                success: true,
                data: mockModerators,
                meta: { total: 50, limit: 20, offset: 0 },
            });

            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByLabelText('Pagination')).toBeInTheDocument();
            });

            expect(screen.getByText(/Page 1 of 3/)).toBeInTheDocument();
        });
    });

    describe('Error handling', () => {
        it('displays error message when loading fails', async () => {
            vi.mocked(moderationApi.listChannelModerators).mockRejectedValue(
                new Error('Failed to load')
            );

            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('Failed to load')).toBeInTheDocument();
            });
        });

        it('allows closing error alert', async () => {
            vi.mocked(moderationApi.listChannelModerators).mockRejectedValue(
                new Error('Failed to load')
            );

            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('Failed to load')).toBeInTheDocument();
            });

            const alert = screen.getByRole('alert');
            const closeButton = alert.querySelector('button');
            if (closeButton) {
                await userEvent.click(closeButton);
                await waitFor(() => {
                    expect(screen.queryByText('Failed to load')).not.toBeInTheDocument();
                });
            }
        });
    });

    describe('Empty state', () => {
        it('shows empty state message when no moderators exist', async () => {
            vi.mocked(moderationApi.listChannelModerators).mockResolvedValue({
                success: true,
                data: [],
                meta: { total: 0, limit: 20, offset: 0 },
            });

            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('No moderators yet')).toBeInTheDocument();
            });
        });
    });

    describe('Accessibility', () => {
        it('has proper ARIA labels on interactive elements', async () => {
            render(<ModeratorManager channelId={mockChannelId} canManage={true} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            expect(screen.getByLabelText('Add new moderator')).toBeInTheDocument();
            expect(screen.getByLabelText('Search moderators')).toBeInTheDocument();
        });

        it('has proper table structure with headers', async () => {
            render(<ModeratorManager channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('John Doe')).toBeInTheDocument();
            });

            const table = screen.getByRole('table');
            expect(table).toBeInTheDocument();

            const headers = screen.getAllByRole('columnheader');
            expect(headers.length).toBeGreaterThan(0);
        });
    });
});
