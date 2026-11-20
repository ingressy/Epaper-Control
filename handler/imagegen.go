package handler

import (
	"Control/types"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Response struct {
	Lessons []types.Lesson `json:"lessons"`
	Room    string         `json:"room"`
}

type ImageResponse struct {
	Status string `json:"status"`
	Image  string `json:"image"` // "0xff, 0xff, 0xff, ..."
}

const outputDir = "handler/cache/"
const hexDir = "handler/image_hex/"

var httpClient = &http.Client{Timeout: 10 * time.Second}

func PrepareJSON() {
	data, err := os.ReadFile("untis/room.json")
	if err != nil {
		slog.Error("room.json nicht lesbar", "error", err)
		return
	}

	var rooms []types.Room
	if err := json.Unmarshal(data, &rooms); err != nil {
		slog.Error("room.json Parsefehler", "error", err)
		return
	}

	// Ordner erstellen falls nicht vorhanden
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		slog.Error("Ordner konnte nicht erstellt werden:", "error", err)
		return
	}

	for _, r := range rooms {
		data, err := os.ReadFile("untis/cache/" + r.Room + ".json")
		if err != nil {
			continue
		}

		var resp Response
		if err := json.Unmarshal(data, &resp); err != nil {
			slog.Error("Cache Parsefehler", "room", r.Room, "error", err)
			continue
		}
		//time cheat
		//today := now().Format("2006-01-02")
		today := time.Now().Format("2006-01-02")

		filtered := []types.Lesson{}
		for _, lesson := range resp.Lessons {
			if lesson.Date == today {
				filtered = append(filtered, lesson)
			}
		}
		resp.Lessons = filtered

		// JSON formatieren
		formatted, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			slog.Error("Marshal Fehler", "room", r.Room, "error", err)
			continue
		}

		// Pfad zusammensetzen: "data/roomName.json"
		filename := filepath.Join(outputDir, r.Room+".json")
		err = os.WriteFile(filename, formatted, 0644)
		if err != nil {
			slog.Error("Datei schreiben", "room", r.Room, "error", err)
			return
		}

	}

}

func Getpicturehex() {
	data, err := os.ReadFile("untis/room.json")
	if err != nil {
		slog.Error("room.json nicht lesbar", "error", err)
		return
	}

	var rooms []types.Room
	if err := json.Unmarshal(data, &rooms); err != nil {
		slog.Error("room.json Parsefehler", "error", err)
		return
	}

	//create folder for image hex json
	err = os.MkdirAll(hexDir, 0755)
	if err != nil {
		slog.Error("Ordner konnte nicht erstellt werden:", "error", err)
		return
	}

	for _, r := range rooms {
		formatted, err := os.ReadFile(filepath.Join(outputDir, r.Room+".json"))
		if err != nil {
			slog.Error("Cache nicht lesbar", "room", r.Room, "error", err)
			continue
		}
		// generate image
		httpResp, err := httpClient.Post("http://172.20.0.4:72/generate", "application/json", bytes.NewBuffer(formatted))
		if err != nil {
			slog.Error("POST Fehler", "room", r.Room, "error", err)
			continue
		}
		_, err = io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		if err != nil {
			slog.Error("Antwort lesen", "error", err)
			continue
		}

		if httpResp.StatusCode != http.StatusOK {
			slog.Error("ImageGen POST Status", "room", r.Room, "status", httpResp.StatusCode)
			continue
		}

		// get image hex
		httpResp, err = httpClient.Get("http://172.20.0.4:72/image?room=" + r.Room)
		if err != nil {
			slog.Error("Fehler", "error", err)
			continue
		}

		body, err := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		if err != nil {
			slog.Error("Antwort lesen", "error", err)
			continue
		}

		if httpResp.StatusCode != http.StatusOK {
			slog.Error("ImageGen Get Status", "room", r.Room, "status", httpResp.StatusCode)
			continue
		}

		// JSON parsen
		var imgResp ImageResponse
		if err := json.Unmarshal(body, &imgResp); err != nil {
			slog.Error("JSON Parse", "error", err)
			continue
		}

		// Hex-String direkt speichern
		filename := filepath.Join(hexDir, r.Room+".hex")
		if err := os.WriteFile(filename, []byte(imgResp.Image), 0644); err != nil {
			slog.Error("Schreiben", "error", err)
			continue
		}
		slog.Info("Gespeichert", "file", filename)
	}

}
