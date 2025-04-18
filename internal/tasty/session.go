package tasty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	SessionURL = "/sessions"
)

type User struct {
	Email      string `json:"email"`
	Username   string `json:"username"`
	Externalid string `json:"external-id"`
}

type Data struct {
	User              User      `json:"user"`
	SessionToken      string    `json:"session-token"`
	SessionExpiration time.Time `json:"session-expiration"`
}

type AuthResponse struct {
	Data    Data   `json:"data"`
	Context string `json:"context"`
}

func (c *TastyAPI) Authenticate(uname, pass string) error {
	authURL := fmt.Sprintf("%s%s", c.baseurl, SessionURL)
	authData := map[string]string{
		"login":    uname,
		"password": pass,
	}
	authBody, err := json.Marshal(authData)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(authURL, "application/json", bytes.NewReader(authBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

	}
	return nil
}
