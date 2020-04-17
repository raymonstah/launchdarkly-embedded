package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/boltdb/bolt"
	"github.com/urfave/cli/v2"
	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
	"gopkg.in/launchdarkly/go-server-sdk.v4/lddynamodb"
	"log"
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

func action(_ *cli.Context) error {
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ldConfig := ld.DefaultConfig
	s := session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String("http://localhost:8000")}))
	ddbClient := dynamodb.New(s)
	dynamoFactory, err := lddynamodb.NewDynamoDBFeatureStoreFactory("ld-table", lddynamodb.CacheTTL(0), lddynamodb.DynamoClient(ddbClient))
	if err != nil {
		return fmt.Errorf("unable to create dynamo feature store factory: %w", err)
	}
	ldConfig.FeatureStoreFactory = dynamoFactory
	//ldConfig.FeatureStoreFactory = NewBoltFeatureStoreFactory(db)
	ldClient, err := ld.MakeCustomClient(sdkKey, ldConfig, 5*time.Second)
	if err != nil {
		return fmt.Errorf("unable to make LaunchDarkly client: %w", err)
	}
	defer ldClient.Close()

	ticker := time.NewTicker(time.Millisecond)
	for {
		select {
		case <-ticker.C:
			start := time.Now()
			variation, err := ldClient.BoolVariation("test-flag", ld.NewAnonymousUser("blah"), true)
			if err != nil {
				fmt.Errorf("unable to get bool variation: %w", err)
			}
			fmt.Printf("evaluated variation as %v in %v\n", variation, time.Now().Sub(start).String())
		}

	}
}
