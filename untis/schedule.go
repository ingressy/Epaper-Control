package untis

import "github.com/robfig/cron/v3"

func schedule() {

	// pause 09:40 - 10:00
	// 11:30 - 11:45
	// 13:15 - 13:45
	// 15:15 - 15:30
	// 17:00 - 17:15

	c := cron.New()

	c.AddFunc("30 9 * * *", func() {
	})

	c.AddFunc("10 10 * * *", func() {
	})

	c.Start()

	select {}

}
