package main

import (
	"Control/watchdog"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//ruft den Watchdog für den ImageGen Mircoservice auf
	watchdog.CheckImageGen(r)

	r.Run("0.0.0.0:70")
}
