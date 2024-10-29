package hn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	t.Run("Submissions", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedURL := "/user/foouser/submitted.json"
			if r.URL.Path != expectedURL {
				t.Errorf("Expected to request '%s', got: %s", expectedURL, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[1,2,5]`))
		}))
		defer server.Close()

		baseURL = server.URL
		actual, err := FetchSubmissions(context.Background(), "foouser")
		if err != nil {
			t.Fatal(err)
		}
		expected := []int{1, 2, 5}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expected:\n  %v\ngot:\n  %v", expected, actual)
		}
	})

	t.Run("Story", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedURL := "/item/42.json"
			if r.URL.Path != expectedURL {
				t.Errorf("Expected to request '%s', got: %s", expectedURL, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
  "by": "whoishiring",
  "descendants": 3,
  "id": 42,
  "kids": [55, 56, 57],
  "score": 999,
  "text": "boilerplate",
  "time": 1727794816,
  "title": "Ask HN: Who is hiring? (October 2024)",
  "type": "story"
}`))
		}))
		defer server.Close()

		baseURL = server.URL
		actual, err := FetchStory(context.Background(), 42)
		if err != nil {
			t.Fatal(err)
		}
		expected := &Story{
			Id:            42,
			Kids:          []int{55, 56, 57},
			Time:          1727794816,
			GoTime:        time.Unix(1727794816, 0),
			Title:         "Ask HN: Who is hiring? (October 2024)",
			FetchedTime:   actual.FetchedTime, // no good way to test these
			FetchedGoTime: actual.FetchedGoTime,
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expected:\n  %v\ngot:\n  %v", expected, actual)
		}
	})

	t.Run("Comment", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedURL := "/item/77.json"
			if r.URL.Path != expectedURL {
				t.Errorf("Expected to request '%s', got: %s", expectedURL, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
  "by": "bro",
  "id": 77,
  "parent": 999,
  "text": "goodjob",
  "time": 1727794816,
  "type": "comment"
}`))
		}))
		defer server.Close()

		baseURL = server.URL
		actual, err := FetchComment(context.Background(), 77)
		if err != nil {
			t.Fatal(err)
		}
		expected := &Comment{
			Id:            77,
			Parent:        999,
			Text:          "goodjob",
			Time:          1727794816,
			GoTime:        time.Unix(1727794816, 0),
			FetchedTime:   actual.FetchedTime, // no good way to test these
			FetchedGoTime: actual.FetchedGoTime,
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expected:\n  %#v\ngot:\n  %#v", expected, actual)
		}
	})
}
