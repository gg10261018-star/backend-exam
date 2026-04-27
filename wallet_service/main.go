package wallet_service

import (
	"log"
	"wallet_service/db"
	"wallet_service/wallet"

	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	// 初始化 DB
	if err := db.Init(); err != nil {
		log.Fatal("db init failed:", err)
	}
	defer db.Close()

	// Router
	wallet.Router(engine)

	log.Fatal(engine.Run(":8080"))
}
