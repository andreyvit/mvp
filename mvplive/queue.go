package mvplive

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/andreyvit/mvp/flake"
)

const (
	debugLog = false
)

type QueueOptions struct {
	TTL                   time.Duration
	MaxMessagesPerChannel int
}

type Queue struct {
	ttl   time.Duration
	limit int

	mut  sync.Mutex
	cond *sync.Cond
	msgs map[Channel][]*Msg
}

func NewQueue(opt QueueOptions) *Queue {
	if opt.TTL == 0 {
		opt.TTL = time.Minute
	}
	if opt.MaxMessagesPerChannel == 0 {
		opt.MaxMessagesPerChannel = 100
	}
	q := &Queue{
		ttl:   opt.TTL,
		limit: opt.MaxMessagesPerChannel,
		msgs:  make(map[Channel][]*Msg),
	}
	q.cond = sync.NewCond(&q.mut)
	return q
}

func (q *Queue) Push(ch Channel, msg *Msg) {
	q.mut.Lock()
	defer q.mut.Unlock()
	q.pushWithLockHeld(ch, msg, q.cutoffIDByID(msg.ID))
	q.cond.Broadcast()
}

func (q *Queue) Await(ctx context.Context, ch Channel, afterID flake.ID) []*Msg {
	q.mut.Lock()
	defer q.mut.Unlock()

	defer translateContextDoneToCondBroadcast(ctx, q.cond)

	if debugLog {
		log.Printf("mvplive: channel %v: Await(after=%v): start", ch, afterID)
	}
	for {
		msgs := q.messagesAfterWithLockHeld(ch, afterID, q.cutoffIDByTime(time.Now()))
		if debugLog {
			log.Printf("mvplive: channel %v: Await(after=%v): msgs=%d", ch, afterID, len(msgs))
		}
		if msgs != nil || ctx.Err() != nil {
			if debugLog {
				log.Printf("mvplive: channel %v: Await(after=%v): end", ch, afterID)
			}
			return msgs
		}
		q.cond.Wait()
	}
}

func translateContextDoneToCondBroadcast(ctx context.Context, cond *sync.Cond) (cancel func()) {
	if ctx == nil {
		return nop
	}
	donec := ctx.Done()
	if donec == nil {
		return nop
	}

	cancelc := make(chan struct{})

	go func() {
		select {
		case <-donec:
			cond.Broadcast()
		case <-cancelc:
			break
		}
	}()

	return func() { close(cancelc) }
}

func nop() {}

func (q *Queue) MessagesAfter(ch Channel, afterID flake.ID, now time.Time) []*Msg {
	q.mut.Lock()
	defer q.mut.Unlock()
	return q.messagesAfterWithLockHeld(ch, afterID, q.cutoffIDByTime(now))
}

func (q *Queue) messagesAfterWithLockHeld(ch Channel, afterID, cutoffID flake.ID) []*Msg {
	msgs := q.msgs[ch]

	lower := q.determineLowerBound(msgs, cutoffID, 0)
	if lower > 0 {
		if debugLog {
			log.Printf("mvplive: channel %v: trimming %d while retriving", ch, lower)
		}
		copy(msgs, msgs[lower:])
		msgs = msgs[:len(msgs)-lower]
		q.msgs[ch] = msgs
	}

	for i, msg := range msgs {
		if msg.ID > afterID {
			result := make([]*Msg, len(msgs)-i)
			copy(result, msgs[i:])
			return result
		}
	}
	return nil
}

func (q *Queue) pushWithLockHeld(ch Channel, msg *Msg, cutoffID flake.ID) {
	msgs := q.msgs[ch]
	lower := q.determineLowerBound(msgs, cutoffID, 1)
	if lower > 0 {
		if debugLog {
			log.Printf("mvplive: channel %v: trimming %d while pushing", ch, lower)
		}
		copy(msgs, msgs[lower:])
		msgs = msgs[:len(msgs)-lower]
	}
	msgs = append(msgs, msg)
	q.msgs[ch] = msgs
}

func (q *Queue) determineLowerBound(msgs []*Msg, cutoffID flake.ID, aboutToBeAdded int) int {
	n := len(msgs)
	lower := 0
	if q.limit > 0 && (n+aboutToBeAdded) > q.limit {
		lower = (n + aboutToBeAdded) - q.limit
	}
	for lower < n && msgs[lower].ID < cutoffID {
		lower++
	}
	return lower
}

func (q *Queue) cutoffIDByID(id flake.ID) flake.ID {
	ms := id.Milliseconds()
	delta := uint64(q.ttl.Milliseconds())
	if ms > delta {
		return flake.Build(ms-delta, 0, 0)
	} else {
		return 0
	}
}

func (q *Queue) cutoffIDByTime(now time.Time) flake.ID {
	return flake.MinAt(now.Add(-q.ttl))
}
