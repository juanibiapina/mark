package domain

import "slices"

type Context struct {
	items []ContextItem
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) Items() []ContextItem {
	return c.items
}

func (c *Context) AddItem(item ContextItem) {
	c.items = append(c.items, item)
}

func (c *Context) DeleteItem(index int) {
	c.items = slices.Delete(c.items, index, index+1)
}

func (c *Context) Message() string {
	var message string

	for _, item := range c.items {
		message += item.Message()
		message += "\n\n"
	}

	return message
}
