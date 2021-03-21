package pdgzf

import (
	_ "embed"
	"fmt"
	"testing"
)

var queryString = `{"where":{"keywords":"","township":"310115125","projectId":null,"typeName":null,"rent":null},"pageIndex":0,"pageSize":10}`

//go:embed login.credential
var loginBody string

//go:embed login.jsessionid
var loginJSessionID string

func Test_getHouses(t *testing.T) {
	cookies, err := Login(loginBody, loginJSessionID)
	if err != nil {
		panic(err)
	}
	houses := GetHouses(queryString, cookies)
	fmt.Println(houses)
}

func Test_getQueue(t *testing.T) {
	cookies, err := Login(loginBody, loginJSessionID)
	if err != nil {
		panic(err)
	}
	fmt.Println(GetQueue(9852, cookies))
}

func TestLogin(t *testing.T) {
	cookies, err := Login(loginBody, loginJSessionID)
	if err != nil {
		panic(err)
	}
	fmt.Println(cookies)
}

func TestGetLoginArgs(t *testing.T) {
	fmt.Println(GetLoginArgs(`{"account":"xxx","password":"xxxx","captcha":"%s"}`,
		func(imageBase64 string) (result string, err error) { return "test", nil }))
}
