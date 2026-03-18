package scanapp

type dispatchObserver interface {
	OnGateWaitStart(cidr string, taskIndex int)
	OnGateReleased(cidr string, taskIndex int)
	OnBucketWaitStart(cidr string, taskIndex int)
	OnBucketAcquired(cidr string, taskIndex int)
	OnTaskEnqueued(cidr string, taskIndex int)
}

type noopDispatchObserver struct{}

func (noopDispatchObserver) OnGateWaitStart(string, int) {}

func (noopDispatchObserver) OnGateReleased(string, int) {}

func (noopDispatchObserver) OnBucketWaitStart(string, int) {}

func (noopDispatchObserver) OnBucketAcquired(string, int) {}

func (noopDispatchObserver) OnTaskEnqueued(string, int) {}
