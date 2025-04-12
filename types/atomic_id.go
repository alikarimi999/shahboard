package types

import (
	"sync/atomic"
)

type AtomicObjectId struct {
	ptr atomic.Pointer[ObjectId]
}

func NewAtomicObjectId(id ObjectId) *AtomicObjectId {
	var a AtomicObjectId
	a.ptr.Store(&id)
	return &a
}

func (a *AtomicObjectId) Load() ObjectId {
	ptr := a.ptr.Load()
	if ptr == nil {
		return ObjectZero
	}
	return *ptr
}

func (a *AtomicObjectId) Store(id ObjectId) {
	a.ptr.Store(&id)
}

func (a *AtomicObjectId) Swap(newId ObjectId) ObjectId {
	old := a.ptr.Swap(&newId)
	if old == nil {
		return ObjectZero
	}
	return *old
}

func (a *AtomicObjectId) SetZero() {
	zero := ObjectZero
	a.ptr.Store(&zero)
}
