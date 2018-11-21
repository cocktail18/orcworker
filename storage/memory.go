package storage

import (
	"container/list"
	"github.com/cocktail18/orcworker"
	"sync"
)

type MemoryStorage struct {
	list list.List
	lock sync.RWMutex
	set  map[string]struct{}
}

func NewMemoryStorage() *MemoryStorage {
	m := &MemoryStorage{}
	m.list.Init()
	m.set = make(map[string]struct{})
	return m
}

func (memoryStorage *MemoryStorage) EnQueue(seed *orcworker.Seed) error {
	memoryStorage.lock.Lock()
	defer memoryStorage.lock.Unlock()
	memoryStorage.list.PushBack(seed)
	hash, err := seed.Sha256()
	if err != nil {
		return err
	}
	memoryStorage.set[hash] = struct{}{}
	return nil
}

func (memoryStorage *MemoryStorage) DeQueue() (*orcworker.Seed, error) {
	memoryStorage.lock.Lock()
	defer memoryStorage.lock.Unlock()
	ele := memoryStorage.list.Front()
	if ele == nil {
		return nil, orcworker.ERR_SEEDS_EMPTY
	}
	seed := memoryStorage.list.Remove(ele)
	return seed.(*orcworker.Seed), nil
}

func (memoryStorage *MemoryStorage) QueueCapacity() (int, error) {
	memoryStorage.lock.RLock()
	defer memoryStorage.lock.RUnlock()
	l := memoryStorage.list.Len()
	return l, nil
}

func (memoryStorage *MemoryStorage) IsContain(seed *orcworker.Seed) (bool, error) {
	memoryStorage.lock.RLock()
	defer memoryStorage.lock.RUnlock()
	hash, err := seed.Sha256()
	if err != nil {
		return false, err
	}
	_, ok := memoryStorage.set[hash]
	if ok {
		return true, nil
	}
	return false, nil
}

