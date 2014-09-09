package gitkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	relyingPartyURL  = "https://www.googleapis.com/identitytoolkit/v3/relyingparty"
	tokenEndpointURL = "https://accounts.google.com/o/oauth2/token"
)

type relyingParty struct {
	hc *http.Client

	cfg *GitkitServerConfig
}

func newRelyingParty(cfg *GitkitServerConfig) *relyingParty {
	rp := &relyingParty{
		hc:  &http.Client{},
		cfg: cfg,
	}
	return rp
}

func (rp *relyingParty) DeleteAccount(userID string) error {
	res := make(map[string]string)
	return rp.invokeWithServiceAccount("POST", "/deleteAccount",
		map[string]string{"localId": userID}, &res)
}

func (rp *relyingParty) DownloadAccount(nextPageToken string, maxResults uint) ([]*User, string, error) {
	param := make(map[string]interface{})
	param["maxResults"] = maxResults
	if nextPageToken != "" {
		param["nextPageToken"] = nextPageToken
	}
	result := &struct {
		Users         []*User `json:"users"`
		NextPageToken string  `json:"nextPageToken"`
	}{}
	err := rp.invokeWithServiceAccount("POST", "/downloadAccount", param, result)
	if err != nil {
		return nil, "", err
	}
	return result.Users, result.NextPageToken, nil
}

func (rp *relyingParty) GetAccountInfoByID(userID string) (*User, error) {
	result := &struct {
		Users []*User `json:"users"`
	}{}
	err := rp.invokeWithServiceAccount("GET", "/getAccountInfo", map[string]interface{}{
		"localId": []string{userID},
	}, result)
	if err != nil {
		return nil, err
	}
	return result.Users[0], nil
}

func (rp *relyingParty) GetAccountInfoByEmail(email string) (*User, error) {
	result := &struct {
		Users []*User `json:"users"`
	}{}
	err := rp.invokeWithServiceAccount("GET", "/getAccountInfo", map[string]interface{}{
		"email": []string{email},
	}, result)
	if err != nil {
		return nil, err
	}
	return result.Users[0], nil
}

func (rp *relyingParty) GetOobConfirmationCode(requestBody map[string]string) (string, error) {
	res := make(map[string]string)
	err := rp.invokeWithServiceAccount("POST", "/getOobConfirmationCode", requestBody, &res)
	if err != nil {
		return "", err
	}
	return res["oobCode"], nil
}

func (rp *relyingParty) GetPublicKeys() (ret map[string]string, err error) {
	ret = make(map[string]string)
	if rp.cfg.ServerAPIKey != "" {
		api := "/publicKeys?key=" + rp.cfg.ServerAPIKey
		err = rp.invoke("GET", api, nil, &ret)
	} else {
		err = rp.invokeWithServiceAccount("GET", "/publicKeys", nil, &ret)
	}
	return
}

type uploadAccountResult struct {
	Error []struct {
		Index   int    `json:"index"`
		Message string `json:"message"`
	} `json:"error"`
}

func (rp *relyingParty) UploadAccount(hashAlgorithm string, signerKey, saltSeparator []byte, rounds, memoryCost int, users []*User) (*uploadAccountResult, error) {
	param := map[string]interface{}{
		"hashAlgorithm": hashAlgorithm,
		"signerKey":     signerKey,
		"saltSeparator": saltSeparator,
		"rounds":        rounds,
		"memoryCost":    memoryCost,
		"users":         users,
	}
	res := &uploadAccountResult{}
	err := rp.invokeWithServiceAccount("POST", "/uploadAccount", param, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (rp *relyingParty) doInvoke(method, api string, param, result interface{}, needServiceAccount bool) (err error) {
	var body io.Reader
	if param != nil {
		reqBody, err := json.Marshal(param)
		if err != nil {
			return err
		}
		body = bytes.NewReader(reqBody)
	}
	req, err := http.NewRequest(method, relyingPartyURL+api, body)
	if err != nil {
		return err
	}
	if needServiceAccount {
		accessToken, err := rp.getAccessToken()
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", "Bearer "+accessToken)
	}
	resp, err := rp.hc.Do(req)
	if err != nil {
		return err
	}
	err = parseError(api, resp)
	if err != nil {
		return err
	}
	return decodeJSON(resp.Body, result)
}

func (rp *relyingParty) invokeWithServiceAccount(method, api string, param, result interface{}) (err error) {
	return rp.doInvoke(method, api, param, result, true)
}

func (rp *relyingParty) invoke(method, api string, param, result interface{}) (err error) {
	return rp.doInvoke(method, api, param, result, false)
}

func (rp *relyingParty) getAccessToken() (string, error) {
	assertion, err := rp.generateAssertion()
	if err != nil {
		return "", err
	}
	param := url.Values{
		"assertion":  {assertion},
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
	}
	fmt.Println(param)
	resp, err := rp.hc.PostForm(tokenEndpointURL, param)
	if err != nil {
		return "", err
	}
	err = parseError("/oauth2/token", resp)
	if err != nil {
		return "", err
	}
	result := make(map[string]string)
	err = decodeJSON(resp.Body, result)
	if err != nil {
		return "", err
	}
	return result["access_token"], nil
}

func (rp *relyingParty) generateAssertion() (string, error) {
	now := time.Now().Unix()
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims["iss"] = rp.cfg.ServiceAccountEmail
	token.Claims["scope"] = "https://www.googleapis.com/auth/identitytoolkit"
	token.Claims["aud"] = tokenEndpointURL
	token.Claims["lat"] = now
	token.Claims["exp"] = now + 3600
	return token.SignedString(rp.cfg.priKey)
}

func decodeJSON(r io.ReadCloser, v interface{}) error {
	dec := json.NewDecoder(r)
	defer r.Close()
	return dec.Decode(v)
}

func parseError(api string, resp *http.Response) error {
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		errResp := &struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}{}
		err := decodeJSON(resp.Body, errResp)
		var errMsg string
		if err == nil {
			errMsg = errResp.Error.Message
		} else {
			errMsg = resp.Status
		}
		return fmt.Errorf("gitkit: %s %d %s", api, resp.StatusCode, errMsg)
	} else if resp.StatusCode >= 500 {
		return fmt.Errorf("gitkit: %s %d %s", api, resp.StatusCode, resp.Status)
	}
	return nil
}
