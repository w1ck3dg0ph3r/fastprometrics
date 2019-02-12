# Fastprometrics

Go fasthttp prometheus metrics middleware

## Example
```go
package main

import (
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	metrics "github.com/w1ck3dg0ph3r/fastprometrics"
)

func main() {
	router := fasthttprouter.New()
	handler := router.Handler

	handler = metrics.Add(handler,
		metrics.WithPath("/metrics"),
		metrics.WithSubsystem("http"),
	)

	router.GET("/ping", func(ctx *fasthttp.RequestCtx) {
		ctx.SetBodyString("pong")
	})
	
	fasthttp.ListenAndServe("127.0.0.1:8080", handler)
}

```
