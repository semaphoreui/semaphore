//go:build ignore

// wsprobe logs in to one Semaphore node and prints every WebSocket message it receives for a
// while. Run it against node B while a task runs on node A: the "update" events prove that the
// Redis Pub/Sub broadcaster relays them across nodes.
//
//	go run AGENTS/tools/redis-sentinel-stand/wsprobe.go -base http://localhost:3111 -for 40s
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	base := flag.String("base", "http://localhost:3111", "node to listen on")
	dur := flag.Duration("for", 30*time.Second, "how long to listen")
	flag.Parse()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	body, _ := json.Marshal(map[string]string{"auth": "admin", "password": "admin123"})
	resp, err := client.Post(*base+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode/100 != 2 {
		fmt.Fprintln(os.Stderr, "login failed:", err, resp)
		os.Exit(1)
	}
	resp.Body.Close()

	u, _ := url.Parse(*base)
	wsURL := "ws" + strings.TrimPrefix(*base, "http") + "/api/ws"
	dialer := websocket.Dialer{Jar: jar}
	conn, _, err := dialer.Dial(wsURL, http.Header{"Origin": {u.Scheme + "://" + u.Host}})
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial failed:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Printf("%s listening on %s for %s\n", time.Now().UTC().Format("15:04:05"), wsURL, *dur)

	deadline := time.Now().Add(*dur)
	count := 0
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(deadline)
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		count++
		var m map[string]any
		if json.Unmarshal(msg, &m) == nil {
			fmt.Printf("%s type=%v task_id=%v status=%v\n", time.Now().UTC().Format("15:04:05"), m["type"], m["task_id"], m["status"])
		} else {
			fmt.Printf("%s %s\n", time.Now().UTC().Format("15:04:05"), string(msg))
		}
	}
	fmt.Printf("%d message(s) received\n", count)
}
