package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// FetchSlaves returns the slaves for the discovery service at the specified endpoint
func FetchSlaves(endpoint, username, cluster string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/u/%s/%s", endpoint, username, cluster))
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	var addrs []string
	if err := json.NewDecoder(resp.Body).Decode(&addrs); err != nil {
		return nil, err
	}

	return addrs, nil
}

// RegisterSlave adds a new slave identified by the slaveID into the discovery service
// the default TTL is 30 secs
func RegisterSlave(endpoint, username, cluster, slaveID, addr string) error {
	buf := strings.NewReader(addr)

	_, err := http.Post(fmt.Sprintf("%s/u/%s/%s/%s", endpoint, username, cluster, slaveID), "application/json", buf)
	return err
}
