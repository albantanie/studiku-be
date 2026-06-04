package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := New(nil)

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Register panicked: %v", recovered)
		}
	}()

	Register(router, handler)
}
