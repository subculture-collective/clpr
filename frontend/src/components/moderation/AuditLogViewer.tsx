import { useEffect, useMemo, useState, useCallback } from 'react';
import { format, subDays } from 'date-fns';
import {
    getAuditLogs,
    exportAuditLogs,
    type AuditLogEntry,
} from '../../lib/moderation-api';
import {
    Shield,
    Ban,
    AlertTriangle,
    CheckCircle,
    XCircle,
    Clock,
    Download,
    RefreshCw,
    Search,
    Filter,
    ChevronLeft,
    ChevronRight,
} from 'lucide-react';

interface AuditLogFilters {
    actor: string;
    action: string;
    target: string;
    startDate: string;
    endDate: string;
    search: string;
    limit: number;
    offset: number;
}

// Action values the backend writes to moderation_audit_logs. The API filters
// by exact match, so every value here must be one the backend stores.
const AUDIT_ACTIONS: ReadonlyArray<readonly [value: string, label: string]> = [
    ['ban_user', 'User banned'],
    ['unban_user', 'User unbanned'],
    ['suspend_comment_privileges', 'Comment privileges suspended'],
    ['lift_comment_suspension', 'Comment suspension lifted'],
    ['enable_comment_review', 'Comment review turned on'],
    ['disable_comment_review', 'Comment review turned off'],
    ['update_user_role', 'User role changed'],
    ['update_user_karma', 'User karma changed'],
    ['award_badge', 'Badge awarded'],
    ['remove_badge', 'Badge removed'],
    ['approve', 'Submission approved'],
    ['reject', 'Submission rejected'],
    ['resolve_appeal', 'Appeal resolved'],
    ['add_moderator', 'Channel moderator added'],
    ['remove_moderator', 'Channel moderator removed'],
    ['update_moderator_permissions', 'Channel moderator role changed'],
    ['clip_hidden', 'Clip hidden'],
    ['clip_unhidden', 'Clip unhidden'],
    ['clip_metadata_updated', 'Clip details updated'],
    ['dmca_notice_received', 'DMCA notice received'],
    ['dmca_notice_reviewed', 'DMCA notice reviewed'],
    ['dmca_takedown_processed', 'DMCA takedown processed'],
    ['dmca_strike_issued', 'DMCA strike issued'],
    ['dmca_user_suspended', 'DMCA user suspended'],
    ['dmca_user_terminated', 'DMCA user terminated'],
    ['dmca_counter_notice_received', 'DMCA counter-notice received'],
    ['dmca_counter_notice_forwarded', 'DMCA counter-notice forwarded'],
    ['dmca_content_reinstated', 'DMCA content reinstated'],
    ['dmca_lawsuit_filed', 'DMCA lawsuit filed'],
    ['account_deletion_requested', 'Account deletion requested'],
    ['account_deletion_cancelled', 'Account deletion cancelled'],
];

const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// The API rejects actor and target values that are not UUIDs, so anything
// else is left out of the request.
const validUserId = (value: string): string => {
    const trimmed = value.trim();
    return UUID_PATTERN.test(trimmed) ? trimmed : '';
};

const isIncompleteUserId = (value: string): boolean =>
    value.trim() !== '' && validUserId(value) === '';

const ID_FILTER_DEBOUNCE_MS = 300;

// Utility function to truncate IDs for display
const truncateId = (id: string, length: number = 8): string => {
    return id.length > length ? `${id.substring(0, length)}...` : id;
};

export function AuditLogViewer() {
    const [logs, setLogs] = useState<AuditLogEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [total, setTotal] = useState(0);
    const [exporting, setExporting] = useState(false);
    const [filters, setFilters] = useState<AuditLogFilters>({
        actor: '',
        action: '',
        target: '',
        startDate: format(subDays(new Date(), 30), 'yyyy-MM-dd'),
        endDate: format(new Date(), 'yyyy-MM-dd'),
        search: '',
        limit: 50,
        offset: 0,
    });

    // Actor and target reach the API only as complete UUIDs, after a pause in typing.
    const [appliedIds, setAppliedIds] = useState({ actor: '', target: '' });
    useEffect(() => {
        const timer = setTimeout(() => {
            const next = {
                actor: validUserId(filters.actor),
                target: validUserId(filters.target),
            };
            setAppliedIds((prev) =>
                prev.actor === next.actor && prev.target === next.target ? prev : next
            );
        }, ID_FILTER_DEBOUNCE_MS);
        return () => clearTimeout(timer);
    }, [filters.actor, filters.target]);

    // Convert date filters to RFC3339 format for API consistency
    const apiFilters = useMemo(
        () => ({
            actor: appliedIds.actor,
            target: appliedIds.target,
            action: filters.action,
            search: filters.search,
            limit: filters.limit,
            offset: filters.offset,
            startDate: filters.startDate
                ? new Date(filters.startDate + 'T00:00:00').toISOString()
                : '',
            endDate: filters.endDate
                ? new Date(filters.endDate + 'T23:59:59').toISOString()
                : '',
        }),
        [
            appliedIds,
            filters.action,
            filters.search,
            filters.limit,
            filters.offset,
            filters.startDate,
            filters.endDate,
        ]
    );

    const loadLogs = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            const response = await getAuditLogs(apiFilters);
            setLogs(response.logs);
            setTotal(response.total);
        } catch (err) {
            setError('Failed to load audit logs');
            console.error('Audit logs error:', err);
        } finally {
            setLoading(false);
        }
    }, [apiFilters]);

    useEffect(() => {
        loadLogs();
    }, [loadLogs]);

    const handleExportCSV = async () => {
        try {
            setExporting(true);
            // Convert date filters to RFC3339 format for API consistency
            const blob = await exportAuditLogs({
                actor: validUserId(filters.actor) || undefined,
                action: filters.action || undefined,
                target: validUserId(filters.target) || undefined,
                startDate: filters.startDate
                    ? new Date(filters.startDate + 'T00:00:00').toISOString()
                    : undefined,
                endDate: filters.endDate
                    ? new Date(filters.endDate + 'T23:59:59').toISOString()
                    : undefined,
                search: filters.search || undefined,
            });

            // Download file
            const link = document.createElement('a');
            const url = URL.createObjectURL(blob);
            link.setAttribute('href', url);
            link.setAttribute(
                'download',
                `audit-logs-${format(new Date(), 'yyyy-MM-dd-HHmmss')}.csv`
            );
            link.style.visibility = 'hidden';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            URL.revokeObjectURL(url);
        } catch (err) {
            console.error('Export error:', err);
            setError('Failed to export audit logs');
        } finally {
            setExporting(false);
        }
    };

    const getActionIcon = (action: string) => {
        const iconClass = 'h-4 w-4';
        switch (action.toLowerCase()) {
            case 'ban':
            case 'ban_user':
                return <Ban className={iconClass} />;
            case 'unban':
            case 'unban_user':
                return <CheckCircle className={iconClass} />;
            case 'timeout':
            case 'user_timeout':
                return <Clock className={iconClass} />;
            case 'warn':
            case 'warning_issued':
                return <AlertTriangle className={iconClass} />;
            case 'approve':
            case 'approved':
                return <CheckCircle className={iconClass} />;
            case 'reject':
            case 'rejected':
                return <XCircle className={iconClass} />;
            case 'sync':
            case 'sync_bans':
                return <RefreshCw className={iconClass} />;
            default:
                return <Shield className={iconClass} />;
        }
    };

    const getActionColor = (action: string) => {
        switch (action.toLowerCase()) {
            case 'ban':
            case 'ban_user':
            case 'reject':
            case 'rejected':
                return 'bg-red-100 text-red-800';
            case 'unban':
            case 'unban_user':
            case 'approve':
            case 'approved':
                return 'bg-green-100 text-green-800';
            case 'timeout':
            case 'user_timeout':
                return 'bg-yellow-100 text-yellow-800';
            case 'warn':
            case 'warning_issued':
                return 'bg-orange-100 text-orange-800';
            case 'sync':
            case 'sync_bans':
                return 'bg-blue-100 text-blue-800';
            default:
                return 'bg-blue-100 text-blue-800';
        }
    };

    const handlePageChange = (newOffset: number) => {
        setFilters({ ...filters, offset: newOffset });
    };

    const currentPage = Math.floor(filters.offset / filters.limit) + 1;
    const totalPages = Math.ceil(total / filters.limit);

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex gap-2">
                    <button
                        onClick={loadLogs}
                        disabled={loading}
                        className="inline-flex items-center gap-2 rounded bg-gray-600 px-4 py-2 text-sm font-medium text-white hover:bg-gray-700 disabled:cursor-not-allowed disabled:bg-gray-400"
                        title="Refresh logs"
                    >
                        <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                        <span className="hidden sm:inline">Refresh</span>
                    </button>
                    <button
                        onClick={handleExportCSV}
                        disabled={logs.length === 0 || exporting}
                        className="inline-flex items-center gap-2 rounded bg-primary-400 px-4 py-2 text-sm font-medium text-background hover:bg-primary-300 disabled:cursor-not-allowed disabled:bg-gray-400"
                        title="Export to CSV"
                    >
                        <Download className="h-4 w-4" />
                        <span className="hidden sm:inline">{exporting ? 'Exporting...' : 'Export CSV'}</span>
                    </button>
                </div>
            </div>

            {/* Filters */}
            <div className="rounded-lg border border-gray-200 bg-white p-4">
                <div className="mb-3 flex items-center gap-2 text-sm font-medium text-gray-700">
                    <Filter className="h-4 w-4" />
                    <span>Filters</span>
                </div>
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
                    <div>
                        <label
                            htmlFor="action-filter"
                            className="mb-1 block text-sm font-medium text-gray-700"
                        >
                            Action
                        </label>
                        <select
                            id="action-filter"
                            value={filters.action}
                            onChange={(e) =>
                                setFilters({ ...filters, action: e.target.value, offset: 0 })
                            }
                            className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                        >
                            <option value="">All Actions</option>
                            {AUDIT_ACTIONS.map(([value, label]) => (
                                <option key={value} value={value}>
                                    {label}
                                </option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label
                            htmlFor="actor-filter"
                            className="mb-1 block text-sm font-medium text-gray-700"
                        >
                            Actor (User ID)
                        </label>
                        <input
                            id="actor-filter"
                            type="text"
                            value={filters.actor}
                            onChange={(e) =>
                                setFilters({ ...filters, actor: e.target.value, offset: 0 })
                            }
                            placeholder="Filter by actor ID..."
                            aria-describedby={isIncompleteUserId(filters.actor) ? 'actor-filter-hint' : undefined}
                            className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                        />
                        {isIncompleteUserId(filters.actor) && (
                            <p id="actor-filter-hint" className="mt-1 text-xs text-gray-600">
                                Enter the full user ID to filter.
                            </p>
                        )}
                    </div>
                    <div>
                        <label
                            htmlFor="target-filter"
                            className="mb-1 block text-sm font-medium text-gray-700"
                        >
                            Target (User ID)
                        </label>
                        <input
                            id="target-filter"
                            type="text"
                            value={filters.target}
                            onChange={(e) =>
                                setFilters({ ...filters, target: e.target.value, offset: 0 })
                            }
                            placeholder="Filter by target ID..."
                            aria-describedby={isIncompleteUserId(filters.target) ? 'target-filter-hint' : undefined}
                            className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                        />
                        {isIncompleteUserId(filters.target) && (
                            <p id="target-filter-hint" className="mt-1 text-xs text-gray-600">
                                Enter the full user ID to filter.
                            </p>
                        )}
                    </div>
                    <div>
                        <label
                            htmlFor="start-date-filter"
                            className="mb-1 block text-sm font-medium text-gray-700"
                        >
                            Start Date
                        </label>
                        <input
                            id="start-date-filter"
                            type="date"
                            value={filters.startDate}
                            onChange={(e) =>
                                setFilters({
                                    ...filters,
                                    startDate: e.target.value,
                                    offset: 0,
                                })
                            }
                            className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                        />
                    </div>
                    <div>
                        <label
                            htmlFor="end-date-filter"
                            className="mb-1 block text-sm font-medium text-gray-700"
                        >
                            End Date
                        </label>
                        <input
                            id="end-date-filter"
                            type="date"
                            value={filters.endDate}
                            onChange={(e) =>
                                setFilters({ ...filters, endDate: e.target.value, offset: 0 })
                            }
                            className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                        />
                    </div>
                    <div className="md:col-span-2 lg:col-span-2 xl:col-span-4">
                        <label
                            htmlFor="search-filter"
                            className="mb-1 block text-sm font-medium text-gray-700"
                        >
                            Search (Reason)
                        </label>
                        <div className="relative">
                            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                            <input
                                id="search-filter"
                                type="text"
                                value={filters.search}
                                onChange={(e) =>
                                    setFilters({ ...filters, search: e.target.value, offset: 0 })
                                }
                                placeholder="Search in reason field..."
                                className="w-full rounded border border-gray-300 py-2 pl-10 pr-3 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                            />
                        </div>
                    </div>
                    <div>
                        <label
                            htmlFor="limit-filter"
                            className="mb-1 block text-sm font-medium text-gray-700"
                        >
                            Per Page
                        </label>
                        <select
                            id="limit-filter"
                            value={filters.limit}
                            onChange={(e) =>
                                setFilters({
                                    ...filters,
                                    limit: parseInt(e.target.value),
                                    offset: 0,
                                })
                            }
                            className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                        >
                            <option value="25">25</option>
                            <option value="50">50</option>
                            <option value="100">100</option>
                        </select>
                    </div>
                </div>
            </div>

            {/* Loading State */}
            {loading && (
                <div className="flex h-64 items-center justify-center rounded-lg border border-gray-200 bg-white">
                    <div className="text-center">
                        <RefreshCw className="mx-auto h-8 w-8 animate-spin text-gray-400" />
                        <div className="mt-2 text-lg text-gray-600">Loading audit logs...</div>
                    </div>
                </div>
            )}

            {/* Error State */}
            {error && !loading && (
                <div className="rounded-lg border border-red-200 bg-red-50 p-4">
                    <p className="text-red-800">{error}</p>
                    <button
                        onClick={loadLogs}
                        className="mt-2 text-sm text-red-600 underline hover:text-red-800"
                    >
                        Retry
                    </button>
                </div>
            )}

            {/* Timeline View */}
            {!loading && !error && (
                <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
                    {/* Table Header */}
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                        Timestamp
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                        Action
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                        Actor
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                        Target
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                        Entity Type
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                        Reason
                                    </th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-200 bg-white">
                                {logs.length === 0 ? (
                                    <tr>
                                        <td
                                            colSpan={6}
                                            className="px-6 py-12 text-center text-sm text-gray-500"
                                        >
                                            <Shield className="mx-auto h-12 w-12 text-gray-300" />
                                            <p className="mt-2 font-medium">No audit logs found</p>
                                            <p className="mt-1 text-xs">Try adjusting your filters</p>
                                        </td>
                                    </tr>
                                ) : (
                                    logs.map((log) => (
                                        <tr key={log.id} className="hover:bg-gray-50">
                                            <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-900">
                                                <div className="flex flex-col">
                                                    <span className="font-medium">
                                                        {format(
                                                            new Date(log.createdAt),
                                                            'MMM dd, yyyy'
                                                        )}
                                                    </span>
                                                    <span className="text-xs text-gray-500">
                                                        {format(
                                                            new Date(log.createdAt),
                                                            'HH:mm:ss'
                                                        )}
                                                    </span>
                                                </div>
                                            </td>
                                            <td className="whitespace-nowrap px-6 py-4 text-sm">
                                                <span
                                                    className={`inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-semibold ${getActionColor(
                                                        log.action
                                                    )}`}
                                                >
                                                    {getActionIcon(log.action)}
                                                    {log.action}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 text-sm text-gray-900">
                                                <div className="flex flex-col">
                                                    <span className="font-medium">
                                                        {log.actor.username || 'Unknown'}
                                                    </span>
                                                    <span className="text-xs text-gray-500">
                                                        {truncateId(log.actor.id)}
                                                    </span>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 text-sm text-gray-900">
                                                <div className="flex flex-col">
                                                    <span className="font-medium">
                                                        {log.target.username || 'Unknown'}
                                                    </span>
                                                    <span className="text-xs text-gray-500">
                                                        {truncateId(log.target.id)}
                                                    </span>
                                                </div>
                                            </td>
                                            <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                                                {log.entityType}
                                            </td>
                                            <td className="max-w-md px-6 py-4 text-sm text-gray-500">
                                                <div className="line-clamp-2" title={log.reason}>
                                                    {log.reason || '-'}
                                                </div>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}

            {/* Pagination */}
            {!loading && !error && totalPages > 1 && (
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div className="text-sm text-gray-700">
                        Showing {filters.offset + 1} to{' '}
                        {Math.min(filters.offset + filters.limit, total)} of {total} results
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() =>
                                handlePageChange(
                                    Math.max(0, filters.offset - filters.limit)
                                )
                            }
                            disabled={filters.offset === 0}
                            className="inline-flex items-center gap-1 rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:bg-gray-100 disabled:text-gray-400"
                        >
                            <ChevronLeft className="h-4 w-4" />
                            <span className="hidden sm:inline">Previous</span>
                        </button>
                        <span className="flex items-center px-3 text-sm text-gray-700">
                            Page {currentPage} of {totalPages}
                        </span>
                        <button
                            onClick={() =>
                                handlePageChange(filters.offset + filters.limit)
                            }
                            disabled={filters.offset + filters.limit >= total}
                            className="inline-flex items-center gap-1 rounded border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:bg-gray-100 disabled:text-gray-400"
                        >
                            <span className="hidden sm:inline">Next</span>
                            <ChevronRight className="h-4 w-4" />
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}
