package flipperui

// FlipperOpts represents an optional configuration for the flipper model.
type FlipperOpts func(*Model)

// WithScreenshotResolution sets the resolution of the screenshot.
func WithScreenshotResolution(width, height int) FlipperOpts {
	return func(m *Model) {
		m.screenshotResolution.width = width
		m.screenshotResolution.height = height
	}
}
