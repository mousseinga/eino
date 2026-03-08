package middleware

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	"ai-eino-interview-agent/api/response"
	"ai-eino-interview-agent/internal/errors"

	"github.com/cloudwego/hertz/pkg/app"
)

// Recovery 恢复中间件，捕获 Panic
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.Printf("[PANIC] Recovered: %v\n%s", r, stack)

				appErr := errors.NewInternalError(
					"Internal server error",
					fmt.Errorf("panic: %v\nStack:\n%s", r, string(stack)),
				)

				response.Error(ctx, c, appErr.HTTPStatus, appErr.Message)
			}
		}()

		c.Next(ctx)
	}
}
