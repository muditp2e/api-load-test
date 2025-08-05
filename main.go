package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Catch-all route: matches any path and method
	router.Any("/*path", func(c *gin.Context) {
		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("Error reading body:", err)
		} else if len(bodyBytes) > 0 {
			fmt.Println("Request Body:")
			fmt.Println(string(bodyBytes))
		} else {
			fmt.Println("No request body.")
		}

		// Respond with 200 OK
		c.String(http.StatusOK, "OK")
	})

	// Start the server
	err := router.Run(":8087")
	if err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
