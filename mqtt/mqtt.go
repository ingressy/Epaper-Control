package mqtt

import (
	"Control/handler"
	"Control/influx"
	"Control/types"
	"Control/untis"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	awakeTimes = make(map[string]time.Time)
	awakeMu    sync.Mutex
)

func ConnecttoMQTT() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://172.20.0.5:1883")
	opts.SetClientID("Epaper-Control")

	//auto reconnect
	opts.SetAutoReconnect(true)
	opts.OnConnect = func(c mqtt.Client) {
		subscribe(c)
	}

	client := mqtt.NewClient(opts)

	for {
		token := client.Connect()
		token.Wait()
		if token.Error() == nil {
			break
		}
		slog.Error("MQTT Verbindung fehlgeschlagen, versuche es in 10s erneut", "error", token.Error())
		time.Sleep(10 * time.Second)
	}

	untis.Get_room_from_json()
	handler.PrepareJSON()
	handler.Getpicturehex()

	select {}
}

func subscribe(client mqtt.Client) {
	client.Subscribe("+/awake", 0, onAwake)
	client.Subscribe("+/gn", 0, onGN)
}

func onAwake(c mqtt.Client, msg mqtt.Message) {
	parts := strings.Split(msg.Topic(), "/")
	if len(parts) < 2 {
		return
	}
	id := parts[0]
	awakeMu.Lock()
	awakeTimes[id] = time.Now()
	awakeMu.Unlock()

	payload := strings.Split(string(msg.Payload()), ",")
	if len(payload) < 3 {
		return
	}
	batteryvolt := payload[1]
	errorcode := payload[2]

	slog.Info("EPD wach", "id", id, "akku", batteryvolt, "errorcode", errorcode)

	battery, _ := strconv.ParseFloat(batteryvolt, 8)
	influx.SaveBatteryInflux(id, battery)

	room := types.GetRoomfromID(id)
	wasNight, _ := types.GetNightsleep(id)
	handler.IsNightSleep(id, room)

	isNight, err := types.GetNightsleep(id)
	slog.Debug("NightSleep State", "id", id, "isNight", isNight, "err", err)

	if !isNight {
		//slog.Debug("Es ist normaler Tag und dem EPD wird ein Bild zugesendet!")
		SendImage(c, id)
	} else if !wasNight && isNight {
		//slog.Debug("Übergang Tag→Nacht, sende letztes Bild")
		SendImage(c, id)
	} else {
		//slog.Debug("Es ist Nacht und das EPD soll schlafen!")
		sendsleep(c, id)
	}
}

func onGN(c mqtt.Client, msg mqtt.Message) {
	topicParts := strings.Split(msg.Topic(), "/")
	if len(topicParts) < 2 {
		return
	}
	id := topicParts[0]

	// Payload: "goodnight,600"
	parts := strings.Split(string(msg.Payload()), ",")
	if len(parts) < 2 {
		return
	}

	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		slog.Error("Fehler", "error", err)
		return
	}

	// Zeit messen
	awakeMu.Lock()
	start, ok := awakeTimes[id]
	if ok {
		delete(awakeTimes, id)
	}
	awakeMu.Unlock()

	if ok {
		duration := time.Since(start)
		slog.Info("EPD Refresh-Zeit", "id", id, "duration", duration)

		influx.SaveRefreshTimeInflux(id, duration)
	}

	slog.Info("ESP32 schläft", "seconds", seconds)
	wakeTime := time.Now().Add(time.Duration(seconds) * time.Second)
	slog.Info("Wacht auf um", "time", wakeTime.Format("15:04:05"))

	c.Publish(id+"/sleep", 1, true, "")
}

func sendsleep(c mqtt.Client, id string) {
	responseTopic := id + "/sleep"

	var sekunden int

	if !types.Config.Wartung {
		sekunden = handler.Getwakeuptime(types.GetRoomfromID(id)) * 60
	} else {
		sekunden = types.Config.Wartung_sleep_time * 60
	}

	if sekunden < 0 {
		slog.Error("Negative Schlafdauer berechnet, sende keinen Sleep-Befehl", "id", id, "sekunden", sekunden)
		return
	}

	send := c.Publish(responseTopic, 1, true, fmt.Sprintf("%d", sekunden))
	send.Wait()
	slog.Info("EPD geht schlafen")

	//check if handler/cache/room.json is emtpy = night for epd
	//room := types.GetRoomfromID(id)
	//handler.IsNightSleep(id, room)
}
