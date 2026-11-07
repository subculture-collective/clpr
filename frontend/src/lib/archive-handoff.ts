/**
 * Rekolekt clip handoff, contract version 1 (`rekolekt.clip-handoff`).
 *
 * The archive exports an MP4 and this JSON document describing which
 * recording, range and transcript were selected. clpr reads the document into a
 * submission draft; the server repeats every check here and also matches the
 * uploaded bytes to `media.sha256`. The provenance is declared by the archive,
 * not verified rights or origin.
 */

import { apiClient } from './api';

export const ARCHIVE_HANDOFF_SCHEMA = 'rekolekt.clip-handoff';
export const ARCHIVE_HANDOFF_VERSION = 1;
export const MAX_ARCHIVE_HANDOFF_BYTES = 512 * 1024;
export const MAX_ARCHIVE_CAPTION_CUES = 200;
export const MAX_ARCHIVE_CUE_CHARS = 2000;

export interface ArchiveCaptionCue {
  start_ms: number;
  end_ms: number | null;
  text: string;
  text_truncated?: boolean;
}

export interface ArchiveHandoff {
  schema: typeof ARCHIVE_HANDOFF_SCHEMA;
  schema_version: typeof ARCHIVE_HANDOFF_VERSION;
  export_id: string;
  generated_at: number;
  archive: { name: string; origin: string };
  recording: { video_id: string; youtube_id: string | null };
  passage: {
    url: string;
    player_url: string;
    start_ms: number;
    end_ms: number;
    duration_ms: number;
  };
  transcript: {
    status: 'captured' | 'unavailable';
    source?: 'whisper' | 'youtube';
    transcript_id?: string | null;
    language?: string | null;
    kind?: 'machine' | 'caption-auto' | 'caption-manual';
    segment_count?: number;
    revision_sha256?: string;
    timing?: 'segment';
    captured_at?: number;
  };
  captions: {
    timing: 'clip-relative' | 'unavailable';
    cues: ArchiveCaptionCue[];
    cues_truncated: boolean;
    text_sha256: string | null;
  };
  media: { type: 'video/mp4'; filename: string; bytes: number; sha256: string };
  provenance: {
    basis: 'declared';
    statement: string;
    media_source: {
      kind: 'youtube-range' | 'registered-original';
      youtube_id?: string;
      original_sha256?: string;
      timeline_aligned?: boolean;
    };
  };
}

/** What a submission draft keeps for an archive import. */
export interface ArchiveImportDraft {
  /** The document exactly as imported; it is sent to the server unchanged. */
  raw: string;
  handoff: ArchiveHandoff;
  /** The submitter's cues; the archive's originals stay in `handoff`. */
  cues: ArchiveCaptionCue[];
}

/** The record the server stores under source_metadata.archive_provenance. */
export interface ArchiveProvenanceRecord {
  export_id: string;
  archive: { name: string; origin: string };
  passage: ArchiveHandoff['passage'];
  transcript: ArchiveHandoff['transcript'];
  captions: {
    source_cues: ArchiveCaptionCue[];
    cues: ArchiveCaptionCue[];
    edited: boolean;
    language?: string | null;
  };
  provenance: ArchiveHandoff['provenance'];
}

export class ArchiveHandoffError extends Error {}

const SHA256 = /^[0-9a-f]{64}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function fail(message: string): never {
  throw new ArchiveHandoffError(message);
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isInt(value: unknown): value is number {
  return typeof value === 'number' && Number.isInteger(value);
}

function text(value: unknown, field: string, max: number, required = true): string {
  if (typeof value !== 'string' || (required && value.trim() === '')) {
    fail(`The archive handoff is missing ${field}`);
  }
  if ([...value].length > max) {
    fail(`The archive handoff ${field} is longer than ${max} characters`);
  }
  return value;
}

function httpsURL(value: unknown, field: string): URL {
  if (typeof value !== 'string' || value === '' || value.length > 2048) {
    fail(`The archive handoff ${field} is missing or too long`);
  }
  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    fail(`The archive handoff ${field} must be an https URL`);
  }
  if (parsed.protocol !== 'https:' || !parsed.hostname) {
    fail(`The archive handoff ${field} must be an https URL`);
  }
  if (parsed.username || parsed.password) {
    fail(`The archive handoff ${field} must not contain credentials`);
  }
  return parsed;
}

/** Checks cue bounds against a clip of `durationMs`; returns the first problem. */
export function validateCaptionCues(cues: ArchiveCaptionCue[], durationMs: number): string | null {
  if (cues.length > MAX_ARCHIVE_CAPTION_CUES) {
    return `Captions can have at most ${MAX_ARCHIVE_CAPTION_CUES} cues`;
  }
  let previousStart = 0;
  for (const [index, cue] of cues.entries()) {
    const n = index + 1;
    if (!isInt(cue.start_ms) || cue.start_ms < 0 || cue.start_ms > durationMs) {
      return `Caption cue ${n} starts outside the clip`;
    }
    if (cue.start_ms < previousStart) {
      return `Caption cue ${n} starts before the cue above it`;
    }
    previousStart = cue.start_ms;
    if (
      cue.end_ms !== null &&
      (!isInt(cue.end_ms) || cue.end_ms <= cue.start_ms || cue.end_ms > durationMs)
    ) {
      return `Caption cue ${n} has an end time outside the clip or before its start`;
    }
    if (typeof cue.text !== 'string') {
      return `Caption cue ${n} has no text`;
    }
    if ([...cue.text].length > MAX_ARCHIVE_CUE_CHARS) {
      return `Caption cue ${n} is longer than ${MAX_ARCHIVE_CUE_CHARS} characters`;
    }
  }
  return null;
}

async function sha256Hex(data: BufferSource): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', data);
  return Array.from(new Uint8Array(digest), byte => byte.toString(16).padStart(2, '0')).join('');
}

/**
 * The producer's cue digest: SHA-256 over one compact JSON line
 * `[start_ms,end_ms,text]` per cue. JSON.stringify escapes exactly what
 * Python's json.dumps(ensure_ascii=False) does.
 */
export function cueTextSHA256(cues: ArchiveCaptionCue[]): Promise<string> {
  const lines = cues.map(cue => `${JSON.stringify([cue.start_ms, cue.end_ms, cue.text])}\n`).join('');
  return sha256Hex(new TextEncoder().encode(lines));
}

/** SHA-256 of a selected file, to compare with the handoff before upload. */
export async function fileSHA256(file: Blob): Promise<string> {
  return sha256Hex(await file.arrayBuffer());
}

/**
 * Parses and validates a handoff document. Throws ArchiveHandoffError with a
 * message the submitter can act on.
 */
export async function parseArchiveHandoff(raw: string): Promise<ArchiveHandoff> {
  if (new TextEncoder().encode(raw).length > MAX_ARCHIVE_HANDOFF_BYTES) {
    fail(`The archive handoff is larger than ${MAX_ARCHIVE_HANDOFF_BYTES / 1024} KiB`);
  }
  let doc: unknown;
  try {
    doc = JSON.parse(raw);
  } catch {
    fail('The archive handoff is not valid JSON');
  }
  if (!isObject(doc) || doc.schema !== ARCHIVE_HANDOFF_SCHEMA) {
    fail(`This file is not a ${ARCHIVE_HANDOFF_SCHEMA} document`);
  }
  if (doc.schema_version !== ARCHIVE_HANDOFF_VERSION) {
    fail(
      `Unsupported archive handoff version; clpr reads version ${ARCHIVE_HANDOFF_VERSION}. Export the clip again from the archive.`,
    );
  }
  if (typeof doc.export_id !== 'string' || !UUID.test(doc.export_id)) {
    fail('The archive handoff has an invalid export_id');
  }

  const archive = isObject(doc.archive) ? doc.archive : fail('The archive handoff is missing archive');
  text(archive.name, 'archive.name', 200);
  const origin = httpsURL(archive.origin, 'archive.origin');
  if ((origin.pathname !== '/' && origin.pathname !== '') || origin.search || origin.hash) {
    fail('The archive handoff origin must be only a scheme and host');
  }

  const recording = isObject(doc.recording) ? doc.recording : fail('The archive handoff is missing recording');
  text(recording.video_id, 'recording.video_id', 128);

  const passage = isObject(doc.passage) ? doc.passage : fail('The archive handoff is missing passage');
  const { start_ms: start, end_ms: end, duration_ms: duration } = passage;
  if (!isInt(start) || !isInt(end) || !isInt(duration) || start < 0 || end <= start) {
    fail('The archive handoff has an invalid passage range: end_ms must be after start_ms, and start_ms cannot be negative');
  }
  if (duration !== end - start) {
    fail('The archive handoff passage duration does not match its start and end');
  }
  for (const field of ['url', 'player_url'] as const) {
    if (httpsURL(passage[field], `passage.${field}`).host !== origin.host) {
      fail(`The archive handoff passage.${field} is not on the archive's origin`);
    }
  }

  const transcript = isObject(doc.transcript) ? doc.transcript : fail('The archive handoff is missing transcript');
  if (transcript.status !== 'captured' && transcript.status !== 'unavailable') {
    fail('The archive handoff transcript status must be captured or unavailable');
  }
  if (transcript.status === 'captured') {
    if (transcript.source !== 'whisper' && transcript.source !== 'youtube') {
      fail('The archive handoff transcript source must be whisper or youtube');
    }
    if (typeof transcript.revision_sha256 !== 'string' || !SHA256.test(transcript.revision_sha256)) {
      fail('The archive handoff transcript is missing a valid revision_sha256');
    }
  }

  const captions = isObject(doc.captions) ? doc.captions : fail('The archive handoff is missing captions');
  if (captions.timing !== 'clip-relative' && captions.timing !== 'unavailable') {
    fail('The archive handoff has unsupported caption timing');
  }
  if (!Array.isArray(captions.cues)) {
    fail('The archive handoff captions have no cue list');
  }
  const cues = captions.cues as ArchiveCaptionCue[];
  if (cues.some(cue => !isObject(cue))) {
    fail('The archive handoff has a malformed caption cue');
  }
  const cueProblem = validateCaptionCues(cues, duration);
  if (cueProblem) {
    fail(cueProblem);
  }
  if (cues.length > 0) {
    if (typeof captions.text_sha256 !== 'string' || (await cueTextSHA256(cues)) !== captions.text_sha256) {
      fail(
        'The archive handoff captions do not match their text_sha256 digest; the file may have been edited. Download the handoff again.',
      );
    }
  }

  const media = isObject(doc.media) ? doc.media : fail('The archive handoff is missing media');
  if (media.type !== 'video/mp4') {
    fail('The archive handoff media type must be video/mp4');
  }
  if (!isInt(media.bytes) || media.bytes <= 0) {
    fail('The archive handoff media size must be positive');
  }
  if (typeof media.sha256 !== 'string' || !SHA256.test(media.sha256)) {
    fail('The archive handoff media sha256 must be 64 lowercase hex characters');
  }
  text(media.filename, 'media.filename', 255);

  const provenance = isObject(doc.provenance) ? doc.provenance : fail('The archive handoff is missing provenance');
  if (provenance.basis !== 'declared') {
    fail('The archive handoff provenance basis must be "declared"');
  }
  const source = isObject(provenance.media_source) ? provenance.media_source : {};
  if (source.kind !== 'youtube-range' && source.kind !== 'registered-original') {
    fail('The archive handoff has an unsupported media source kind');
  }

  return doc as unknown as ArchiveHandoff;
}

/** Compares a selected file with the handoff; returns an error message or null. */
export async function checkArchiveMedia(file: File, handoff: ArchiveHandoff): Promise<string | null> {
  const mismatch = `The selected video does not match the archive handoff: it expects ${handoff.media.filename} (${handoff.media.bytes} bytes). Choose the MP4 downloaded with this handoff.`;
  if (file.size !== handoff.media.bytes) {
    return mismatch;
  }
  return (await fileSHA256(file)) === handoff.media.sha256 ? null : mismatch;
}

/** Reads the stored record from a clip's source_metadata, if it has one. */
export function archiveProvenanceFromMetadata(metadata: unknown): ArchiveProvenanceRecord | null {
  if (!isObject(metadata) || !isObject(metadata.archive_provenance)) {
    return null;
  }
  const record = metadata.archive_provenance;
  if (!isObject(record.archive) || !isObject(record.passage) || !isObject(record.captions)) {
    return null;
  }
  try {
    if (new URL(String(record.passage.url)).protocol !== 'https:') {
      return null;
    }
  } catch {
    return null;
  }
  return record as unknown as ArchiveProvenanceRecord;
}

export function clipCaptionsURL(clipId: string): string {
  const base = (apiClient.defaults.baseURL ?? '/api/v1').replace(/\/$/, '');
  return `${base}/clips/${encodeURIComponent(clipId)}/captions`;
}

/** Formats clip-relative or absolute milliseconds as m:ss.s. */
export function formatArchiveMs(ms: number): string {
  const totalSeconds = ms / 1000;
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = (totalSeconds % 60).toFixed(1).padStart(4, '0');
  return `${minutes}:${seconds}`;
}
