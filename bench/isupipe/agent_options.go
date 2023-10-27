package isupipe

import (
	"net/http"

	"github.com/isucon/isucandar/agent"
)

func WithClient(c *http.Client) agent.AgentOption {
	return func(a *agent.Agent) error {
		a.HttpClient = c

		return nil
	}
}
