package subscriptions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/services/subscription"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func GetSubscription(w http.ResponseWriter, r *http.Request) {
	key, err := helpers.Store(r).GetOption("subscription_key")
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if key == "" {
		helpers.WriteJSON(w, 404, nil)
		return
	}

	token, err := subscription.GetToken(helpers.Store(r))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, 200, token)
}

func Activate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key string `json:"key"`
	}

	err := helpers.Store(r).SetOption("subscription_key", req.Key)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	buf, err := json.Marshal(req)
	if err != nil {
		log.Error(err)
		return
	}

	resp, err := http.Post(
		fmt.Sprintf(
			"https://cloud.semui.co/billing/subscriptions/%s/activate",
			req.Key,
		),
		"application/json",
		bytes.NewBuffer(buf),
	)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	var res struct {
		token string
	}

	if err = json.Unmarshal(body, &res); err != nil {
		log.Error(err)
		return
	}

	token, err := subscription.ParseToken(res.token)
	if err != nil {
		log.Error(err)
		return
	}

	err = helpers.Store(r).SetOption("subscription_token", res.token)
	if err != nil {
		log.Error(err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, token)
}
