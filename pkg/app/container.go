package app

// Container is a view that can be rendered to a specific width and height.
type Container interface {
	Render(m App, width, height int) string
}
