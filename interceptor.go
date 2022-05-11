package httplib

import (
	"context"
	"time"
)

type HttpHandler func(ctx context.Context, method []byte, url string, params []byte, data []byte, timeout time.Duration,
	contentType string, opts ...Option) (response *Response, err error)

type HttpInterceptor func(ctx context.Context, method []byte, url string, params []byte, data []byte, timeout time.Duration,
	contentType string, handler HttpHandler, opts ...Option) (response *Response, err error)
