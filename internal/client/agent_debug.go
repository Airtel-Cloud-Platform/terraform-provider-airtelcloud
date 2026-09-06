package client

import (
	"encoding/json"
	"fmt"
	"os"
)

func AgentDebugLog(location, message, hypothesisID string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"location":     location,
		"message":      message,
		"hypothesisId": hypothesisID,
		"data":         data,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return
	}

	f, err := os.OpenFile(
		"/tmp/agent-debug.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = fmt.Fprintln(f, string(b))
}
