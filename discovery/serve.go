package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/gorilla/mux"
)

var serveCommand = cli.Command{
	Name:  "serve",
	Usage: "serve the REST api for the discovery service",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "addr", Value: ":8080", Usage: "ip and port to serve the HTTP api"},
		cli.IntFlag{Name: "ttl", Value: 30, Usage: "set the default ttl that is required for slave information"},
	},
	Action: serveAction,
}

type requestInfo struct {
	Username string
	Cluster  string
	SlaveID  string
}

func newRequestInfo(r *http.Request) requestInfo {
	vars := mux.Vars(r)

	return requestInfo{
		Username: vars["username"],
		Cluster:  vars["cluster"],
		SlaveID:  vars["slave"],
	}
}

type server struct {
	r      *mux.Router
	client *etcd.Client
	ttl    uint64
}

func newServer(context *cli.Context) http.Handler {
	s := &server{
		r:      mux.NewRouter(),
		client: getEtcdClient(context),
		ttl:    uint64(context.Int("ttl")),
	}

	// list the slaves in the cluster
	s.r.HandleFunc("/u/{username:.*}/{cluster:.*}", s.listClusterSlaves).Methods("GET")

	// update slave information for the cluster
	s.r.HandleFunc("/u/{username:.*}/{cluster:.*}/{slave:.*}", s.updateSlave).Methods("POST")

	// delete the cluster
	s.r.HandleFunc("/u/{username:.*}/{cluster:.*}", s.deleteCluster).Methods("DELETE")

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

// listClusterSlaves returns a list of all the slave's addresses in the cluster
// for the specific user and cluster name
//
// GET /u/crosbymichael/testcluster
//
// ["192.168.56.1:2375"]
func (s *server) listClusterSlaves(w http.ResponseWriter, r *http.Request) {
	var (
		ips []string

		info = newRequestInfo(r)
	)

	resp, err := s.client.Get(filepath.Join("/citadel/discovery", info.Username, info.Cluster, "slaves"), true, true)
	if err != nil {
		logger.WithField("error", err).Error("list cluster slaves")

		writeError(w, err, info)
		return
	}

	for _, n := range resp.Node.Nodes {
		ips = append(ips, n.Value)
	}

	w.Header().Set("content-type", "application/json")
	if len(ips) == 0 {
		if _, err := w.Write([]byte("[]")); err != nil {
			logger.WithField("error", err).Error("encode slave ips")
		}

		return
	}

	if err := json.NewEncoder(w).Encode(ips); err != nil {
		logger.WithField("error", err).Error("encode slave ips")
	}
}

func (s *server) updateSlave(w http.ResponseWriter, r *http.Request) {
	info := newRequestInfo(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.WithField("error", err).Error("read request body for addr")

		writeError(w, err, info)
		return
	}

	if _, err := s.client.Set(filepath.Join("/citadel/discovery", info.Username, info.Cluster, "slaves", info.SlaveID), string(data), s.ttl); err != nil {
		logger.WithField("error", err).Error("read request body for addr")

		writeError(w, err, info)
		return
	}
}

func (s *server) deleteCluster(w http.ResponseWriter, r *http.Request) {
	info := newRequestInfo(r)

	if _, err := s.client.Delete(filepath.Join("/citadel/discovery", info.Username, info.Cluster), true); err != nil {
		logger.WithField("error", err).Error("list cluster slaves")

		writeError(w, err, info)
		return
	}
}

func writeError(w http.ResponseWriter, err error, info requestInfo) {
	if isNotFound(err) {
		http.Error(w, fmt.Sprintf("cluster for %s/%s not found", info.Username, info.Cluster), http.StatusNotFound)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// isNotFound returns true if the error is an etcd key not found error
func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "Key not found")
}

func serveAction(context *cli.Context) {
	s := newServer(context)

	logger.WithField("addr", context.String("addr")).Info("start discovery service")

	if err := http.ListenAndServe(context.String("addr"), s); err != nil {
		logger.Fatal(err)
	}
}
