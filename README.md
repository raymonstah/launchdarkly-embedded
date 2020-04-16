# launchdarkly-embedded

## Purpose
The purpose of this project was to play around with embedded databases in Go. I've heard a bit about them and wanted to learn more. In this project specifically, I used [BoltDB](https://github.com/boltdb/bolt).

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

## Things left to do:
* Write unit tests
* Documentation