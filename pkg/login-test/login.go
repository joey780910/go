package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type UserLoginInfo struct {
	Grant_type    string
	Username      string
	Password      string
	Client_id     string
	Client_secret string
	Scope         string
}

type MyAToken struct {
	Token_type  string
	Expires_in  int
	AccessToken string
}

type TripData struct {
	Status string
	Name   string
	TripId int
}

func StartLogin(allUserLoginInfo []UserLoginInfo, loopTimes int, loginTimes int) {
	for j := 0; j < loopTimes; j++ {
		for i := 0; i < loginTimes; i++ {
			go Login(allUserLoginInfo[rand.Intn(len(allUserLoginInfo))], j*loginTimes+i)
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

// Login 登入
func Login(userLoginInfo UserLoginInfo, number int) {
	baseURL := "http://conciergeapi.hooloop.com"
	requestURL := baseURL + "/oauth/token"
	// 要 POST的 参数
	form := url.Values{
		"grant_type":    {userLoginInfo.Grant_type},
		"username":      {userLoginInfo.Username},
		"password":      {userLoginInfo.Password},
		"client_id":     {userLoginInfo.Client_id},
		"client_secret": {userLoginInfo.Client_secret},
		"scope":         {userLoginInfo.Scope},
	}

	// func Post(url string, bodyType string, body io.Reader) (resp *Response, err error) {
	start := time.Now()
	//fmt.Println(start)
	body := bytes.NewBufferString(form.Encode())
	myClient := http.Client{Timeout: time.Duration(time.Second * 60)}
	rsp, err := myClient.Post(requestURL, "application/x-www-form-urlencoded", body)
	if err != nil {
		//panic(err)
		fmt.Println(err)
	}
	defer rsp.Body.Close()

	content, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		//panic(err)
		fmt.Println(err)
	} else {
		elapsed := time.Since(start).Seconds()
		fmt.Println(number, ":", userLoginInfo.Client_id, len(string(content)), "spend", elapsed, "sec")
		// fmt.Println(string(content))

		// Get access token
		myAToken := MyAToken{}
		err := json.Unmarshal(content, &myAToken)
		if err != nil {
			fmt.Println("error:", err)
		}

		// Get Base
		baseData := GetBaseWithAuth(myAToken.AccessToken)
		fmt.Println(baseData.Placeid)

		// Get All Place
		placeData := GetPlaceWithAuth(myAToken.AccessToken)
		fmt.Println(placeData.Marks[0].Name)

		placeids := []string{}
		for i := 0; i < 6; i++ {
			placeids = append(placeids, placeData.Marks[rand.Intn(len(placeData.Marks)-1)+1].Placeid)
		}

		// Get route
		GetRouteStepsWithAuth(myAToken.AccessToken, baseData.Placeid, placeids)

		// 預設路線
		GetTripWithAuth(myAToken.AccessToken)

		// trip id
		PutTripRecordWithAuth(myAToken.AccessToken, placeids)
	}
}

// GetWithAuth http Get
func GetWithAuth(accessToken string, url string) (retbody []byte) {
	baseURL := "http://conciergeapi.hooloop.com"
	requestURL := baseURL + url

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	retbody = body
	return retbody
}

type BaseData struct {
	// Language          string
	// AvailableLanguage []string
	Placeid string
	// MapCenter         coordinate
	// MapScale          int
}

// GetBaseWithAuth 取得Base資訊
func GetBaseWithAuth(accessToken string) (retBaseData BaseData) {
	body := GetWithAuth(accessToken, "/base/en")

	baseData := BaseData{}
	err := json.Unmarshal(body, &baseData)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(string(body))

	retBaseData = baseData
	return retBaseData
}

type PlaceData struct {
	Marks []Mark
}

type Mark struct {
	Placeid         string
	Icon            string
	Phone           string
	Business_hours  BusinessHours
	Url             string
	Photo_reference []string
	Placetype       []int
	Name            string
	BriefInfo       string
	Content         string
	Address         string
}

// BusinessHours 營業時段
type BusinessHours struct {
	Periods []period
}

type period struct {
	Day   string
	Open  string
	Close string
}

type Coordinate struct {
	Lat float64
	Lng float64
}

// GetPlaceWithAuth 取得所有PlaceID資訊
func GetPlaceWithAuth(accessToken string) (retPlaceData PlaceData) {
	body := GetWithAuth(accessToken, "/place/en")

	placeData := PlaceData{}
	err := json.Unmarshal(body, &placeData)
	if err != nil {
		fmt.Println(err)
	}
	retPlaceData = placeData
	return retPlaceData
}

// GetTripWithAuth 取得預設路線
func GetTripWithAuth(accessToken string) {
	GetWithAuth(accessToken, "/predefinedtrip/en")
}

// GetRouteStepsWithAuth 取得路徑
func GetRouteStepsWithAuth(accessToken string, hotelPlaceID string, placeids []string) {
	viaPoints := ""
	for i := 0; i < len(placeids); i++ {
		viaPoints += placeids[i] + ","
	}
	viaPoints = strings.TrimRight(viaPoints, ",")

	body := GetWithAuth(accessToken, "/route/"+hotelPlaceID+"?viapoint="+viaPoints)
	fmt.Println(string(body))
}

// PutTripRecordWithAuth 取得tripID sample {"status":"OK","name":null,"id":239}
func PutTripRecordWithAuth(accessToken string, placeids []string) (retTripData TripData) {
	viaPoints := ""
	for i := 0; i < len(placeids); i++ {
		viaPoints += placeids[i] + ","
	}
	viaPoints = strings.TrimRight(viaPoints, ",")

	baseURL := "http://conciergeapi.hooloop.com"
	requestURL := baseURL + "/triprecord"

	req, err := http.NewRequest("PUT", requestURL, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("viapoint", viaPoints)
	req.Header.Set("language", "en")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println(string(body))
	tripData := TripData{}
	err = json.Unmarshal(body, &tripData)
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println(string(body))
	retTripData = tripData
	return retTripData
}
