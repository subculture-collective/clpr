package search

import (
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin/binding"
)

// assertMatchesGin binds rawQuery with gin's binding.Query, which the handler
// used before, and with bindSearchQuery, and fails if the results differ.
func assertMatchesGin(t *testing.T, rawQuery string) {
	t.Helper()
	request := httptest.NewRequest("GET", "/api/v1/search", nil)
	request.URL.RawQuery = rawQuery

	var want models.SearchRequest
	wantErr := binding.Query.Bind(request, &want)
	var got models.SearchRequest
	gotErr := bindSearchQuery(request.URL.Query(), &got)

	if (wantErr == nil) != (gotErr == nil) {
		t.Fatalf("query %q: gin error = %v, bindSearchQuery error = %v", rawQuery, wantErr, gotErr)
	}
	if wantErr == nil && !reflect.DeepEqual(want, got) {
		t.Fatalf("query %q:\n gin             = %+v\n bindSearchQuery = %+v", rawQuery, want, got)
	}
}

var searchQuerySeeds = []string{
	"",
	"q=ace",
	"q=valorant+ace&type=clips&sort=recent&page=2&limit=50",
	"q=%20padded%20&type=&sort=",
	"q=!",
	"q=a&q=b",
	"said=cleanest+ace",
	"said=",
	"game_id=&twitch_category_id=&creator_id=&language=&date_from=&date_to=",
	"game_id=123&twitch_category_id=456&creator_id=789&language=en",
	"date_from=2026-01-01&date_to=2026-02-01",
	"tags=featured",
	"tags=a&tags=b&tags=",
	"tags=",
	"min_votes=5",
	"min_votes=",
	"min_votes=%205%20",
	"min_votes=-3",
	"min_votes=+3",
	"min_votes=abc",
	"min_votes=1.5",
	"min_votes=99999999999999999999",
	"page=",
	"page=%20",
	"page=0",
	"page=-1",
	"page=x",
	"limit=101",
	"limit=1_000",
	"limit=0x10",
	"limit=20&limit=x",
	"limit=x&limit=20",
	"Q=upper&LIMIT=5",
	"unknown=1",
	"q=%zz",
	"q=a;b",
	"q=ace&page=2&min_votes=x",
}

func TestBindSearchQueryMatchesGin(t *testing.T) {
	for _, rawQuery := range searchQuerySeeds {
		assertMatchesGin(t, rawQuery)
	}
}

// TestBindSearchQueryCoversEveryFormField fails when SearchRequest gains a
// form field that bindSearchQuery does not read.
func TestBindSearchQueryCoversEveryFormField(t *testing.T) {
	fields := reflect.TypeOf(models.SearchRequest{})
	for i := range fields.NumField() {
		key := fields.Field(i).Tag.Get("form")
		if key == "" {
			t.Fatalf("SearchRequest.%s has no form tag", fields.Field(i).Name)
		}
		query := url.Values{key: {"7"}}
		var got models.SearchRequest
		if err := bindSearchQuery(query, &got); err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		if reflect.ValueOf(got).Field(i).IsZero() {
			t.Errorf("bindSearchQuery ignores form key %q (SearchRequest.%s)", key, fields.Field(i).Name)
		}
		assertMatchesGin(t, query.Encode())
	}
}

func FuzzBindSearchQueryMatchesGin(f *testing.F) {
	for _, rawQuery := range searchQuerySeeds {
		f.Add(rawQuery)
	}
	f.Fuzz(func(t *testing.T, rawQuery string) {
		assertMatchesGin(t, rawQuery)
	})
}
