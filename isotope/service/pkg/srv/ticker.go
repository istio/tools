package srv

import (
	"context"
	"net/http"
	"sync"
)

// Ticker generates a default value and a tick value
type StatusTicker struct {
	Probability float64
	StatusChan  chan int

	active bool
	once   sync.Once
}

func (t *StatusTicker) Start(ctx context.Context) {
	// Start ticker only once
	t.once.Do(func() {
		t.active = true
		// Start goroutine cleaning up
		go func(ctx context.Context, t *StatusTicker) {
			<-ctx.Done()
			close(t.StatusChan)
			t.active = false
		}(ctx, t)

		// Start goroutine generating the tick
		go func(t *StatusTicker) {
			var every int
			if t.Probability > 0 {
				every = 10000 / int(t.Probability*10000)
			}

			counter := 1
			for t.active {
				if t.Probability == 0 || counter < every {
					t.StatusChan <- http.StatusOK
				} else {
					t.StatusChan <- http.StatusInternalServerError
					counter = 1
				}
				counter++
			}
		}(t)
	})
}
