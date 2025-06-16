package cronjob

type Job struct {
	ctx          *Context
	handlerChain []Handler
}

func NewJob(c *CronJob, handler Handler) *Job {
	j := &Job{}
	j.handlerChain = append(c.handlers, handler)
	return j
}

func (s *Job) Run() {
	s.ctx = &Context{}
	s.ctx.reset(nil)
	s.ctx.handlerChain = s.handlerChain
	s.ctx.Next()
}
