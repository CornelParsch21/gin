package main

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/gin-gonic/gin"
)

// ReusableReadCloser wraps a byte slice and implements io.ReadCloser.
// It resets the read pointer to the beginning of the slice whenever EOF is reached,
// allowing the body to be read multiple times.
type ReusableReadCloser struct {
	mu         sync.Mutex
	body       []byte
	readCloser io.ReadCloser
}

// Read reads from the underlying reader. If EOF is reached, it resets the reader.
func (r *ReusableReadCloser) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.readCloser == nil {
		r.readCloser = io.NopCloser(bytes.NewBuffer(r.body))
	}
	n, err = r.readCloser.Read(p)
	if err == io.EOF {
		r.readCloser = nil
	}
	return
}

// Close closes the reader and resets it.
func (r *ReusableReadCloser) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.readCloser != nil {
		r.readCloser.Close()
		r.readCloser = nil
	}
	return nil
}

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// RequestBodyBuffer is a middleware that buffers the request body,
// allowing it to be read multiple times by downstream handlers.
func RequestBodyBuffer() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request == nil || c.Request.Body == nil {
			c.Next()
			return
		}

		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer func() {
			buf.Reset()
			bufferPool.Put(buf)
		}()

		_, err := buf.ReadFrom(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		bodyBytes := buf.Bytes()
		c.Set(gin.BodyBytesKey, bodyBytes)

		c.Request.Body = &ReusableReadCloser{body: bodyBytes}
		c.Next()
	}
}

func main() {
	fmt.Println("Hello, Bounty Hunter!")
}
