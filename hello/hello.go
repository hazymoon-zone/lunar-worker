// Service hello implements a simple hello world REST API.
package hello

import (
	"context"
)

// This is a simple REST API that responds with a personalized greeting.
//
//encore:api public path=/hello/:name
func World(ctx context.Context, name string) (*Response, error) {
	msg := "Hello, " + name + "!"
	return &Response{Message: msg}, nil
}

type Response struct {
	Message string
}
