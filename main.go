package main

import (
	"Control/untis"
	"Control/watchdog"
)

func main() {
	//r := gin.Default()

	//ruft den Watchdog für den ImageGen Mircoservice auf
	//watchdog.CheckImageGen(r)
	watchdog.CheckUntis()

	//r.Run("0.0.0.0:70")
	untis.Get_data("2.305")

}
