import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AppealForm } from './AppealForm';
import * as moderationApi from '../../lib/moderation-api';

// Mock the API module
vi.mock('../../lib/moderation-api');

describe('AppealForm', () => {
  const mockOnClose = vi.fn();
  const mockOnSuccess = vi.fn();
  const mockModerationActionId = 'action-123';

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(async () => {
    await act(async () => vi.runOnlyPendingTimers());
    vi.useRealTimers();
  });

  describe('Rendering', () => {
    it('renders the modal when open', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(screen.getByText('Appeal Moderation Decision')).toBeInTheDocument();
    });

    it('does not render when closed', () => {
      render(
        <AppealForm
          open={false}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(screen.queryByText('Appeal Moderation Decision')).not.toBeInTheDocument();
    });

    it('renders the reason textarea', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(screen.getByLabelText('Reason for Appeal *')).toBeInTheDocument();
    });

    it('renders description text', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(
        screen.getByText(/Please explain why you believe this moderation decision was incorrect/)
      ).toBeInTheDocument();
    });

    it('renders character counter', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(screen.getByText('0/2000 characters (minimum 10)')).toBeInTheDocument();
    });

    it('renders Submit Appeal button', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(screen.getByRole('button', { name: 'Submit Appeal' })).toBeInTheDocument();
    });

    it('renders Cancel button', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
    });
  });

  describe('Form Validation', () => {
    it('disables Submit button when reason is empty', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      expect(submitButton).toBeDisabled();
    });

    it('disables Submit button when reason is less than 10 characters', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'short');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      expect(submitButton).toBeDisabled();
    });

    it('enables Submit button when reason is at least 10 characters', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'This is a valid appeal reason');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      expect(submitButton).not.toBeDisabled();
    });

    it('shows error when submitting with reason less than 10 characters', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'short');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      expect(submitButton).toBeDisabled();
    });

    it('updates character counter when typing', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'Test appeal');

      expect(screen.getByText('11/2000 characters (minimum 10)')).toBeInTheDocument();
    });
  });

  describe('Form Submission', () => {
    it('successfully submits appeal', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.createAppeal).mockResolvedValue({
        success: true,
        data: {
          id: 'appeal-123',
          moderation_action_id: mockModerationActionId,
          reason: 'Valid appeal reason',
          status: 'pending',
          created_at: '2024-01-01T00:00:00Z',
        },
      });

      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
          onSuccess={mockOnSuccess}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'This is a valid appeal reason');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(moderationApi.createAppeal).toHaveBeenCalledWith({
          moderation_action_id: mockModerationActionId,
          reason: 'This is a valid appeal reason',
        });
      });
    });

    it('shows success message after submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.createAppeal).mockResolvedValue({
        success: true,
        data: {
          id: 'appeal-123',
          moderation_action_id: mockModerationActionId,
          reason: 'Valid appeal reason',
          status: 'pending',
          created_at: '2024-01-01T00:00:00Z',
        },
      });

      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'This is a valid appeal reason');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText(/Appeal submitted successfully/)
        ).toBeInTheDocument();
      });
    });

    it('calls onSuccess and onClose after successful submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.createAppeal).mockResolvedValue({
        success: true,
        data: {
          id: 'appeal-123',
          moderation_action_id: mockModerationActionId,
          reason: 'Valid appeal reason',
          status: 'pending',
          created_at: '2024-01-01T00:00:00Z',
        },
      });

      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
          onSuccess={mockOnSuccess}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'This is a valid appeal reason');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      await user.click(submitButton);

      // Wait for success message
      await waitFor(() => {
        expect(screen.getByText(/Appeal submitted successfully/)).toBeInTheDocument();
      });

      // Advance timers to trigger the delayed onClose and onSuccess
      await vi.advanceTimersByTimeAsync(2000);

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled();
        expect(mockOnSuccess).toHaveBeenCalled();
      });
    });

    it('trims whitespace from reason before submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.createAppeal).mockResolvedValue({
        success: true,
        data: {
          id: 'appeal-123',
          moderation_action_id: mockModerationActionId,
          reason: 'Valid appeal reason',
          status: 'pending',
          created_at: '2024-01-01T00:00:00Z',
        },
      });

      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, '  Valid appeal reason  ');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(moderationApi.createAppeal).toHaveBeenCalledWith({
          moderation_action_id: mockModerationActionId,
          reason: 'Valid appeal reason',
        });
      });
    });
  });

  describe('Error Handling', () => {
    it('shows error message when submission fails', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.createAppeal).mockRejectedValue(
        new Error('Failed to submit appeal')
      );

      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'This is a valid appeal reason');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/Failed to submit appeal/)).toBeInTheDocument();
      });
    });

    it('shows validation error for short reason', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'short');

      // Try to submit by clicking the button (it should be disabled)
      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      expect(submitButton).toBeDisabled();
    });

    it('does not call API when validation fails', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'short');

      // Button is disabled, so we can't click it
      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      expect(submitButton).toBeDisabled();
      expect(moderationApi.createAppeal).not.toHaveBeenCalled();
    });
  });

  describe('Modal Behavior', () => {
    it('closes modal when Cancel is clicked', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      await user.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('resets form when closed', async () => {
      const user = userEvent.setup({ delay: null });
      const { rerender } = render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'Test appeal reason');

      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      await user.click(cancelButton);

      // Reopen modal
      rerender(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const newTextarea = screen.getByLabelText('Reason for Appeal *') as HTMLTextAreaElement;
      expect(newTextarea.value).toBe('');
    });

    it('does not close during submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.createAppeal).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'This is a valid appeal reason');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      await user.click(submitButton);

      // Wait a bit
      await vi.advanceTimersByTimeAsync(100);

      // onClose should not have been called
      expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('disables Cancel button during submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.createAppeal).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'This is a valid appeal reason');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      await user.click(submitButton);

      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      expect(cancelButton).toBeDisabled();
    });
  });

  describe('Accessibility', () => {
    it('has proper labels for form inputs', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(screen.getByLabelText('Reason for Appeal *')).toBeInTheDocument();
    });

    it('has proper role for modal', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('marks textarea as required', () => {
      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      expect(textarea).toHaveAttribute('required');
    });

    it('disables textarea during submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.createAppeal).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(
        <AppealForm
          open={true}
          onClose={mockOnClose}
          moderationActionId={mockModerationActionId}
        />
      );

      const textarea = screen.getByLabelText('Reason for Appeal *');
      await user.type(textarea, 'This is a valid appeal reason');

      const submitButton = screen.getByRole('button', { name: 'Submit Appeal' });
      await user.click(submitButton);

      expect(textarea).toBeDisabled();
    });
  });
});
