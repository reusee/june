// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import "sync"

const lockShardBits = 7

var lockShardMask = func() byte {
	var ret byte
	for i := 0; i < lockShardBits; i++ {
		ret &= 1 << i
	}
	return ret
}()

type _EntityLocks struct {
	shards [lockShardBits]*_EntityLockShard
}

type _EntityLockShard struct {
	*sync.Cond
	locks map[Key]struct{}
}

func newEntityLocks() *_EntityLocks {
	locks := new(_EntityLocks)
	for i := range locks.shards {
		locks.shards[i] = &_EntityLockShard{
			Cond:  sync.NewCond(new(sync.Mutex)),
			locks: make(map[Key]struct{}),
		}
	}
	return locks
}

func (l *_EntityLocks) Lock(key Key) (
	unlock func(),
) {
	shard := l.shards[key.Hash[len(key.Hash)-1]&lockShardMask]
	shard.L.Lock()
	defer shard.L.Unlock()
	for {
		_, ok := shard.locks[key]
		if ok {
			shard.Wait()
			continue
		}
		break
	}
	shard.locks[key] = struct{}{}
	unlock = func() {
		shard.L.Lock()
		delete(shard.locks, key)
		shard.Broadcast()
		shard.L.Unlock()
	}
	return
}

func (_ Def) EntityLocks() *_EntityLocks {
	return newEntityLocks()
}
