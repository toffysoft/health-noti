package notify

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/toffysoft/health-noti.git/conf"
)

func Notify(m string) {
	var c conf.Conf

	c.GetConf("conf.yaml")
	data := url.Values{}
	data.Set("message", m)

	req, err := http.NewRequest("POST", "https://notify-api.line.me/api/notify", strings.NewReader(data.Encode()))
	req.Header.Set("Authorization", "Bearer "+c.LineToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s \n", err)
	}
	defer resp.Body.Close()
}
