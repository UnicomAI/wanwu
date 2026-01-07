package trans

import (
	"fmt"
)

// TaskState 任务状态
type TaskState int32

const (
	TaskStateInit     TaskState = 0 // 已创建
	TaskStatePending  TaskState = 1 // 排队中
	TaskStateRunning  TaskState = 2 // 运行中
	TaskStateDeleting TaskState = 3 // 删除中
	TaskStatePause    TaskState = 4 // 暂停
	TaskStateFinished TaskState = 5 // 结束
	TaskStateFailed   TaskState = 6 // 失败
)

// TaskMark 任务标记，任务期望的操作
type TaskMark int32

const (
	TaskMarkNone   TaskMark = 0 // 无标记
	TaskMarkRun    TaskMark = 1 // 标记需要开始
	TaskMarkDelete TaskMark = 2 // 标记需要删除
	TaskMarkPause  TaskMark = 3 // 标记需要暂停
)

// TaskEvent 任务事件
type TaskEvent int32

const (
	EventNone TaskEvent = 0
)

// 用户事件
const (
	EventUserRun    TaskEvent = 1 // 用户运行任务
	EventUserDelete TaskEvent = 2 // 用户删除任务
	EventUserPause  TaskEvent = 3 // 用户暂停任务
)

// 系统事件
const (
	EventSysExecute TaskEvent = 4 // 系统执行任务
	EventSysPause   TaskEvent = 6 // 系统暂停任务
)

// 任务事件
const (
	EventTaskFinished TaskEvent = 7 // 任务结束
	EventTaskFailed   TaskEvent = 8 // 任务失败
)

// TaskStatus 任务{状态, 标记}组合
type TaskStatus struct {
	S TaskState
	M TaskMark
}

func CheckTransfer(taskStatus TaskStatus, event TaskEvent) (TaskStatus, bool, error) {
	switch taskStatus {

	// 任务创建初始状态，不会排队、运行
	case TaskStatus{S: TaskStateInit, M: TaskMarkNone}:
		switch event {
		case EventUserRun:
			// 进入pending.runQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkRun}, false, nil
		case EventUserDelete:
			// 直接删除任务记录
			return TaskStatus{}, true, nil
		case EventUserPause:
			// 直接暂停
			return TaskStatus{S: TaskStatePause, M: TaskMarkPause}, false, nil
		}

	// 任务在pending.runQueue中排队，用户可见为排队中
	case TaskStatus{S: TaskStatePending, M: TaskMarkRun}:
		switch event {
		case EventUserDelete:
			// 离开pending.runQueue，进入pending.deleteQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkDelete}, false, nil
		case EventUserPause:
			// 离开pending.runQueue，直接暂停
			return TaskStatus{S: TaskStatePause, M: TaskMarkPause}, false, nil
		case EventSysExecute:
			// 离开pending.runQueue，进入内存执行run
			return TaskStatus{S: TaskStateRunning, M: TaskMarkRun}, false, nil
		default:
		}

	// 任务在pending.deleteQueue中排队，用户不可见
	case TaskStatus{S: TaskStatePending, M: TaskMarkDelete}:
		switch event {
		case EventSysExecute:
			// 离开pending.deleteQueue，进入内存执行delete
			return TaskStatus{S: TaskStateDeleting, M: TaskMarkDelete}, false, nil
		default:
		}

	// 任务在内存中执行run，用户可见为运行中
	case TaskStatus{S: TaskStateRunning, M: TaskMarkRun}:
		switch event {
		case EventUserDelete:
			// 用户标记删除
			return TaskStatus{S: TaskStateRunning, M: TaskMarkDelete}, false, nil
		case EventUserPause:
			// 用户标记暂停
			return TaskStatus{S: TaskStateRunning, M: TaskMarkPause}, false, nil
		case EventSysPause:
			// 停止执行，离开内存
			return TaskStatus{S: TaskStatePause, M: TaskMarkRun}, false, nil
		case EventTaskFinished:
			// 执行结束，离开内存
			return TaskStatus{S: TaskStateFinished, M: TaskMarkRun}, false, nil
		case EventTaskFailed:
			// 执行失败，离开内存
			return TaskStatus{S: TaskStateFailed, M: TaskMarkRun}, false, nil
		default:
		}

	// 任务在内存中执行run，但用户标记删除，用户可见为取消中
	case TaskStatus{S: TaskStateRunning, M: TaskMarkDelete}:
		switch event {
		case EventSysPause, EventTaskFinished, EventTaskFailed:
			// 停止执行/执行结束，离开内存，进入pending.deleteQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkDelete}, false, nil
		default:
		}

	// 任务在内存中执行run，用户标记暂停，用户可见为取消中
	case TaskStatus{S: TaskStateRunning, M: TaskMarkPause}:
		switch event {
		case EventSysPause:
			// 停止执行，离开内存
			return TaskStatus{S: TaskStatePause, M: TaskMarkPause}, false, nil
		case EventTaskFinished:
			// 执行结束，离开内存
			return TaskStatus{S: TaskStateFinished, M: TaskMarkRun}, false, nil
		case EventTaskFailed:
			// 执行失败，离开内存
			return TaskStatus{S: TaskStateFailed, M: TaskMarkRun}, false, nil
		default:
		}

	// 任务在内存中执行delete，用户不可见
	case TaskStatus{S: TaskStateDeleting, M: TaskMarkDelete}:
		switch event {
		case EventSysPause:
			// 停止执行，离开内存
			return TaskStatus{S: TaskStatePause, M: TaskMarkDelete}, false, nil
		case EventTaskFinished:
			// 执行结束，直接删除任务记录
			return TaskStatus{}, true, nil
		case EventTaskFailed:
			// 执行结束，离开内存
			return TaskStatus{S: TaskStateFailed, M: TaskMarkDelete}, false, nil
		}

	// 任务暂停run，但用户可见为运行中
	case TaskStatus{S: TaskStatePause, M: TaskMarkRun}:
		switch event {
		case EventUserPause:
			// 直接暂停
			return TaskStatus{S: TaskStatePause, M: TaskMarkPause}, false, nil
		case EventUserDelete:
			// 进入pending.deleteQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkDelete}, false, nil
		case EventSysExecute:
			// 进入内存执行run
			return TaskStatus{S: TaskStateRunning, M: TaskMarkRun}, false, nil
		default:
		}

	// 任务暂停delete，用户不可见
	case TaskStatus{S: TaskStatePause, M: TaskMarkDelete}:
		switch event {
		case EventSysExecute:
			// 进入内存执行delete
			return TaskStatus{S: TaskStateDeleting, M: TaskMarkDelete}, false, nil
		default:
		}

	// 任务暂停run，用户可见为暂停中
	case TaskStatus{S: TaskStatePause, M: TaskMarkPause}:
		switch event {
		case EventUserRun:
			// 进入pending.runQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkRun}, false, nil
		case EventUserDelete:
			// 进入pending.deleteQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkDelete}, false, nil
		default:
		}

	// 任务执行run结束，用户可见为结束
	case TaskStatus{S: TaskStateFinished, M: TaskMarkRun}:
		switch event {
		case EventUserDelete:
			// 进入pending.deleteQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkDelete}, false, nil
		default:
		}

	// 任务执行run失败，用户可见为失败
	case TaskStatus{S: TaskStateFailed, M: TaskMarkRun}:
		switch event {
		case EventUserRun:
			// 失败后重启，进入pending.runQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkRun}, false, nil
		case EventUserDelete:
			// 进入pending.deleteQueue
			return TaskStatus{S: TaskStatePending, M: TaskMarkDelete}, false, nil
		default:
		}

	default:
	}

	return TaskStatus{}, false, fmt.Errorf("{state %v, mark %v} event %v transfer invalid",
		taskStatus.S, taskStatus.M, event)
}
