package connection

import "sync"

type ConnectionPool struct {
	pool   map[string]*Connection
	rwlock *sync.RWMutex
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		pool:   make(map[string]*Connection),
		rwlock: &sync.RWMutex{},
	}
}

func (cp *ConnectionPool) Put(key string, conn *Connection) {
	cp.rwlock.Lock()
	defer cp.rwlock.Unlock()
	cp.pool[conn.PodUID] = conn
}

func (cp *ConnectionPool) Delete(key string) {
	cp.rwlock.Lock()
	defer cp.rwlock.Unlock()
	conn, ok := cp.pool[key]
	if ok {
		conn.Close()
	}
	delete(cp.pool, key)
}

func (cp *ConnectionPool) GetConnections() (*map[string]*Connection, func()) {
	rlock := cp.rwlock.RLocker()
	rlock.Lock()
	return &cp.pool, rlock.Unlock
}
