package jstat

import (
	"encoding/json"
	"os"
	"os/user"
)

type User struct{}

func (mod *User) Init() error {
	return nil
}

func (mod *User) Run() (json.RawMessage, error) {
	var (
		currentUser *user.User
		host        string
		err         error
	)

	currentUser, err = user.Current()
	if err != nil {
		return nil, err
	}

	host, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		UID, GID, Name, Host string
	}{
		UID:  currentUser.Uid,
		GID:  currentUser.Gid,
		Name: currentUser.Username,
		Host: host,
	})
}

func (mod *User) Sleep() error {
	select {}
}

func (mod *User) Cleanup() error {
	return nil
}

func NewUser() *User {
	return &User{}
}
