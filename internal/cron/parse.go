package cron

import (
	"time"

	robfigcron "github.com/robfig/cron/v3"
)

var defaultParser = robfigcron.NewParser(
	robfigcron.Minute | robfigcron.Hour | robfigcron.Dom | robfigcron.Month | robfigcron.Dow | robfigcron.Descriptor,
)

func parseSchedule(expr string) (robfigcron.Schedule, error) {
	return defaultParser.Parse(expr)
}

func loadTimezone(tz string) (*time.Location, error) {
	return time.LoadLocation(tz)
}
