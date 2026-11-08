package services

import (
	"context"
	"os"

	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
)

// WordTranscriber transcribes a WAV file with word timings.
type WordTranscriber interface {
	TranscribeWords(ctx context.Context, wavPath string) (*WhisperResult, error)
}

// WhisperCaptioner transcribes a clip file already on the render host for
// Short captions (#81). Its temporary WAV goes under WorkDir and is deleted
// at once.
type WhisperCaptioner struct {
	Whisper    WordTranscriber
	FFmpegPath string
	WorkDir    string
}

// Whisper's own confidence signals: a segment above this no-speech
// probability, or below this average log-probability, is likely music or a
// hallucination and gets no caption.
const (
	captionMaxNoSpeech  = 0.6
	captionMinLogprob   = -1.0
	captionLanguageCode = "en"
)

// Words returns the spoken words of an English clip, or none for a clip in
// another language or without speech.
func (c *WhisperCaptioner) Words(ctx context.Context, mediaPath string) ([]countdownrender.Word, error) {
	wav, err := ExtractAudioFromURL(ctx, c.FFmpegPath, mediaPath, c.WorkDir)
	if err != nil {
		return nil, err
	}
	defer os.Remove(wav)
	result, err := c.Whisper.TranscribeWords(ctx, wav)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Language != captionLanguageCode {
		return nil, nil
	}
	var words []countdownrender.Word
	for _, seg := range result.Segments {
		if seg.NoSpeechProb > captionMaxNoSpeech || seg.AvgLogprob < captionMinLogprob {
			continue
		}
		for _, w := range seg.Words {
			words = append(words, countdownrender.Word{Start: w.Start, End: w.End, Text: w.Word})
		}
	}
	return words, nil
}
