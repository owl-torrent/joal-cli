package stop

import "context"

type Request interface {
	Ctx() context.Context
	// AwaitDone waits for a routine to call NotifyDone, returns an error if the context expires before a call to NotifyDone
	AwaitDone() error
	NotifyDone()
}

type request struct {
	ctx          context.Context
	doneStopping chan struct{}
}

func NewRequest(ctx context.Context) Request {
	return &request{
		ctx:          ctx,
		doneStopping: make(chan struct{}),
	}
}

func (r *request) AwaitDone() error {
	select {
	case <-r.doneStopping:
		return nil
	case <-r.ctx.Done():
		return r.ctx.Err()
	}
}

func (r *request) NotifyDone() {
	r.doneStopping <- struct{}{}
}

func (r *request) Ctx() context.Context {
	return r.ctx
}

type Chan = chan Request

func NewChan() Chan {
	return make(chan Request)
}
