package mqtt

import (
	"Control/types"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func SendImage(c mqtt.Client, id string) {
	time.Sleep(5 * time.Second)
	//if wartung is enable send wartungs image
	if types.Config.Wartung {
		imageData, err := loadImage("mqtt/wartung.png")

		if err != nil {
			slog.Error("Fehler", "error", err)
			return
		}

		responseTopic := id + "/image"
		slog.Info("Wartungsmodus aktiv")

		token := c.Publish(responseTopic, 0, false, imageData)
		slog.Info("Bild gesendet", "topic", responseTopic)
		token.Wait()

		time.Sleep(5 * time.Second)
		sendsleep(c, id)
	} else {
		hexDir := "handler/image_hex"
		room := types.GetRoomfromID(id)
		data, err := os.ReadFile(filepath.Join(hexDir, room+".hex"))
		if err != nil {
			slog.Error("Hex nicht lesbar", "error", err)
			return
		}

		// Hex-String → []byte
		parts := strings.Split(string(data), ", ")
		imageBytes := make([]byte, len(parts))
		for i, p := range parts {
			p = strings.TrimSpace(p)
			val, err := strconv.ParseUint(strings.TrimPrefix(p, "0x"), 16, 8)
			if err != nil {
				slog.Error("Hex nicht lesbar", "error", err)
			}
			imageBytes[i] = byte(val)
		}
		slog.Info("Sende Bytes", "bytes", len(imageBytes), "id", id)

		// Per MQTT senden
		topic := id + "/image"
		token := c.Publish(topic, 0, false, imageBytes)
		token.Wait()

		time.Sleep(5 * time.Second)
		sendsleep(c, id)
	}
}
