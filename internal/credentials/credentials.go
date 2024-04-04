package credentials

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"credential"`
}

var ErrEmptyCredentials = errors.New("empty credentials")

func Get() (Credentials, error) {
	resp, err := http.Get("http://speed.cloudflare.com/turn-creds")
	if err != nil {
		return Credentials{}, errors.Wrap(err, "failed to get credentials")
	}
	defer resp.Body.Close()

	var creds Credentials
	err = json.NewDecoder(resp.Body).Decode(&creds)
	if err != nil {
		return Credentials{}, errors.Wrap(err, "failed to decode credentials")
	}

	if creds.Username == "" || creds.Password == "" {
		return Credentials{}, ErrEmptyCredentials
	}

	return creds, nil
}
