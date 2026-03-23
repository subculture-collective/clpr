import { useState, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import { useAuth, useToast } from '@/hooks';
import { useFocusTrap } from '@/hooks/useFocusTrap';
import type { WatchHistoryEntry } from '@/types/watchHistory';

type FilterType = 'all' | 'completed' | 'in-progress';

export function WatchHistoryPage() {
  const [history, setHistory] = useState<WatchHistoryEntry[]>([]);
  const [filter, setFilter] = useState<FilterType>('all');
  const [isLoading, setIsLoading] = useState(true);
  const [showClearConfirm, setShowClearConfirm] = useState(false);
  const { isAuthenticated } = useAuth();
  const { showToast } = useToast();
  const modalRef = useRef<HTMLDivElement>(null);
  const clearButtonRef = useRef<HTMLButtonElement>(null);

  // Apply focus trap to modal when open
  useFocusTrap(modalRef, showClearConfirm);

  // Fetch watch history
  useEffect(() => {
    if (!isAuthenticated) {
      setIsLoading(false);
      return;
    }

    const fetchHistory = async () => {
      try {
        setIsLoading(true);
        const response = await fetch(`/api/v1/watch-history?filter=${filter}`, {
          credentials: 'include',
        });

        if (!response.ok) {
          showToast('Failed to fetch watch history', 'error');
          console.error('Failed to fetch watch history');
          return;
        }

        const data = await response.json();
        setHistory(data.history || []);
      } catch (error) {
        showToast('Error loading watch history', 'error');
        console.error('Error fetching watch history:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchHistory();
  }, [filter, isAuthenticated, showToast]);

  const handleClearHistory = async () => {
    try {
      const response = await fetch('/api/v1/watch-history', {
        method: 'DELETE',
        credentials: 'include',
      });

      if (!response.ok) {
        showToast('Failed to clear watch history', 'error');
        console.error('Failed to clear watch history');
        return;
      }

      setHistory([]);
      setShowClearConfirm(false);
      showToast('Watch history cleared successfully', 'success');

      // Return focus to the clear button
      if (clearButtonRef.current) {
        clearButtonRef.current.focus();
      }
    } catch (error) {
      showToast('Error clearing watch history', 'error');
      console.error('Error clearing watch history:', error);
    }
  };

  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins === 0) {
      return 'just now';
    } else if (diffMins < 60) {
      return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
    } else if (diffHours < 24) {
      return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
    } else if (diffDays < 7) {
      return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">
          <h1 className="text-3xl font-bold mb-4">Watch History</h1>
          <p className="text-muted-foreground mb-6">Sign in to view your watch history</p>
          <Link
            to="/login"
            className="inline-block px-6 py-3 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
          >
            Sign In
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6 gap-4">
        <h1 className="text-3xl font-bold">Watch History</h1>
        <button
          ref={clearButtonRef}
          onClick={() => setShowClearConfirm(true)}
          className="px-4 py-2 text-red-500 hover:text-red-400 transition-colors"
        >
          Clear History
        </button>
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2 mb-6 border-b border-border">
        <button
          onClick={() => setFilter('all')}
          className={`px-4 py-2 font-medium transition-colors ${
            filter === 'all'
              ? 'text-purple-500 border-b-2 border-purple-500'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          All
        </button>
        <button
          onClick={() => setFilter('in-progress')}
          className={`px-4 py-2 font-medium transition-colors ${
            filter === 'in-progress'
              ? 'text-purple-500 border-b-2 border-purple-500'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          In Progress
        </button>
        <button
          onClick={() => setFilter('completed')}
          className={`px-4 py-2 font-medium transition-colors ${
            filter === 'completed'
              ? 'text-purple-500 border-b-2 border-purple-500'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Completed
        </button>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-purple-500"></div>
          <p className="mt-4 text-muted-foreground">Loading watch history...</p>
        </div>
      ) : history.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground">No watch history found</p>
        </div>
      ) : (
        <div className="grid gap-4">
          {history.map((entry) => (
            <WatchHistoryCard key={entry.id} entry={entry} formatTime={formatTime} formatDate={formatDate} />
          ))}
        </div>
      )}

      {/* Clear History Confirmation Modal */}
      {showClearConfirm && (
        <div
          className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="clear-history-title"
          onKeyDown={(e) => {
            if (e.key === 'Escape') {
              setShowClearConfirm(false);
              if (clearButtonRef.current) {
                clearButtonRef.current.focus();
              }
            }
          }}
        >
          <div ref={modalRef} className="bg-background rounded-lg p-6 max-w-md w-full">
            <h2 id="clear-history-title" className="text-xl font-bold mb-4">Clear Watch History?</h2>
            <p className="text-muted-foreground mb-6">
              This will permanently delete your entire watch history. This action cannot be undone.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => {
                  setShowClearConfirm(false);
                  if (clearButtonRef.current) {
                    clearButtonRef.current.focus();
                  }
                }}
                className="px-4 py-2 bg-surface text-white rounded hover:bg-surface-hover transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleClearHistory}
                className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
              >
                Clear History
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

interface WatchHistoryCardProps {
  entry: WatchHistoryEntry;
  formatTime: (seconds: number) => string;
  formatDate: (dateString: string) => string;
}

function WatchHistoryCard({ entry, formatTime, formatDate }: WatchHistoryCardProps) {
  if (!entry.clip) {
    return null;
  }

  const { clip } = entry;

  // Clamp progress percent between 0 and 100
  const progressPercent = Math.max(0, Math.min(100, entry.progress_percent));

  return (
    <Link
      to={`/clips/${clip.id}`}
      className="flex gap-4 bg-background p-4 rounded-lg hover:bg-surface transition-colors"
    >
      {/* Thumbnail */}
      <div className="relative flex-shrink-0 w-48 h-28 bg-surface rounded overflow-hidden">
        {clip.thumbnail_url && (
          <img
            src={clip.thumbnail_url}
            alt={clip.title}
            className="w-full h-full object-cover"
            loading="lazy"
          />
        )}
        {/* Progress Indicator */}
        <div
          className="absolute bottom-0 left-0 right-0 h-1 bg-surface"
          role="progressbar"
          aria-valuenow={Math.round(progressPercent)}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label={`${Math.round(progressPercent)}% watched`}
        >
          <div
            className="h-full bg-purple-600"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <h3 className="font-semibold text-lg mb-1 truncate">{clip.title}</h3>
        <p className="text-sm text-muted-foreground mb-2">
          {clip.broadcaster_name} • {clip.game_name}
        </p>

        {/* Progress Info */}
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          {entry.completed ? (
            <span className="text-green-500 flex items-center gap-1">
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                <path
                  fillRule="evenodd"
                  d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                  clipRule="evenodd"
                />
              </svg>
              Completed
            </span>
          ) : (
            <>
              <span>
                {formatTime(entry.progress_seconds)} / {formatTime(entry.duration_seconds)} •{' '}
                {Math.floor(progressPercent)}% watched
              </span>
              <span className="text-purple-500 flex items-center gap-1 font-medium">
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                  <path d="M6.3 2.841A1.5 1.5 0 004 4.11V15.89a1.5 1.5 0 002.3 1.269l9.344-5.89a1.5 1.5 0 000-2.538L6.3 2.84z" />
                </svg>
                Resume
              </span>
            </>
          )}
          <span>Watched {formatDate(entry.watched_at)}</span>
        </div>
      </div>
    </Link>
  );
}
