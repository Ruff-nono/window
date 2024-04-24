package breaker

import (
	"github.com/benbjohnson/clock"
	"sync"
	"time"
)

type (
	RollingWindowOption func(rollingWindow *RollingWindow)

	RollingWindow struct {
		lock          sync.RWMutex
		windowBuckets int
		win           *window
		bucketTime    time.Duration
		offset        int
		ignoreCurrent bool
		lastTime      time.Time // start time of the last bucket
		clock         clock.Clock
	}
)

func NewRollingWindow(windowTime time.Duration, windowBuckets int, opts ...RollingWindowOption) *RollingWindow {
	if windowBuckets < 1 {
		panic("windowBuckets must be greater than 0")
	}
	bucketTime := time.Duration(windowTime.Nanoseconds() / int64(windowBuckets))
	if bucketTime < 1 {
		panic("bucketTime must be greater than 0")
	}
	ck := clock.New()
	w := &RollingWindow{
		windowBuckets: windowBuckets,
		win:           newWindow(windowBuckets),
		bucketTime:    bucketTime,
		lastTime:      ck.Now(),
		clock:         ck,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (rw *RollingWindow) MarkSuccess() {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	rw.updateOffset()
	rw.win.succeed(rw.offset)
}

func (rw *RollingWindow) MarkFailed() {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	rw.updateOffset()
	rw.win.fail(rw.offset)
}

func (rw *RollingWindow) Statistics() (success, failure, total int64) {
	rw.Reduce(func(b *Bucket) {
		success += b.Success
		failure += b.Failure
		total += b.Total
	})
	return
}

func (rw *RollingWindow) Reduce(fn func(b *Bucket)) {
	rw.lock.RLock()
	defer rw.lock.RUnlock()

	span := rw.span()
	var diff int // 需要统计的bucket数量
	if span == 0 && rw.ignoreCurrent {
		diff = rw.windowBuckets - 1
	} else {
		// exclude expire buckets
		diff = rw.windowBuckets - span
	}

	if diff > 0 {
		offset := (rw.offset + span + 1) % rw.windowBuckets
		rw.win.reduce(offset, diff, fn)
	}
}

// 计算经过了多少时间间隔
func (rw *RollingWindow) span() int {
	offset := int(rw.clock.Since(rw.lastTime) / rw.bucketTime)
	if 0 <= offset && offset < rw.windowBuckets {
		return offset
	}
	return rw.windowBuckets
}

func (rw *RollingWindow) updateOffset() {
	span := rw.span()
	if span <= 0 {
		return
	}

	offset := rw.offset
	for i := 0; i < span; i++ {
		rw.win.resetBucket((offset + i + 1) % rw.windowBuckets)
	}
	// update
	rw.offset = (offset + span) % rw.windowBuckets
	rw.lastTime = rw.clock.Now().Add(-rw.clock.Since(rw.lastTime) % rw.bucketTime)
}

func (w *window) reduce(start, count int, fn func(b *Bucket)) {
	for i := 0; i < count; i++ {
		// 自定义统计函数
		fn(w.buckets[(start+i)%w.size])
	}
}
