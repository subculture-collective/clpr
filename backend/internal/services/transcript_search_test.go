package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/opensearch"
	"github.com/google/uuid"
)

func TestClipSearchSaidFilterAndSnippets(t *testing.T) {
	clipID := uuid.New()
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"hits":{"total":{"value":1},"hits":[{"_source":{"id":"`+clipID.String()+`","title":"ace"},
			"highlight":{"transcript":["that was the \ue000cleanest\ue001 \ue000ace\ue001 ever"]}}]}}`)
	}))
	defer server.Close()
	client, err := opensearch.NewClient(&opensearch.Config{URL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	said := "  cleanest ace "
	resp, err := NewOpenSearchService(client).Search(context.Background(),
		&models.SearchRequest{Type: "clips", Said: &said, Page: 1, Limit: 10, Sort: "relevance"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results.Clips) != 1 || resp.Results.Clips[0].TranscriptSnippet == nil ||
		*resp.Results.Clips[0].TranscriptSnippet != "that was the \ue000cleanest\ue001 \ue000ace\ue001 ever" {
		t.Fatalf("clips = %+v", resp.Results.Clips)
	}
	raw, _ := json.Marshal(body)
	text := string(raw)
	for _, want := range []string{
		`"match_phrase":{"transcript":{"query":"cleanest ace","slop":0}}`,
		`"_source":{"excludes":["transcript"]}`,
		`"highlight":{"fields":{"transcript":`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("search body lacks %s:\n%s", want, text)
		}
	}
	if strings.Contains(text, "multi_match") {
		t.Errorf("a said-only search added a text query: %s", text)
	}
}

func TestClipQuerySearchesTranscripts(t *testing.T) {
	query := (&OpenSearchService{}).buildClipQuery(&models.SearchRequest{Query: "no way"})
	raw, _ := json.Marshal(query)
	if !strings.Contains(string(raw), `"transcript"`) || strings.Contains(string(raw), "match_phrase") {
		t.Fatalf("query = %s", raw)
	}
}

func TestCopyTranscriptSnippetsKeepsThemThroughReranking(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	snippet := "\ue000hello\ue001"
	reranked := copyTranscriptSnippets([]models.Clip{{ID: b}, {ID: a}}, []models.Clip{{ID: a, TranscriptSnippet: &snippet}, {ID: b}})
	if reranked[1].TranscriptSnippet == nil || *reranked[1].TranscriptSnippet != snippet || reranked[0].TranscriptSnippet != nil {
		t.Fatalf("reranked = %+v", reranked)
	}
}

func TestClipTranscriptNeverSerialized(t *testing.T) {
	text := "the whole transcript"
	raw, _ := json.Marshal(models.Clip{ID: uuid.New(), Transcript: &text})
	if strings.Contains(string(raw), text) {
		t.Fatalf("clip JSON carries the transcript: %s", raw)
	}
}

func TestUpdateClipTranscriptsToleratesClipsNotYetIndexed(t *testing.T) {
	for _, tc := range []struct {
		response string
		wantErr  bool
	}{
		{`{"errors":true,"items":[{"update":{"status":200}},{"update":{"status":404,"error":{"type":"document_missing_exception"}}}]}`, false},
		{`{"errors":true,"items":[{"update":{"status":400,"error":{"type":"mapper_parsing_exception"}}}]}`, true},
	} {
		var body string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			body = string(raw)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, tc.response)
		}))
		client, err := opensearch.NewClient(&opensearch.Config{URL: server.URL})
		if err != nil {
			t.Fatal(err)
		}
		id := uuid.New()
		err = NewSearchIndexerService(client).UpdateClipTranscripts(context.Background(), map[uuid.UUID]string{id: "no way"})
		server.Close()
		if (err != nil) != tc.wantErr {
			t.Fatalf("response %s: err = %v", tc.response, err)
		}
		if !strings.Contains(body, `"update":{"_id":"`+id.String()+`","_index":"clips"}`) || !strings.Contains(body, `{"doc":{"transcript":"no way"}}`) {
			t.Fatalf("bulk body = %s", body)
		}
	}
}
