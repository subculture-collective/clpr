/**
 * Offline Indicator Component
 * 
 * Displays network status and sync information to users
 * Shows when offline and provides feedback on pending operations
 */

import { useNetworkStatus } from '@/hooks/useNetworkStatus';
import { useSyncManager } from '@/hooks/useSyncManager';
import { useTranslation } from 'react-i18next';
import { SyncStatus } from '@/lib/sync-manager';
import { useOverlapsTwitchPlayer } from '@/hooks/useTwitchPlayerLayer';

export function OfflineIndicator() {
  const { t } = useTranslation();
  const { online, queuedRequestCount } = useNetworkStatus();
  const { syncState, pendingCount, triggerSync } = useSyncManager();
  const visible = !(online && queuedRequestCount === 0 && pendingCount === 0 && syncState.status !== SyncStatus.ERROR);
  // Twitch forbids covering its players; step aside while over one.
  const [coversPlayer, rootRef] = useOverlapsTwitchPlayer(visible);

  // Don't show anything if online and no pending operations
  if (!visible) {
    return null;
  }

  const totalPending = queuedRequestCount + pendingCount;
  const isSyncing = syncState.isSyncing;

  return (
    <div ref={rootRef} className={`fixed bottom-4 left-4 right-4 md:left-auto md:right-4 md:w-96 z-50 ${coversPlayer ? 'invisible' : ''}`}>
      {/* Offline Banner */}
      {!online && (
        <div 
          className="mb-2 bg-orange-500 text-white px-4 py-3 rounded-lg shadow-lg flex items-center gap-3"
          role="alert"
          aria-live="polite"
        >
          <svg
            className="w-5 h-5 flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414"
            />
          </svg>
          <div className="flex-1">
            <p className="font-medium">{t('offline.youreOffline')}</p>
            {totalPending > 0 && (
              <p className="text-sm opacity-90">
                {t('offline.operationsWillSync', { count: totalPending })}
              </p>
            )}
          </div>
        </div>
      )}

      {/* Sync Status */}
      {online && (isSyncing || totalPending > 0 || syncState.status === SyncStatus.ERROR) && (
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 px-4 py-3 rounded-lg shadow-lg">
          <div className="flex items-center gap-3">
            {/* Status Icon */}
            {isSyncing && (
              <svg
                className="animate-spin h-5 w-5 text-blue-500"
                fill="none"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
            )}
            {!isSyncing && syncState.status === SyncStatus.SUCCESS && totalPending === 0 && (
              <svg
                className="h-5 w-5 text-green-500"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            )}
            {syncState.status === SyncStatus.ERROR && (
              <svg
                className="h-5 w-5 text-red-500"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            )}

            {/* Status Text */}
            <div className="flex-1">
              {isSyncing && (
                <p className="text-sm text-gray-700 dark:text-gray-300">
                  {t('offline.syncing', { count: totalPending })}
                </p>
              )}
              {!isSyncing && syncState.status === SyncStatus.SUCCESS && totalPending === 0 && (
                <p className="text-sm text-gray-700 dark:text-gray-300">
                  {t('offline.allSynced')}
                </p>
              )}
              {!isSyncing && totalPending > 0 && syncState.status !== SyncStatus.ERROR && (
                <p className="text-sm text-gray-700 dark:text-gray-300">
                  {t('offline.changesPending', { count: totalPending })}
                </p>
              )}
              {syncState.status === SyncStatus.ERROR && (
                <>
                  <p className="text-sm text-red-600 dark:text-red-400 font-medium">
                    {t('offline.syncFailed')}
                  </p>
                  {syncState.error && (
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      {syncState.error.message}
                    </p>
                  )}
                </>
              )}
              {syncState.lastSyncAt && syncState.status === SyncStatus.SUCCESS && (
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  {t('offline.lastSynced', { time: new Date(syncState.lastSyncAt).toLocaleTimeString() })}
                </p>
              )}
            </div>

            {/* Action Button */}
            {!isSyncing && (syncState.status === SyncStatus.ERROR || totalPending > 0) && (
              <button
                onClick={triggerSync}
                className="text-sm font-medium text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
              >
                {t('common.retry')}
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
