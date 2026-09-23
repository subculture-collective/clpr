import { useEffect, useState, useCallback, useMemo } from 'react';
import { format } from 'date-fns';
import {
    getChannelBans,
    getAllBans,
    unbanUser,
    type ChatBan,
} from '../../lib/chat-api';
import { Modal } from '@/components/ui/Modal';

interface BanListViewerProps {
    channelId?: string;
    /**
     * Whether the user can manage bans (revoke them)
     * @default false
     */
    canManage?: boolean;
}

type SortField = 'created_at' | 'username' | 'expires_at';
type SortDirection = 'asc' | 'desc';

interface BanFilters {
    user: string;
    reason: string;
    dateFrom: string;
    dateTo: string;
    status: 'all' | 'active' | 'expired';
}

export function BanListViewer({
    channelId,
    canManage = false,
}: BanListViewerProps) {
    const [bans, setBans] = useState<ChatBan[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [limit] = useState(50);
    const [sortField, setSortField] = useState<SortField>('created_at');
    const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
    const [filters, setFilters] = useState<BanFilters>({
        user: '',
        reason: '',
        dateFrom: '',
        dateTo: '',
        status: 'all',
    });

    // Revoke ban modal state
    const [revokeModalOpen, setRevokeModalOpen] = useState(false);
    const [banToRevoke, setBanToRevoke] = useState<ChatBan | null>(null);
    const [revoking, setRevoking] = useState(false);

    // View details modal state
    const [detailsModalOpen, setDetailsModalOpen] = useState(false);
    const [selectedBan, setSelectedBan] = useState<ChatBan | null>(null);

    // Bulk actions state
    const [selectedBanIds, setSelectedBanIds] = useState<Set<string>>(
        new Set(),
    );
    const [bulkActionModalOpen, setBulkActionModalOpen] = useState(false);
    const [bulkActionType, setBulkActionType] = useState<'unban' | null>(null);
    const [bulkActionInProgress, setBulkActionInProgress] = useState(false);

    const loadBans = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            const useAll = !channelId || channelId === 'all';
            const response =
                useAll ?
                    await getAllBans(page, limit)
                :   await getChannelBans(channelId, page, limit);
            const normalizedBans = response.bans.map(ban => ({
                ...ban,
                target_username: ban.target_username?.replace(
                    /^user(\d{2})/,
                    'user-$1',
                ),
            }));
            setBans(normalizedBans);
            setTotal(response.total);
        } catch (err) {
            setError('Failed to load bans');
            console.error('Bans loading error:', err);
        } finally {
            setLoading(false);
        }
    }, [channelId, page, limit]);

    useEffect(() => {
        loadBans();
    }, [loadBans]);

    // Filter and sort bans
    const filteredAndSortedBans = useMemo(() => {
        let result = [...bans];

        // Apply filters
        if (filters.user) {
            const searchTerm = filters.user.toLowerCase();
            result = result.filter(ban => {
                const username = ban.target_username?.toLowerCase() || '';
                const userId = ban.user_id?.toLowerCase() || '';
                return (
                    username.includes(searchTerm) || userId.includes(searchTerm)
                );
            });
        }

        if (filters.reason) {
            const searchTerm = filters.reason.toLowerCase();
            result = result.filter(ban =>
                ban.reason?.toLowerCase().includes(searchTerm),
            );
        }

        if (filters.dateFrom) {
            const fromDate = new Date(filters.dateFrom);
            result = result.filter(ban => new Date(ban.created_at) >= fromDate);
        }

        if (filters.dateTo) {
            const toDate = new Date(filters.dateTo);
            toDate.setHours(23, 59, 59, 999);
            result = result.filter(ban => new Date(ban.created_at) <= toDate);
        }

        if (filters.status !== 'all') {
            const now = new Date();
            result = result.filter(ban => {
                const expiresAt =
                    ban.expires_at ? new Date(ban.expires_at) : null;
                const isPermanent = !expiresAt;
                const isExpired = expiresAt ? expiresAt < now : false;
                const isActive = expiresAt ? expiresAt >= now : false;

                if (filters.status === 'expired') {
                    return isExpired;
                }
                // Treat permanent bans as active
                return isActive || isPermanent;
            });
        }

        // Apply sorting
        result.sort((a, b) => {
            let aValue: string | number | undefined;
            let bValue: string | number | undefined;

            if (sortField === 'created_at') {
                aValue = new Date(a.created_at).getTime();
                bValue = new Date(b.created_at).getTime();
            } else if (sortField === 'username') {
                aValue = a.target_username || '';
                bValue = b.target_username || '';
            } else if (sortField === 'expires_at') {
                const maxTime = Number.MAX_SAFE_INTEGER;
                aValue =
                    a.expires_at ? new Date(a.expires_at).getTime() : maxTime;
                bValue =
                    b.expires_at ? new Date(b.expires_at).getTime() : maxTime;
            }

            if (aValue === undefined || bValue === undefined) return 0;

            if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1;
            if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1;
            return 0;
        });

        return result;
    }, [bans, filters, sortField, sortDirection]);

    const handleSort = (field: SortField) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('desc');
        }
    };

    const handleSortKeyPress = (
        field: SortField,
        event: React.KeyboardEvent,
    ) => {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            handleSort(field);
        }
    };

    const handleRevokeBan = async () => {
        if (!banToRevoke) return;

        try {
            setRevoking(true);
            // Unban by ban ID to match API contract
            await unbanUser(banToRevoke.id);
            setRevokeModalOpen(false);
            setBanToRevoke(null);
            setPage(1); // Reset to first page after revocation
            setSuccessMessage(
                `Ban revoked successfully for ${
                    banToRevoke.target_username || banToRevoke.user_id
                }`,
            );
            await loadBans(); // Reload the list
        } catch (err) {
            setError(
                err instanceof Error ? err.message : 'Failed to revoke ban',
            );
            setSuccessMessage(null);
            setRevokeModalOpen(false);
            setBanToRevoke(null);
        } finally {
            setRevoking(false);
        }
    };

    const handleExportCSV = () => {
        if (filteredAndSortedBans.length === 0) return;

        // Create CSV header
        const headers = [
            'User',
            'Banned By',
            'Reason',
            'Banned At',
            'Expires At',
            'Status',
        ];

        // Create CSV rows
        const rows = filteredAndSortedBans.map(ban => {
            const now = new Date();
            const isExpired = ban.expires_at && new Date(ban.expires_at) < now;
            const status =
                ban.expires_at ?
                    isExpired ? 'Expired'
                    :   'Active'
                :   'Permanent';

            return [
                ban.target_username || ban.user_id,
                ban.banned_by_username || ban.banned_by,
                ban.reason || '',
                format(new Date(ban.created_at), 'yyyy-MM-dd HH:mm:ss'),
                ban.expires_at ?
                    format(new Date(ban.expires_at), 'yyyy-MM-dd HH:mm:ss')
                :   'Never',
                status,
            ];
        });

        // Combine into CSV string
        const csvContent = [
            headers.join(','),
            ...rows.map(row =>
                row
                    .map(cell => `"${String(cell).replace(/"/g, '""')}"`)
                    .join(','),
            ),
        ].join('\n');

        // Download file
        const blob = new Blob([csvContent], {
            type: 'text/csv;charset=utf-8;',
        });
        const link = document.createElement('a');
        const url = URL.createObjectURL(blob);

        try {
            link.setAttribute('href', url);
            link.setAttribute(
                'download',
                `channel-bans-${channelId || 'all-channels'}-${format(
                    new Date(),
                    'yyyy-MM-dd',
                )}.csv`,
            );
            link.style.visibility = 'hidden';
            document.body.appendChild(link);
            link.click();
        } finally {
            if (link.parentNode) {
                document.body.removeChild(link);
            }
            URL.revokeObjectURL(url);
        }
    };

    const handlePageChange = (newPage: number) => {
        setPage(newPage);
        // Clear selection when changing pages to avoid confusion
        setSelectedBanIds(new Set());
    };

    // Bulk selection handlers
    const handleSelectAll = () => {
        const activeBanIds = activeBans.map(ban => ban.id);

        if (
            selectedBanIds.size === activeBanIds.length &&
            activeBanIds.length > 0
        ) {
            // If all are selected, deselect all
            setSelectedBanIds(new Set());
        } else {
            // Select all active bans
            setSelectedBanIds(new Set(activeBanIds));
        }
    };

    const handleSelectBan = (banId: string) => {
        const newSelection = new Set(selectedBanIds);
        if (newSelection.has(banId)) {
            newSelection.delete(banId);
        } else {
            newSelection.add(banId);
        }
        setSelectedBanIds(newSelection);
    };

    const handleBulkUnban = async () => {
        if (selectedBanIds.size === 0) return;

        try {
            setBulkActionInProgress(true);
            const banIdsArray = Array.from(selectedBanIds);

            // Execute unbans sequentially to avoid overwhelming the server
            let successful = 0;
            let failed = 0;

            for (const banId of banIdsArray) {
                try {
                    await unbanUser(banId);
                    successful++;
                } catch (error) {
                    console.error(`Failed to unban ${banId}:`, error);
                    failed++;
                }
            }

            setBulkActionModalOpen(false);
            setSelectedBanIds(new Set());

            if (failed === 0) {
                setSuccessMessage(
                    `Successfully unbanned ${successful} user${successful !== 1 ? 's' : ''}`,
                );
            } else {
                setError(
                    `Unbanned ${successful} user${successful !== 1 ? 's' : ''}, but ${failed} failed`,
                );
            }

            setPage(1); // Reset to first page
            await loadBans(); // Reload the list
        } catch (err) {
            setError(
                err instanceof Error ?
                    err.message
                :   'Failed to perform bulk unban',
            );
        } finally {
            setBulkActionInProgress(false);
        }
    };

    const openBulkActionModal = (actionType: 'unban') => {
        setBulkActionType(actionType);
        setBulkActionModalOpen(true);
    };

    const totalPages = Math.ceil(total / limit);
    const currentPage = page;

    const getBanStatus = (ban: ChatBan) => {
        if (!ban.expires_at) return 'Permanent';
        const now = new Date();
        const expiresAt = new Date(ban.expires_at);
        return expiresAt < now ? 'Expired' : 'Active';
    };

    const getBanStatusColor = (status: string) => {
        switch (status) {
            case 'Active':
                return 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400';
            case 'Expired':
                return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400';
            case 'Permanent':
                return 'bg-purple-100 text-purple-800 dark:bg-purple-900/20 dark:text-purple-400';
            default:
                return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400';
        }
    };

    // Memoize active (non-expired) bans for bulk selection
    const activeBans = useMemo(() => {
        return filteredAndSortedBans.filter(
            ban => getBanStatus(ban) !== 'Expired',
        );
    }, [filteredAndSortedBans]);

    if (loading) {
        return (
            <div className='flex h-64 items-center justify-center'>
                <div className='text-lg text-gray-600 dark:text-gray-400'>
                    Loading bans...
                </div>
            </div>
        );
    }

    if (error && bans.length === 0) {
        return (
            <div className='rounded-lg border border-red-200 bg-red-50 dark:bg-red-900/20 dark:border-red-800 p-4'>
                <p className='text-red-800 dark:text-red-400'>{error}</p>
                <button
                    onClick={loadBans}
                    className='mt-2 text-sm text-red-600 dark:text-red-400 underline hover:text-red-800 dark:hover:text-red-300'
                >
                    Retry
                </button>
            </div>
        );
    }

    return (
        <div className='space-y-6'>
            {error && bans.length > 0 && (
                <div
                    role='alert'
                    className='rounded-lg border border-red-200 bg-red-50 dark:bg-red-900/20 dark:border-red-800 p-3 text-sm text-red-800'
                >
                    {error}
                </div>
            )}
            {successMessage && (
                <div
                    role='alert'
                    className='rounded-lg border border-green-200 bg-green-50 dark:bg-green-900/20 dark:border-green-800 p-3 text-sm text-green-800 dark:text-green-400'
                >
                    {successMessage}
                </div>
            )}
            <div className='flex items-center justify-between'>
                <h2 className='text-2xl font-bold text-gray-900 dark:text-gray-100'>
                    Channel Bans
                </h2>
                <button
                    onClick={handleExportCSV}
                    disabled={filteredAndSortedBans.length === 0}
                    className='rounded bg-primary-400 px-4 py-2 text-sm font-medium text-background hover:bg-primary-300 disabled:cursor-not-allowed disabled:bg-gray-400'
                    aria-label='Export bans to CSV'
                >
                    Export to CSV
                </button>
            </div>

            {/* Filters */}
            <div className='rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4'>
                <div className='grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-5'>
                    <div>
                        <label
                            htmlFor='user-filter'
                            className='mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300'
                        >
                            User
                        </label>
                        <input
                            id='user-filter'
                            type='text'
                            value={filters.user}
                            onChange={e =>
                                setFilters({ ...filters, user: e.target.value })
                            }
                            placeholder='Search by username'
                            className='w-full rounded border border-gray-300 dark:border-gray-600 px-3 py-2 text-sm dark:bg-gray-700 dark:text-white'
                        />
                    </div>
                    <div>
                        <label
                            htmlFor='reason-filter'
                            className='mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300'
                        >
                            Reason
                        </label>
                        <input
                            id='reason-filter'
                            type='text'
                            value={filters.reason}
                            onChange={e =>
                                setFilters({
                                    ...filters,
                                    reason: e.target.value,
                                })
                            }
                            placeholder='Search by reason'
                            className='w-full rounded border border-gray-300 dark:border-gray-600 px-3 py-2 text-sm dark:bg-gray-700 dark:text-white'
                        />
                    </div>
                    <div>
                        <label
                            htmlFor='date-from-filter'
                            className='mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300'
                        >
                            From Date
                        </label>
                        <input
                            id='date-from-filter'
                            type='date'
                            value={filters.dateFrom}
                            onChange={e =>
                                setFilters({
                                    ...filters,
                                    dateFrom: e.target.value,
                                })
                            }
                            className='w-full rounded border border-gray-300 dark:border-gray-600 px-3 py-2 text-sm dark:bg-gray-700 dark:text-white'
                        />
                    </div>
                    <div>
                        <label
                            htmlFor='date-to-filter'
                            className='mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300'
                        >
                            To Date
                        </label>
                        <input
                            id='date-to-filter'
                            type='date'
                            value={filters.dateTo}
                            onChange={e =>
                                setFilters({
                                    ...filters,
                                    dateTo: e.target.value,
                                })
                            }
                            className='w-full rounded border border-gray-300 dark:border-gray-600 px-3 py-2 text-sm dark:bg-gray-700 dark:text-white'
                        />
                    </div>
                    <div>
                        <label
                            htmlFor='status-filter'
                            className='mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300'
                        >
                            Status
                        </label>
                        <select
                            id='status-filter'
                            value={filters.status}
                            onChange={e =>
                                setFilters({
                                    ...filters,
                                    status: e.target.value as
                                        | 'all'
                                        | 'active'
                                        | 'expired',
                                })
                            }
                            className='w-full rounded border border-gray-300 dark:border-gray-600 px-3 py-2 text-sm dark:bg-gray-700 dark:text-white'
                        >
                            <option value='all'>All</option>
                            <option value='active'>Active</option>
                            <option value='expired'>Expired</option>
                        </select>
                    </div>
                </div>
            </div>

            {/* Error Alert */}
            {error && (
                <div
                    role='alert'
                    aria-live='assertive'
                    aria-atomic='true'
                    className='rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 p-4'
                >
                    <p className='text-red-800 dark:text-red-400'>{error}</p>
                </div>
            )}

            {/* Bulk Actions Toolbar */}
            {canManage && selectedBanIds.size > 0 && (
                <div className='rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-4'>
                    <div className='flex items-center justify-between'>
                        <div className='text-sm font-medium text-blue-800 dark:text-blue-400'>
                            {selectedBanIds.size} ban
                            {selectedBanIds.size !== 1 ? 's' : ''} selected
                        </div>
                        <div className='flex gap-2'>
                            <button
                                onClick={() => setSelectedBanIds(new Set())}
                                className='px-3 py-1.5 text-sm font-medium text-blue-700 dark:text-blue-300 hover:text-blue-900 dark:hover:text-blue-100'
                            >
                                Clear Selection
                            </button>
                            <button
                                onClick={() => openBulkActionModal('unban')}
                                disabled={bulkActionInProgress}
                                className='rounded bg-red-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:bg-gray-400'
                            >
                                {bulkActionInProgress ?
                                    'Processing...'
                                :   'Bulk Unban'}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Table */}
            <div className='overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800'>
                <table
                    className='min-w-full divide-y divide-gray-200 dark:divide-gray-700'
                    role='table'
                >
                    <thead className='bg-gray-50 dark:bg-gray-900'>
                        <tr>
                            {canManage && (
                                <th
                                    className='px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400'
                                    role='columnheader'
                                >
                                    <input
                                        type='checkbox'
                                        checked={
                                            selectedBanIds.size > 0 &&
                                            selectedBanIds.size ===
                                                activeBans.length
                                        }
                                        onChange={handleSelectAll}
                                        className='h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500'
                                        aria-label='Select all active bans'
                                        disabled={activeBans.length === 0}
                                    />
                                </th>
                            )}
                            <th
                                className='px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800'
                                onClick={() => handleSort('username')}
                                onKeyPress={e =>
                                    handleSortKeyPress('username', e)
                                }
                                tabIndex={0}
                                role='columnheader'
                                aria-sort={
                                    sortField === 'username' ?
                                        sortDirection === 'asc' ?
                                            'ascending'
                                        :   'descending'
                                    :   'none'
                                }
                            >
                                User{' '}
                                {sortField === 'username' &&
                                    (sortDirection === 'asc' ? '↑' : '↓')}
                            </th>
                            <th
                                className='px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400'
                                role='columnheader'
                            >
                                Banned By
                            </th>
                            <th
                                className='px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400'
                                role='columnheader'
                            >
                                Reason
                            </th>
                            <th
                                className='px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800'
                                onClick={() => handleSort('created_at')}
                                onKeyPress={e =>
                                    handleSortKeyPress('created_at', e)
                                }
                                tabIndex={0}
                                role='columnheader'
                                aria-sort={
                                    sortField === 'created_at' ?
                                        sortDirection === 'asc' ?
                                            'ascending'
                                        :   'descending'
                                    :   'none'
                                }
                            >
                                Banned At{' '}
                                {sortField === 'created_at' &&
                                    (sortDirection === 'asc' ? '↑' : '↓')}
                            </th>
                            <th
                                className='px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800'
                                onClick={() => handleSort('expires_at')}
                                onKeyPress={e =>
                                    handleSortKeyPress('expires_at', e)
                                }
                                tabIndex={0}
                                role='columnheader'
                                aria-sort={
                                    sortField === 'expires_at' ?
                                        sortDirection === 'asc' ?
                                            'ascending'
                                        :   'descending'
                                    :   'none'
                                }
                            >
                                Expires At{' '}
                                {sortField === 'expires_at' &&
                                    (sortDirection === 'asc' ? '↑' : '↓')}
                            </th>
                            <th
                                className='px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400'
                                role='columnheader'
                            >
                                Status
                            </th>
                            <th
                                className='px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400'
                                role='columnheader'
                            >
                                Actions
                            </th>
                        </tr>
                    </thead>
                    <tbody className='divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-800'>
                        {filteredAndSortedBans.length === 0 ?
                            <tr>
                                <td
                                    colSpan={canManage ? 8 : 7}
                                    className='px-6 py-4 text-center text-sm text-gray-500 dark:text-gray-400'
                                >
                                    No bans available
                                </td>
                            </tr>
                        :   filteredAndSortedBans.map(ban => {
                                const status = getBanStatus(ban);
                                const isExpired = status === 'Expired';
                                const isSelected = selectedBanIds.has(ban.id);
                                return (
                                    <tr
                                        key={ban.id}
                                        className='hover:bg-gray-50 dark:hover:bg-gray-700'
                                    >
                                        {canManage && (
                                            <td className='px-6 py-4 whitespace-nowrap'>
                                                <input
                                                    type='checkbox'
                                                    checked={isSelected}
                                                    onChange={() =>
                                                        handleSelectBan(ban.id)
                                                    }
                                                    disabled={isExpired}
                                                    className='h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50'
                                                    aria-label={`Select ban for ${
                                                        ban.target_username ||
                                                        ban.user_id
                                                    }`}
                                                />
                                            </td>
                                        )}
                                        <td className='px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100'>
                                            {ban.target_username || ban.user_id}
                                        </td>
                                        <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400'>
                                            {ban.banned_by_username ||
                                                ban.banned_by}
                                        </td>
                                        <td className='px-6 py-4 text-sm text-gray-500 dark:text-gray-400 max-w-xs truncate'>
                                            {ban.reason || '-'}
                                        </td>
                                        <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400'>
                                            {format(
                                                new Date(ban.created_at),
                                                'MMM dd, yyyy HH:mm',
                                            )}
                                        </td>
                                        <td className='px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400'>
                                            {ban.expires_at ?
                                                format(
                                                    new Date(ban.expires_at),
                                                    'MMM dd, yyyy HH:mm',
                                                )
                                            :   'Never'}
                                        </td>
                                        <td className='px-6 py-4 whitespace-nowrap'>
                                            <span
                                                className={`inline-flex px-2 py-1 text-xs font-semibold ${getBanStatusColor(
                                                    status,
                                                )}`}
                                            >
                                                {status}
                                            </span>
                                        </td>
                                        <td className='px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2'>
                                            <button
                                                onClick={() => {
                                                    setSelectedBan(ban);
                                                    setDetailsModalOpen(true);
                                                }}
                                                className='text-blue-600 dark:text-blue-400 hover:text-blue-900 dark:hover:text-blue-300'
                                                aria-label={`View details for ${
                                                    ban.target_username ||
                                                    ban.user_id
                                                }`}
                                            >
                                                Details
                                            </button>
                                            {canManage &&
                                                status !== 'Expired' && (
                                                    <button
                                                        onClick={() => {
                                                            setBanToRevoke(ban);
                                                            setRevokeModalOpen(
                                                                true,
                                                            );
                                                        }}
                                                        className='text-red-600 dark:text-red-400 hover:text-red-900 dark:hover:text-red-300'
                                                        aria-label={`Revoke ban for ${
                                                            ban.target_username ||
                                                            ban.user_id
                                                        }`}
                                                    >
                                                        Revoke
                                                    </button>
                                                )}
                                        </td>
                                    </tr>
                                );
                            })
                        }
                    </tbody>
                </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
                <div
                    className='flex items-center justify-between'
                    aria-label='Pagination'
                >
                    <div className='text-sm text-gray-700 dark:text-gray-300'>
                        Showing page {currentPage} of {totalPages} (
                        {filteredAndSortedBans.length} bans shown
                        {filteredAndSortedBans.length !== total ?
                            ` out of ${total} total`
                        :   ''}
                        )
                    </div>
                    <div className='flex gap-2'>
                        <button
                            onClick={() => handlePageChange(currentPage - 1)}
                            disabled={currentPage === 1}
                            className='rounded bg-gray-200 dark:bg-gray-700 px-3 py-1 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600 disabled:cursor-not-allowed disabled:opacity-50'
                            aria-label='Previous page'
                        >
                            Previous
                        </button>
                        <button
                            onClick={() => handlePageChange(currentPage + 1)}
                            disabled={currentPage === totalPages}
                            className='rounded bg-gray-200 dark:bg-gray-700 px-3 py-1 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600 disabled:cursor-not-allowed disabled:opacity-50'
                            aria-label='Next page'
                        >
                            Next
                        </button>
                    </div>
                </div>
            )}

            {/* Revoke Ban Modal */}
            <Modal
                open={revokeModalOpen}
                onClose={() => setRevokeModalOpen(false)}
                title='Revoke Ban'
                size='md'
            >
                <div className='space-y-4'>
                    <p className='text-sm text-gray-600 dark:text-gray-400'>
                        Are you sure you want to revoke the ban for{' '}
                        <strong>
                            {banToRevoke?.target_username ||
                                banToRevoke?.user_id}
                        </strong>
                        ?
                    </p>
                    {banToRevoke?.reason && (
                        <div className='rounded bg-gray-50 dark:bg-gray-700 p-3'>
                            <p className='text-xs text-gray-500 dark:text-gray-400'>
                                Original ban reason:
                            </p>
                            <p className='text-sm text-gray-700 dark:text-gray-300'>
                                {banToRevoke.reason}
                            </p>
                        </div>
                    )}
                    <div className='flex justify-end space-x-3 pt-4'>
                        <button
                            type='button'
                            onClick={() => setRevokeModalOpen(false)}
                            disabled={revoking}
                            className='px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 disabled:opacity-50'
                        >
                            Cancel
                        </button>
                        <button
                            type='button'
                            onClick={handleRevokeBan}
                            disabled={revoking}
                            className='px-4 py-2 text-sm font-medium text-white bg-red-600 border border-transparent rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50'
                        >
                            {revoking ? 'Revoking...' : 'Revoke Ban'}
                        </button>
                    </div>
                </div>
            </Modal>

            {/* View Details Modal */}
            <Modal
                open={detailsModalOpen}
                onClose={() => setDetailsModalOpen(false)}
                title='Ban Details'
                size='md'
            >
                {selectedBan && (
                    <div className='space-y-4'>
                        <div className='grid grid-cols-2 gap-4'>
                            <div>
                                <p className='text-xs text-gray-500 dark:text-gray-400'>
                                    Banned User
                                </p>
                                <p className='text-sm font-medium text-gray-900 dark:text-gray-100'>
                                    {selectedBan.target_username ||
                                        selectedBan.user_id}
                                </p>
                            </div>
                            <div>
                                <p className='text-xs text-gray-500 dark:text-gray-400'>
                                    Banned By
                                </p>
                                <p className='text-sm font-medium text-gray-900 dark:text-gray-100'>
                                    {selectedBan.banned_by_username ||
                                        selectedBan.banned_by}
                                </p>
                            </div>
                            <div>
                                <p className='text-xs text-gray-500 dark:text-gray-400'>
                                    Banned At
                                </p>
                                <p className='text-sm font-medium text-gray-900 dark:text-gray-100'>
                                    {format(
                                        new Date(selectedBan.created_at),
                                        'MMM dd, yyyy HH:mm:ss',
                                    )}
                                </p>
                            </div>
                            <div>
                                <p className='text-xs text-gray-500 dark:text-gray-400'>
                                    Expires At
                                </p>
                                <p className='text-sm font-medium text-gray-900 dark:text-gray-100'>
                                    {selectedBan.expires_at ?
                                        format(
                                            new Date(selectedBan.expires_at),
                                            'MMM dd, yyyy HH:mm:ss',
                                        )
                                    :   'Never (Permanent)'}
                                </p>
                            </div>
                            <div className='col-span-2'>
                                <p className='text-xs text-gray-500 dark:text-gray-400'>
                                    Status
                                </p>
                                <span
                                    className={`inline-flex mt-1 px-2 py-1 text-xs font-semibold ${getBanStatusColor(
                                        getBanStatus(selectedBan),
                                    )}`}
                                >
                                    {getBanStatus(selectedBan)}
                                </span>
                            </div>
                        </div>
                        {selectedBan.reason && (
                            <div>
                                <p className='text-xs text-gray-500 dark:text-gray-400'>
                                    Reason
                                </p>
                                <p className='text-sm text-gray-900 dark:text-gray-100 mt-1 p-3 bg-gray-50 dark:bg-gray-700 rounded'>
                                    {selectedBan.reason}
                                </p>
                            </div>
                        )}
                        <div className='flex justify-end pt-4'>
                            <button
                                type='button'
                                onClick={() => setDetailsModalOpen(false)}
                                className='px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500'
                            >
                                Close
                            </button>
                        </div>
                    </div>
                )}
            </Modal>

            {/* Bulk Action Confirmation Modal */}
            <Modal
                open={bulkActionModalOpen}
                onClose={() =>
                    !bulkActionInProgress && setBulkActionModalOpen(false)
                }
                title='Confirm Bulk Action'
                size='md'
            >
                <div className='space-y-4'>
                    <p className='text-sm text-gray-600 dark:text-gray-400'>
                        {bulkActionType === 'unban' && (
                            <>
                                Are you sure you want to unban{' '}
                                <strong>{selectedBanIds.size}</strong> user
                                {selectedBanIds.size !== 1 ? 's' : ''}?
                            </>
                        )}
                    </p>
                    <div className='rounded bg-yellow-50 dark:bg-yellow-900/20 p-3'>
                        <p className='text-xs text-yellow-800 dark:text-yellow-400'>
                            This action cannot be undone. The selected users
                            will immediately regain access to the channel.
                        </p>
                    </div>
                    {bulkActionInProgress && (
                        <div className='flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400'>
                            <div className='h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-600'></div>
                            <span>Processing bulk action...</span>
                        </div>
                    )}
                    <div className='flex justify-end space-x-3 pt-4'>
                        <button
                            type='button'
                            onClick={() => setBulkActionModalOpen(false)}
                            disabled={bulkActionInProgress}
                            className='px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 disabled:opacity-50 disabled:cursor-not-allowed'
                        >
                            Cancel
                        </button>
                        <button
                            type='button'
                            onClick={handleBulkUnban}
                            disabled={bulkActionInProgress}
                            className='px-4 py-2 text-sm font-medium text-white bg-red-600 border border-transparent rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed'
                        >
                            {bulkActionInProgress ?
                                'Processing...'
                            :   'Confirm Bulk Unban'}
                        </button>
                    </div>
                </div>
            </Modal>
        </div>
    );
}
