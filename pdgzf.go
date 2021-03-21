package pdgzf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"
)

// House 房源信息
type House struct {
	ID              int64   `json:"id"`
	FullName        string  `json:"fullName"`
	TypeName        int     `json:"typeName"`
	Queue           []Queue `json:"queue"`
	QueueCount      int     `json:"queueCount"`
	SelectStartTime string  `json:"selectStartTime"`
}

type Period struct {
	Name      string `json:"name"`
	StartTime string `json:"startTime"`
}

type Queue struct {
	Qualification QueueItem `json:"qualification"`
	Status        string    `json:"status"`
	Position      int       `json:"position"`
	Period        Period    `json:"period"`
}

// QueueItem 排队信息
type QueueItem struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
}

// Login 登陆验证码不好破...这里直接复制登陆请求(一个漏洞)
// 抓取一次成功的登陆请求可以无限复用， JSESSIONID有效期待验证,
// loginReqStr是login的body
// JSESSIONID从cookie中找
func Login(loginReqStr, jSessionID string) (cookies []*http.Cookie, err error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://select.pdgzf.com/api/v1.0/app/gzf/user/login"), bytes.NewBuffer([]byte(loginReqStr)))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("Cookie", fmt.Sprintf(`JSESSIONID=%s`, jSessionID))
	req.Header.Set("Content-Type", "application/json")
	cli := http.Client{}
	rsp, err := cli.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return rsp.Cookies(), nil
}

// GetHouses 根据请求查询房源列表
// reqStr https://select.pdgzf.com/houseLists 页面定制查询条件，复制POST参数JSON字符串
// cookies 从Login中拿，也可以自己从header中取GZFAuthentication（有效期24h），存入http.Cookie结构中
// 如果无需排队数据，cookies参数填空即可
func GetHouses(reqStr string, cookies []*http.Cookie) []*House {
	houses := []*House{}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://select.pdgzf.com/api/v1.0/app/gzf/house/list"), bytes.NewBuffer([]byte(reqStr)))
	if err != nil {
		log.Println(err)
		return houses
	}
	for _, cookie := range cookies {
		if cookie.Name == "GZFAuthentication" {
			v, _ := url.QueryUnescape(cookie.Value)
			req.Header.Set("GZFAuthentication", v)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	cli := http.Client{}
	rsp, err := cli.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	rb, err := io.ReadAll(rsp.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	ret := make(map[string]interface{})
	err = json.Unmarshal(rb, &ret)
	if err != nil {
		log.Println(err)
		return nil
	}
	now := time.Now()
	now = now.Add(-95 * time.Minute)
	housesData := gjson.Get(string(rb), "data.data")

	for _, houseData := range housesData.Array() {
		if houseData.IsObject() {
			house := &House{}
			err := json.Unmarshal([]byte(houseData.Raw), house)
			if err != nil {
				log.Println(err)
				continue
			}
			houses = append(houses, house)
		}
	}
	return houses
}

// GetQueue 获取房源排队队列
// houseID 房源ID
// cookies 从Login中拿，也可以自己从header中取GZFAuthentication（有效期24h）
// 如果无需排队数据，cookies参数填空即可
func GetQueue(houseID int, cookies []*http.Cookie) []*QueueItem {
	queue := []*QueueItem{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://select.pdgzf.com/api/v1.0/app/gzf/house/%d", houseID), nil)
	if err != nil {
		log.Println(err)
		return queue
	}
	for _, cookie := range cookies {
		if cookie.Name == "GZFAuthentication" {
			v, _ := url.QueryUnescape(cookie.Value)
			req.Header.Set("GZFAuthentication", v)
		}
		//req.AddCookie(cookie)
	}
	cli := http.Client{}
	rsp, err := cli.Do(req)
	rb, err := io.ReadAll(rsp.Body)
	if err != nil {
		log.Println(err)
		return queue
	}
	ret := make(map[string]interface{})
	err = json.Unmarshal(rb, &ret)
	if err != nil {
		log.Println(err)
		return queue
	}
	queueListData := gjson.Get(string(rb), "data.queue")
	for _, queueItemData := range queueListData.Array() {
		queueItem := &QueueItem{}
		err = json.Unmarshal([]byte(queueItemData.Get("qualification").Raw), queueItem)
		if err != nil {
			log.Println(err)
			return queue
		}
		queue = append(queue, queueItem)
	}
	return queue
}
