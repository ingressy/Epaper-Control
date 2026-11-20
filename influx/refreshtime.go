package influx

import (
	"context"
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func SaveRefreshTimeInflux(id string, d time.Duration) {
	writeAPI := influxClient.WriteAPIBlocking(
		"epaper",
		"display",
	)
	p := influxdb2.NewPointWithMeasurement("refreshtime").
		AddTag("device", id).
		AddField("seconds", d.Seconds()).
		SetTime(time.Now())
	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		slog.Error("Influx RefreshTime Write Fehler", "error", err)
	}
}
