import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { SaidFilter, TranscriptSnippet } from './TranscriptSearch';

describe('TranscriptSnippet', () => {
  it('marks the matched words', () => {
    render(<TranscriptSnippet snippet={'that was the cleanest ace ever'} />);
    const marks = screen.getAllByText(/cleanest|ace/, { selector: 'mark' });
    expect(marks.map(m => m.textContent)).toEqual(['cleanest', 'ace']);
    expect(screen.getByTestId('transcript-snippet')).toHaveTextContent('Said in the clip: …that was the cleanest ace ever…');
  });

  it('shows text without marks as is', () => {
    render(<TranscriptSnippet snippet='<b>not html</b>' />);
    expect(screen.getByTestId('transcript-snippet')).toHaveTextContent('<b>not html</b>');
  });
});

describe('SaidFilter', () => {
  it('submits the trimmed phrase and clears it', () => {
    const onChange = vi.fn();
    const { rerender } = render(<SaidFilter onChange={onChange} />);
    fireEvent.change(screen.getByLabelText('Said in the clip'), { target: { value: '  no way  ' } });
    fireEvent.click(screen.getByRole('button', { name: 'Find' }));
    expect(onChange).toHaveBeenLastCalledWith('no way');

    rerender(<SaidFilter value='no way' onChange={onChange} />);
    fireEvent.click(screen.getByRole('button', { name: /Clear/ }));
    expect(onChange).toHaveBeenLastCalledWith(undefined);

    rerender(<SaidFilter value='other words' onChange={onChange} />);
    expect(screen.getByLabelText('Said in the clip')).toHaveValue('other words');
  });
});
