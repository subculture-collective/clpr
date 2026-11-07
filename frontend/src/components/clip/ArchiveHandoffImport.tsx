import { useId, useState } from 'react';
import { Alert, Button } from '../ui';
import {
    ArchiveHandoffError,
    formatArchiveMs,
    parseArchiveHandoff,
    type ArchiveCaptionCue,
    type ArchiveHandoff,
    type ArchiveImportDraft,
} from '../../lib/archive-handoff';

export interface ArchiveHandoffImportProps {
    /** The import already in the draft, if any. */
    current: ArchiveImportDraft | null;
    /** True when the draft has content an import would replace. */
    draftHasContent: boolean;
    disabled?: boolean;
    onApply: (importDraft: ArchiveImportDraft) => void;
    onRemove: () => void;
    onCuesChange: (cues: ArchiveCaptionCue[]) => void;
}

function cueRange(cue: ArchiveCaptionCue): string {
    return cue.end_ms === null ?
            `${formatArchiveMs(cue.start_ms)} – next cue`
        :   `${formatArchiveMs(cue.start_ms)} – ${formatArchiveMs(cue.end_ms)}`;
}

function sameCues(a: ArchiveCaptionCue[], b: ArchiveCaptionCue[]): boolean {
    return (
        a.length === b.length &&
        a.every(
            (cue, i) =>
                cue.start_ms === b[i].start_ms &&
                cue.end_ms === b[i].end_ms &&
                cue.text === b[i].text,
        )
    );
}

function HandoffSummary({ handoff }: { handoff: ArchiveHandoff }) {
    const { passage, transcript, captions, media } = handoff;
    return (
        <dl className='grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1 text-sm'>
            <dt className='text-muted-foreground'>Archive</dt>
            <dd>{handoff.archive.name}</dd>
            <dt className='text-muted-foreground'>Passage</dt>
            <dd>
                {formatArchiveMs(passage.start_ms)} – {formatArchiveMs(passage.end_ms)} of the
                recording ({formatArchiveMs(passage.duration_ms)})
            </dd>
            <dt className='text-muted-foreground'>Transcript</dt>
            <dd>
                {transcript.status === 'captured' ?
                    <>
                        {transcript.source} ({transcript.kind}
                        {transcript.language ? `, ${transcript.language}` : ''}), revision{' '}
                        <code>{transcript.revision_sha256?.slice(0, 12)}</code>
                    </>
                :   'No transcript was available'}
            </dd>
            <dt className='text-muted-foreground'>Captions</dt>
            <dd>
                {captions.cues.length} cue{captions.cues.length === 1 ? '' : 's'}
                {captions.cues_truncated ? ' (the archive cut the list short)' : ''}
            </dd>
            <dt className='text-muted-foreground'>Video</dt>
            <dd className='break-all'>
                {media.filename} ({media.bytes.toLocaleString()} bytes)
            </dd>
        </dl>
    );
}

/**
 * Imports a Rekolekt clip handoff into the submission draft. Importing only
 * fills the draft; the submitter still chooses the MP4 and submits.
 */
export function ArchiveHandoffImport({
    current,
    draftHasContent,
    disabled = false,
    onApply,
    onRemove,
    onCuesChange,
}: ArchiveHandoffImportProps) {
    const inputId = useId();
    const [pending, setPending] = useState<{ raw: string; handoff: ArchiveHandoff } | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [inputKey, setInputKey] = useState(0);

    const handleFile = async (file: File | null) => {
        setError(null);
        setPending(null);
        if (!file) return;
        try {
            const raw = await file.text();
            setPending({ raw, handoff: await parseArchiveHandoff(raw) });
        } catch (err) {
            setError(
                err instanceof ArchiveHandoffError ? err.message : 'The archive handoff could not be read',
            );
        } finally {
            setInputKey(key => key + 1);
        }
    };

    const apply = () => {
        if (!pending) return;
        onApply({
            raw: pending.raw,
            handoff: pending.handoff,
            cues: pending.handoff.captions.cues.map(cue => ({ ...cue })),
        });
        setPending(null);
    };

    const replacing = draftHasContent || current !== null;

    return (
        <div className='rounded-xl border border-border/70 p-4 space-y-3' data-testid='archive-handoff-import'>
            <div>
                <label htmlFor={inputId} className='block text-sm font-medium mb-1'>
                    Archive handoff (optional)
                </label>
                <p className='text-xs text-muted-foreground mb-2'>
                    If the clip came from an archive export, import the JSON file downloaded with it. It
                    adds the source passage and captions to this draft. You still choose the MP4 below,
                    and nothing is submitted until you press Submit.
                </p>
                <input
                    key={inputKey}
                    id={inputId}
                    type='file'
                    accept='application/json,.json'
                    disabled={disabled}
                    onChange={e => void handleFile(e.currentTarget.files?.[0] ?? null)}
                    className='block w-full text-sm file:mr-4 file:rounded-md file:border-0 file:bg-primary-500 file:px-4 file:py-2 file:text-sm file:font-medium file:text-white'
                />
            </div>

            {error && (
                <Alert variant='error' role='alert'>
                    {error}
                </Alert>
            )}

            {pending && (
                <div className='rounded-lg border border-primary-500/30 bg-primary-500/5 p-3 space-y-3' data-testid='archive-handoff-review'>
                    <p className='text-sm font-medium'>Review the archive handoff</p>
                    <HandoffSummary handoff={pending.handoff} />
                    <p className='text-xs text-muted-foreground'>{pending.handoff.provenance.statement}</p>
                    {replacing && (
                        <Alert variant='warning'>
                            Your draft already has content. Replacing it clears the title, reason, tags
                            and any earlier archive import.
                        </Alert>
                    )}
                    <div className='flex gap-2'>
                        <Button type='button' size='sm' onClick={apply} disabled={disabled}>
                            {replacing ? 'Replace draft' : 'Use in draft'}
                        </Button>
                        <Button type='button' size='sm' variant='ghost' onClick={() => setPending(null)}>
                            Cancel
                        </Button>
                    </div>
                </div>
            )}

            {current && !pending && (
                <div className='space-y-3' data-testid='archive-handoff-current'>
                    <HandoffSummary handoff={current.handoff} />
                    <p className='text-sm'>
                        <a
                            href={current.handoff.passage.url}
                            target='_blank'
                            rel='noopener noreferrer nofollow'
                            className='text-link underline'
                        >
                            Open the source passage
                        </a>
                    </p>
                    <p className='text-xs text-muted-foreground'>
                        Archive provenance is declared by the archive, not verified rights. Moderators
                        review this clip like any other upload.
                    </p>
                    <CaptionCueEditor
                        cues={current.cues}
                        sourceCues={current.handoff.captions.cues}
                        disabled={disabled}
                        onChange={onCuesChange}
                    />
                    <Button type='button' size='sm' variant='ghost' onClick={onRemove} disabled={disabled}>
                        Remove archive import
                    </Button>
                </div>
            )}
        </div>
    );
}

interface CaptionCueEditorProps {
    cues: ArchiveCaptionCue[];
    sourceCues: ArchiveCaptionCue[];
    disabled: boolean;
    onChange: (cues: ArchiveCaptionCue[]) => void;
}

function CaptionCueEditor({ cues, sourceCues, disabled, onChange }: CaptionCueEditorProps) {
    if (sourceCues.length === 0) {
        return <p className='text-sm text-muted-foreground'>The archive sent no captions for this clip.</p>;
    }
    const edited = !sameCues(cues, sourceCues);
    return (
        <fieldset className='space-y-2'>
            <legend className='text-sm font-medium'>Captions</legend>
            <p className='text-xs text-muted-foreground'>
                Times are relative to the clip. Your edits are kept apart from the archive's original
                captions.
            </p>
            {cues.length === 0 && <p className='text-sm text-muted-foreground'>All captions removed.</p>}
            <ol className='space-y-2'>
                {cues.map((cue, index) => (
                    <li key={`${cue.start_ms}-${index}`} className='flex items-start gap-2'>
                        <span className='w-36 shrink-0 pt-2 font-mono text-xs text-muted-foreground'>
                            {cueRange(cue)}
                        </span>
                        <textarea
                            aria-label={`Caption ${index + 1} text`}
                            value={cue.text}
                            disabled={disabled}
                            rows={1}
                            maxLength={2000}
                            onChange={e =>
                                onChange(
                                    cues.map((item, i) =>
                                        i === index ? { ...item, text: e.target.value } : item,
                                    ),
                                )
                            }
                            className='min-w-0 flex-1 rounded-md border border-border bg-background px-2 py-1 text-sm'
                        />
                        <Button
                            type='button'
                            size='sm'
                            variant='ghost'
                            aria-label={`Remove caption ${index + 1}`}
                            disabled={disabled}
                            onClick={() => onChange(cues.filter((_, i) => i !== index))}
                        >
                            Remove
                        </Button>
                    </li>
                ))}
            </ol>
            {edited && (
                <Button
                    type='button'
                    size='sm'
                    variant='ghost'
                    disabled={disabled}
                    onClick={() => onChange(sourceCues.map(cue => ({ ...cue })))}
                >
                    Restore the archive's captions
                </Button>
            )}
        </fieldset>
    );
}
