# QueueBee

An open source Slack App for managing user queues in slack channels.

## Motivation

Slack's great at letting us organize ourselves, yet sometimes a bit more than ad-hoc organization is helpful.

If multiple people want to deploy an application, each on their own so they can take the time to verify their changes before passing it on to the next, QueueBee can be helpful to keep track of who's turn it is.

## Usage

### Slash commands

QueueBee comes with the following slash commands:

- `/qbee join`: Join the queue
- `/qbee leave`: Leave the queue
- `/qbee skip`: Skip your turn. Unlike leaving the queue, it will just put you in second place when you were in first.
- `/qbee list`: List the current queue.
- `/qbee ack`: Acknowledge your turn. When it's your turn you have a set (configurable) time to accept your turn. If you don't, then your turn is skipped and the next person can go first. That way if you were in the queue but are now in a meeting and unable to use your turn, nobody's blocked.
- `/qbee config`: Configure the queue. Currently you can configure:
    - the initial time given to accept your turn
    - how long a participant can hold its turn for
- `/qbee help`: Shows a help message listing the commands

### Interactivity

TODO: wait for this to be on GH so I can upload the images.

## Installation

### Docker

The app can be run in a docker container (a sample Dockerfile is provided) or built as a single executable and run on any server.

**Note:** The sample container does not currently include `litestream` for DB backups and replication.

### Config

The app takes in the following 4 env variables as config:

```env
SLACK_SECRET=
SLACK_SIGNING_SECRET=
LITEQ_DB="liteq.db" # path to a database file that stores delayed jobs (warn a user that time is about to expire)
QUEUEBEE_DB="queuebee.db" # main DB file
```

### DB Backups

QueueBee uses SQLite as its database. You can setup DB backups easily via [litestream](https://litestream.io/). A sample `litestream.yml` config is provided that will create a DB snapshot every hour and store 72h worth of snapshots.

### Slack

To setup the app with slack, you will need to:

- go to https://api.slack.com/apps and create an app
- When prompted, select `From an app manifest`
- Use the provided sample `slack-manifest.json`. You need to replace `$domain` with the base url of where `queuebee` is running.
- You will then be able to get the app's secret and signing secret. Those are required as env variables to run `queuebee`.
- Upload the app's logo. `queuebee.png` is provided in this repository.
