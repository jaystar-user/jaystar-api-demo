package job

import (
	"context"
	"github.com/SeanZhenggg/go-utils/logger"
	"jaystar/internal/controller/job"
	"jaystar/internal/controller/job/middleware"
	"jaystar/internal/cronjob"
	"jaystar/internal/utils/osUtil"
	"log"
)

type IJob interface {
	Init()
	Start()
	Stop(ctx context.Context)
}

type Job struct {
	cronJob *cronjob.CronJob `wire:"-"`
	ctrl    *job.JobController
	mw      middleware.IJobMiddleware
	logger  logger.ILogger
}

func ProvideJob(ctrl *job.JobController, mw middleware.IJobMiddleware) IJob {
	return &Job{
		cronJob: cronjob.ProvideCronJob(),
		ctrl:    ctrl,
		mw:      mw,
	}
}

func (j *Job) Init() {
	j.cronJob.Use(j.mw.Handle)

	if !osUtil.IsLocal() {
		err := j.cronJob.AddSchedule("0 0 4 * * *", j.ctrl.SemesterSettlement)
		if err != nil {
			log.Fatalf("cron job AddSchedule failed: %v", err)
		}
	}
}

func (j *Job) Start() {
	j.cronJob.Start()
}

func (j *Job) Stop(ctx context.Context) {
	cronCtx := j.cronJob.Stop()

	select {
	case <-cronCtx.Done():
	case <-ctx.Done():
		j.logger.Warn(ctx, "cron job graceful stop failed")
	}
}
