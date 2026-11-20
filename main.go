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
	"reflect"
	"sync"
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

	var cronIDs []cron.EntryID
	var registered []string
	var cronMu sync.Mutex

	types.Loadconfig()
	types.Loadepd()
	handler.LoggingHandler()
	slog.Info("Loading Config Files finished!")

	c := cron.New()

	syncCronTask := func() {
		cronMu.Lock()
		defer cronMu.Unlock()
		cfg := types.GetConfig()
		if reflect.DeepEqual(cfg.Task_time_cron, registered) {
			return
		}
		for _, id := range cronIDs {
			c.Remove(id)
		}

		cronIDs = nil
		registered = nil

		for _, task := range cfg.Task_time_cron {
			task := task
			slog.Info(task + " registiert ...")

			id, err := c.AddFunc(task, func() {
				slog.Info("Cron wird ausgeführt:", "task", task)
				untis.Get_room_from_json()
				handler.PrepareJSON()
				handler.Getpicturehex()
			})
			if err != nil {
				slog.Error("Cron Task ungültig", "task", task, "error", err)
				continue
			}
			cronIDs = append(cronIDs, id)
			registered = append(registered, task)
		}
	}
	syncCronTask()

	if _, err := c.AddFunc("*/1 * * * *", syncCronTask); err != nil {
		slog.Error("Sync-Cron konnte nicht registiert werden", "error", err)
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
	handler.StopLogging()
	slog.Info("Server wird beendet")
}
