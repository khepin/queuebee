package queries

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (q *Queue) AddUser(userID string) error {
	if q.UserList == "" {
		ul := []string{userID}
		l, err := serializeList(ul)
		if err != nil {
			return err
		}
		q.UserList = l
		return nil
	}

	ul, err := deserializeList(q.UserList)
	if err != nil {
		return err
	}
	for _, u := range ul {
		if u == userID {
			return UserAlreadyInQueueError("user is already in the queue (" + userID + ")")
		}
	}
	ul = append(ul, userID)
	l, err := serializeList(ul)
	if err != nil {
		return err
	}
	q.UserList = l
	return nil
}

type UserAlreadyInQueueError string

func (u UserAlreadyInQueueError) Error() string {
	return string(u)
}

func deserializeList(ul string) ([]string, error) {
	var u []string
	err := json.Unmarshal([]byte(ul), &u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func serializeList(ul []string) (string, error) {
	b, err := json.Marshal(ul)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (q *Queue) ToSaveParams() SaveQueueParams {
	return SaveQueueParams{
		QueueID:          q.QueueID,
		ChannelID:        q.ChannelID,
		ChannelName:      q.ChannelName,
		TeamDomain:       q.TeamDomain,
		TeamID:           q.TeamID,
		UserList:         q.UserList,
		RequireAckBefore: q.RequireAckBefore,
		Config:           q.Config,
		AckedBy:          q.AckedBy,
	}
}

func (q *Queue) GetFirstUser() string {
	users, err := deserializeList(q.UserList)
	if err != nil {
		return ""
	}
	if len(users) == 0 {
		return ""
	}
	return users[0]
}

func (q *Queue) GetUsersInLine() []string {
	users, err := deserializeList(q.UserList)
	if err != nil {
		return nil
	}
	if len(users) < 2 {
		return nil
	}

	return users[1:]
}

func (q *Queue) RemoveUser(userID string) (bool, error) {
	ul, err := deserializeList(q.UserList)
	if err != nil {
		return false, err
	}
	for i, u := range ul {
		if u == userID {
			ul = append(ul[:i], ul[i+1:]...)
			l, err := serializeList(ul)
			if err != nil {
				return false, err
			}
			q.UserList = l
			return true, nil
		}
	}
	return false, nil
}

func (q *Queue) SkipTurn(userId string) (bool, error) {
	ul, err := deserializeList(q.UserList)
	if err != nil {
		return false, err
	}

	for i, u := range ul {
		if u == userId {
			if len(ul) > i+1 {
				ul[i], ul[i+1] = ul[i+1], ul[i]
			}
			l, err := serializeList(ul)
			if err != nil {
				return false, err
			}
			q.UserList = l
			return true, nil
		}
	}
	return false, nil
}

func (q *Queue) IsEmpty() bool {
	users, err := deserializeList(q.UserList)
	if err != nil {
		return true
	}
	return len(users) == 0
}

type Config struct {
	InitialAckTimeout     int64   `json:"initial_ack_timeout"`
	SubsequentAckTimeouts []int64 `json:"subsequent_ack_timeouts"`
}

func (c Config) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Config) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, c)
	case string:
		return json.Unmarshal([]byte(src), c)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}

func (c Config) SubsequentAckTimeoutsStr() string {
	timeoutStrs := []string{}
	for _, timeout := range c.SubsequentAckTimeouts {
		timeoutStrs = append(timeoutStrs, strconv.FormatInt(timeout, 10))
	}
	return strings.Join(timeoutStrs, ",")
}
