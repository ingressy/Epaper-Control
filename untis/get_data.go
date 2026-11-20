package untis

import (
	"Control/types"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const outputDir = "untis/cache/"

var httpClient = &http.Client{Timeout: 10 * time.Second}

func Get_room_from_json() {
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

	for _, r := range rooms {
		Get_data(r.Room)
	}
}

func Get_data(room string) {
	resp, err := httpClient.Get("http://172.20.0.3:71/untis?room=" + room)
	if err != nil {
		slog.Error("Untis Microservice nicht erreichbar", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Untis Mircoservice Status", "room", room, "status", resp.StatusCode)
		return
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		slog.Error("JSON Decoding", "error", err)
		return
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		slog.Error("JSON MarshalIndent", "error", err)
		return
	}

	// Ordner erstellen falls nicht vorhanden
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		slog.Error("Ordner erstellen", "error", err)
		return
	}

	// Pfad zusammensetzen: "data/roomName.json"
	filename := filepath.Join(outputDir, room+".json")
	err = os.WriteFile(filename, formatted, 0644)
	if err != nil {
		slog.Error("Datei schreiben", "error", err)
		return
	}
}
