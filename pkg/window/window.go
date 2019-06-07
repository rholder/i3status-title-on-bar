package window

type WindowAPI interface {
	ActiveWindowTitle() string
	BeginTitleChangeDetection(onChange func(), onError func(error)) error
}
