package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/boltdb/bolt"
	ldbolt "github.com/raymonstah/launchdarkly-embedded/src/bolt"
	"github.com/urfave/cli/v2"
	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
	"gopkg.in/launchdarkly/go-server-sdk.v4/lddynamodb"
	ldredis "gopkg.in/launchdarkly/go-server-sdk.v4/redis"
	"log"
	"net/http"
	"os"
	"time"
)

var sdkKey string

func main() {
	app := &cli.App{
		Action: action,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "sdk-key",
				Required:    true,
				Destination: &sdkKey,
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func useBolt(db *bolt.DB) ld.FeatureStoreFactory {
	factory, err :=  ldbolt.NewBoltFeatureStoreFactory(db)
	if err != nil {
		log.Fatal(err)
	}
	return factory
}

func useDynamoLocal() ld.FeatureStoreFactory {
	s := session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String("http://localhost:8000")}))
	ddbClient := dynamodb.New(s)
	dynamoFactory, err := lddynamodb.NewDynamoDBFeatureStoreFactory("ld-table", lddynamodb.CacheTTL(0), lddynamodb.DynamoClient(ddbClient))
	if err != nil {
		log.Fatal(err)
	}
	return dynamoFactory
}

func useRedis() ld.FeatureStoreFactory {
	store, err := ldredis.NewRedisFeatureStoreFactory(
		ldredis.HostAndPort("localhost", 6379),
		ldredis.Prefix("my-key-prefix"),
		ldredis.CacheTTL(0))
	if err != nil {
		log.Fatal(err)
	}
	return store
}
func action(_ *cli.Context) error {
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	connected := connectedToLD()

	ldConfig := ld.DefaultConfig
	ldConfig.UseLdd = !connected
	ldConfig.FeatureStoreFactory = useBolt(db)
	//ldConfig.FeatureStoreFactory = useRedis()
	//ldConfig.FeatureStoreFactory = useDynamoLocal()
	ldClient, err := ld.MakeCustomClient(sdkKey, ldConfig, 5*time.Second)
	if err != nil {
		return fmt.Errorf("unable to make LaunchDarkly client: %w", err)
	}
	defer ldClient.Close()

	benchmark(ldClient)
	return nil
}

func benchmark(ldClient *ld.LDClient) {
	ticker := time.NewTicker(1 * time.Second)
	totalTime := time.Duration(0)
	totalTicks := 1000
	for i := 0; i < totalTicks; i++ {
		select {
		case <-ticker.C:
			start := time.Now()
			str, err := ldClient.StringVariation("string-flag", ld.NewAnonymousUser("blah"), "")
			if err != nil {
				fmt.Println("unable to get string variation", err)
			}
			elapsed := time.Now().Sub(start)
			fmt.Printf("evaluated variation as %q in %v\n", str, elapsed.String())
			totalTime += elapsed
		}
	}

	fmt.Printf("average evaluation time: %v\n", (totalTime / time.Duration(totalTicks)).String())
}

func connectedToLD() bool {
	_, err := http.Get("https://app.launchdarkly.com")
	if err != nil {
		return false
	}
	return true
}
