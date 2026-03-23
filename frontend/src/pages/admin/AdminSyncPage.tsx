import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Container, Card, CardHeader, CardBody, Button } from '../../components';
import { RefreshCw, Clock, CheckCircle, AlertCircle, XCircle, PlayCircle, Info } from 'lucide-react';
import axios from 'axios';

interface SyncStatus {
  last_sync_time: string;
  last_sync_status: 'success' | 'failed' | 'running';
  total_synced: number;
  failed_count: number;
  is_syncing: boolean;
}

export function AdminSyncPage() {
  const queryClient = useQueryClient();
  const [triggering, setTriggering] = useState(false);

  // Fetch sync status
  const { data: status, isLoading, error, refetch } = useQuery<SyncStatus>({
    queryKey: ['sync-status'],
    queryFn: async () => {
      const response = await axios.get('/api/v1/admin/sync/status');
      return response.data;
    },
    refetchInterval: 10000, // Auto-refresh every 10 seconds
  });

  // Trigger manual sync mutation
  const triggerSyncMutation = useMutation({
    mutationFn: async () => {
      const response = await axios.post('/api/v1/admin/sync/clips', {});
      return response.data;
    },
    onMutate: () => {
      setTriggering(true);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sync-status'] });
      setTimeout(() => {
        setTriggering(false);
        refetch();
      }, 2000);
    },
    onError: () => {
      setTriggering(false);
    },
  });

  const getStatusIcon = (status?: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'failed':
        return <XCircle className="w-5 h-5 text-red-500" />;
      case 'running':
        return <RefreshCw className="w-5 h-5 text-blue-500 animate-spin" />;
      default:
        return <AlertCircle className="w-5 h-5 text-gray-500" />;
    }
  };

  const getStatusBadge = (status?: string) => {
    const classes = {
      success: 'bg-green-500/20 text-green-400',
      failed: 'bg-red-500/20 text-red-400',
      running: 'bg-blue-500/20 text-blue-400',
    };

    return (
      <span
        className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium ${
          classes[status as keyof typeof classes] || 'bg-gray-500/20 text-gray-400'
        }`}
      >
        {getStatusIcon(status)}
        {status || 'unknown'}
      </span>
    );
  };

  return (
    <Container className="py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">Sync Controls</h1>
        <p className="text-muted-foreground">
          Manually trigger and monitor Twitch clip synchronization
        </p>
      </div>

      {error && (
        <Card className="mb-6">
          <CardBody>
            <div className="bg-destructive/10 border border-destructive text-destructive px-4 py-3 rounded">
              <p className="font-bold">Error loading sync status</p>
              <p>Unable to connect to sync service. Please try again later.</p>
            </div>
          </CardBody>
        </Card>
      )}

      {/* Sync Status Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardBody>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Last Sync Status</p>
                {isLoading ? (
                  <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-20 mt-2 animate-pulse"></div>
                ) : (
                  <div className="mt-2">{getStatusBadge(status?.last_sync_status)}</div>
                )}
              </div>
              {getStatusIcon(status?.last_sync_status)}
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Last Sync Time</p>
                {isLoading ? (
                  <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-32 mt-2 animate-pulse"></div>
                ) : status?.last_sync_time ? (
                  <p className="text-lg font-semibold mt-1">
                    {new Date(status.last_sync_time).toLocaleString()}
                  </p>
                ) : (
                  <p className="text-lg font-semibold mt-1 text-muted-foreground">Never</p>
                )}
              </div>
              <Clock className="w-5 h-5 text-muted-foreground" />
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Synced</p>
                {isLoading ? (
                  <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-16 mt-2 animate-pulse"></div>
                ) : (
                  <p className="text-2xl font-bold mt-1">{status?.total_synced || 0}</p>
                )}
              </div>
              <CheckCircle className="w-5 h-5 text-green-500" />
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Failed</p>
                {isLoading ? (
                  <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-12 mt-2 animate-pulse"></div>
                ) : (
                  <p className="text-2xl font-bold mt-1">{status?.failed_count || 0}</p>
                )}
              </div>
              <XCircle className="w-5 h-5 text-red-500" />
            </div>
          </CardBody>
        </Card>
      </div>

      {/* Manual Sync Control */}
      <Card className="mb-6">
        <CardHeader>
          <h2 className="text-xl font-semibold">Manual Sync Control</h2>
        </CardHeader>
        <CardBody>
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium mb-1">Trigger Manual Sync</p>
              <p className="text-sm text-muted-foreground">
                Start a manual synchronization of Twitch clips. This will fetch the latest clips from
                configured channels.
              </p>
              {status?.is_syncing && (
                <p className="text-sm text-blue-400 mt-2 flex items-center gap-2">
                  <RefreshCw className="w-4 h-4 animate-spin" />
                  Sync is currently running...
                </p>
              )}
            </div>
            <Button
              onClick={() => triggerSyncMutation.mutate()}
              disabled={status?.is_syncing || triggering || isLoading}
              variant="primary"
              className="ml-4"
            >
              {triggering ? (
                <>
                  <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                  Starting...
                </>
              ) : status?.is_syncing ? (
                <>
                  <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                  Running
                </>
              ) : (
                <>
                  <PlayCircle className="w-4 h-4 mr-2" />
                  Trigger Sync
                </>
              )}
            </Button>
          </div>

          {triggerSyncMutation.isSuccess && (
            <div className="mt-4 p-3 bg-green-500/10 border border-green-500/20 text-green-400 rounded">
              <p className="text-sm">
                ✓ Sync triggered successfully! Status will update automatically.
              </p>
            </div>
          )}

          {triggerSyncMutation.isError && (
            <div className="mt-4 p-3 bg-red-500/10 border border-red-500/20 text-red-400 rounded">
              <p className="text-sm">
                ✗ Failed to trigger sync. Please check logs or try again later.
              </p>
            </div>
          )}
        </CardBody>
      </Card>

      {/* Sync Information */}
      <Card>
        <CardHeader>
          <h2 className="text-xl font-semibold">Sync Information</h2>
        </CardHeader>
        <CardBody>
          <div className="space-y-3 text-sm">
            <div className="flex items-center justify-between py-2 border-b border-border">
              <span className="text-muted-foreground">Service</span>
              <span className="font-medium">Twitch Clip Sync</span>
            </div>
            <div className="flex items-center justify-between py-2 border-b border-border">
              <span className="text-muted-foreground">Auto-refresh</span>
              <span className="font-medium">Every 10 seconds</span>
            </div>
            <div className="flex items-center justify-between py-2 border-b border-border">
              <span className="text-muted-foreground">Current Status</span>
              {isLoading ? (
                <div className="h-5 bg-gray-200 dark:bg-gray-700 rounded w-20 animate-pulse"></div>
              ) : (
                getStatusBadge(status?.last_sync_status)
              )}
            </div>
            <div className="flex items-center justify-between py-2">
              <span className="text-muted-foreground">Is Syncing</span>
              {isLoading ? (
                <div className="h-5 bg-gray-200 dark:bg-gray-700 rounded w-12 animate-pulse"></div>
              ) : (
                <span className="font-medium">{status?.is_syncing ? 'Yes' : 'No'}</span>
              )}
            </div>
          </div>

          <div className="mt-4 p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded text-sm">
            <p className="font-medium mb-1 flex items-center gap-1.5"><Info size={16} strokeWidth={1.75} /> Note</p>
            <p>
              Automatic synchronization runs on a schedule defined in the backend configuration.
              Manual triggers are useful for testing or forcing an immediate sync.
            </p>
          </div>
        </CardBody>
      </Card>
    </Container>
  );
}
