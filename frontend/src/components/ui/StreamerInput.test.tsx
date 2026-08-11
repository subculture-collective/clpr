import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { StreamerInput } from './StreamerInput';

// Mock the search API
vi.mock('../../lib/search-api', () => ({
    searchApi: {
        getSuggestions: vi.fn(),
    },
}));

describe('StreamerInput', () => {
    const mockOnChange = vi.fn();

    it('renders with default props', () => {
        render(<StreamerInput value="" onChange={mockOnChange} />);

        const input = screen.getByRole('textbox');
        expect(input).toBeInTheDocument();
        expect(input).toHaveAttribute('placeholder', 'Enter creator name...');
    });

    it('shows auto-detected badge when autoDetected is true', () => {
        render(
            <StreamerInput
                value=""
                onChange={mockOnChange}
                autoDetected={true}
            />
        );

        expect(
            screen.getByText('Will be auto-detected from clip')
        ).toBeInTheDocument();
    });

    it('does not show badge when autoDetected is false', () => {
        render(
            <StreamerInput
                value=""
                onChange={mockOnChange}
                autoDetected={false}
            />
        );

        expect(
            screen.queryByText('Will be auto-detected from clip')
        ).not.toBeInTheDocument();
    });

    it('calls onChange when user types', async () => {
        const user = userEvent.setup();
        render(<StreamerInput value="" onChange={mockOnChange} />);

        const input = screen.getByRole('textbox');
        await user.type(input, 'x');

        expect(mockOnChange).toHaveBeenCalled();
    });

    it('displays the provided value', () => {
        render(<StreamerInput value="xQc" onChange={mockOnChange} />);

        const input = screen.getByRole('textbox');
        expect(input).toHaveValue('xQc');
    });

    it('respects disabled prop', () => {
        render(
            <StreamerInput value="" onChange={mockOnChange} disabled={true} />
        );

        const input = screen.getByRole('textbox');
        expect(input).toBeDisabled();
    });

    it('respects required prop', () => {
        render(
            <StreamerInput value="" onChange={mockOnChange} required={true} />
        );

        const input = screen.getByRole('textbox');
        expect(input).toBeRequired();
    });

    it('shows appropriate helper text based on autoDetected state', () => {
        const { rerender } = render(
            <StreamerInput
                value=""
                onChange={mockOnChange}
                autoDetected={false}
            />
        );

        expect(
            screen.getByText('Type to search for creators or enter manually')
        ).toBeInTheDocument();

        rerender(
            <StreamerInput
                value=""
                onChange={mockOnChange}
                autoDetected={true}
            />
        );

        expect(
            screen.getByText(
                'Creator will be detected from the clip URL. You can override by typing here.'
            )
        ).toBeInTheDocument();
    });

    it('uses custom id when provided', () => {
        render(
            <StreamerInput
                value=""
                onChange={mockOnChange}
                id="custom-streamer-id"
            />
        );

        const input = screen.getByRole('textbox');
        expect(input).toHaveAttribute('id', 'custom-streamer-id');
    });

    it('shows required asterisk when required is true', () => {
        render(
            <StreamerInput value="" onChange={mockOnChange} required={true} />
        );

        expect(screen.getByText('*')).toBeInTheDocument();
    });
});
