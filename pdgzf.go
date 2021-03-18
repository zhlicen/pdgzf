package pdgzf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

type Queue struct {
	Qualification QueueItem `json:"qualification"`
}

// QueueItem 排队信息
type QueueItem struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
}

// GetHouses 根据请求查询房源列表
// reqStr https://select.pdgzf.com/houseLists 页面定制查询条件，复制POST参数JSON字符串
// asscessToken 登陆请求回复中获得，不填无法获取到 Qualification
func GetHouses(reqStr string, accessToken string) []*House {
	houses := []*House{}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://select.pdgzf.com/api/v1.0/app/gzf/house/list"), bytes.NewBuffer([]byte(reqStr)))
	if err != nil {
		log.Println(err)
		return houses
	}
	if accessToken != "" {
		req.Header.Set("gzfauthentication", accessToken)
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
// asscessToken 登陆请求回复中获得
func GetQueue(houseID int, accessToken string) []*QueueItem {
	queue := []*QueueItem{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://select.pdgzf.com/api/v1.0/app/gzf/house/%d", houseID), nil)
	if err != nil {
		log.Println(err)
		return queue
	}
	req.Header.Set("gzfauthentication", accessToken)
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
