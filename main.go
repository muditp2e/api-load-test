package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/contracts/deploy", func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		fmt.Println("Received request body:")
		fmt.Println(string(bodyBytes))

		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"contractAddress": "0xb4db8fa7ba17ad9b9356a9cf23ccfae749056fe1",
			"transactionHash": "",
		})
	})

	r.POST("/transactions", func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		fmt.Println("Received request body:")
		fmt.Println(string(bodyBytes))

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"transaction": "0x95f115491b9374fa56cf8a42b76bd5818fcea1df225d981a0a4a26a54184e9d0",
			"queue_id":    "idemp-1749730421583",
		})
	})

	r.POST("/block/monitor", func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		fmt.Println("Received request body:")
		fmt.Println(string(bodyBytes))

		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})

	r.POST("/webhooks/oraclePrice", func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		fmt.Println("Received request body:")
		fmt.Println(string(bodyBytes))

		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"subscriptionId": "8b880665-2a78-4137-ad74-f5037f87f136",
			"status":         "active",
			"createdAt":      "2025-06-19T05:59:08.863Z",
		})
	})

	r.Run(":8086") // Runs on localhost:8080
}
