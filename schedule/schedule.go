package schedule

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/toffysoft/health-noti.git/conf"
	"github.com/toffysoft/health-noti.git/notify"
)

var c conf.Conf

func handleResponse(resp *resty.Response, err error, e conf.Endpoint) {

	f := "2 Jan 2006 15:04:05"

	// // Explore response object
	// fmt.Println("Response Info:", e.Path)
	// fmt.Println("Error      :", err)
	// fmt.Println("Status Code:", resp.StatusCode())
	// fmt.Println("Status     :", resp.Status())
	// fmt.Println("Time       :", resp.Time())
	// fmt.Println("Received At:", resp.ReceivedAt())
	// fmt.Println("Body       :\n", resp)
	// fmt.Println()s

	// // Explore trace info
	// fmt.Println("Request Trace Info:")
	ti := resp.Request.TraceInfo()
	// fmt.Println("DNSLookup    :", ti.DNSLookup)
	// fmt.Println("ConnTime     :", ti.ConnTime)
	// fmt.Println("TLSHandshake :", ti.TLSHandshake)
	// fmt.Println("ServerTime   :", ti.ServerTime)
	// fmt.Println("ResponseTime :", ti.ResponseTime)
	// fmt.Println("TotalTime    :", ti.TotalTime)
	// fmt.Println("IsConnReused :", ti.IsConnReused)
	// fmt.Println("IsConnWasIdle:", ti.IsConnWasIdle)
	// fmt.Println("ConnIdleTime :", ti.ConnIdleTime)

	if err != nil {
		s := fmt.Sprintf("%s => Request Timeout Reason : %s / (%s)", e.Path, time.Now().Format(f), err)
		notify.Notify(s)
	} else if resp.Time().Milliseconds() > e.TimeLimit || resp.StatusCode() != 200 {

		// respBody := ""
		// if resp.StatusCode() != 200 {
		// 	respBody = resp.String()
		// 	- Response Body   	: %s
		// }

		// s := fmt.Sprintf(`%s => Response : %d ms | Response Status : %s / (%s)`, e.Path, resp.Time().Milliseconds(), resp.Status(), time.Now().Format(f))
		// s := fmt.Sprintf(`%s => Response : %d ms | Response Status : %s / (%s)
		// 	- DNSLookup    		: %s
		// 	- ConnectionTime  	: %s
		// 	- TLSHandshake    	: %s
		// 	- ServerTime    	: %s
		// 	- ResponseTime    	: %s
		// 	- TotalTime    		: %s
		// 	- IsConnReused    	: %t
		// 	- IsConnWasIdle    	: %t
		// 	- ConnIdleTime    	: %s`, e.Path, resp.Time().Milliseconds(), resp.Status(), time.Now().Format(f),
		// 	ti.DNSLookup, ti.ConnTime, ti.TLSHandshake, ti.ServerTime, ti.ResponseTime, ti.TotalTime, ti.IsConnReused, ti.IsConnWasIdle, ti.ConnIdleTime)
		s := fmt.Sprintf(`%s => Response : %d ms | Response Status : %s / (%s)
			- DNSLookup    		: %s
			- ConnectionTime  	: %s
			- TLSHandshake    	: %s
			- ServerTime    	: %s
			- ResponseTime    	: %s
			- TotalTime    		: %s`, e.Path, resp.Time().Milliseconds(), resp.Status(), time.Now().Format(f),
			ti.DNSLookup, ti.ConnTime, ti.TLSHandshake, ti.ServerTime, ti.ResponseTime, ti.TotalTime)
		notify.Notify(string(s))

		if resp.StatusCode() == 429 {
			h, _ := json.MarshalIndent(resp.RawResponse.Header, "", " ")
			msg := fmt.Sprintf("Response Header : => %s", string(h))
			notify.Notify(msg)
		}

	}
}

func Run() {
	var authToken string
	var propertyId string
	var propertyUnitId string

	var authRespBody map[string]interface{}
	var propertyRespBody map[string]interface{}

	client := resty.New()
	client.SetTimeout(1 * time.Minute)
	for {
		c.GetConf("conf.yaml")

		authResp, err := client.R().
			EnableTrace().
			SetBody(c.AuthenticationEndpoint.Body).Post(c.BaseURL + c.AuthenticationEndpoint.Path)

		handleResponse(authResp, err, c.AuthenticationEndpoint)

		if authResp.StatusCode() == 200 {

			_ = json.Unmarshal(authResp.Body(), &authRespBody)

			if val, ok := authRespBody["token"].(string); ok {
				authToken = val
			}

			propertyResp, err := client.R().EnableTrace().SetHeader("Authorization", "Bearer "+authToken).Get(c.BaseURL + c.PropertyEndpoint.Path)
			handleResponse(propertyResp, err, c.PropertyEndpoint)
			if authResp.StatusCode() == 200 {

				_ = json.Unmarshal(propertyResp.Body(), &propertyRespBody)

				if bodyData, ok := propertyRespBody["data"].(map[string]interface{}); ok {

					if userProperty, ok := bodyData["user_property"].([]interface{}); ok {

						if len(userProperty) > 0 {

							if property, ok := userProperty[0].(map[string]interface{}); ok {

								propertyId, _ = property["property_id"].(string)
								propertyUnitId, _ = property["property_unit_id"].(string)
							}

							if propertyId != "" && propertyUnitId != "" {

								for _, e := range c.Endpoints {
									var resp *resty.Response
									var err error
									var req *resty.Request

									req = client.R().EnableTrace().SetHeader("Authorization", "Bearer "+authToken)

									if e.Method == "post" {
										body := map[string]interface{}{}

										for k, v := range e.Body {
											body[k] = v
										}

										if e.RequiredProperty {
											body["property_id"] = propertyId
										}

										if e.RequiredPropertyUnit {
											body["property_unit_id"] = propertyUnitId
										}

										resp, err = req.SetBody(body).Post(c.BaseURL + e.Path)
									} else {

										query := map[string]string{}

										for k, v := range e.Query {
											query[k], _ = v.(string)
										}

										if e.RequiredProperty {
											query["property_id"] = propertyId
										}

										if e.RequiredPropertyUnit {
											query["property_unit_id"] = propertyUnitId
										}

										resp, err = req.SetQueryParams(query).Get(c.BaseURL + e.Path)
									}

									handleResponse(resp, err, e)
								}

							} else {
								msg := fmt.Sprintf("Error :  propertyId (%s) and propertyUnitId (%s)", propertyId, propertyUnitId)
								notify.Notify(msg)
							}

						} else {

							email, _ := c.AuthenticationEndpoint.Body["email"]
							msg := fmt.Sprintf("Error :  account %s has no property", email)
							notify.Notify(msg)
						}
					}
				}
			}
		}

		time.Sleep(time.Minute * c.IntervalTime)
	}
}
