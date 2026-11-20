package subsonic

import (
	"context"
	"math/rand/v2"

	"github.com/waygu-music-plugin/sdk/pluginapi"
)

// PlayList plays tracks from the given list starting at the specified index.
func PlayList(p PlayerPort, l *List, index int) {
	if p == nil || l == nil || l.Pages == nil {
		return
	}
	pluginapi.PlayPagedList(p, l.Pages, index)
}

// ShuffleList starts playing the list in random order.
func ShuffleList(ctx context.Context, p PlayerPort, l *List, post func(func())) {
	if ctx.Err() != nil || p == nil || l == nil || l.Pages == nil {
		return
	}
	if post == nil {
		post = func(fn func()) { fn() }
	}
	l.Pages.When(0, func(ok bool) {
		if !ok || ctx.Err() != nil {
			return
		}
		total := l.Pages.Total()
		seed := rand.Int64()
		pages := randomPages(total, seed, 0)

		feed := pluginapi.ShuffledQueueFeed(l.Pages, pages, seed)
		feed(ctx, pluginapi.QueueFeedCursor{}, func(first []pluginapi.Track, err error) {
			post(func() {
				if ctx.Err() != nil {
					return
				}
				if err == nil && len(first) > 0 {
					p.PlayList(first, 0, func(ctx context.Context, cursor pluginapi.QueueFeedCursor, got func([]pluginapi.Track, error)) {
						cursor.Page++
						feed(ctx, cursor, got)
					})
				}
			})
		})
	})
}

func randomPages(total int, seed int64, base int) []int {
	if total <= 0 {
		return nil
	}
	n := (total + PageSize - 1) / PageSize
	r := rand.New(rand.NewPCG(uint64(seed), 0x9e3779b97f4a7c15))
	out := r.Perm(n)
	if base >= 0 && base < n {
		for i, pg := range out {
			if pg == base {
				out[0], out[i] = out[i], out[0]
				break
			}
		}
	}
	return out
}
