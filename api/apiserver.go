package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/api/models"
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

func RunAPI(bind string, b *balancer.Balancer) {
	logrus.Infof("Starting api server on " + bind)

	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		onUISocket(c.Writer, c.Request, b)
	})

	router.Run(bind)
}

func onUISocket(w http.ResponseWriter, r *http.Request, b *balancer.Balancer) {
	logrus.Debugf("UI joined")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Warnf("failed to set websocket upgrade: %+v", err)
		return
	}

	go handleFromUI(b, c)
	go handleToUI(b, c)
}

func handleToUI(b *balancer.Balancer, c *websocket.Conn) {
	for {
		var msg models.BaseModel = <- b.Scheduler.ToUi

		err := c.WriteJSON(msg)
		if err != nil {
			logrus.Errorf("web socket to UI was closed, resending message", err)
			b.Scheduler.ToUi <- msg
			break
		}
	}
}

func handleFromUI(b *balancer.Balancer, c *websocket.Conn) {
	for {
		var msg models.BaseModel
		err := c.ReadJSON(&msg)
		if err != nil {
			logrus.Warnf("error reading json on websocket: %v", err)
			break
		}

		logrus.Infof("message from client: ", msg)

		err = c.WriteJSON(msg)
		if err != nil {
			logrus.Warnf("error sending message to UI on websocket: ", err)
		}
	}
}
