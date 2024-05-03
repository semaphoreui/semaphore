package subscriptions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/services/subscription"
	"io"
	"net/http"
)

func GetSubscription(w http.ResponseWriter, r *http.Request) {
	store := helpers.Store(r)

	key, err := store.GetOption("subscription_key")
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if key == "" {
		helpers.WriteJSON(w, 200, subscription.Token{})
		return
	}

	token, err := subscription.GetToken(store)

	if errors.Is(err, db.ErrNotFound) {
		token.Key = key
	} else if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, 200, token)
}

func Activate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key string `json:"key"`
	}

	if !helpers.Bind(w, r, &req) {
		helpers.WriteErrorStatus(w, "Invalid request", http.StatusBadRequest)
		return
	}

	buf, err := json.Marshal(req)
	if err != nil {
		helpers.WriteError(w, err)
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

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if resp.StatusCode == 409 {
		helpers.WriteErrorStatus(w, "Subscription key already activated.", resp.StatusCode)
		return
	}

	if resp.StatusCode != 200 {
		helpers.WriteErrorStatus(w, "Invalid subscription key.", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var res struct {
		Token string `json:"token"`
	}

	if err = json.Unmarshal(body, &res); err != nil {
		helpers.WriteError(w, err)
		return
	}

	token, err := subscription.ParseToken(res.Token)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).SetOption("subscription_key", req.Key)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).SetOption("subscription_token", res.Token)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, token)
}
