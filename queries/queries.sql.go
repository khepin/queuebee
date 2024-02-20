// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package queries

import (
	"context"
)

const findQueue = `-- name: FindQueue :one
SELECT
    queue_id, channel_id, channel_name, team_domain, team_id, user_list, require_ack_before, acked_by, config
FROM
    queues
WHERE
    queue_id = ?
`

func (q *Queries) FindQueue(ctx context.Context, queueID string) (*Queue, error) {
	row := q.db.QueryRowContext(ctx, findQueue, queueID)
	var i Queue
	err := row.Scan(
		&i.QueueID,
		&i.ChannelID,
		&i.ChannelName,
		&i.TeamDomain,
		&i.TeamID,
		&i.UserList,
		&i.RequireAckBefore,
		&i.AckedBy,
		&i.Config,
	)
	return &i, err
}

const saveQueue = `-- name: SaveQueue :exec
UPDATE
    queues
SET
    channel_id = ?,
    channel_name = ?,
    team_domain = ?,
    team_id = ?,
    user_list = ?,
    require_ack_before = ?,
    acked_by = ?,
    config = ?
WHERE
    queue_id = ?
`

type SaveQueueParams struct {
	ChannelID        string
	ChannelName      string
	TeamDomain       string
	TeamID           string
	UserList         string
	RequireAckBefore int64
	AckedBy          string
	Config           Config
	QueueID          string
}

func (q *Queries) SaveQueue(ctx context.Context, arg SaveQueueParams) error {
	_, err := q.db.ExecContext(ctx, saveQueue,
		arg.ChannelID,
		arg.ChannelName,
		arg.TeamDomain,
		arg.TeamID,
		arg.UserList,
		arg.RequireAckBefore,
		arg.AckedBy,
		arg.Config,
		arg.QueueID,
	)
	return err
}

const upsertQueue = `-- name: UpsertQueue :exec
INSERT INTO
    queues (
        queue_id,
        channel_id,
        channel_name,
        team_domain,
        team_id,
        user_list,
        config
    )
VALUES
    (
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?
    ) ON CONFLICT (queue_id) DO NOTHING
`

type UpsertQueueParams struct {
	QueueID     string
	ChannelID   string
	ChannelName string
	TeamDomain  string
	TeamID      string
	UserList    string
	Config      Config
}

func (q *Queries) UpsertQueue(ctx context.Context, arg UpsertQueueParams) error {
	_, err := q.db.ExecContext(ctx, upsertQueue,
		arg.QueueID,
		arg.ChannelID,
		arg.ChannelName,
		arg.TeamDomain,
		arg.TeamID,
		arg.UserList,
		arg.Config,
	)
	return err
}