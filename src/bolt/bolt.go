package ldbolt

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
	"gopkg.in/launchdarkly/go-server-sdk.v4/utils"
	"time"
)

type boltFeatureStore struct {
	boltDB *bolt.DB
}

//InitInternal dumps all data into bolt
func (b boltFeatureStore) InitInternal(allData map[ld.VersionedDataKind]map[string]ld.VersionedData) error {
	fmt.Println("InitInternal...")
	count := 0
	err := b.boltDB.Update(func(tx *bolt.Tx) error {
		for versionedData, data := range allData {
			if err := tx.DeleteBucket([]byte(versionedData.GetNamespace())); err != nil {
				// idempotent
			}

			bucket, err := tx.CreateBucketIfNotExists([]byte(versionedData.GetNamespace()))
			if err != nil {
				return fmt.Errorf("unable to create bucket %v: %w", versionedData.GetNamespace(), err)
			}

			for key, d := range data {
				raw, err := json.Marshal(d)
				if err != nil {
					return fmt.Errorf("unable to marshal json: %w", err)
				}

				if err := bucket.Put([]byte(key), raw); err != nil {
					return fmt.Errorf("unable to put data in Bolt: %w", err)
				}
				count++
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to init all data: %w", err)
	}

	fmt.Println("InitInternal finished, wrote ", count)
	return nil
}

// GetInternal assumes that the inner bucket exists
func (b boltFeatureStore) GetInternal(kind ld.VersionedDataKind, key string) (ld.VersionedData, error) {
	defer func(t time.Time) {
		fmt.Printf("took %v microseconds to get item %v\n", time.Now().Sub(t).Microseconds(), key)
	}(time.Now())

	var result ld.VersionedData
	err := b.boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(kind.GetNamespace()))
		rawData := bucket.Get([]byte(key))
		if rawData == nil {
			result = nil
			return nil
		}
		versionedData, err := utils.UnmarshalItem(kind, rawData)
		if err != nil {
			return fmt.Errorf("unable to unmarshal item from Bolt: %w", err)
		}
		result = versionedData
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to view: %w", err)
	}

	return result, nil
}

// GetAllInternal assumes a bucket for the namespace in bolt already exists
func (b boltFeatureStore) GetAllInternal(kind ld.VersionedDataKind) (map[string]ld.VersionedData, error) {
	var results map[string]ld.VersionedData
	err := b.boltDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(kind.GetNamespace()))
		err := bucket.ForEach(func(k, v []byte) error {
			data, err := utils.UnmarshalItem(kind, v)
			if err != nil {
				return fmt.Errorf("unable to unmarshal item from Bolt: %w", err)
			}

			results[string(k)] = data
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to run ForEach on bucket: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to view bucket: %w", err)
	}

	return results, nil
}

// UpsertInternal adds or updates a single item. If an item with the same key already
// exists, it should update it only if the new item's GetVersion() value is greater
// than the old one. It should return the final state of the item, i.e. if the update
// succeeded then it returns the item that was passed in, and if the update failed due
// to the version check then it returns the item that is currently in the data store
// (this ensures that caching works correctly).
//
// Note that deletes are implemented by using UpsertInternal to store an item whose
// Deleted property is true.
func (b boltFeatureStore) UpsertInternal(kind ld.VersionedDataKind, item ld.VersionedData) (ld.VersionedData, error) {
	fmt.Println("UpsertInternal..")
	defer fmt.Println("UpsertInternal finished")

	err := b.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(kind.GetNamespace()))
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("unable to marshal item to json: %w", err)
		}

		if err := bucket.Put([]byte(item.GetKey()), data); err != nil {
			return fmt.Errorf("unable to put item in Bolt: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update item: %w", err)
	}

	return item, nil
}

// InitializedInternal returns true if the data store contains a complete data set,
// meaning that InitInternal has been called at least once. In a shared data store, it
// should be able to detect this even if InitInternal was called in a different process,
// i.e. the test should be based on looking at what is in the data store. The method
// does not need to worry about caching this value; FeatureStoreWrapper will only call
// it when necessary.
func (b boltFeatureStore) InitializedInternal() bool {
	bucketName := []byte("init-ld")
	var got []byte
	err := b.boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("unable to create bucket `init-ld`: %w", err)
		}
		return nil
	})
	if err != nil {
		return false
	}

	err = b.boltDB.View(func(tx *bolt.Tx) error {

		got = tx.Bucket(bucketName).Get([]byte("init-key"))
		return nil
	})
	if err != nil {
		return false
	}

	return got != nil
}

// GetCacheTTL returns the length of time that data should be retained in an in-memory
// cache. This cache is maintained by FeatureStoreWrapper. If GetCacheTTL returns zero,
// there will be no cache. If it returns a negative number, the cache never expires.
func (b boltFeatureStore) GetCacheTTL() time.Duration {
	return 0
}

func NewBoltFeatureStoreFactory(db *bolt.DB) (ld.FeatureStoreFactory, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		for _, kind := range ld.VersionedDataKinds {
			_, err := tx.CreateBucketIfNotExists([]byte(kind.GetNamespace()))
			if err != nil {
				return fmt.Errorf("unable to create bucket %q: %w", kind.GetNamespace(), err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error creating buckets: %w", err)
	}
	return func(config ld.Config) (ld.FeatureStore, error) {
		boltFeatureStore := boltFeatureStore{boltDB: db}
		return utils.NewFeatureStoreWrapperWithConfig(boltFeatureStore, config), nil
	}, nil
}
