package gitkit

import (
	"encoding/json"
	"errors"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
)

type GitkitServerConfig struct {
	ClientID                     string `json:"clientId"`
	ServiceAccountEmail          string `json:"serviceAccountEmail"`
	ServiceAccountPrivateKeyFile string `json:"serviceAccountPrivateKeyFile"`
	WidgetURL                    string `json:"widgetUrl"`
	CookieName                   string `json:"cookieName"`
	ServerAPIKey                 string `json:"serverApiKey"`

	priKey  []byte
	pubKeys map[string][]byte
}

// Client GITKit client
type Client struct {
	cfg *GitkitServerConfig

	rp *relyingParty
}

// NewClientFromJSON GitkitClient.createFromJson
func NewClientFromJSON(configFile string) (*Client, error) {
	cfgData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	cfg := new(GitkitServerConfig)
	err = json.Unmarshal(cfgData, cfg)
	if err != nil {
		return nil, err
	}
	c := &Client{
		cfg: cfg,
	}

	c.rp = newRelyingParty(cfg)

	cfg.priKey, err = ioutil.ReadFile(cfg.ServiceAccountPrivateKeyFile)
	if err != nil {
		return nil, err
	}

	keys, err := c.GetPublicKeys()
	if err != nil {
		return nil, err
	}
	cfg.pubKeys = make(map[string][]byte)
	for k, v := range keys {
		cfg.pubKeys[k] = []byte(v)
	}
	return c, nil
}

func (c *Client) ValidateTokenInRequest(r *http.Request) (*User, error) {
	cookie, err := r.Cookie(c.cfg.CookieName)
	if err != nil {
		return nil, errors.New("gitkit: " + err.Error())
	}
	return c.ValidateToken(cookie.Value)
}

// ValidateToken https://developers.google.com/identity-toolkit/v3/required-endpoints#id_token_desc
func (c *Client) ValidateToken(token string) (*User, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		key, ok := c.cfg.pubKeys[t.Header["kid"].(string)]
		if ok {
			return key, nil
		}
		return nil, errors.New(fmt.Sprint("gitkit: no pub key with kid=", t.Header["kid"]))
	})
	if err != nil {
		return nil, err
	}
	// aud := t.Claims["aud"].(string)
	// if aud != c.cfg.WebClientID {
	// 	return nil, errors.New("gitkit: Token aud not match your client ID. aud=" + aud)
	// }
	return c.GetAccountInfoByID(t.Claims["user_id"].(string))
}

func (c *Client) DeleteAccount(userID string) error {
	return c.rp.DeleteAccount(userID)
}

func (c *Client) DownloadAccount(nextPageToken string, maxResults uint) ([]*User, string, error) {
	return c.rp.DownloadAccount(nextPageToken, maxResults)
}

func (c *Client) GetAccountInfoByEmail(email string) (*User, error) {
	return c.rp.GetAccountInfoByEmail(email)
}

func (c *Client) GetAccountInfoByID(userID string) (*User, error) {
	return c.rp.GetAccountInfoByID(userID)
}

func (c *Client) GetOobConfirmationCode(requestBody map[string]string) (string, error) {
	return c.rp.GetOobConfirmationCode(requestBody)
}

func (c *Client) GetPublicKeys() (ret map[string]string, err error) {
	return c.rp.GetPublicKeys()
}

func (c *Client) UploadAccount(hashAlgorithm string, signerKey, saltSeparator []byte, rounds, memoryCost int, users []*User) (*uploadAccountResult, error) {
	return c.rp.UploadAccount(hashAlgorithm, signerKey, saltSeparator, rounds, memoryCost, users)
}
