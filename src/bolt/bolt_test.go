package ldbolt_test

import (
	"github.com/boltdb/bolt"
	ldbolt "github.com/raymonstah/launchdarkly-embedded/src/bolt"
	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
	"log"
	"time"
)

func ExampleBoltFeatureStore() {
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ldConfig := ld.DefaultConfig
	boltFeatureStoreFactory, err := ldbolt.NewBoltFeatureStoreFactory(db)
	if err != nil {
		log.Fatal(err)
	}

	ldConfig.FeatureStoreFactory = boltFeatureStoreFactory
	ldClient, err := ld.MakeCustomClient("SDK_KEY_HERE", ldConfig, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer ldClient.Close()
}