package extension

import (
	"github.com/boltdb/bolt"

	mqtt "github.com/goiiot/libmqtt"
)

type BoltPersist struct {
	bucket *bolt.Bucket
}

// Name of BoltPersist is "BoltPersist"
func (b *BoltPersist) Name() string {
	if b == nil {
		return "<nil>"
	}

	return "BoltPersist"
}

func (b *BoltPersist) Store(key string, p mqtt.Packet) error {
	if b == nil {
		return nil
	}

	b.bucket.Put([]byte(key), p.Bytes())
	return nil
}

func (b *BoltPersist) Load(key string) (mqtt.Packet, bool) {
	if b == nil {
		return nil, false
	}

	return nil, false
}

func (b *BoltPersist) Range(f func(string, mqtt.Packet) bool) {
	if b == nil {
		return
	}
}

func (b *BoltPersist) Delete(key string) error {
	if b == nil {
		return nil
	}

	return nil
}
