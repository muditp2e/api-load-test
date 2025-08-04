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

	r.GET("/getKwalaChains", func(c *gin.Context) {
		fmt.Println("Received request /getKwalaChains:")
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"1": gin.H{
					"oracleAddress":        "0xD31a59c85aE9D8edEFeC411D448f90841571b89c",
					"tokenName":            "Solana",
					"oracleChainId":        1,
					"slippage":             "0.5",
					"tokenContractAddress": "0xD31a59c85aE9D8edEFeC411D448f90841571b89c",
					"tokenContractChainId": 2,
				},
				"2": gin.H{
					"oracleAddress":        "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
					"tokenName":            "Ethereum",
					"oracleChainId":        2,
					"slippage":             "0.5",
					"tokenContractAddress": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
					"tokenContractChainId": 2,
				},
				"3": gin.H{
					"oracleAddress":        "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
					"tokenName":            "Bitcoin",
					"oracleChainId":        3,
					"slippage":             "0.5",
					"tokenContractAddress": "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
					"tokenContractChainId": 2,
				},
				"4": gin.H{
					"oracleAddress":        "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
					"tokenName":            "BNB",
					"oracleChainId":        4,
					"slippage":             "0.5",
					"tokenContractAddress": "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
					"tokenContractChainId": 2,
				},
				"5": gin.H{
					"oracleAddress":        "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
					"tokenName":            "AMOY POLYGON",
					"oracleChainId":        104,
					"slippage":             "0.5",
					"tokenContractAddress": "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
					"tokenContractChainId": 80002,
				},
				"6": gin.H{
					"oracleAddress":        "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
					"tokenName":            "KALP",
					"oracleChainId":        105,
					"slippage":             "0.5",
					"tokenContractAddress": "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
					"tokenContractChainId": 80000,
				},
			},
		})
	})

	r.Run(":8087") // Runs on localhost:8080
}
