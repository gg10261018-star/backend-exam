package wallet

import (
	"database/sql"
	"net/http"
	"strconv"
	"wallet_service/db"

	"github.com/gin-gonic/gin"
)

type Account struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type TransferResult struct {
	FromID      int64   `json:"from_id"`
	FromName    string  `json:"from_name"`
	FromBalance float64 `json:"from_balance"`
	ToID        int64   `json:"to_id"`
	ToName      string  `json:"to_name"`
	ToBalance   float64 `json:"to_balance"`
}

func Router(router gin.IRouter) {
	g := router.Group("/api")

	g.POST("/accounts", createAccount)
	g.GET("/accounts/:id", getAccount)
	g.POST("/transfer", transfer)
}

func createAccount(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var acc Account
	err := db.DB.QueryRow(
		`INSERT INTO accounts (name) VALUES ($1) RETURNING id, name, balance`,
		req.Name,
	).Scan(&acc.ID, &acc.Name, &acc.Balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, acc)
}

func getAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var acc Account
	err = db.DB.QueryRow(
		`SELECT id, name, balance FROM accounts WHERE id = $1`, id,
	).Scan(&acc.ID, &acc.Name, &acc.Balance)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, acc)
}

func transfer(c *gin.Context) {
	var req struct {
		FromID string `json:"from_id" binding:"required"`
		ToID   string `json:"to_id"   binding:"required"`
		Amount string `json:"amount"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	fromID, err1 := strconv.ParseInt(req.FromID, 10, 64)
	toID, err2 := strconv.ParseInt(req.ToID, 10, 64)
	amount, err3 := strconv.ParseFloat(req.Amount, 64)
	if err1 != nil || err2 != nil || err3 != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
		return
	}
	if fromID == toID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot transfer to self"})
		return
	}

	first, second := fromID, toID
	if fromID > toID {
		first, second = toID, fromID
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer tx.Rollback()

	var firstAcc, secondAcc Account
	if err = tx.QueryRow(
		`SELECT id, name, balance FROM accounts WHERE id = $1 FOR UPDATE`, first,
	).Scan(&firstAcc.ID, &firstAcc.Name, &firstAcc.Balance); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = tx.QueryRow(
		`SELECT id, name, balance FROM accounts WHERE id = $1 FOR UPDATE`, second,
	).Scan(&secondAcc.ID, &secondAcc.Name, &secondAcc.Balance); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fromAcc, toAcc := firstAcc, secondAcc
	if firstAcc.ID == toID {
		fromAcc, toAcc = secondAcc, firstAcc
	}

	if fromAcc.Balance < amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient balance"})
		return
	}

	if _, err = tx.Exec(
		`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromAcc.ID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err = tx.Exec(
		`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toAcc.ID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit error"})
		return
	}

	c.JSON(http.StatusOK, TransferResult{
		FromID:      fromAcc.ID,
		FromName:    fromAcc.Name,
		FromBalance: fromAcc.Balance - amount,
		ToID:        toAcc.ID,
		ToName:      toAcc.Name,
		ToBalance:   toAcc.Balance + amount,
	})
}
