package breaker

type Bucket struct {
	Failure int64
	Success int64
	Total   int64
}

func (b *Bucket) succeed() {
	b.Total++
	b.Success++
}

func (b *Bucket) fail() {
	b.Total++
	b.Failure++
}

func (b *Bucket) reset() {
	b.Total = 0
	b.Success = 0
	b.Failure = 0
}

type window struct {
	buckets []*Bucket
	size    int
}

func newWindow(size int) *window {
	buckets := make([]*Bucket, size)
	for i := 0; i < size; i++ {
		buckets[i] = new(Bucket)
	}
	return &window{
		buckets: buckets,
		size:    size,
	}
}

func (w *window) succeed(offset int) {
	w.buckets[offset%w.size].succeed()
}

func (w *window) fail(offset int) {
	w.buckets[offset%w.size].fail()
}

func (w *window) resetBucket(offset int) {
	w.buckets[offset%w.size].reset()
}
