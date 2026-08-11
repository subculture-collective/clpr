package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

// WhisperService calls hasanara's whisper_runner.py as a subprocess
// to transcribe audio chunks via faster-whisper or openai-whisper.
type WhisperService struct {
	pythonPath string
	runnerPath string
}

// WhisperSegment represents a single transcribed segment with timing.
type WhisperSegment struct {
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	AvgLogprob float64 `json:"avg_logprob"`
}

// WhisperResult is the structured output from a transcription run.
type WhisperResult struct {
	Segments []WhisperSegment `json:"segments"`
	Language string           `json:"language"`
	FullText string           `json:"full_text"`
}

// NewWhisperService creates a WhisperService. runnerDir is the directory
// containing the whisper_runner.py symlink (e.g. "backend/whisper").
func NewWhisperService(pythonPath, runnerDir string) *WhisperService {
	return &WhisperService{
		pythonPath: pythonPath,
		runnerPath: filepath.Join(runnerDir, "whisper_runner.py"),
	}
}

// TranscribeAudio runs whisper_runner.py on a WAV file and returns the
// structured transcription result.
func (s *WhisperService) TranscribeAudio(ctx context.Context, wavPath string) (*WhisperResult, error) {
	cmd := exec.CommandContext(ctx, s.pythonPath, s.runnerPath, wavPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("whisper transcription failed: %w\nstderr: %s", err, stderr.String())
	}

	var result WhisperResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parsing whisper output: %w\nraw output: %s", err, string(output))
	}
	return &result, nil
}
