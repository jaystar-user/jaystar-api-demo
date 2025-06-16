package cronjob

import (
	"context"
	"github.com/robfig/cron/v3"
)

type CronJob struct {
	pkgCron  *cron.Cron `wire:"-"`
	handlers []Handler
}

func ProvideCronJob() *CronJob {
	return &CronJob{
		pkgCron:  cron.New(cron.WithSeconds()),
		handlers: make([]Handler, 0),
	}
}

func (j *CronJob) AddSchedule(spec string, cmd func(ctx *Context)) error {
	job := NewJob(j, cmd)
	_, err := j.pkgCron.AddJob(spec, job)
	return err
}

func (j *CronJob) Use(handlers ...Handler) {
	j.handlers = append(j.handlers, handlers...)
}

func (j *CronJob) Start() {
	j.pkgCron.Start()
}

func (j *CronJob) Stop() context.Context {
	return j.pkgCron.Stop()
}
