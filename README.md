# launchdarkly-embedded

## Purpose
The purpose of this project was to play around with embedded databases in Go. 
I've heard a bit about them and wanted to learn more. 
In this project specifically, I used [BoltDB](https://github.com/boltdb/bolt).

## Reasons to Use an Embedded Database
* Typically very fast (more on this later..)
* Everything is local
* Simple / Maintenance free
* Easy to export / move around
* Single file

## Reasons Not to Use an Embedded Database
* No ad-hoc querying
* Structure is limited
* No Security / Replication / Access-control built-in

## What does this have to with LaunchDarkly?
LaunchDarkly has a Go SDK. In that SDK, you can define a custom feature store.
As of 04/20, we support Redis, Dynamo, and Consul.
You can learn more [here](https://docs.launchdarkly.com/sdk/concepts/feature-store).

Why use a persistent feature store? Well the LaunchDarkly docs explains it.
>The main reason to do this is to accelerate flag updates when your application has to restart, 
>and after restarting, it takes longer to establish a connection to LaunchDarkly than you want. 
>If you have a persistent feature store that has already been populated, the SDK can still evaluate
>flags using the last known flag state from the store until newer data is available from LaunchDarkly.

You can even use Relay mode, which allows you to run your app without connecting to LaunchDarkly at all.
```go
config.UseLdd = true
```

## Things left to do:
* Write unit tests
* Documentation