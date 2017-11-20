package api

import (
	"net/http"

	"sync"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"time"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var mux sync.Mutex
var uiConnection *websocket.Conn

func RunAPI(bind string, b *balancer.Balancer) {
	logrus.Infof("Starting api server on " + bind)

	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		onUISocket(c.Writer, c.Request, b)
	})

	go sendStatisticsToUI(b)

	router.Run(bind)
}

func onUISocket(w http.ResponseWriter, r *http.Request, b *balancer.Balancer) {
	logrus.Debugf("UI joined")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Warnf("failed to set websocket upgrade: %+v", err)
		return
	}

	mux.Lock()
	uiConnection = c
	mux.Unlock()
}

func sendStatisticsToUI(b *balancer.Balancer) {
	toUiTicker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <- toUiTicker.C:
			mux.Lock()
			if uiConnection != nil {
				err := uiConnection.WriteJSON(b.Scheduler.GetStats())
				if err != nil {
					logrus.Errorf("connection to UI was closed, will not send updates now", err)
					uiConnection = nil
				}
			}
			mux.Unlock()
		}
	}
}
