package oauth2

import (
	"bytes"
	"encoding/json"
	"github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/commons"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var atLocker = &sync.RWMutex{}
var daemonToken string
var lastRefresh time.Time
var expiresIn time.Duration
var client = &http.Client{}

func RefreshToken() bool {
	atLocker.Lock()
	defer atLocker.Unlock()

	//if we just refreshed in the last minute, don't refresh the token
	if lastRefresh.Add(1 * time.Minute).After(time.Now()) {
		return false
	}

	clientId := config.GetString("clientId")
	clientSecret := config.GetString("clientSecret")

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	encodedData := data.Encode()
	request, _ := http.NewRequest("POST", config.GetString("authServer"), bytes.NewBufferString(encodedData))

	request.Header.Add("Authorization", "Bearer "+daemonToken)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))
	response, err := client.Do(request)
	defer commons.CloseResponse(response)
	if err != nil {
		logging.Error("Error talking to auth server", err)
		return false
	}

	var responseData requestResponse
	err = json.NewDecoder(response.Body).Decode(&responseData)

	if responseData.Error != "" {
		return false
	}

	daemonToken = responseData.AccessToken
	lastRefresh = time.Now()
	expiresIn = responseData.ExpiresIn

	return true
}

func RefreshIfStale() {
	//we know the token only lasts about an hour,
	//so we'll check to see if we know the cache is old
	atLocker.RLock()
	oldCache := lastRefresh.Add(expiresIn).Before(time.Now())
	atLocker.RUnlock()
	if oldCache {
		RefreshToken()
	}
}

type requestResponse struct {
	AccessToken string        `json:"access_token"`
	ExpiresIn   time.Duration `json:"expires_in"`
	Error       string        `json:"error"`
}
