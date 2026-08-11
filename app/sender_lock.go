package app

import (
	"hash/fnv"
	"sync"
)

const senderLockShardCount = 4096

type senderLockShard struct {
	mu sync.Mutex
}

var senderLockShards [senderLockShardCount]senderLockShard

func getSenderLockShard(address string) *senderLockShard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(address))

	return &senderLockShards[uint32(h.Sum32())%senderLockShardCount]
}

func lockSender(address string) *senderLockShard {
	shard := getSenderLockShard(address)
	shard.mu.Lock()

	return shard
}

func unlockSender(shard *senderLockShard) {
	shard.mu.Unlock()
}
