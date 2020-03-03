package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toffysoft/health-noti.git/notify"
	"github.com/toffysoft/health-noti.git/schedule"
)

type Webhook struct {
	Endpoint               string      `json:"endpoint"`
	MonitorDashboardLink   string      `json:"MONITOR_DASHBOARD_LINK"`
	Monitortype            string      `json:"MONITORTYPE"`
	StatusChangeAttributes []Attribute `json:"STATUS_CHANGE_ATTRIBUTES"`
	FailedAttributes       []Attribute `json:"FAILED_ATTRIBUTES"`
	MonitorID              int64       `json:"MONITOR_ID"`
	Status                 string      `json:"STATUS"`
	Monitorname            string      `json:"MONITORNAME"`
	CT                     string      `json:"ct"`
	FailedLocations        string      `json:"FAILED_LOCATIONS"`
	IncidentReason         string      `json:"INCIDENT_REASON"`
	OutageTimeUnixFormat   string      `json:"OUTAGE_TIME_UNIX_FORMAT"`
	Monitorurl             string      `json:"MONITORURL"`
	Timezone               string      `json:"TIMEZONE"`
	MonitorGroupname       string      `json:"MONITOR_GROUPNAME"`
	Pollfrequency          int64       `json:"POLLFREQUENCY"`
	IncidentTime           string      `json:"INCIDENT_TIME"`
	IncidentTimeISO        string      `json:"INCIDENT_TIME_ISO"`
}

type Attribute struct {
	Attributeid   string  `json:"attributeid"`
	AlertType     string  `json:"alertType"`
	AttributeName string  `json:"attributeName"`
	ChildID       int64   `json:"childId"`
	ChildName     string  `json:"childName"`
	Status        *string `json:"status,omitempty"`
}

func main() {

	go schedule.Run()

	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	router.Use(cors)

	router.POST("/api/webhook", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Printf("%s \n", err)
		}
		defer c.Request.Body.Close()

		var w Webhook
		err = json.Unmarshal(body, &w)
		if err != nil {
			fmt.Printf("%s \n", err)
		}

		m := w.Status + " => " + w.Monitorurl + " : " + w.IncidentReason + "( " + w.IncidentTime + " )"

		go notify.Notify(m)

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "OK"})
	})

	router.Run(":8080")
}

func cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "authorization, Authorization, origin, content-type, accept")
	c.Header("Allow", "HEAD,GET,OPTIONS")
	c.Header("Content-Type", "application/json")
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
