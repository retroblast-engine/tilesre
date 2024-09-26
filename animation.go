package tilesre

import "time"

// Animation represents a series of tiles that make up an animation.
type Animation struct {
	Frames     []int           // Frames is a slice of tile IDs that make up the animation.
	Index      int             // Index is the current frame index.
	Duration   []time.Duration // Duration is how long each frame should be displayed.
	LastChange time.Time       // LastChange is the time when the frame last changed.
}

// NextFrame returns the next frame of the animation and resets to the first frame if it's the last frame.
func (a *Animation) NextFrame() int {
	timePassed := time.Since(a.LastChange)
	if timePassed >= a.Duration[a.Index] {
		a.Index++
		if a.Index >= len(a.Frames) {
			a.Index = 0
		}
		a.LastChange = time.Now()
	}

	return a.Frames[a.Index]
}
