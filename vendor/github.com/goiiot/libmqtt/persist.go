/*
 * Copyright GoIIoT (https://github.com/goiiot)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package libmqtt

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// PacketDroppedByStrategy used when persist store packet while strategy
	// don't allow that persist
	PacketDroppedByStrategy = errors.New("packet persist dropped by strategy ")
)

// PersistStrategy defines the details to be complied in persist methods
type PersistStrategy struct {
	// Interval applied to file/database persist
	// if this field is set to 0, means do persist per action
	// default value is 1s
	Interval time.Duration

	// MaxCount applied to all persist method
	// if this field set to 0, means no persist limit
	// for memory persist, means max in memory count
	// for file/database persist, means max entry in file/memory
	// default value is 0
	MaxCount uint32

	// DropOnExceed defines how to tackle with packets incoming
	// when max count is reached, default value is false
	DropOnExceed bool

	// DuplicateReplace defines whether duplicated key should
	// override previous one, default value is true
	DuplicateReplace bool
}

// DefaultPersistStrategy will create a default PersistStrategy
// Interval = 1s, MaxCount = 0, DropOnExceed = false, DuplicateReplace = true
func DefaultPersistStrategy() *PersistStrategy {
	return &PersistStrategy{
		Interval:         time.Second,
		MaxCount:         0,
		DropOnExceed:     false,
		DuplicateReplace: true,
	}
}

// PersistMethod defines the behavior of persist methods
type PersistMethod interface {
	// Name of what persist strategy used
	Name() string

	// Store a packet with key
	Store(key string, p Packet) error

	// Load a packet from stored data according to the key
	Load(key string) (Packet, bool)

	// Range over data stored, return false to break the range
	Range(func(key string, p Packet) bool)

	// Delete
	Delete(key string) error

	// Destroy stored data
	Destroy() error
}

// NonePersist defines no persist storage
var NonePersist = &nonePersist{}

type nonePersist struct{}

func (n *nonePersist) Name() string                          { return "nonePersist" }
func (n *nonePersist) Store(key string, p Packet) error      { return nil }
func (n *nonePersist) Load(key string) (Packet, bool)        { return nil, false }
func (n *nonePersist) Range(func(key string, p Packet) bool) {}
func (n *nonePersist) Delete(key string) error               { return nil }
func (n *nonePersist) Destroy() error                        { return nil }

// NewMemPersist create a in memory persist method with provided strategy
// if no strategy provided (nil), then the default strategy will be used
func NewMemPersist(strategy *PersistStrategy) *MemPersist {
	p := &MemPersist{
		data: &sync.Map{},
		n:    0,
	}

	if strategy == nil {
		p.strategy = DefaultPersistStrategy()
	} else {
		p.strategy = strategy
	}
	return p
}

// MemPersist is the in memory persist method
type MemPersist struct {
	data     *sync.Map
	n        uint32
	strategy *PersistStrategy
}

// Name of this persist method
func (m *MemPersist) Name() string {
	if m == nil {
		return "<nil>"
	}
	return "MemPersist"
}

// Store a key packet pair, in memory persist always return nil (no error)
func (m *MemPersist) Store(key string, p Packet) error {
	if m == nil {
		return nil
	}

	if m.strategy.MaxCount > 0 &&
		atomic.LoadUint32(&m.n) >= m.strategy.MaxCount &&
		m.strategy.DropOnExceed {
		// packet dropped
		return PacketDroppedByStrategy
	}

	if _, loaded := m.data.LoadOrStore(key, p); !loaded {
		atomic.AddUint32(&m.n, 1)
	} else if m.strategy.DuplicateReplace {
		m.data.Store(key, p)
	}
	return nil
}

// Load a packet with key, return nil, false when no packet found
func (m *MemPersist) Load(key string) (Packet, bool) {
	if m == nil {
		return nil, false
	}

	if p, ok := m.data.Load(key); ok {
		if p != nil {
			return p.(Packet), true
		}
	} else {
		return nil, false
	}

	return nil, true
}

// Range over all packet persisted
func (m *MemPersist) Range(f func(key string, p Packet) bool) {
	if m == nil || f == nil {
		return
	}

	m.data.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(Packet))
	})
}

// Delete a persisted packet with key
func (m *MemPersist) Delete(key string) error {
	if m == nil {
		return nil
	}

	m.data.Delete(key)
	return nil
}

// Destroy persist storage
func (m *MemPersist) Destroy() error {
	if m == nil {
		return nil
	}

	m.data = &sync.Map{}
	return nil
}

const (
	fileSuffix = ".mqtt"
)

// NewFilePersist will create a file persist method with provided
// dirPath and strategy, if no strategy provided (nil), then the
// default strategy will be used
func NewFilePersist(dirPath string, strategy *PersistStrategy) *FilePersist {
	p := &FilePersist{
		dirPath:  dirPath,
		inMemBuf: &sync.Map{},
		bytesBuf: &bytes.Buffer{},
	}

	if strategy != nil {
		p.strategy = strategy
	} else {
		p.strategy = DefaultPersistStrategy()
	}

	// init file packet size
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return filepath.SkipDir
		}

		if strings.HasSuffix(info.Name(), fileSuffix) {
			p.n++
		}
		return nil
	})

	return p
}

// FilePersist is the file persist method
type FilePersist struct {
	dirPath   string
	inMemBuf  *sync.Map
	inMemSize uint32
	bytesBuf  *bytes.Buffer
	strategy  *PersistStrategy
	n         uint32
}

// Name of this persist method
func (m *FilePersist) Name() string {
	if m == nil {
		return "<nil>"
	}

	return "FilePersist"
}

// Store a key packet pair, error happens when file access failed
func (m *FilePersist) Store(key string, p Packet) error {
	if m == nil {
		return nil
	}

	if m.strategy.MaxCount > 0 && m.strategy.DropOnExceed &&
		atomic.LoadUint32(&m.n)+atomic.LoadUint32(&m.inMemSize) >= m.strategy.MaxCount {
		// packet dropped
		return PacketDroppedByStrategy
	}

	if !m.exists(key) || m.strategy.DuplicateReplace {
		if m.strategy.Interval > 0 {
			// has persist interval
			if atomic.LoadUint32(&m.inMemSize) == 0 {
				// schedule a file save action according to the strategy
				defer func() {
					go m.worker()
				}()
			}
			m.inMemBuf.Store(key, p)
			atomic.AddUint32(&m.inMemSize, 1)
		} else {
			// persist every time
			return m.store(key, p)
		}
	}

	return nil
}

// Load a packet with key, return nil, false when no packet found
func (m *FilePersist) Load(key string) (Packet, bool) {
	if m == nil {
		return nil, false
	}

	packet, err := m.getPacketFromFile(m.getFilename(key))
	if err != nil {
		return nil, false
	}

	return packet, true
}

// Range over all packet persisted
func (m *FilePersist) Range(ranger func(key string, p Packet) bool) {
	if m == nil || ranger == nil {
		return
	}

	filepath.Walk(m.dirPath, func(path string, info os.FileInfo, err error) error {
		// error happened or is dir
		if err != nil || info.IsDir() {
			return filepath.SkipDir
		}

		// not libmqtt packet file
		if !strings.HasSuffix(info.Name(), fileSuffix) {
			return nil
		}

		// decode packet
		pkt, err := m.getPacketFromFile(path)
		if err != nil {
			return nil
		}

		ranger(strings.TrimSuffix(info.Name(), fileSuffix), pkt)

		return nil
	})
}

// Delete a persisted packet with key
func (m *FilePersist) Delete(key string) error {
	if m == nil {
		return nil
	}

	return os.Remove(path.Join(m.dirPath, key))
}

// Destroy persist storage
func (m *FilePersist) Destroy() error {
	if m == nil {
		return nil
	}

	return os.RemoveAll(m.dirPath)
}

func (m *FilePersist) getPacketFromFile(path string) (Packet, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	packet, err := DecodeOnePacket(V311, bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func (m *FilePersist) exists(key string) bool {
	_, err := os.Open(m.getFilename(key))
	if err != nil && os.IsNotExist(err) {
		// no such packet file
		return false
	}
	return true
}

func (m *FilePersist) store(key string, p Packet) error {
	f, err := os.Create(m.getFilename(key))
	if err != nil {
		return err
	}
	defer f.Close()

	// write packet bytes to file
	w := bufio.NewWriter(f)
	err = EncodeOnePacket(V311, p, w)
	if err != nil {
		println(err.Error())
		return err
	}
	err = w.Flush()
	if err != nil {
		println(err.Error())
		return err
	}

	atomic.AddUint32(&m.n, 1)
	atomic.StoreUint32(&m.inMemSize, atomic.LoadUint32(&m.inMemSize)-1)
	return nil
}

func (m *FilePersist) worker() {
	time.Sleep(m.strategy.Interval)

	persistedKeys := make([]string, 0)
	m.inMemBuf.Range(func(key, value interface{}) bool {
		k := key.(string)
		p, ok := value.(Packet)
		if !ok {
			return true
		}

		m.store(k, p)
		persistedKeys = append(persistedKeys, k)
		return true
	})

	for _, k := range persistedKeys {
		m.inMemBuf.Delete(k)
	}

	if atomic.LoadUint32(&m.inMemSize) > 0 {
		m.worker()
	}
}

func (m *FilePersist) getFilename(key string) string {
	return path.Join(m.dirPath, key+fileSuffix)
}
