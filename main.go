// main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
)

type BridgeStatusResponse struct {
	TxID      string `json:"txId"`
	Status    string `json:"status"`
	UpdatedAt int64  `json:"updatedAt"`
}

var db *sql.DB

func main() {
	var err error
	// Connect to MySQL
	dsn := "root:password123@tcp(localhost:3306)/bridge"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	defer db.Close()

	router := gin.New()
	router.Use(gin.Logger())
	allowedOrigins := []string{"*"}

	router.Use(cors.New(cors.Config{
		AllowOrigins:  allowedOrigins,
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length", "Access-Control-Allow-Origin"},
	}))
	router.Use(secure.New(secure.Config{
		SSLRedirect:           false,
		IsDevelopment:         false,
		STSSeconds:            315360000,
		STSIncludeSubdomains:  false,
		FrameDeny:             false,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		// IENoOpen:              true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	}))

	router.GET("/tx/status", handleGetTxStatus)
	router.GET("/tx/pending", handlePendingTxs)
	router.POST("/tx/add", handleInsertTx)
	router.POST("/tx/update", handleUpdateTxStatus)

	err = router.Run(":6500")
	if err != nil {
		fmt.Printf("Error starting the server: %v", err)
		os.Exit(1)
	}
}

// Fetch all non-TxSuccess rows
func handlePendingTxs(c *gin.Context) {
	rows, err := db.Query(`SELECT txID, status FROM bridge_token WHERE status != 'TxSuccess' and status != 'TxFailed'`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []BridgeStatusResponse
	for rows.Next() {
		var tx BridgeStatusResponse
		if err := rows.Scan(&tx.TxID, &tx.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, tx)
	}

	c.JSON(http.StatusOK, results)
}

// Insert new transaction
func handleInsertTx(c *gin.Context) {
	var req BridgeStatusResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	_, err := db.Exec(
		`INSERT INTO bridge_token (txID, status, created_at, updated_at) VALUES (?, ?, NOW(), NOW())`,
		req.TxID, req.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed: " + err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// Get single transaction status
func handleGetTxStatus(c *gin.Context) {
	txID := c.Query("txId")
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing txId parameter"})
		return
	}

	var status string
	err := db.QueryRow("SELECT status FROM bridge_token WHERE txID = ?", txID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, BridgeStatusResponse{
		TxID:   txID,
		Status: status,
	})
}

// Update transaction status
func handleUpdateTxStatus(c *gin.Context) {
	var req BridgeStatusResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	t := time.UnixMilli(req.UpdatedAt)

	_, err := db.Exec(`UPDATE bridge_token SET status = ?, updated_at = ? WHERE txID = ?`, req.Status, t, req.TxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed: " + err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
