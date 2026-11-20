package influx

import (
	"Control/types"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var influxClient influxdb2.Client

func InitInflux() {
	influxClient = influxdb2.NewClient(
		"http://172.20.0.6:8086",
		types.GetConfig().InfluxToken,
	)
}

func SaveBatteryInflux(id string, battery float64) {
	writeAPI := influxClient.WriteAPI(
		"epaper",
		"display",
	)

	p := influxdb2.NewPointWithMeasurement("battery").AddTag("device", id).AddField("volt", battery).SetTime(time.Now())

	writeAPI.WritePoint(p)
	writeAPI.Flush()
}
