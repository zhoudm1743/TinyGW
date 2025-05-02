package channel

// PriorityChannel 具有优先级的双通道
type PriorityChannel interface {
	Start()                                // 开启任务
	Stop()                                 // 结束任务
	DispatchPriorTask(task any)            // 指派高优先级任务
	DispatchNormalTask(task any)           // 指派正常任务
	SetPriorWorker(worker func(task any))  // 设置任务执行器
	SetNormalWorker(worker func(task any)) // 设置任务执行器
	NormalTaskIsFull() bool                // 采集任务队列是否已满
}

var _ PriorityChannel = (*priorityChannel)(nil)

type priorityChannel struct {
	priorChan    chan any
	normalChan   chan any
	stopChan     chan struct{}
	priorWorker  func(any)
	normalWorker func(any)
}

func NewPriorityChannel() PriorityChannel {
	return &priorityChannel{
		priorChan:  make(chan any, 64),
		normalChan: make(chan any, 128),
		stopChan:   make(chan struct{}, 1),
	}
}

func (pc *priorityChannel) Start() {
	go pc.Worker()
}

func (pc *priorityChannel) Stop() {
	pc.stopChan <- struct{}{}
}
func (pc *priorityChannel) DispatchPriorTask(task any) {
	pc.priorChan <- task
}

func (pc *priorityChannel) DispatchNormalTask(task any) {
	pc.normalChan <- task
}

func (pc *priorityChannel) SetPriorWorker(worker func(task any)) {
	pc.priorWorker = worker
}

func (pc *priorityChannel) SetNormalWorker(worker func(task any)) {
	pc.normalWorker = worker
}

func (pc *priorityChannel) NormalTaskIsFull() bool {
	return len(pc.normalChan) >= 128
}

// Worker1 算法参见：https://blog.csdn.net/hurray123/article/details/50038329/
func (pc *priorityChannel) Worker1() {
	for {
		select {
		case <-pc.stopChan:
			return
		case priorTask := <-pc.priorChan:
			pc.priorWorker(priorTask)
		default:
			select {
			case priorTask := <-pc.priorChan:
				pc.priorWorker(priorTask)
			case normalTask := <-pc.normalChan:
				pc.normalWorker(normalTask)
			}
		}
	}
}

// Worker 另外一种写法：https://www.csdn.net/tags/MtTaQg4sMTk5LWJsb2cO0O0O.html
//
//	https://www.zhihu.com/question/469460715/answer/1974574115
func (pc *priorityChannel) Worker() {
	for {
		select {
		case <-pc.stopChan:
			return
		case priorTask := <-pc.priorChan:
			pc.priorWorker(priorTask)
		case normalTask := <-pc.normalChan:
		priority:
			for {
				select {
				case priorTask := <-pc.priorChan:
					pc.priorWorker(priorTask)
				default:
					break priority
				}
			}
			pc.normalWorker(normalTask)
		}
	}
}
