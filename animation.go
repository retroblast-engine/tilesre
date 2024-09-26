package tilesre

import "time"

// Animation represents a series of tiles that make up an animation.
type Animation struct {
	Frames     []int
	Index      int
	Duration   []time.Duration // how long the current frame should be displayed
	LastChange time.Time       // is updated to the current time each time the frame changes
}

// NextFrame returns the next frame of the animation and resets to the first frame if it's the last frame.
func (s *Animation) NextFrame() int {
	timePassed := time.Since(s.LastChange)
	if timePassed >= s.Duration[s.Index] {
		s.Index++
		if s.Index >= len(s.Frames) {
			s.Index = 0
		}
		s.LastChange = time.Now()
	}

	return s.Frames[s.Index]
}
