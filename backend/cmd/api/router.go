package main

import "github.com/gin-gonic/gin"

// newRouter builds the API engine. Hierarchical tag slugs such as
// "content/funny" travel as one encoded path segment (content%2Ffunny), so
// routes match on the raw path and then unescape parameter values.
func newRouter() *gin.Engine {
	r := gin.New()
	r.UseRawPath = true
	r.UnescapePathValues = true
	return r
}
