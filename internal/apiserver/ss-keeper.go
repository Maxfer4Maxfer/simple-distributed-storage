package apiserver

import "sync"

type storageServerKeeper struct {
	storageServers                 map[string]StorageServer
	storageServerClientCreatorFunc StorageServerClientCreatorFunc
	sync.RWMutex
}

func (k *storageServerKeeper) get(address string) StorageServer {
	k.RLock()
	ss, ok := k.storageServers[address]
	k.RUnlock()

	if !ok {
		ss = k.storageServerClientCreatorFunc(address)

		k.Lock()
		k.storageServers[address] = ss
		k.Unlock()
	}

	return ss
}
