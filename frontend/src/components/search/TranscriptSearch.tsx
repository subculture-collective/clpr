import { useState, type FormEvent } from 'react';
import { MessageSquareQuote, X } from 'lucide-react';

const MATCH_START = '';
const MATCH_END = '';

/**
 * The part of a clip's transcript that matched a search (#87). The API marks
 * each matched word between U+E000 and U+E001; they are shown as <mark>.
 */
export function TranscriptSnippet({ snippet }: { snippet: string }) {
  const parts: { text: string; match: boolean }[] = [];
  for (const piece of snippet.split(MATCH_START)) {
    const end = piece.indexOf(MATCH_END);
    if (end === -1) {
      if (piece) parts.push({ text: piece, match: false });
      continue;
    }
    parts.push({ text: piece.slice(0, end), match: true });
    if (end + 1 < piece.length) parts.push({ text: piece.slice(end + 1), match: false });
  }
  return (
    <p className='mt-2 flex gap-2 text-sm text-muted-foreground' data-testid='transcript-snippet'>
      <MessageSquareQuote size={16} className='mt-0.5 shrink-0' aria-hidden='true' />
      <span>
        <span className='sr-only'>Said in the clip: </span>
        &hellip;
        {parts.map((part, i) =>
          part.match ? (
            <mark key={i} className='bg-brand/25 px-0.5 text-foreground'>
              {part.text}
            </mark>
          ) : (
            <span key={i}>{part.text}</span>
          ),
        )}
        &hellip;
      </span>
    </p>
  );
}

/** A phrase that must have been said in the clip, matched against transcripts. */
export function SaidFilter({ value, onChange }: { value?: string; onChange: (said?: string) => void }) {
  const [draft, setDraft] = useState(value ?? '');
  const [shown, setShown] = useState(value);
  // Follow a value changed from outside, for example by the URL.
  if (shown !== value) {
    setShown(value);
    setDraft(value ?? '');
  }
  const submit = (event: FormEvent) => {
    event.preventDefault();
    onChange(draft.trim() || undefined);
  };
  return (
    <form onSubmit={submit} className='mb-6 flex flex-wrap items-center gap-2' role='search' aria-label='Search what was said'>
      <label htmlFor='said-filter' className='text-sm font-medium'>
        Said in the clip
      </label>
      <input
        id='said-filter'
        type='search'
        value={draft}
        maxLength={200}
        onChange={event => setDraft(event.target.value)}
        placeholder='a phrase someone said'
        className='min-h-11 min-w-0 flex-1 border border-border bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500'
      />
      <button type='submit' className='min-h-11 border border-border px-4 text-sm font-medium hover:border-primary'>
        Find
      </button>
      {value && (
        <button
          type='button'
          onClick={() => onChange(undefined)}
          className='inline-flex min-h-11 items-center gap-1 px-2 text-sm text-muted-foreground hover:text-foreground'
        >
          <X size={14} aria-hidden='true' /> Clear
        </button>
      )}
    </form>
  );
}
