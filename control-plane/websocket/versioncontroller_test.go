package websocket

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/clustering"
	mock_dao "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/dao"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_ResetConnections_IfWatcherIsNotStopped_WatcherDoCatchQuitSignal(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mock_dao.NewMockDao(ctrl)

	vController := &VersionController{
		watchers: NewWatcherSet(),
		data:     dao,
	}

	done := make(chan error)
	watcher := newWatcher()
	go waitWatchersQuitWithTimeout(watcher, 5*time.Second, done)

	vController.getWatchers().Push(watcher)
	vController.ResetConnections(clustering.NodeInfo{}, clustering.Role(1))

	assert.True(t, waitDone(done), "watcher has not got the quit signal")
}

func Test_ResetConnections_IfWatcherIsStopped_WatcherDoNotCatchQuitSignal(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mock_dao.NewMockDao(ctrl)

	vController := &VersionController{
		watchers: NewWatcherSet(),
		data:     dao,
	}

	done := make(chan error)
	watcher := newWatcher()
	go waitWatchersQuitWithTimeout(watcher, 5*time.Second, done)

	vController.getWatchers().Push(watcher)
	watcher.stopped = true // mark watcher like it has been stopped
	vController.ResetConnections(clustering.NodeInfo{}, clustering.Role(1))

	assert.False(t, waitDone(done), "watcher has got the quit signal. it should not, because the watcher is stopped")
}

func waitDone(done chan error) bool {
	err := <-done
	if err != nil {
		fmt.Printf("done came with error: %v", err)
		return false
	}
	return true
}

func waitWatchersQuitWithTimeout(w *watcher, duration time.Duration, done chan error) {
	select {
	case <-time.After(duration):
		done <- errors.New("time out: watcher has not received quit signal")
	case <-w.quit:
		done <- nil
	}
}
