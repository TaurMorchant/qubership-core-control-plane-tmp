package queue

import (
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"sync"
)

const TaskExecutorBuffSize = 100

var logger logging.Logger

func init() {
	logger = logging.GetLogger("queue")
}

type GroupTaskExecutor struct {
	groupQueues *sync.Map
}

func NewGroupTaskExecutor() *GroupTaskExecutor {
	return &GroupTaskExecutor{groupQueues: &sync.Map{}}
}

func (exec *GroupTaskExecutor) Execute(group interface{}, task *AsyncTask) {
	taskExecutor, _ := exec.groupQueues.LoadOrStore(group, NewTaskExecutor(TaskExecutorBuffSize))
	taskExecutor.(*TaskExecutor).Execute(task)
}

func (exec *GroupTaskExecutor) ShutdownGroup(group interface{}) {
	if taskExecutor, found := exec.groupQueues.Load(group); found {
		taskExecutor.(*TaskExecutor).Shutdown()
	}
	exec.groupQueues.Delete(group)
}

func (exec *GroupTaskExecutor) ShutdownAll() {
	exec.groupQueues.Range(func(group, taskExecutor interface{}) bool {
		taskExecutor.(*TaskExecutor).Shutdown()
		exec.groupQueues.Delete(group)
		return true
	})
}

type AsyncTask struct {
	task func(args ...interface{})
	args []interface{}
}

func (task *AsyncTask) Run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("AsyncTask failed and recovered: %v", r)
		}
	}()
	task.task(task.args)
}

func NewAsyncTask(task func(args ...interface{}), args ...interface{}) *AsyncTask {
	return &AsyncTask{task: task, args: args}
}

type TaskExecutor struct {
	queue chan taskQueueMsg
}

type taskQueueMsg struct {
	asyncTask *AsyncTask
	terminal  bool
}

func NewTaskExecutor(buffSize int) *TaskExecutor {
	exec := TaskExecutor{queue: make(chan taskQueueMsg, buffSize)}
	go exec.waitForTasksAndExecute()
	return &exec
}

func (exec *TaskExecutor) Shutdown() {
	exec.queue <- taskQueueMsg{terminal: true}
}

func (exec *TaskExecutor) Execute(task *AsyncTask) {
	exec.queue <- taskQueueMsg{terminal: false, asyncTask: task}
}

func (exec *TaskExecutor) waitForTasksAndExecute() {
	for {
		task := <-exec.queue
		if task.terminal {
			return
		}
		task.asyncTask.Run()
	}
}
