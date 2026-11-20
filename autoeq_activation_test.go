package subsonic

import (
	"testing"

	"github.com/waygu-music-plugin/sdk/pluginapi"
	"github.com/waygu-music-plugin/sdk/pluginapi/eq"
	"github.com/waygu-music-plugin/sdk/plugintest"
)

// A headphone profile from a previous activation cannot change playback after returning.
func TestAutoEQResultFollowsActivation(t *testing.T) {
	plugintest.CheckAutoEQActivation(t, Platforms()[0], func(instance pluginapi.Adapter, auto *eq.AutoEQ, profile eq.Meta) {
		h := instance.(*Adapter).Home()
		if h.eqUI.autoeqBusy != "" {
			t.Fatal("the previous request left the profile picker busy")
		}
		h.eqUI.autoeq = auto
		h.applyAutoEQ(profile)
	})
}
