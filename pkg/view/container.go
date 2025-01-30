package view

// Container is a component that can be rendered to a specific width and height
type Container interface {
	Render(width, height int) string
}
