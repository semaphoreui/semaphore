package subscription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

func RefreshToken(store db.Store) {
	key, err := store.GetOption("subscription_key")
	if err != nil {
		log.Error(err)
		return
	}

	if key == "" {
		return
	}

	token, err := store.GetOption("subscription_token")

	if err != nil {
		log.Error(err)
		return
	}

	var req struct {
		token string
	}

	req.token = token

	buf, err := json.Marshal(req)
	if err != nil {
		log.Error(err)
		return
	}

	resp, err := http.Post(
		fmt.Sprintf(
			"https://cloud.semui.co/billing/subscriptions/%s/validate",
			key,
		),
		"application/json",
		bytes.NewBuffer(buf),
	)

	if err != nil {
		log.Error(err)
		return
	} else if resp.StatusCode != 200 {
		log.Error(fmt.Errorf("Can not verify key! Response code: " + strconv.Itoa(resp.StatusCode)))
		return
	}

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

	_, err = ParseToken(res.token)
	if err != nil {
		log.Error(err)
		return
	}

	err = store.SetOption("subscription_token", res.token)
	if err != nil {
		log.Error(err)
		return
	}
}

func StartValidationCron(store db.Store) {
	c := cron.New()

	_, err := c.AddFunc("0 1 * * *", func() {
		RefreshToken(store)
	})

	if err != nil {
		log.Error(err)
		return
	}

	c.Start()
}
