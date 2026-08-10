package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type TestPayload struct {
	Message string `json:"message"`
}

func TestRequestBodyBuffer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestBodyBuffer())

	r.POST("/test", func(c *gin.Context) {
		var p1 TestPayload
		if err := c.ShouldBindJSON(&p1); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var p2 TestPayload
		if err := c.ShouldBindJSON(&p2); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if p1.Message != p2.Message {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "payloads do not match"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": p1.Message})
	})

	payload := TestPayload{Message: "hello"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRequestBodyBufferDirectRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestBodyBuffer())

	r.POST("/test", func(c *gin.Context) {
		body1, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		body2, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if string(body1) != string(body2) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "bodies do not match"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"body": string(body1)})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("hello world"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}
