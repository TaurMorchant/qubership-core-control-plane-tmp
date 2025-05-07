package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type tmStub struct {
	mutex            sync.Mutex
	receivedRequests map[string][]*http.Request
	responseQueue    map[string][]*http.Response
}

func StartTmStub(port int) *tmStub {
	tmStub := tmStub{
		receivedRequests: make(map[string][]*http.Request),
		responseQueue:    make(map[string][]*http.Response),
	}

	r := mux.NewRouter()
	r.PathPrefix("/api/v4/tenant-manager").HandlerFunc(tmStub.tenantApiStubHandler)
	r.HandleFunc("/clear", func(writer http.ResponseWriter, request *http.Request) {
		tmStub.clear()
		writer.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	tmAddr := fmt.Sprintf(":%d", port)
	log.Infof("Starting tenant-manager stub on port %s", tmAddr)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Panic in tenant-manager stub thread: %v", r)
			}
		}()
		log.Errorf("Stopped tenant-manager stub; Error: %v", http.ListenAndServe(tmAddr, r))
	}()

	tmUrl := fmt.Sprintf("http://localhost%s", tmAddr)
	ready := false
	deadline := time.Now().Add(60 * time.Second)
	for !ready && time.Now().Before(deadline) {
		resp, err := http.DefaultClient.Get(tmUrl + "/health")
		if err != nil {
			log.InfoC(ctx, "Failed to check tenant-manager stub status: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			ready = true
		} else {
			log.InfoC(ctx, "Tenant-manager stub liveness check failed... Status Code: %v", resp.StatusCode)
			time.Sleep(100 * time.Millisecond)
		}
	}
	if !ready {
		log.PanicC(ctx, "Tenant-manager stub is still not UP after timeout exceeded")
	}
	log.InfoC(ctx, "Started tenant-manager stub with URL %v", tmUrl)
	return &tmStub
}

func (tmStub *tmStub) tenantApiStubHandler(w http.ResponseWriter, r *http.Request) {
	tmStub.mutex.Lock()
	defer tmStub.mutex.Unlock()
	log.Info("Handle request %s", r.URL.String())

	requestPath := r.URL.Path
	requestCopy := copyRequest(r)
	if _, ok := tmStub.receivedRequests[requestPath]; !ok {
		tmStub.receivedRequests[requestPath] = make([]*http.Request, 0)
	}
	tmStub.receivedRequests[requestPath] = append(tmStub.receivedRequests[requestPath], requestCopy)

	if responses, ok := tmStub.responseQueue[requestPath]; ok {
		resp := *responses[0]
		responses = responses[1:]

		for headerName, headerVal := range resp.Header {
			w.Header().Set(headerName, strings.Join(headerVal, ","))
		}
		w.WriteHeader(resp.StatusCode)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		if _, err := w.Write(body); err != nil {
			panic(err)
		}
		return
	}

	RespondWithJson(w, 200, map[string]string{
		"message": "OK",
	})
}

func copyRequest(r *http.Request) *http.Request {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	requestCopy := r.Clone(r.Context())
	if len(body) != 0 {
		requestCopy.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	return requestCopy
}

func (tmStub *tmStub) clear() {
	tmStub.mutex.Lock()
	defer tmStub.mutex.Unlock()

	tmStub.receivedRequests = make(map[string][]*http.Request)
	tmStub.responseQueue = make(map[string][]*http.Response)
}

func (tmStub *tmStub) verifyRequests(t *testing.T, method, path string, num int) {
	tmStub.mutex.Lock()
	defer tmStub.mutex.Unlock()

	requestsToPath := tmStub.receivedRequests[path]
	if requestsToPath == nil {
		assert.Equal(t, num, 0)
		return
	}
	requestsNum := 0
	for _, req := range requestsToPath {
		if req.Method == method {
			requestsNum++
		}
	}
	assert.Equal(t, num, requestsNum)
}
