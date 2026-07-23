package queue

import (
	"context"
	"easycelery/internal/task"
	"log/slog"
	"sync"
	"time"
)

type Scheduler struct {
	mu sync.Mutex

	heap *DelayedHeap

	queue Queue

	wakeup chan struct{}

	timer *time.Timer
}

func NewScheduler(queue Queue) *Scheduler {
	return &Scheduler{
		heap:   NewDelayedHeap(),
		queue:  queue,
		wakeup: make(chan struct{}),
	}
}

type DelayedHeap struct {
	mu           sync.Mutex
	delayedTasks []*DelayedTask
}

func NewDelayedHeap() *DelayedHeap {
	return &DelayedHeap{
		delayedTasks: make([]*DelayedTask, 0),
	}
}

type DelayedTask struct {
	task *task.Task

	executeAt time.Time

	index int
}

func (d *DelayedTask) Before(other *DelayedTask) bool {
	return d.executeAt.Before(other.executeAt)
}

func (d *DelayedTask) After(other *DelayedTask) bool {
	return d.executeAt.After(other.executeAt)
}

func (d *DelayedTask) Equal(other *DelayedTask) bool {
	return d.executeAt.Equal(other.executeAt)
}
func (h *DelayedHeap) Push(task *DelayedTask) {
	h.mu.Lock()
	defer h.mu.Unlock()
	task.index = len(h.delayedTasks) - 1
	h.delayedTasks = append(h.delayedTasks, task)

	h.shiftUp(len(h.delayedTasks) - 1)
}

func (h *DelayedHeap) Peek() *DelayedTask {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.delayedTasks) == 0 {
		return nil
	}
	return h.delayedTasks[0]
}

func (h *DelayedHeap) Pop() *DelayedTask {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.delayedTasks) == 0 {
		return nil
	}

	if len(h.delayedTasks) == 1 {
		poppedTask := h.delayedTasks[0]
		h.delayedTasks = h.delayedTasks[1:]
		return poppedTask
	}

	poppedTask := h.delayedTasks[0]
	last := len(h.delayedTasks) - 1
	h.delayedTasks[0] = h.delayedTasks[last]
	h.delayedTasks = h.delayedTasks[:last]
	h.delayedTasks[0].index = 0

	h.shiftDown(0)

	return poppedTask
}

func (h *DelayedHeap) right(i int) int {
	return 2*i + 2
}

func (h *DelayedHeap) left(i int) int {
	return 2*i + 1
}

func (h *DelayedHeap) parent(i int) int {
	return (i - 1) / 2
}

func (h *DelayedHeap) shiftUp(index int) {
	if index == 0 {
		return
	}

	parentIdx := h.parent(index)
	parent := h.get(parentIdx)
	elem := h.get(index)

	if elem.Before(parent) {
		h.swap(parentIdx, index)
		h.shiftUp(parentIdx)
	}
}

func (h *DelayedHeap) swap(i1 int, i2 int) {
	h.delayedTasks[i1], h.delayedTasks[i2] = h.delayedTasks[i2], h.delayedTasks[i1]

	h.delayedTasks[i1].index = i1
	h.delayedTasks[i2].index = i2
}

func (h *DelayedHeap) get(index int) *DelayedTask {
	if index < len(h.delayedTasks) && index >= 0 {
		return h.delayedTasks[index]
	}
	return nil
}

func (h *DelayedHeap) shiftDown(index int) {
	rightIdx, leftIdx := h.right(index), h.left(index)
	right := h.get(rightIdx)
	left := h.get(leftIdx)
	elem := h.get(index)
	if elem == nil {
		return
	}
	if right != nil && (left == nil || right.Before(left)) {
		if right.Before(elem) {
			h.swap(rightIdx, index)
			h.shiftDown(rightIdx)
			return
		}
	} else if left != nil {
		if left.Before(elem) {
			h.swap(leftIdx, index)
			h.shiftDown(leftIdx)
			return
		}
	}
}

func (s *Scheduler) Add(task *task.Task, delay time.Duration) {
	delayedTask := &DelayedTask{
		task:      task,
		executeAt: time.Now().Add(delay),
	}
	s.heap.Push(delayedTask)

	select {
	case s.wakeup <- struct{}{}:
	default:
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	for {
		var timerC <-chan time.Time

		if s.timer != nil {
			timerC = s.timer.C
		}
		select {
		case <-timerC:
			if s.heap.Peek().executeAt.After(time.Now()) {
				break
			}
			delayedTask := s.heap.Pop()
			s.queue.Push(delayedTask.task)
			peek := s.heap.Peek()
			if peek != nil {
				s.timer.Reset(time.Until(s.heap.Peek().executeAt))
			}
		case <-s.wakeup:
			if s.timer != nil {
				peek := s.heap.Peek()
				if peek != nil {
					s.timer.Reset(time.Until(s.heap.Peek().executeAt))
				}
			} else {
				peek := s.heap.Peek()
				if peek != nil {
					s.timer = time.NewTimer(time.Until(s.heap.Peek().executeAt))
				}
			}
		case <-ctx.Done():
			slog.Info("Scheduler stopping due to context stop")
			return
		}
	}
}
