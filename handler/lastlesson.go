package handler

import (
	"Control/types"
	"encoding/json"
	"log/slog"
	"os"
)

func IsNightSleep(id string, room string) {
	data, err := os.ReadFile("handler/cache/" + room + ".json")
	if err != nil {
		return
	}
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}
	if len(resp.Lessons) == 0 {
		slog.Debug("Es gibt kein Unterricht in diesen Raum mehr, für das EPD:", "id", id)
		if err := types.SetNightSleep(id, true); err != nil {
			slog.Error("NightSleep setzen fehlgeschlagen", "id", id, "error", err)
		}
	} else {
		if err := types.SetNightSleep(id, false); err != nil {
			slog.Error("NightSleep setzen fehlgeschlagen", "id", id, "error", err)
		}
	}
}
