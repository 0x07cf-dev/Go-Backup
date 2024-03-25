package notify

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/0x07cf-dev/go-backup/internal/logger"
)

type Notifier struct {
	// ntfy.sh
	Host  string
	Topic string
	Token string
	// Health
	HealthMonitors []HealthMonitors
}

func NewNotifierFromEnv() (*Notifier, error) {
	host := os.Getenv("NTFY_HOST")
	topic := os.Getenv("NTFY_TOPIC")
	token := os.Getenv("NTFY_TOKEN")

	// Find all health monitors defined in environment
	var monitors []HealthMonitors
	var monitorEnvVars = map[string]HealthMonitors{
		"NTFY_HEALTHCHECKS": HealthChecksIO,
		"NTFY_BETTERUPTIME": BetterUptime,
	}
	for envVar, monitor := range monitorEnvVars {
		if m, ok := os.LookupEnv(envVar); ok {
			monitors = append(monitors, monitor)
			logger.Debugf("%s = %s", envVar, m)
		}
	}

	return NewNotifier(host, topic, token, monitors...)
}

func NewNotifier(host string, topic string, token string, healthMonitors ...HealthMonitors) (*Notifier, error) {
	_, err := url.ParseRequestURI(host)
	if err != nil {
		return nil, fmt.Errorf("invalid ntfy.sh host")
	}
	if topic == "" {
		return nil, fmt.Errorf("invalid ntfy.sh topic")
	}
	logger.Infof("Notifier: %s/%s", host, topic)
	return &Notifier{
		Host:           host,
		Topic:          topic,
		Token:          token,
		HealthMonitors: healthMonitors,
	}, nil
}

func (notifier *Notifier) SendHeartbeats(endpoint string, withLog bool) (string, error) {
	resultCh := make(chan string, len(notifier.HealthMonitors))
	errCh := make(chan error, len(notifier.HealthMonitors))

	for _, mon := range notifier.HealthMonitors {
		// POST log file
		var buf bytes.Buffer
		if withLog && monitorParams[mon].Method == "POST" {
			file, err := os.Open(logger.LogPath)
			if err != nil {
				fmt.Println("Error opening file:", err)
			}
			defer file.Close()
			buf.ReadFrom(file)
		}

		// Ping uptime monitor
		resp, err := mon.Ping(endpoint, &buf)
		if err != nil {
			errCh <- err
		}
		resultCh <- resp
	}

	// Collect successes
	close(resultCh)
	var ret strings.Builder
	for res := range resultCh {
		ret.WriteString(res)
	}

	// Collect errors
	var errStrings []string
	for err := range errCh {
		errStrings = append(errStrings, err.Error())
	}

	if len(errStrings) > 0 {
		return ret.String(), fmt.Errorf(strings.Join(errStrings, "; "))
	}

	return ret.String(), nil
}

func (notifier *Notifier) Send(title string, body string, tags []string, opts ...MsgOptFunc) (string, error) {
	return notifier.SendMessage(newMessage(notifier.Topic, title, body, tags, opts...))
}

func (notifier *Notifier) SendMessage(message *Message) (string, error) {
	jsonData, err := message.Marshal()
	if err != nil {
		return "", err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if len(notifier.Token) > 0 {
		headers["Authorization"] = "Bearer " + notifier.Token
	}

	resp, err := httpPost(notifier.Host, bytes.NewBuffer(jsonData), 10, 5, headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ret := resp.Status
	if resp.StatusCode == http.StatusUnauthorized {
		if len(notifier.Token) > 0 {
			ret = "invalid auth token"
		}
		ret = "authentication required"
	}
	return ret, nil
}

func httpPost(url string, body *bytes.Buffer, timeout int, retries int, headers map[string]string) (*http.Response, error) {
	return httpRequest("POST", url, body, timeout, retries, headers)
}

func httpRequest(method string, url string, body *bytes.Buffer, timeout int, retries int, headers map[string]string) (*http.Response, error) {
	var client = &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Retry until request succeeds or exceeds max attempts
	for i := 0; i < retries; i++ {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, err
		}

		// Set request headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Failure
		resp.Body.Close()

		// Wait before retrying
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("maximum attempts exceeded")
}
