package breaker

import (
	"github.com/benbjohnson/clock"
	"testing"
	"time"
)

func TestRollingWindow(t *testing.T) {
	c := clock.NewMock()
	rollingWindow := NewRollingWindow(time.Millisecond*10, 2, func(w *RollingWindow) { w.clock = c; w.lastTime = c.Now() })

	expectSuccess, expectFailure, expectTotal := int64(0), int64(0), int64(0)
	// 0
	rollingWindow.MarkSuccess()
	expectSuccess += 1
	expectTotal += 1
	if s, f, total := rollingWindow.Statistics(); s != expectSuccess && f != expectFailure && total != expectTotal {
		t.Fatalf("error window")
	}

	// 6
	c.Add(6 * time.Millisecond)
	rollingWindow.MarkSuccess()
	expectSuccess += 1
	expectTotal += 1
	if s, f, total := rollingWindow.Statistics(); s != expectSuccess && f != expectFailure && total != expectTotal {
		t.Fatalf("error window")
	}

	// 8
	c.Add(2 * time.Millisecond)
	rollingWindow.MarkFailed()
	expectFailure += 1
	expectTotal += 1
	if s, f, total := rollingWindow.Statistics(); s != expectSuccess && f != expectFailure && total != expectTotal {
		t.Fatalf("error window")
	}

	// 11
	c.Add(3 * time.Millisecond)
	rollingWindow.MarkFailed()
	expectFailure += 1
	expectSuccess -= 1
	if s, f, total := rollingWindow.Statistics(); s != expectSuccess && f != expectFailure && total != expectTotal {
		t.Fatalf("error window")
	}

	// 14
	c.Add(3 * time.Millisecond)
	rollingWindow.MarkFailed()
	expectFailure += 1
	expectTotal += 1
	if s, f, total := rollingWindow.Statistics(); s != expectSuccess && f != expectFailure && total != expectTotal {
		t.Fatalf("error window")
	}

	// 17
	c.Add(3 * time.Millisecond)
	rollingWindow.MarkSuccess()
	expectFailure -= 1
	expectTotal -= 1
	if s, f, total := rollingWindow.Statistics(); s != expectSuccess && f != expectFailure && total != expectTotal {
		t.Fatalf("error window")
	}
}
