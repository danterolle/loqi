package llamacpp

import (
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 2 * time.Second}

func Reachable(baseURL string) bool {
	resp, err := httpClient.Get(baseURL + "/v1/models")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func WaitForReady(seconds int, baseURL string) bool {
	for i := 0; i < seconds; i++ {
		if Reachable(baseURL) {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}
