package schedule

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/toffysoft/health-noti.git/conf"
	"github.com/toffysoft/health-noti.git/notify"
)

func handleResponse(resp *resty.Response, err error, e conf.Endpoint) {
	var c conf.Conf
	c.GetConf("conf.yaml")

	f := "2 Jan 2006 15:04:05"

	if err != nil {
		s := fmt.Sprintf("Alert (%s) => Request Timeout Reason : %s", time.Now().Format(f), err)
		notify.Notify(s)
	} else if resp.Time().Milliseconds() > e.TimeLimit || resp.StatusCode() != 200 {
		s := fmt.Sprintf("Alert (%s)  %s => Response Time : %d ms | Response Status : %s", time.Now().Format(f), c.BaseURL+e.Path, resp.Time().Milliseconds(), resp.Status())
		notify.Notify(s)
	}
}

func Run() {
	var authToken string
	var c conf.Conf
	c.GetConf("conf.yaml")
	client := resty.New()
	client.SetTimeout(1 * time.Minute)
	for {

		authResp, err := client.R().
			EnableTrace().
			SetBody(c.AuthenticationEndpoint.Body).Post(c.BaseURL + c.AuthenticationEndpoint.Path)

		handleResponse(authResp, err, c.AuthenticationEndpoint)

		// // Explore response object
		// fmt.Println("Response Info:")
		// fmt.Println("Error      :", err)
		// fmt.Println("Status Code:", authResp.StatusCode())
		// fmt.Println("Status     :", authResp.Status())
		// fmt.Println("Time       :", authResp.Time())
		// fmt.Println("Received At:", authResp.ReceivedAt())
		// fmt.Println("Body       :\n", authResp)
		// fmt.Println()

		// // Explore trace info
		// fmt.Println("Request Trace Info:")
		// ti := authResp.Request.TraceInfo()
		// fmt.Println("DNSLookup    :", ti.DNSLookup)
		// fmt.Println("ConnTime     :", ti.ConnTime)
		// fmt.Println("TLSHandshake :", ti.TLSHandshake)
		// fmt.Println("ServerTime   :", ti.ServerTime)
		// fmt.Println("ResponseTime :", ti.ResponseTime)
		// fmt.Println("TotalTime    :", ti.TotalTime)
		// fmt.Println("IsConnReused :", ti.IsConnReused)
		// fmt.Println("IsConnWasIdle:", ti.IsConnWasIdle)
		// fmt.Println("ConnIdleTime :", ti.ConnIdleTime)

		var authRespBody map[string]interface{}
		err = json.Unmarshal(authResp.Body(), &authRespBody)
		if err != nil {
			fmt.Printf("%s \n", err)
		}

		if val, ok := authRespBody["token"].(string); ok {
			authToken = val
		}

		if err == nil {

			for _, e := range c.Endpoints {
				var resp *resty.Response
				var err error
				var req *resty.Request

				req = client.R().EnableTrace().SetHeader("Authorization", "Bearer "+authToken)

				if e.Method == "post" {
					resp, err = req.SetBody(e.Body).Post(c.BaseURL + e.Path)
				} else {
					resp, err = req.Get(c.BaseURL + e.Path)
				}
				handleResponse(resp, err, e)
			}
		}

		time.Sleep(time.Minute * c.IntervalTime)
	}
}
