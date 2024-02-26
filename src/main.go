package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func basicAuth(c *gin.Context) {
    user, password, hasAuth := c.Request.BasicAuth()
    if hasAuth && user == "testuser" && password == "testpass" {
        c.Next()
    } else {
        c.Abort()
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
}

func ping(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"pong": true})
}

func main() {
    router := gin.Default()
    router.GET("/ping", basicAuth, ping)

    router.Run("0.0.0.0:8080")
}
