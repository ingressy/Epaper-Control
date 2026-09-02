package handler

import (
	"encoding/json"
	"os"
	"time"
)

// Feste Weckzeiten, nur an Wochentagen (Mo-Fr), HH:MM, 24h
var forcedWakeTimes = []string{"07:30", "08:00"}

func Getwakeuptime(room string) int {
	result := getScheduleBasedWakeup(room)

	if forced := minutesUntilNextForcedWake(); forced < result {
		result = forced
	}

	return result
}

func getScheduleBasedWakeup(room string) int {
	data, err := os.ReadFile("handler/cache/" + room + ".json")
	if err == nil {
		var resp Response
		if err := json.Unmarshal(data, &resp); err == nil && len(resp.Lessons) > 0 {
			now := time.Now()
			nowMinutes := now.Hour()*60 + now.Minute()
			nextlesson := resp.Lessons[0]
			startMinutes := parseTime(nextlesson.StartTime)

			if startMinutes > nowMinutes {
				return startMinutes - nowMinutes
			}
			endMinutes := parseTime(nextlesson.EndTime)
			return endMinutes - nowMinutes
		}
	}

	data, err = os.ReadFile("untis/cache/" + room + ".json")
	if err != nil {
		return 10 * 60
	}
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil || len(resp.Lessons) == 0 {
		return 6 * 60
	}

	nextlesson := resp.Lessons[0]
	nextLessonDate, _ := time.ParseInLocation("2006-01-02", nextlesson.Date, time.Local)
	sleeptime := parseTime(nextlesson.StartTime)
	lessonTime := nextLessonDate.Add(time.Duration(sleeptime) * time.Minute)
	sleepDuration := time.Until(lessonTime)

	if sleepDuration > 6*time.Hour {
		return 6 * 60
	}

	return int(sleepDuration.Minutes())
}

// minutesUntilNextForcedWake sucht die nächste feste Weckzeit,
// die auf einen Wochentag (Mo-Fr) fällt - egal ob heute oder an einem der
// nächsten Tage (falls z.B. Wochenende dazwischen liegt).
func minutesUntilNextForcedWake() int {
	now := time.Now()

	// Bis zu 7 Tage in die Zukunft schauen, um sicher einen Wochentag zu finden
	for dayOffset := 0; dayOffset <= 7; dayOffset++ {
		day := now.AddDate(0, 0, dayOffset)
		weekday := day.Weekday()

		if weekday == time.Saturday || weekday == time.Sunday {
			continue
		}

		for _, t := range forcedWakeTimes {
			wakeMinutes := parseTime(t)
			wakeTime := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()).
				Add(time.Duration(wakeMinutes) * time.Minute)

			if wakeTime.After(now) {
				return int(time.Until(wakeTime).Minutes())
			}
		}
	}

	return int(^uint(0) >> 1) // sollte praktisch nie eintreten
}

func parseTime(t string) int {
	if len(t) != 4 {
		return 10 * 60
	}
	hours := int(t[0]-'0')*10 + int(t[1]-'0')
	mins := int(t[2]-'0')*10 + int(t[3]-'0')
	return hours*60 + mins
}
