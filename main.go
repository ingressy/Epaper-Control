package main

import (
	"Control/handler"
	"Control/influx"
	"Control/mqtt"
	"Control/rest"
	"Control/types"
	"Control/untis"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "Control/docs"
)

// cronManager hält die dynamisch aus der Config geladenen Cron-Jobs.
// sync() prüft jede Minute, ob sich die Config geändert hat.
type cronManager struct {
	c          *cron.Cron
	mu         sync.Mutex
	ids        []cron.EntryID
	registered []string
}

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

	cronMgr := newCronManager()
	cronMgr.Start()

	influx.InitInflux()

	r := setupRouter()

	serverDone := make(chan error, 1)

	go mqtt.ConnecttoMQTT()
	go runHTTPServer(r, serverDone)

	waitForShutdown(cronMgr, serverDone)
}

func newCronManager() *cronManager {
	m := &cronManager{c: cron.New()}
	m.sync()
	if _, err := m.c.AddFunc("*/1 * * * *", m.sync); err != nil {
		slog.Error("Sync-Cron konnte nicht registriert werden", "error", err)
	}
	return m
}

func (m *cronManager) Start() { m.c.Start() }

func (m *cronManager) Stop() { m.c.Stop() }

func (m *cronManager) sync() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := types.GetConfig()

	if slices.Equal(cfg.Task_time_cron, m.registered) {
		return
	}
	for _, id := range m.ids {
		m.c.Remove(id)
	}

	m.ids = nil
	m.registered = nil

	for _, task := range cfg.Task_time_cron {
		task := task
		slog.Info(task + " registiert ...")

		id, err := m.c.AddFunc(task, func() { m.run(task) })
		if err != nil {
			slog.Error("Cron Task ungültig", "task", task, "error", err)
			continue
		}
		m.ids = append(m.ids, id)
		m.registered = append(m.registered, task)
	}

}

func (m *cronManager) run(task string) {
	slog.Info("Cron wird ausgeführt:", "task", task)
	untis.Get_room_from_json()
	handler.PrepareJSON()
	handler.Getpicturehex()
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	auth := r.Group("/")
	auth.Use(rest.JWTMiddleware())

	registerRoutes(auth)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

func registerRoutes(auth *gin.RouterGroup) {
	auth.GET("/get_wartung", rest.REST_GetWartung)
	auth.GET("/get_wartung_sleep", rest.REST_GetWartungSleepTime)
	auth.GET("/get_log_filename", rest.REST_GetLogFileName)
	auth.GET("/get_device/:id", rest.REST_GetDeviceViaID)

	auth.POST("/post_wartung", rest.REST_PostWartung)
	auth.POST("/post_wartung_sleep", rest.REST_PostWartungSleep)
}

func runHTTPServer(r *gin.Engine, done chan<- error) {
	if err := r.Run(types.ListenAddr); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP Server beendet mit Fehler", "error", err)
		done <- err
		return

	}
	done <- nil
}

func waitForShutdown(m *cronManager, serverDone <-chan error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		//normaler Shutdown
	case err := <-serverDone:
		if err != nil {
			slog.Error("HTTP Server vorzeitig beendet, Server fährt herrunter", "error", err)
		}
	}

	m.Stop()
	handler.StopLogging()
	slog.Info("Server wird beendet")
}
