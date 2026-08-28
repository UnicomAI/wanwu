package shared

import (
	"fmt"
	"sync"
)

// HaltState 累计单次 sandbox 会话内的连续 [BLOCKED:...] 次数。
// 连续到达 threshold 触发 haltFn（通常绑定到 context.CancelCauseFunc）主动终止会话，
// 让 eino runner iterator 关闭，上层通过 BuildFinalAgentEvent 下发 error[agent] 兜底消息。
//
// 语义：连续计数——任一非 BLOCKED 输出重置为 0；只有反复 BLOCKED 才会累积到阈值。
// 合法业务下第一次误伤后 LLM 通常一两次内切到正确路径，counter 会被重置，不会误熔断。
//
// 并发：Execute 可能被 eino ADK 从不同 goroutine 调用；haltFn 只调用一次。
type HaltState struct {
	threshold int
	haltFn    func(error)

	mu      sync.Mutex
	counter int
	halted  bool
}

// NewHaltState 构造 HaltState。threshold <= 0 视为禁用（NoteBlocked 永不触发 halt）；
// haltFn 可为 nil（仅记账不通知）。
func NewHaltState(threshold int, haltFn func(error)) *HaltState {
	return &HaltState{
		threshold: threshold,
		haltFn:    haltFn,
	}
}

// NoteBlocked 累计一次 BLOCKED 输出。若刚跨过阈值触发 haltFn 并返回 true；
// 已熔断的后续调用不会重复触发 haltFn，仍返回 false。
//
// 调用契约：haltFn 不得从内部反向调用 HaltState 的任何方法（NoteBlocked/NoteSuccess/Halted），
// 否则会死锁。当前生产用法 haltFn = cancelSession(err) 只是设置 context 取消原因，符合契约。
func (s *HaltState) NoteBlocked() (justHalted bool) {
	if s == nil || s.threshold <= 0 {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.halted {
		return false
	}
	s.counter++
	if s.counter >= s.threshold {
		s.halted = true
		if s.haltFn != nil {
			s.haltFn(fmt.Errorf("consecutive security blocks reached threshold %d", s.threshold))
		}
		return true
	}
	return false
}

// NoteSuccess 重置连续计数。任何非 BLOCKED 输出（成功执行、普通报错）均调用。
// 已熔断状态不重置——一旦熔断本会话内不可恢复。
func (s *HaltState) NoteSuccess() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.halted {
		return
	}
	s.counter = 0
}

// Halted 返回是否已熔断。
func (s *HaltState) Halted() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.halted
}
