package flipperui

type FlipperOpts func(*Model)

func WithScreenshotResolution(width, height int) FlipperOpts {
	return func(m *Model) {
		m.screenshotResolution.width = width
		m.screenshotResolution.height = height
	}
}
