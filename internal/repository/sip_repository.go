package repository

import (
	"container/heap"
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"SIP/internal/model"
	"SIP/internal/util"
)

type SIPRepository interface {
	Create(ctx context.Context, sip model.SIP) (model.SIP, error)
	GetByID(ctx context.Context, sipID string) (model.SIP, error)
	ListByUser(ctx context.Context, userID string) ([]model.SIP, error)
	ListDue(ctx context.Context, now time.Time) ([]model.SIP, error)
	Update(ctx context.Context, sip model.SIP) (model.SIP, error)
}

type InMemorySIPRepository struct {
	mu        sync.RWMutex
	sips      map[string]model.SIP
	sipHeap   sipMinHeap
	heapIndex map[string]*sipHeapItem
}

var (
	sipRepositoryOnce sync.Once
	sipRepositoryInst *InMemorySIPRepository
)

func NewInMemorySIPRepository() *InMemorySIPRepository {
	sipRepositoryOnce.Do(func() {
		sipRepositoryInst = &InMemorySIPRepository{
			sips:      map[string]model.SIP{},
			sipHeap:   sipMinHeap{},
			heapIndex: map[string]*sipHeapItem{},
		}
	})
	return sipRepositoryInst
}

func (r *InMemorySIPRepository) Create(_ context.Context, sip model.SIP) (model.SIP, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sips[sip.SIPID]; ok {
		return model.SIP{}, fmt.Errorf("%w: sip %s already exists", util.ErrConflict, sip.SIPID)
	}
	r.sips[sip.SIPID] = sip
	r.updateHeapForSip(sip)
	return sip, nil
}

func (r *InMemorySIPRepository) GetByID(_ context.Context, sipID string) (model.SIP, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sip, ok := r.sips[sipID]
	if !ok {
		return model.SIP{}, fmt.Errorf("%w: sip %s", util.ErrNotFound, sipID)
	}
	return sip, nil
}

func (r *InMemorySIPRepository) ListByUser(_ context.Context, userID string) ([]model.SIP, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.SIP, 0)
	for _, s := range r.sips {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *InMemorySIPRepository) ListDue(_ context.Context, now time.Time) ([]model.SIP, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]model.SIP, 0)
	dueItems := make([]*sipHeapItem, 0)
	for r.sipHeap.Len() > 0 {
		item := r.sipHeap[0]
		if item.nextRunAt.After(now) {
			break
		}
		item = heap.Pop(&r.sipHeap).(*sipHeapItem)
		dueItems = append(dueItems, item)
		if sip, ok := r.sips[item.sipID]; ok && sip.Status == model.SIPStatusActive && !sip.NextRunAt.After(now) {
			out = append(out, sip)
		}
	}
	for _, item := range dueItems {
		heap.Push(&r.sipHeap, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].NextRunAt.Before(out[j].NextRunAt) })
	return out, nil
}

func (r *InMemorySIPRepository) Update(_ context.Context, sip model.SIP) (model.SIP, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.sips[sip.SIPID]
	if !ok {
		return model.SIP{}, fmt.Errorf("%w: sip %s", util.ErrNotFound, sip.SIPID)
	}
	if sip.Version != stored.Version {
		return model.SIP{}, util.ErrConflict
	}
	sip.Version++
	r.sips[sip.SIPID] = sip
	r.updateHeapForSip(sip)
	return sip, nil
}

type sipHeapItem struct {
	sipID     string
	nextRunAt time.Time
	index     int
}

type sipMinHeap []*sipHeapItem

func (h sipMinHeap) Len() int           { return len(h) }
func (h sipMinHeap) Less(i, j int) bool { return h[i].nextRunAt.Before(h[j].nextRunAt) }
func (h sipMinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *sipMinHeap) Push(x any) {
	item := x.(*sipHeapItem)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *sipMinHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[:n-1]
	return item
}

func (r *InMemorySIPRepository) updateHeapForSip(sip model.SIP) {
	if sip.Status == model.SIPStatusActive {
		r.upsertHeapItem(sip.SIPID, sip.NextRunAt)
		return
	}
	r.removeHeapItem(sip.SIPID)
}

func (r *InMemorySIPRepository) upsertHeapItem(sipID string, nextRunAt time.Time) {
	if item, ok := r.heapIndex[sipID]; ok {
		item.nextRunAt = nextRunAt
		heap.Fix(&r.sipHeap, item.index)
		return
	}
	item := &sipHeapItem{sipID: sipID, nextRunAt: nextRunAt}
	heap.Push(&r.sipHeap, item)
	r.heapIndex[sipID] = item
}

func (r *InMemorySIPRepository) removeHeapItem(sipID string) {
	if item, ok := r.heapIndex[sipID]; ok {
		heap.Remove(&r.sipHeap, item.index)
		delete(r.heapIndex, sipID)
	}
}
