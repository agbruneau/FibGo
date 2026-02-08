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

// brailleDots maps (col 0-1, row 0-3) to the braille dot bit offsets.
// Braille character = U+2800 + sum of activated dot bits.
// Column 0: dots 1,2,3,7 (bits 0,1,2,6)
// Column 1: dots 4,5,6,8 (bits 3,4,5,7)
var brailleDots = [2][4]rune{
	{0x01, 0x02, 0x04, 0x40}, // left column
	{0x08, 0x10, 0x20, 0x80}, // right column
}

// RenderBrailleChart renders values (0..100) as a multi-row braille dot chart.
// Each braille character is 2 columns wide × 4 rows tall in the dot grid.
// The chart has `rows` text rows and `width` character columns.
// values are plotted right-aligned (most recent on the right).
func RenderBrailleChart(values []float64, width, rows int) []string {
	if width <= 0 || rows <= 0 || len(values) == 0 {
		return nil
	}

	// Total dot rows available in the braille grid
	dotRows := rows * 4
	// Total dot columns: each character covers 2 dot columns
	dotCols := width * 2

	// Initialize the braille grid (rows × width characters, all empty U+2800)
	grid := make([][]rune, rows)
	for r := range grid {
		grid[r] = make([]rune, width)
		for c := range grid[r] {
			grid[r][c] = 0x2800
		}
	}

	// Plot each value as a dot in the grid (right-aligned)
	startIdx := 0
	if len(values) > dotCols {
		startIdx = len(values) - dotCols
	}

	for i := startIdx; i < len(values); i++ {
		dotCol := (i - startIdx) + (dotCols - min(len(values), dotCols))
		v := values[i]
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}

		// Map value to dot row (0 = top, dotRows-1 = bottom)
		dotRow := dotRows - 1 - int(v/100.0*float64(dotRows-1))
		if dotRow < 0 {
			dotRow = 0
		}
		if dotRow >= dotRows {
			dotRow = dotRows - 1
		}

		// Convert dot coordinates to character cell + offset within cell
		charCol := dotCol / 2
		charRow := dotRow / 4
		subCol := dotCol % 2
		subRow := dotRow % 4

		if charCol >= 0 && charCol < width && charRow >= 0 && charRow < rows {
			grid[charRow][charCol] |= brailleDots[subCol][subRow]
		}
	}

	// Convert grid to strings
	result := make([]string, rows)
	for r := range grid {
		result[r] = string(grid[r])
	}
	return result
}
