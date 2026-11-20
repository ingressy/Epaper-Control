package main

import (
	"Control/handler"
	"Control/influx"
	"Control/mqtt"
	"Control/rest"
	"Control/types"
	"Control/untis"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "Control/docs"
)

// @title 	DoorSign Control MircoService API
// @version v1.2
// @host 192.168.133.50:80
// @BasePath        /
// @SecurityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	types.Loadconfig()
	types.Loadepd()
	handler.LoggingHandler()
	slog.Info("Loading Config Files finished!")

	c := cron.New()

	for _, task := range types.Config.Task_time_cron {
		task := task

		slog.Info(task + " registiert ...")

		c.AddFunc(task, func() {
			slog.Info("Cron wird ausgeführt:", "task", task)
			untis.Get_room_from_json()
			handler.PrepareJSON()
			handler.Getpicturehex()
		})
	}
	c.Start()
	influx.InitInflux()

	r := gin.Default()

	auth := r.Group("/")
	auth.Use(rest.JWTMiddleware())
	{
		auth.GET("/get_wartung", rest.REST_GetWartung)
		auth.GET("/get_wartung_sleep", rest.REST_GetWartungSleepTime)
		auth.GET("/get_log_filename", rest.REST_GetLogFileName)
		auth.GET("/get_device/:id", rest.REST_GetDeviceViaID)

		auth.POST("/post_wartung", rest.REST_PostWartung)
		auth.POST("/post_wartung_sleep", rest.REST_PostWartungSleep)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	go mqtt.ConnecttoMQTT()

	go func() {
		if err := r.Run("0.0.0.0:80"); err != nil {
			slog.Error("HTTP Server beendet mit Fehler", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Server wird beendet")
}
