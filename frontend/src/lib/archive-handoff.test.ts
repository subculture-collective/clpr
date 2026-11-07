import { createHash } from 'node:crypto';
import { describe, expect, it } from 'vitest';
import fixture from '../test/fixtures/rekolekt-clip-handoff.v1.json';
import {
    ArchiveHandoffError,
    archiveProvenanceFromMetadata,
    checkArchiveMedia,
    cueTextSHA256,
    parseArchiveHandoff,
    validateCaptionCues,
} from './archive-handoff';

// The fixture is the producer's docs/api/examples/clip-handoff.v1.json from
// rekolekt a4d37c304e54ecfbf8822d2c6162d4f88b414483.
type Doc = Record<string, any>; // eslint-disable-line @typescript-eslint/no-explicit-any

function mutated(change: (doc: Doc) => void): string {
    const doc = structuredClone(fixture) as Doc;
    change(doc);
    return JSON.stringify(doc);
}

async function rejection(raw: string): Promise<string> {
    try {
        await parseArchiveHandoff(raw);
    } catch (err) {
        expect(err).toBeInstanceOf(ArchiveHandoffError);
        return (err as Error).message;
    }
    throw new Error('handoff was accepted');
}

describe('parseArchiveHandoff', () => {
    it('accepts the producer fixture', async () => {
        const handoff = await parseArchiveHandoff(JSON.stringify(fixture));
        expect(handoff.passage).toMatchObject({ start_ms: 754000, end_ms: 763000, duration_ms: 9000 });
        expect(handoff.captions.cues).toHaveLength(4);
        expect(handoff.captions.cues[2].end_ms).toBeNull();
    });

    it('recomputes the producer cue digest', async () => {
        expect(await cueTextSHA256(fixture.captions.cues)).toBe(fixture.captions.text_sha256);
    });

    it.each([
        ['an unsupported version', (d: Doc) => (d.schema_version = 2), 'Unsupported archive handoff version'],
        ['another schema', (d: Doc) => (d.schema = 'other'), 'not a rekolekt.clip-handoff'],
        ['an inverted range', (d: Doc) => (d.passage.end_ms = 1), 'invalid passage range'],
        ['a duration mismatch', (d: Doc) => (d.passage.duration_ms = 1), 'duration does not match'],
        ['a javascript passage URL', (d: Doc) => (d.passage.url = 'javascript:alert(1)'), 'must be an https URL'],
        ['a passage on another origin', (d: Doc) => (d.passage.player_url = 'https://evil.example/v/x'), "not on the archive's origin"],
        ['credentials in a URL', (d: Doc) => (d.passage.url = 'https://u:p@archive.example/x'), 'must not contain credentials'],
        ['an edited cue', (d: Doc) => (d.captions.cues[0].text = 'changed'), 'do not match their text_sha256'],
        ['a cue past the clip', (d: Doc) => (d.captions.cues[3].end_ms = 9001), 'Caption cue 4 has an end time'],
        ['an oversize cue', (d: Doc) => (d.captions.cues[0].text = 'x'.repeat(2001)), 'longer than 2000 characters'],
        ['a non-mp4 media type', (d: Doc) => (d.media.type = 'video/webm'), 'must be video/mp4'],
        ['a verified basis', (d: Doc) => (d.provenance.basis = 'verified'), 'basis must be "declared"'],
    ])('rejects %s', async (_name, change, message) => {
        expect(await rejection(mutated(change))).toContain(message);
    });

    it('rejects malformed and oversize documents', async () => {
        expect(await rejection('{nope')).toContain('not valid JSON');
        expect(await rejection(' '.repeat(512 * 1024 + 1))).toContain('larger than');
    });
});

describe('validateCaptionCues', () => {
    it('checks order and bounds', () => {
        expect(validateCaptionCues(fixture.captions.cues, 9000)).toBeNull();
        expect(
            validateCaptionCues(
                [
                    { start_ms: 500, end_ms: 900, text: 'a' },
                    { start_ms: 100, end_ms: 200, text: 'b' },
                ],
                9000,
            ),
        ).toContain('starts before');
    });
});

describe('checkArchiveMedia', () => {
    it('matches only the exact bytes the handoff names', async () => {
        const bytes = new TextEncoder().encode('rendered passage mp4');
        const raw = mutated(d => {
            d.media.bytes = bytes.length;
            d.media.sha256 = createHash('sha256').update(bytes).digest('hex');
        });
        const handoff = await parseArchiveHandoff(raw);

        expect(await checkArchiveMedia(new File([bytes], 'passage.mp4', { type: 'video/mp4' }), handoff)).toBeNull();

        const sameSize = new TextEncoder().encode('rendered passage mp5');
        expect(await checkArchiveMedia(new File([sameSize], 'passage.mp4'), handoff)).toContain(
            'does not match the archive handoff',
        );
        expect(await checkArchiveMedia(new File(['short'], 'passage.mp4'), handoff)).toContain(
            'does not match the archive handoff',
        );
    });
});

describe('archiveProvenanceFromMetadata', () => {
    it('reads the stored record and ignores other metadata', () => {
        const record = {
            export_id: fixture.export_id,
            archive: fixture.archive,
            passage: fixture.passage,
            captions: { source_cues: fixture.captions.cues, cues: fixture.captions.cues, edited: false },
        };
        expect(archiveProvenanceFromMetadata({ archive_provenance: record })?.passage.url).toBe(fixture.passage.url);
        expect(archiveProvenanceFromMetadata({ storage_key: 'x' })).toBeNull();
        expect(
            archiveProvenanceFromMetadata({
                archive_provenance: { ...record, passage: { ...record.passage, url: 'javascript:alert(1)' } },
            }),
        ).toBeNull();
    });
});
