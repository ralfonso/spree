package spree

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	hashids "github.com/speps/go-hashids"
	"github.com/uber-go/zap"
)

const (
	hashidSalt string = "celery"
)

var (
	h *hashids.HashID
)

func init() {
	hd := hashids.NewData()
	hd.Salt = hashidSalt

	h = hashids.NewWithData(hd)
}

type BoltKV struct {
	ll     zap.Logger
	db     *bolt.DB
	bucket string
}

var _ Metadata = &BoltKV{}

func NewBoltKV(dbFile, dbBucketName string, ll zap.Logger) (*BoltKV, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		ll.Error("could not open database file", zap.Error(err))
		return nil, err
	}

	var b *bolt.Bucket
	db.Update(func(tx *bolt.Tx) error {
		b, err = tx.CreateBucketIfNotExists([]byte(dbBucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &BoltKV{
		ll:     ll,
		db:     db,
		bucket: dbBucketName,
	}, nil
}

func (b *BoltKV) GetId(shot *Shot) string {
	now := time.Now().UTC().UnixNano()
	id, _ := h.EncodeInt64([]int64{now, rand.Int63()})
	return id
}

func (b *BoltKV) Close() error {
	b.ll.Info("Shutting down BoltDB")
	return b.db.Close()
}

func (b *BoltKV) PutShot(shot *Shot) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(b.bucket))
		data, err := proto.Marshal(shot)
		if err != nil {
			b.ll.Error("could not marshal proto file in PutFile", zap.Error(err))
			return err
		}

		err = bkt.Put([]byte(shot.Id), data)
		if err != nil {
			b.ll.Error("could not PutFile in BoltDB", zap.Error(err))
			return err
		}
		shot.Path = fmt.Sprintf("/p/%s", shot.Id)
		return nil
	})

	return err
}

func (b *BoltKV) ListShots() ([]*Shot, error) {
	shots := make([]*Shot, 0)
	err := b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(b.bucket))
		c := bkt.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			b.ll.Info("found shot", zap.String("shot.id", string(k)))
			shot := &Shot{}
			err := proto.Unmarshal(v, shot)
			if err != nil {
				b.ll.Error("could not marshal proto file in ListShots", zap.Error(err))
				continue
			}
			shot.Path = fmt.Sprintf("/p/%s", shot.Id)
			shots = append(shots, shot)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return shots, nil
}

func (b *BoltKV) GetShotById(id string) (*Shot, error) {
	shot := &Shot{}
	err := b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(b.bucket))
		v := bkt.Get([]byte(id))
		err := proto.Unmarshal(v, shot)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	shot.Path = fmt.Sprintf("/p/%s", shot.Id)
	return shot, nil
}

func (b *BoltKV) IncrementViews(id string) (*Shot, error) {
	shot := &Shot{}
	var err error
	b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(b.bucket))
		v := bkt.Get([]byte(id))
		err = proto.Unmarshal(v, shot)

		if err != nil {
			return err
		}

		shot.Views++

		data, err := proto.Marshal(shot)
		if err != nil {
			b.ll.Error("could not marshal shot in IncrementViews", zap.Error(err))
			return err
		}

		err = bkt.Put([]byte(shot.Id), data)
		if err != nil {
			b.ll.Error("could not Put() in IncrementViews", zap.Error(err))
			return err
		}

		return nil
	})

	return shot, err
}
