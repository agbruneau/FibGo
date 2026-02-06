package tui

// sparklineChars maps values 0..7 to Unicode block elements ▁▂▃▄▅▆▇█.
var sparklineChars = [8]rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RingBuffer is a fixed-capacity circular buffer for float64 samples.
type RingBuffer struct {
	data  []float64
	head  int
	count int
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &RingBuffer{data: make([]float64, capacity)}
}

// Push adds a sample, overwriting the oldest if full.
func (r *RingBuffer) Push(v float64) {
	r.data[r.head] = v
	r.head = (r.head + 1) % len(r.data)
	if r.count < len(r.data) {
		r.count++
	}
}

// Len returns the number of valid samples.
func (r *RingBuffer) Len() int { return r.count }

// Cap returns the buffer capacity.
func (r *RingBuffer) Cap() int { return len(r.data) }

// Last returns the most recent sample, or 0 if empty.
func (r *RingBuffer) Last() float64 {
	if r.count == 0 {
		return 0
	}
	idx := r.head - 1
	if idx < 0 {
		idx = len(r.data) - 1
	}
	return r.data[idx]
}

// Slice returns samples in chronological order (oldest first).
func (r *RingBuffer) Slice() []float64 {
	if r.count == 0 {
		return nil
	}
	result := make([]float64, r.count)
	start := r.head - r.count
	if start < 0 {
		start += len(r.data)
	}
	for i := range r.count {
		result[i] = r.data[(start+i)%len(r.data)]
	}
	return result
}

// Resize changes the capacity, preserving the most recent samples that fit.
func (r *RingBuffer) Resize(newCap int) {
	if newCap <= 0 {
		newCap = 1
	}
	if newCap == len(r.data) {
		return
	}
	old := r.Slice()
	r.data = make([]float64, newCap)
	r.head = 0
	r.count = 0
	start := 0
	if len(old) > newCap {
		start = len(old) - newCap
	}
	for _, v := range old[start:] {
		r.Push(v)
	}
}

// Reset clears all samples.
func (r *RingBuffer) Reset() {
	r.head = 0
	r.count = 0
}

// RenderSparkline converts values (0..100) into a sparkline string using Unicode blocks.
func RenderSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	runes := make([]rune, len(values))
	for i, v := range values {
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		idx := int(v / 100.0 * 7.0)
		if idx > 7 {
			idx = 7
		}
		runes[i] = sparklineChars[idx]
	}
	return string(runes)
}
