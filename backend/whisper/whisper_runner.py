#!/usr/bin/env python3
"""Transcribe one audio file with faster-whisper and emit JSON on stdout."""

import json
import os
import sys

from faster_whisper import WhisperModel


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: whisper_runner.py AUDIO_PATH", file=sys.stderr)
        return 2

    model_name = os.getenv("WHISPER_MODEL", "base")
    device = os.getenv("WHISPER_DEVICE", "cpu")
    compute_type = os.getenv("WHISPER_COMPUTE_TYPE", "int8")
    model = WhisperModel(
        model_name,
        device=device,
        compute_type=compute_type,
        download_root=os.getenv("WHISPER_MODEL_DIR", "/tmp/clpr-whisper-models"),
    )
    segments_iter, info = model.transcribe(
        sys.argv[1],
        beam_size=int(os.getenv("WHISPER_BEAM_SIZE", "5")),
        temperature=0.0,
        vad_filter=True,
        condition_on_previous_text=False,
    )

    segments = []
    for segment in segments_iter:
        text = segment.text.strip()
        if not text:
            continue
        segments.append(
            {
                "start": segment.start,
                "end": segment.end,
                "text": text,
                "avg_logprob": segment.avg_logprob,
            }
        )

    print(
        json.dumps(
            {
                "segments": segments,
                "language": info.language,
                "full_text": " ".join(segment["text"] for segment in segments),
            },
            ensure_ascii=False,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
