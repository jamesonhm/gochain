package tasty

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	SessionURL = "/sessions"
)

type User struct {
	Email      *string `json:"email"`
	Username   *string `json:"username"`
	Externalid *string `json:"external-id"`
}

type Data struct {
	User              *User     `json:"user"`
	SessionToken      *string   `json:"session-token"`
	SessionExpiration time.Time `json:"session-expiration"`
}

type Session struct {
	Data    *Data   `json:"data"`
	Context *string `json:"context"`
}

type LoginInfo struct {
	Login         string `json:"login"`
	Password      string `json:"password,omitempty"`
	RememberMe    bool   `json:"remember-me,omitempty"`
	RememberToken string `json:"remember-token,omitempty"`
}

func (c *TastyAPI) CreateSession(ctx context.Context, login LoginInfo) error {
	authURL := fmt.Sprintf("%s%s", c.baseurl, SessionURL)

	session := &Session{}
	err := c.request(ctx, http.MethodPost, noAuth, authURL, nil, login, session)
	if err != nil {
		return err
	}
	c.session = session
	return nil
}
