package services

import (
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Helper functions for testing
func createMarkdownProcessor() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Linkify,
			extension.Table,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

func createSanitizer() *bluemonday.Policy {
	sanitizer := bluemonday.UGCPolicy()
	sanitizer.AllowElements("p", "br", "strong", "em", "del", "code", "pre", "blockquote",
		"ul", "ol", "li", "a", "h1", "h2", "h3", "h4", "h5", "h6", "table", "thead",
		"tbody", "tr", "th", "td")
	sanitizer.AllowAttrs("href").OnElements("a")
	sanitizer.AllowAttrs("class").OnElements("code")
	sanitizer.RequireNoFollowOnLinks(true)
	sanitizer.RequireNoReferrerOnLinks(true)
	sanitizer.AddTargetBlankToFullyQualifiedLinks(true)
	return sanitizer
}

func TestRenderMarkdown(t *testing.T) {
	// Create a minimal service instance for testing markdown
	service := &CommentService{}
	service.markdown = createMarkdownProcessor()
	service.sanitizer = createSanitizer()

	tests := []struct {
		name        string
		input       string
		contains    []string
		notContains []string
	}{
		{
			name:     "Bold text",
			input:    "This is **bold** text",
			contains: []string{"<strong>", "bold", "</strong>"},
		},
		{
			name:     "Italic text",
			input:    "This is *italic* text",
			contains: []string{"<em>", "italic", "</em>"},
		},
		{
			name:     "Strikethrough",
			input:    "This is ~~strikethrough~~ text",
			contains: []string{"<del>", "strikethrough", "</del>"},
		},
		{
			name:     "Link",
			input:    "[GitHub](https://github.com)",
			contains: []string{"<a", "href=\"https://github.com\"", "GitHub", "</a>"},
		},
		{
			name:     "Code block",
			input:    "`code here`",
			contains: []string{"<code>", "code here", "</code>"},
		},
		{
			name:     "Blockquote",
			input:    "> This is a quote",
			contains: []string{"<blockquote>", "This is a quote", "</blockquote>"},
		},
		{
			name:     "Unordered list",
			input:    "- Item 1\n- Item 2",
			contains: []string{"<ul>", "<li>", "Item 1", "Item 2", "</li>", "</ul>"},
		},
		{
			name:     "Ordered list",
			input:    "1. First\n2. Second",
			contains: []string{"<ol>", "<li>", "First", "Second", "</li>", "</ol>"},
		},
		{
			name:        "Script tags removed",
			input:       "<script>alert('xss')</script>",
			notContains: []string{"<script>", "alert"},
		},
		{
			name:        "Iframe removed",
			input:       "<iframe src='evil.com'></iframe>",
			notContains: []string{"<iframe>", "evil.com"},
		},
		{
			name:        "Deleted content",
			input:       "[deleted]",
			contains:    []string{"[deleted]"},
			notContains: []string{"<p>"},
		},
		{
			name:        "Removed content",
			input:       "[removed]",
			contains:    []string{"[removed]"},
			notContains: []string{"<p>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.RenderMarkdown(tt.input)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, result)
				}
			}

			for _, notExpected := range tt.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected output NOT to contain %q, got: %s", notExpected, result)
				}
			}
		})
	}
}

func TestCommentValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "Valid comment",
			content: "This is a valid comment",
			wantErr: false,
		},
		{
			name:    "Empty comment",
			content: "",
			wantErr: true,
		},
		{
			name:    "Only whitespace",
			content: "   ",
			wantErr: true,
		},
		{
			name:    "Too long comment",
			content: strings.Repeat("a", MaxCommentLength+1),
			wantErr: true,
		},
		{
			name:    "Max length comment",
			content: strings.Repeat("a", MaxCommentLength),
			wantErr: false,
		},
		{
			name:    "Minimum length comment",
			content: "a",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := strings.TrimSpace(tt.content)
			hasError := len(content) < MinCommentLength || len(content) > MaxCommentLength

			if hasError != tt.wantErr {
				t.Errorf("Content length validation for %q: got error=%v, want error=%v",
					tt.name, hasError, tt.wantErr)
			}
		})
	}
}

func TestCommentConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		min      int
		max      int
	}{
		{
			name:     "Max comment length",
			constant: MaxCommentLength,
			min:      1000,
			max:      50000,
		},
		{
			name:     "Min comment length",
			constant: MinCommentLength,
			min:      1,
			max:      10,
		},
		{
			name:     "Max nesting depth",
			constant: MaxNestingDepth,
			min:      5,
			max:      20,
		},
		{
			name:     "Edit window minutes",
			constant: EditWindowMinutes,
			min:      1,
			max:      60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant < tt.min || tt.constant > tt.max {
				t.Errorf("%s = %d, expected between %d and %d",
					tt.name, tt.constant, tt.min, tt.max)
			}
		})
	}
}

func TestKarmaValues(t *testing.T) {
	tests := []struct {
		name  string
		value int
	}{
		{
			name:  "Karma per comment",
			value: KarmaPerComment,
		},
		{
			name:  "Karma per upvote",
			value: KarmaPerUpvote,
		},
		{
			name:  "Karma per downvote",
			value: KarmaPerDownvote,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify they're defined
			if tt.value == 0 && tt.name != "Karma per downvote" {
				t.Errorf("%s should not be zero", tt.name)
			}
		})
	}
}

func TestMarkRestrictedComments_UsesExactHiddenMessage(t *testing.T) {
	service := &CommentService{}
	userID := uuid.New()
	node := CommentTreeNode{
		CommentWithAuthor: repository.CommentWithAuthor{Comment: models.Comment{UserID: userID}},
		Replies: []CommentTreeNode{{
			CommentWithAuthor: repository.CommentWithAuthor{Comment: models.Comment{UserID: uuid.New()}},
		}},
	}

	service.markRestrictedComments(&node, userID)

	if !node.IsHiddenByCreatorModeration {
		t.Fatal("expected top-level comment to be hidden")
	}
	if node.CreatorModerationMessage != creatorModerationHiddenCommentMessage {
		t.Fatalf("message = %q, want %q", node.CreatorModerationMessage, creatorModerationHiddenCommentMessage)
	}
	if node.Replies[0].IsHiddenByCreatorModeration {
		t.Fatal("expected non-author reply to remain visible")
	}
}
