import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AppealResolutionModal } from './AppealResolutionModal';
import * as moderationApi from '../../lib/moderation-api';

// Mock the API module
vi.mock('../../lib/moderation-api');

describe('AppealResolutionModal', () => {
  const mockOnClose = vi.fn();
  const mockOnSuccess = vi.fn();
  const mockAppeal: moderationApi.ModerationAppeal = {
    id: 'appeal-123',
    moderation_action_id: 'action-456',
    user_id: 'user-789',
    username: 'testuser',
    display_name: 'Test User',
    reason: 'I believe this ban was unfair because I did not violate any rules.',
    status: 'pending',
    created_at: '2024-01-01T00:00:00Z',
    content_type: 'clip',
    content_id: 'clip-123',
    decision_action: 'ban',
    decision_reason: 'Spam violations',
  };

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
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByText('Resolve Appeal')).toBeInTheDocument();
    });

    it('does not render when closed', () => {
      render(
        <AppealResolutionModal
          open={false}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.queryByText('Resolve Appeal')).not.toBeInTheDocument();
    });

    it('renders appeal details', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByText(/testuser/)).toBeInTheDocument();
      expect(screen.getByText(/Test User/)).toBeInTheDocument();
      expect(screen.getAllByText(/ban/).length).toBeGreaterThan(0);
      expect(screen.getByText(/Spam violations/)).toBeInTheDocument();
      expect(screen.getByText(/I believe this ban was unfair/)).toBeInTheDocument();
    });

    it('renders decision buttons', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByRole('radio', { name: /Approve Appeal/ })).toBeInTheDocument();
      expect(screen.getByRole('radio', { name: /Reject Appeal/ })).toBeInTheDocument();
    });

    it('renders resolution textarea', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByLabelText('Resolution Explanation (Optional)')).toBeInTheDocument();
    });

    it('renders Submit Decision button', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByRole('button', { name: 'Submit Decision' })).toBeInTheDocument();
    });

    it('renders Cancel button', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
    });
  });

  describe('Decision Selection', () => {
    it('selects approve decision when clicked', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      expect(approveButton).toHaveAttribute('aria-checked', 'true');
    });

    it('selects reject decision when clicked', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const rejectButton = screen.getByRole('radio', { name: /Reject Appeal/ });
      await user.click(rejectButton);

      expect(rejectButton).toHaveAttribute('aria-checked', 'true');
    });

    it('can change decision from approve to reject', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      const rejectButton = screen.getByRole('radio', { name: /Reject Appeal/ });

      await user.click(approveButton);
      expect(approveButton).toHaveAttribute('aria-checked', 'true');

      await user.click(rejectButton);
      expect(rejectButton).toHaveAttribute('aria-checked', 'true');
      expect(approveButton).toHaveAttribute('aria-checked', 'false');
    });

    it('disables Submit button when no decision is selected', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      expect(submitButton).toBeDisabled();
    });

    it('enables Submit button when decision is selected', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      expect(submitButton).not.toBeDisabled();
    });
  });

  describe('Form Submission', () => {
    it('successfully resolves appeal with approve decision', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockResolvedValue({
        success: true,
        data: {
          ...mockAppeal,
          status: 'approved',
          resolution: 'Appeal approved',
          resolved_at: '2024-01-02T00:00:00Z',
        },
      });

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
          onSuccess={mockOnSuccess}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(moderationApi.resolveAppeal).toHaveBeenCalledWith(mockAppeal.id, {
          decision: 'approve',
          resolution: undefined,
        });
      });
    });

    it('successfully resolves appeal with reject decision', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockResolvedValue({
        success: true,
        data: {
          ...mockAppeal,
          status: 'rejected',
          resolution: 'Appeal rejected',
          resolved_at: '2024-01-02T00:00:00Z',
        },
      });

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
          onSuccess={mockOnSuccess}
        />
      );

      const rejectButton = screen.getByRole('radio', { name: /Reject Appeal/ });
      await user.click(rejectButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(moderationApi.resolveAppeal).toHaveBeenCalledWith(mockAppeal.id, {
          decision: 'reject',
          resolution: undefined,
        });
      });
    });

    it('includes resolution text when provided', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockResolvedValue({
        success: true,
        data: {
          ...mockAppeal,
          status: 'approved',
          resolution: 'After review, we agree this was an error',
          resolved_at: '2024-01-02T00:00:00Z',
        },
      });

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const resolutionTextarea = screen.getByLabelText('Resolution Explanation (Optional)');
      await user.type(resolutionTextarea, 'After review, we agree this was an error');

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(moderationApi.resolveAppeal).toHaveBeenCalledWith(mockAppeal.id, {
          decision: 'approve',
          resolution: 'After review, we agree this was an error',
        });
      });
    });

    it('shows success message after resolution', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockResolvedValue({
        success: true,
        data: {
          ...mockAppeal,
          status: 'approved',
          resolved_at: '2024-01-02T00:00:00Z',
        },
      });

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/Appeal resolved successfully/)).toBeInTheDocument();
      });
    });

    it('calls onSuccess and onClose after successful resolution', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockResolvedValue({
        success: true,
        data: {
          ...mockAppeal,
          status: 'approved',
          resolved_at: '2024-01-02T00:00:00Z',
        },
      });

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
          onSuccess={mockOnSuccess}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/Appeal resolved successfully/)).toBeInTheDocument();
      });

      // Advance timers to trigger the delayed onClose and onSuccess
      await act(async () => {
        await vi.advanceTimersByTimeAsync(2000);
      });

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled();
        expect(mockOnSuccess).toHaveBeenCalled();
      });
    });

    it('trims whitespace from resolution before submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockResolvedValue({
        success: true,
        data: {
          ...mockAppeal,
          status: 'approved',
          resolved_at: '2024-01-02T00:00:00Z',
        },
      });

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const resolutionTextarea = screen.getByLabelText('Resolution Explanation (Optional)');
      await user.type(resolutionTextarea, '  Trimmed resolution  ');

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(moderationApi.resolveAppeal).toHaveBeenCalledWith(mockAppeal.id, {
          decision: 'approve',
          resolution: 'Trimmed resolution',
        });
      });
    });
  });

  describe('Error Handling', () => {
    it('shows error message when resolution fails', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockRejectedValue(
        new Error('Failed to resolve appeal')
      );

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/Failed to resolve appeal/)).toBeInTheDocument();
      });
    });

    it('shows validation error when no decision is selected', async () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      // Try to submit without selecting a decision
      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      expect(submitButton).toBeDisabled();
    });

    it('does not call API when validation fails', async () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      // Button is disabled, so we can't click it
      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      expect(submitButton).toBeDisabled();
      expect(moderationApi.resolveAppeal).not.toHaveBeenCalled();
    });
  });

  describe('Modal Behavior', () => {
    it('closes modal when Cancel is clicked', async () => {
      const user = userEvent.setup({ delay: null });
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      await user.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('resets form when closed', async () => {
      const user = userEvent.setup({ delay: null });
      const { rerender } = render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const resolutionTextarea = screen.getByLabelText('Resolution Explanation (Optional)');
      await user.type(resolutionTextarea, 'Test resolution');

      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      await user.click(cancelButton);

      // Reopen modal
      rerender(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const newApproveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      expect(newApproveButton).toHaveAttribute('aria-checked', 'false');

      const newResolutionTextarea = screen.getByLabelText('Resolution Explanation (Optional)') as HTMLTextAreaElement;
      expect(newResolutionTextarea.value).toBe('');
    });

    it('does not close during submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      // Wait a bit
      await vi.advanceTimersByTimeAsync(100);

      // onClose should not have been called
      expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('disables Cancel button during submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      expect(cancelButton).toBeDisabled();
    });
  });

  describe('Accessibility', () => {
    it('has proper labels for form inputs', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByLabelText('Resolution Explanation (Optional)')).toBeInTheDocument();
    });

    it('has proper role for modal', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('has proper radiogroup role for decision buttons', () => {
      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      expect(screen.getByRole('radiogroup')).toBeInTheDocument();
    });

    it('disables decision buttons during submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      const rejectButton = screen.getByRole('radio', { name: /Reject Appeal/ });
      expect(approveButton).toBeDisabled();
      expect(rejectButton).toBeDisabled();
    });

    it('disables resolution textarea during submission', async () => {
      const user = userEvent.setup({ delay: null });
      vi.mocked(moderationApi.resolveAppeal).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(
        <AppealResolutionModal
          open={true}
          onClose={mockOnClose}
          appeal={mockAppeal}
        />
      );

      const approveButton = screen.getByRole('radio', { name: /Approve Appeal/ });
      await user.click(approveButton);

      const submitButton = screen.getByRole('button', { name: 'Submit Decision' });
      await user.click(submitButton);

      const resolutionTextarea = screen.getByLabelText('Resolution Explanation (Optional)');
      expect(resolutionTextarea).toBeDisabled();
    });
  });
});
