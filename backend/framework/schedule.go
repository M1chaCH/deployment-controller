package framework

import (
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"time"
)

func RunScheduledTask(taskName string, sleepMinutes int, task func() error) {
	go run(taskName, sleepMinutes, task)
}

func run(taskName string, sleepMinutes int, task func() error) {
	originalSleepConfig := sleepMinutes
	sleepConfig := sleepMinutes
	sleepDuration := time.Duration(sleepConfig) * time.Minute
	failed := false
	maxSleepDuration := time.Duration(24) * time.Hour * 3 // max wait 3 days

	logs.Info(nil, "periodically scheduled check (every %d min): '%s'", sleepConfig, taskName)
	for {
		time.Sleep(sleepDuration)

		logs.Debug(nil, "SCHEDULE: running task: "+taskName)
		err := task()

		if err != nil {
			if sleepDuration > maxSleepDuration {
				sleepConfig *= 2
				sleepDuration *= 2
			}
			logs.Warn(nil, "SCHEDULE: Failed to run check '%s' (will try again in %d mins): %v", taskName, sleepConfig, err)
			failed = true
		} else if failed { // previous run failed, but this run ran successfully
			failed = false
			sleepConfig = originalSleepConfig
			sleepDuration = time.Duration(sleepConfig) * time.Minute
			logs.Info(nil, "SCHEDULE: recovered failed check '%s'", taskName)
		}

		if err == nil {
			logs.Debug(nil, "SCHEDULE: successfully ran task: %s, next run in %d minutes", taskName, sleepConfig)
		}
	}
}
