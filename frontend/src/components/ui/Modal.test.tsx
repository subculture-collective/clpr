import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Modal } from './Modal';

describe('Modal', () => {
    const mockOnClose = vi.fn();
    const originalOverflow = document.body.style.overflow;

    beforeEach(() => {
        mockOnClose.mockClear();
        // Reset body overflow before each test
        document.body.style.overflow = 'unset';
    });

    afterEach(() => {
        // Restore original overflow after each test
        document.body.style.overflow = originalOverflow;
    });

    it('renders when open is true', () => {
        render(
            <Modal open={true} onClose={mockOnClose} title="Test Modal">
                <p>Modal content</p>
            </Modal>
        );

        expect(screen.getByText('Test Modal')).toBeInTheDocument();
        expect(screen.getByText('Modal content')).toBeInTheDocument();
        expect(screen.getByRole('dialog', { name: 'Test Modal' })).toBeInTheDocument();
    });

    it('has an accessible fallback name and focus target without a title or controls', () => {
        render(
            <Modal open={true} onClose={mockOnClose} showCloseButton={false}>
                <p>Informational content</p>
            </Modal>
        );

        const dialog = screen.getByRole('dialog', { name: 'Dialog' });
        expect(dialog).toHaveFocus();
        expect(dialog).toHaveAttribute('tabindex', '-1');
    });

    it('restores focus to the invoking control after close', () => {
        const trigger = document.createElement('button');
        document.body.appendChild(trigger);
        trigger.focus();
        const { rerender } = render(
            <Modal open={true} onClose={mockOnClose} title="Focus modal">
                <button>Inside</button>
            </Modal>
        );

        rerender(
            <Modal open={false} onClose={mockOnClose} title="Focus modal">
                <button>Inside</button>
            </Modal>
        );
        expect(trigger).toHaveFocus();
        trigger.remove();
    });

    it('does not render when open is false', () => {
        render(
            <Modal open={false} onClose={mockOnClose} title="Test Modal">
                <p>Modal content</p>
            </Modal>
        );

        expect(screen.queryByText('Test Modal')).not.toBeInTheDocument();
        expect(screen.queryByText('Modal content')).not.toBeInTheDocument();
    });

    it('sets body overflow to hidden when modal opens', () => {
        const { rerender } = render(
            <Modal open={false} onClose={mockOnClose}>
                <p>Content</p>
            </Modal>
        );

        expect(document.body.style.overflow).toBe('unset');

        rerender(
            <Modal open={true} onClose={mockOnClose}>
                <p>Content</p>
            </Modal>
        );

        expect(document.body.style.overflow).toBe('hidden');
    });

    it('resets body overflow when modal closes', () => {
        const { rerender } = render(
            <Modal open={true} onClose={mockOnClose}>
                <p>Content</p>
            </Modal>
        );

        expect(document.body.style.overflow).toBe('hidden');

        rerender(
            <Modal open={false} onClose={mockOnClose}>
                <p>Content</p>
            </Modal>
        );

        expect(document.body.style.overflow).toBe('unset');
    });

    it('resets body overflow when component unmounts while open', () => {
        const { unmount } = render(
            <Modal open={true} onClose={mockOnClose}>
                <p>Content</p>
            </Modal>
        );

        expect(document.body.style.overflow).toBe('hidden');

        // Unmount while modal is still open (simulates navigation)
        unmount();

        // Body overflow should be reset
        expect(document.body.style.overflow).toBe('unset');
    });

    it('calls onClose when close button is clicked', async () => {
        const user = userEvent.setup();
        render(
            <Modal open={true} onClose={mockOnClose} title="Test Modal">
                <p>Content</p>
            </Modal>
        );

        const closeButton = screen.getByLabelText('Close modal');
        await user.click(closeButton);

        expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('calls onClose when Escape key is pressed', async () => {
        const user = userEvent.setup();
        render(
            <Modal open={true} onClose={mockOnClose} title="Test Modal">
                <p>Content</p>
            </Modal>
        );

        await user.keyboard('{Escape}');

        expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('calls onClose when backdrop is clicked and closeOnBackdrop is true', async () => {
        const user = userEvent.setup();
        render(
            <Modal
                open={true}
                onClose={mockOnClose}
                closeOnBackdrop={true}
                title="Test Modal"
            >
                <p>Content</p>
            </Modal>
        );

        // Click on the backdrop (the container div)
        const backdrop = screen.getByRole('dialog').parentElement;
        if (backdrop) {
            await user.click(backdrop);
            expect(mockOnClose).toHaveBeenCalledTimes(1);
        }
    });

    it('does not call onClose when backdrop is clicked and closeOnBackdrop is false', async () => {
        const user = userEvent.setup();
        render(
            <Modal
                open={true}
                onClose={mockOnClose}
                closeOnBackdrop={false}
                title="Test Modal"
            >
                <p>Content</p>
            </Modal>
        );

        // Click on the backdrop (the container div)
        const backdrop = screen.getByRole('dialog').parentElement;
        if (backdrop) {
            await user.click(backdrop);
            expect(mockOnClose).not.toHaveBeenCalled();
        }
    });

    it('does not show close button when showCloseButton is false', () => {
        render(
            <Modal
                open={true}
                onClose={mockOnClose}
                showCloseButton={false}
                title="Test Modal"
            >
                <p>Content</p>
            </Modal>
        );

        expect(screen.queryByLabelText('Close modal')).not.toBeInTheDocument();
    });

    it('handles multiple modal mount/unmount cycles correctly', () => {
        const { rerender, unmount } = render(
            <Modal open={true} onClose={mockOnClose}>
                <p>Content</p>
            </Modal>
        );

        expect(document.body.style.overflow).toBe('hidden');

        // Close modal
        rerender(
            <Modal open={false} onClose={mockOnClose}>
                <p>Content</p>
            </Modal>
        );

        expect(document.body.style.overflow).toBe('unset');

        // Open modal again
        rerender(
            <Modal open={true} onClose={mockOnClose}>
                <p>Content</p>
            </Modal>
        );

        expect(document.body.style.overflow).toBe('hidden');

        // Unmount while open
        unmount();

        expect(document.body.style.overflow).toBe('unset');
    });
});
