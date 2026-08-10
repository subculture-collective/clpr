import { useState, useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { AxiosError } from 'axios';
import { Modal, ModalFooter } from '../ui/Modal';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { createClipFromStream, type CreateClipFromStreamRequest } from '../../lib/stream-api';

export interface ClipCreatorProps {
  streamer: string;
  /** Current playback time in seconds (optional) */
  currentTime?: number;
  /** Stream duration in seconds (optional) */
  duration?: number;
}

export function ClipCreator({ streamer, currentTime = 0, duration = 0 }: ClipCreatorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [startTime, setStartTime] = useState(0);
  const [endTime, setEndTime] = useState(30);
  const [title, setTitle] = useState('');
  const [quality, setQuality] = useState<'source' | '1080p' | '720p'>('1080p');
  const [errorMessage, setErrorMessage] = useState('');
  const navigate = useNavigate();

  const createClipMutation = useMutation({
    mutationFn: (request: CreateClipFromStreamRequest) =>
      createClipFromStream(streamer, request),
    onSuccess: (data) => {
      setIsOpen(false);
      // Show success and redirect after a short delay
      alert('Clip is being processed! Redirecting to clip page...');
      setTimeout(() => {
        navigate(`/clips/${data.clip_id}`);
      }, 1000);
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      const message = error.response?.data?.error || 'Failed to create clip. Please try again.';
      setErrorMessage(message);
    },
  });

  const openCreator = useCallback(() => {
    // Set initial time range around current playback time
    const start = Math.max(0, currentTime - 15);
    const end = duration > 0 ? Math.min(duration, currentTime + 15) : currentTime + 15;
    setStartTime(start);
    setEndTime(end);
    setTitle('');
    setErrorMessage('');
    setIsOpen(true);
  }, [currentTime, duration]);

  const handleSubmit = () => {
    setErrorMessage('');

    // Validate inputs
    if (!title.trim()) {
      setErrorMessage('Please enter a title for your clip');
      return;
    }

    // Validate title length (must match backend requirement of 3-255 chars)
    if (title.trim().length < 3) {
      setErrorMessage('Title must be at least 3 characters long');
      return;
    }

    if (title.trim().length > 255) {
      setErrorMessage('Title must be no more than 255 characters');
      return;
    }

    // Validate that end time is greater than start time
    if (endTime <= startTime) {
      setErrorMessage('End time must be greater than start time');
      return;
    }

    const clipDuration = endTime - startTime;
    if (clipDuration < 5) {
      setErrorMessage('Clip must be at least 5 seconds long');
      return;
    }
    if (clipDuration > 60) {
      setErrorMessage('Clip cannot be longer than 60 seconds');
      return;
    }

    // Create clip
    createClipMutation.mutate({
      streamer_username: streamer,
      start_time: startTime,
      end_time: endTime,
      quality,
      title: title.trim(),
    });
  };

  const clipDuration = endTime - startTime;
  const isProcessing = createClipMutation.isPending;

  return (
    <>
      <Button
        onClick={openCreator}
        variant="primary"
        className="gap-2"
        aria-label="Create clip from stream"
      >
        <svg
          className="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
          />
        </svg>
        Create Clip
      </Button>

      <Modal
        open={isOpen}
        onClose={() => !isProcessing && setIsOpen(false)}
        title="Create Clip from Stream"
        size="lg"
        closeOnBackdrop={!isProcessing}
      >
        <div className="space-y-4">
          {/* Error Message */}
          {errorMessage && (
            <div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-destructive text-sm">
              {errorMessage}
            </div>
          )}

          {/* Title Input */}
          <div>
            <label htmlFor="clip-title" className="block text-sm font-medium mb-2">
              Title
            </label>
            <Input
              id="clip-title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="My awesome clip"
              maxLength={255}
              disabled={isProcessing}
              className="w-full"
            />
          </div>

          {/* Time Range */}
          <div>
            <label className="block text-sm font-medium mb-2">
              Time Range (5-60 seconds)
            </label>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="start-time" className="block text-xs text-muted-foreground mb-1">
                  Start Time (seconds)
                </label>
                <Input
                  id="start-time"
                  type="number"
                  value={startTime.toFixed(1)}
                  onChange={(e) => setStartTime(Math.max(0, parseFloat(e.target.value) || 0))}
                  step="0.1"
                  min="0"
                  disabled={isProcessing}
                  className="w-full"
                />
              </div>
              <div>
                <label htmlFor="end-time" className="block text-xs text-muted-foreground mb-1">
                  End Time (seconds)
                </label>
                <Input
                  id="end-time"
                  type="number"
                  value={endTime.toFixed(1)}
                  onChange={(e) => setEndTime(Math.max(0, parseFloat(e.target.value) || 0))}
                  step="0.1"
                  min="0"
                  disabled={isProcessing}
                  className="w-full"
                />
              </div>
            </div>
            <p className="text-sm text-muted-foreground mt-2">
              Duration: {clipDuration.toFixed(1)} seconds
              {clipDuration < 5 && (
                <span className="text-destructive ml-2">(minimum 5 seconds)</span>
              )}
              {clipDuration > 60 && (
                <span className="text-destructive ml-2">(maximum 60 seconds)</span>
              )}
            </p>
          </div>

          {/* Quality Selection */}
          <div>
            <label htmlFor="quality" className="block text-sm font-medium mb-2">
              Quality
            </label>
            <select
              id="quality"
              value={quality}
              onChange={(e) => setQuality(e.target.value as 'source' | '1080p' | '720p')}
              disabled={isProcessing}
              className="w-full px-3 py-2 bg-background border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="source">Source (Best Quality)</option>
              <option value="1080p">1080p</option>
              <option value="720p">720p</option>
            </select>
          </div>

          {/* Processing Indicator */}
          {isProcessing && (
            <div className="rounded-lg bg-muted p-4">
              <div className="flex items-center gap-3 mb-3">
                <div className="animate-spin rounded-full h-5 w-5 border-t-2 border-b-2 border-primary"></div>
                <p className="text-sm font-medium">Creating your clip...</p>
              </div>
              <div className="w-full bg-background rounded-full h-2 overflow-hidden">
                <div className="bg-primary h-2 rounded-full animate-pulse" style={{ width: '60%' }} />
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                This may take up to 30 seconds. Please don't close this window.
              </p>
            </div>
          )}
        </div>

        <ModalFooter>
          <Button
            variant="ghost"
            onClick={() => setIsOpen(false)}
            disabled={isProcessing}
          >
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            disabled={isProcessing || !title.trim() || clipDuration < 5 || clipDuration > 60}
          >
            {isProcessing ? 'Creating...' : 'Create Clip'}
          </Button>
        </ModalFooter>
      </Modal>
    </>
  );
}
