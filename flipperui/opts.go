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

// WithFgColor sets the foreground color of the flipper screen.
func WithFgColor(color string) FlipperOpts {
	return func(m *Model) {
		m.fgColor = color
	}
}

// WithBgColor sets the background color of the flipper screen.
func WithBgColor(color string) FlipperOpts {
	return func(m *Model) {
		m.bgColor = color
	}
}
