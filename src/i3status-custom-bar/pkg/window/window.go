package window

import "io"

type WindowAPI interface {
	ActiveWindowTitle() string
	BeginTitleChangeDetection(stderr io.Writer, onChange func()) error
}
