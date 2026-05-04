package watchdog

import (
	"Control/handler"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func PingImageGen(r *gin.Engine) {
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

func PingUntis() {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("http://0.0.0.0:71/health")
	if err != nil {
		fmt.Println("Untis Microservice nicht erreichbar")
		return
	}
	defer resp.Body.Close()

	var health struct {
		Status string `json:"status"`
	}

	err = json.NewDecoder(resp.Body).Decode(&health)
	if err != nil {
		fmt.Println("Ungültige Antwort vom Microservice")
		return
	}

	switch health.Status {
	case "green":
		fmt.Println("Untis Microservice green")
	case "yellow":
		fmt.Println("Untis Microservice yellow")
	case "red":
		fmt.Println("Untis Microservice red")
	default:
		fmt.Println("Untis Microservice unbekannter Status")
	}
}
