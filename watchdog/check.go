package watchdog

import (
	"Control/handler"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CheckImageGen(r *gin.Engine) {
	go func() {
		for {
			now := time.Now()
			//jede Minute
			next := now.Truncate(time.Minute).Add(time.Minute)
			time.Sleep(time.Until(next))

			go pingImageGen(r)
		}
	}()
}

func pingImageGen(r *gin.Engine) {
	r.GET("IMAGEGEN:bla/health", func(c *gin.Context) {

		client := http.Client{
			Timeout: 2 * time.Second,
		}

		resp, err := client.Get("http://IMAGEGEN:bla/health")
		if err != nil {
			//down
			handler.HandleImageGendown()
			return
		}
		defer resp.Body.Close()

		var health struct {
			Status string `json:"status"`
		}
		json.NewDecoder(resp.Body).Decode(&health)

		switch health.Status {

		case "green":
			return
		case "yellow":
			return
		case "red":
			handler.HandleImageGendown()
			return
		default:
			handler.HandleImageGendown()
			return

		}
	})
}
