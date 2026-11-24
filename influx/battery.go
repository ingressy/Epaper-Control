package influx

import (
	"Control/types"
	"context"
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var influxClient influxdb2.Client

func InitInflux() {
	influxClient = influxdb2.NewClient(
		types.InfluxURL,
		types.GetConfig().InfluxToken,
	)
}

func SaveBatteryInflux(id string, battery float64) {
	writeAPI := influxClient.WriteAPIBlocking(
		"epaper",
		"display",
	)

	p := influxdb2.NewPointWithMeasurement("battery").AddTag("device", id).AddField("volt", battery).SetTime(time.Now())

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		slog.Error("Influx Battery Write Fehler", "error", err)
	}
}
