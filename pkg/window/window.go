package window

// API defines the functions necessary to monitor window activity.
type API interface {

	// ActiveWindowTitle returns the currently active window's title.
	ActiveWindowTitle() string

	// DetectWindowTitleChanges blocks and starts detecting changes in window
	// titles. When a change is detected, the onChange function is called and
	// when a non-fatal error occurs the onError function is called for that
	// error.
	DetectWindowTitleChanges(onChange func(), onError func(error)) error
}
