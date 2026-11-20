package subsonic

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/waygu-music-plugin/sdk/pluginapi"
	backend "github.com/waygu-music-plugin/subsonic/internal/client"
)

// Starting a track reports its original Subsonic ID without blocking the UI, and leaving cancels the report.
func TestSubsonicStartReportFollowsActivation(t *testing.T) {
	report := make(chan url.Values, 1)
	stopped := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/scrobble" {
			report <- r.URL.Query()
			select {
			case <-r.Context().Done():
				close(stopped)
			case <-release:
			}
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, `<subsonic-response status="ok" version="1.16.1"></subsonic-response>`)
	}))
	defer server.Close()
	defer close(release)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	client := backend.NewClient(backend.Config{
		Address: server.URL, Username: "fixture-user", Password: "fixture-password",
	})
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	adapter := NewAdapter(nil, backend.ProfileSubsonic, nil)
	adapter.state.SetClient(client)
	adapter.state.Catalog().Recent = &RecentStore{}
	adapter.home.scope = ctx

	returned := make(chan struct{})
	go func() {
		defer close(returned)
		adapter.PlaybackMedia().Started(&pluginapi.Track{
			ID: 25, PlatformTrack: pluginapi.PlatformTrack{Source: "subsonic", Ref: "server-track"},
		})
	}()
	select {
	case query := <-report:
		if query.Get("id") != "server-track" || query.Get("submission") != "false" ||
			query.Get("u") != "fixture-user" || query.Get("t") == "" {
			t.Fatalf("start report used unexpected track, authentication or submission mode: %v", query)
		}
	case <-ctx.Done():
		t.Fatal("the Subsonic server did not receive the start report")
	}
	select {
	case <-returned:
	case <-ctx.Done():
		t.Fatal("start callback blocked on the Subsonic server")
	}
	cancel()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("leaving the platform did not cancel the start report")
	}
}
