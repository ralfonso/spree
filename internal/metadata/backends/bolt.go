package backends

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/ralfonso/spree/internal/metadata"
	"github.com/speps/go-hashids"
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
	db     *bolt.DB
	bucket string
}

func NewBoltKV(dbFile, dbBucketName string) (*BoltKV, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.WithError(err).Error("could not open database file")
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
		db:     db,
		bucket: dbBucketName,
	}, nil
}

func (b *BoltKV) PutFile(file *metadata.File) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucket))
		id, _ := b.NextSequence()
		file.Id, _ = h.Encode([]int{int(id)})

		jsonFile, err := json.Marshal(file)
		if err != nil {
			log.WithError(err).Error("could not marshal json file in PutFile")
			return err
		}

		err = b.Put([]byte(file.Id), jsonFile)
		if err != nil {
			log.WithError(err).Error("could not PutFile in BoltDB")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (b *BoltKV) ListFiles() ([]metadata.File, error) {
	files := make([]metadata.File, 0)
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var file *metadata.File
			err := json.Unmarshal(v, &file)
			if err != nil {
				return err
			}
			files = append(files, *file)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}