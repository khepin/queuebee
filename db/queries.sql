-- name: FindQueue :one
SELECT
    *
FROM
    queues
WHERE
    queue_id = ?;

-- name: UpsertQueue :exec
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
    ) ON CONFLICT (queue_id) DO NOTHING;

-- name: SaveQueue :exec
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
    queue_id = ?;
