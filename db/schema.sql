CREATE TABLE queues (
    -- queue_id: made of team_id:channel_id
    queue_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    channel_name TEXT NOT NULL,
    team_domain TEXT NOT NULL,
    team_id TEXT NOT NULL,
    user_list TEXT NOT NULL,
    require_ack_before INT NOT NULL DEFAULT 0,
    acked_by TEXT NOT NULL DEFAULT '',
    config TEXT NOT NULL DEFAULT '{}',
    PRIMARY KEY (queue_id)
);
