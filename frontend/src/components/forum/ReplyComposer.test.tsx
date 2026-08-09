import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ReplyComposer } from './ReplyComposer';
import { allowTestConsole } from '../../test/setup';

describe('ReplyComposer', () => {
  const mockOnSubmit = vi.fn();
  const mockOnCancel = vi.fn();

  beforeEach(() => {
    mockOnSubmit.mockClear();
    mockOnCancel.mockClear();
  });

  it('renders textarea and buttons', () => {
    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const textarea = screen.getByRole('textbox', { name: 'Reply' });
    expect(textarea).toHaveAccessibleDescription(/Markdown formatting is supported/i);
    expect(screen.getByRole('button', { name: 'Post Reply' })).toBeInTheDocument();
  });

  it('allows typing in textarea', () => {
    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const textarea = screen.getByPlaceholderText(/Write your reply/i);
    fireEvent.change(textarea, { target: { value: 'Test reply content' } });

    expect(textarea).toHaveValue('Test reply content');
  });

  it('disables submit button when content is empty', () => {
    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const submitButton = screen.getByText('Post Reply');
    expect(submitButton).toBeDisabled();
  });

  it('enables submit button when content is provided', () => {
    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const textarea = screen.getByPlaceholderText(/Write your reply/i);
    const submitButton = screen.getByText('Post Reply');

    fireEvent.change(textarea, { target: { value: 'Test reply' } });

    expect(submitButton).not.toBeDisabled();
  });

  it('calls onSubmit with content when form is submitted', async () => {
    mockOnSubmit.mockResolvedValue(undefined);

    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const textarea = screen.getByPlaceholderText(/Write your reply/i);
    const submitButton = screen.getByText('Post Reply');

    fireEvent.change(textarea, { target: { value: 'Test reply content' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith('Test reply content');
    });
  });

  it('clears textarea after successful submission', async () => {
    mockOnSubmit.mockResolvedValue(undefined);

    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const textarea = screen.getByPlaceholderText(/Write your reply/i) as HTMLTextAreaElement;
    const submitButton = screen.getByText('Post Reply');

    fireEvent.change(textarea, { target: { value: 'Test reply' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(textarea.value).toBe('');
    });
  });

  it('shows cancel button when onCancel is provided', () => {
    render(<ReplyComposer onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

    expect(screen.getByText('Cancel')).toBeInTheDocument();
  });

  it('calls onCancel when cancel button is clicked', () => {
    render(<ReplyComposer onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

    const cancelButton = screen.getByText('Cancel');
    fireEvent.click(cancelButton);

    expect(mockOnCancel).toHaveBeenCalled();
  });

  it('shows loading state during submission', async () => {
    mockOnSubmit.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const textarea = screen.getByPlaceholderText(/Write your reply/i);
    const submitButton = screen.getByText('Post Reply');

    fireEvent.change(textarea, { target: { value: 'Test reply' } });
    fireEvent.click(submitButton);

    expect(screen.getByText('Posting...')).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
    expect(submitButton).toHaveAttribute('aria-busy', 'true');

    // Wait for the async operation to complete to prevent unmount during state update
    await waitFor(() => {
      expect(screen.getByText('Post Reply')).toBeInTheDocument();
    });
  });

  it('announces submission errors and associates them with the textarea', async () => {
    allowTestConsole('error', /Failed to post reply:.*network unavailable/);
    mockOnSubmit.mockRejectedValue(new Error('network unavailable'));
    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const textarea = screen.getByRole('textbox', { name: 'Reply' });
    fireEvent.change(textarea, { target: { value: 'Test reply' } });
    fireEvent.click(screen.getByRole('button', { name: 'Post Reply' }));

    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent(/could not be posted/i);
    expect(textarea).toHaveAttribute('aria-invalid', 'true');
    expect(textarea).toHaveAccessibleDescription(/could not be posted/i);
  });

  it('uses custom placeholder when provided', () => {
    render(
      <ReplyComposer
        onSubmit={mockOnSubmit}
        placeholder="Custom placeholder"
      />
    );

    expect(screen.getByPlaceholderText('Custom placeholder')).toBeInTheDocument();
  });

  it('uses custom submit label when provided', () => {
    render(
      <ReplyComposer onSubmit={mockOnSubmit} submitLabel="Submit" />
    );

    expect(screen.getByText('Submit')).toBeInTheDocument();
  });

  it('applies mobile styles when isMobile is true', () => {
    const { container } = render(
      <ReplyComposer onSubmit={mockOnSubmit} isMobile={true} />
    );

    const form = container.querySelector('form');
    expect(form?.className).toContain('rounded-t-lg');
  });

  it('trims whitespace from content before submission', async () => {
    mockOnSubmit.mockResolvedValue(undefined);

    render(<ReplyComposer onSubmit={mockOnSubmit} />);

    const textarea = screen.getByPlaceholderText(/Write your reply/i);
    const submitButton = screen.getByText('Post Reply');

    fireEvent.change(textarea, { target: { value: '  Test reply  ' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith('Test reply');
    });
  });
});
