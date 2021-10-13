package querypage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"projects/platToolv2/internal/head"
)

// CachedConfig is the global variable of config
var CachedConfig head.Config

// Init function
func init() {
	configPath := "./config/header.json"
	CachedConfig = head.LoadConfig(configPath)
}

// Query is the struct including query info
type Query struct {
	Word             string
	Accuracy         string
	Field            string
	Scope            string
	ApiURL           string
	AdvanceSearchURL string
	TimeoutSec       int
	SleepMilliSec    int
	Headers          map[string]string
	BaseToken        string
	FormData         url.Values
}

// SetFormData get the form data and set it to the struct field
func (q *Query) SetFormData() {
	data := map[string][]string{
		"pageIndex": {"0"},
		"pageSize":  {"8"},
		"siteID":    {"1"},
	}
	cond := make([]map[string]string, 0)
	if q.Field == "title" {
		cond = []map[string]string{
			{"accuracy": q.Accuracy, "field": q.Field, "value": q.Word, "logic": "and"},
			{"accuracy": q.Accuracy, "field": q.Field, "logic": "or", "value": q.Word},
		}
	} else if q.Field == "content" {
		cond = []map[string]string{
			{"accuracy": q.Accuracy, "field": q.Field, "value": q.Word, "logic": "and"},
		}
	} else {
		panic("Failed to get form data.")
	}
	condByte, err := json.Marshal(cond)
	if err != nil {
		panic(err)
	}
	condStr := string(condByte)
	data["conditions"] = []string{condStr}

	// Set the FormData field
	q.FormData = data
}

// Get random sleeptime according to the scale
func (q *Query) GetRandomSleeptime() time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(q.SleepMilliSec+rand.Intn(1000)) * time.Millisecond
}

// MustPost request by POST method
func (q *Query) MustPost(attemptNum int) ([]byte, error) {
	pageToken := head.GetPageToken(q.BaseToken, q.ApiURL)
	q.Headers["token"] = pageToken
	q.Headers["Referer"] = q.AdvanceSearchURL

	var byteByte []byte
	var err error
	for t := 0; t < attemptNum; t++ {
		byteByte, err = head.PostForm(q.ApiURL, q.Headers, q.FormData, q.TimeoutSec)
		if err != nil {
			log.Println(err)
			time.Sleep(q.GetRandomSleeptime())
		} else {
			break
		}
	}
	time.Sleep(q.GetRandomSleeptime())
	return byteByte, err
}

func GetData(byteData []byte, dataPtr *map[string]interface{}) error {
	jsonData := make(map[string]interface{})
	json.Unmarshal(byteData, &jsonData)

	var data map[string]interface{}
	data, ok := jsonData["data"].(map[string]interface{})
	if ok != true {
		err := errors.New("Failed to parse the jsonData(step1).")
		return err
	}
	*dataPtr = data
	err := error(nil)
	return err
}

func GetDataList(byteData []byte, dataList *[]map[string]interface{}) error {
	data := make(map[string]interface{})
	err := GetData(byteData, &data)
	if err != nil {
		return err
	}

	midd1, ok := data["datalist"].([]interface{})
	if ok != true {
		err := errors.New("Failed to parse the jsonData(step2).")
		return err
	}

	for _, v := range midd1 {
		midd2, ok := v.(map[string]interface{})
		if ok != true {
			err := errors.New("Failed to parse the jsonData(step3).")
			return err
		}
		*dataList = append(*dataList, midd2)
	}

	err = error(nil)
	return err
}

func GetTotalNum(byteData []byte, totalPtr *int) error {
	data := make(map[string]interface{})
	err := GetData(byteData, &data)
	if err != nil {
		return err
	}
	total := data["total"].(float64)
	*totalPtr = int(total)

	err = error(nil)
	return err
}

// QueryReply is the struct include infomation we need
type QueryReply struct {
	Title          string
	Num            int
	Presented      bool
	FirstPageReply string
}

func (q *Query) GetInfo() *QueryReply {
	infoPtr := new(QueryReply)
	byteData, err := q.MustPost(5)
	if err != nil {
		return infoPtr
	}

	var dataList []map[string]interface{}
	err = GetDataList(byteData, &dataList)
	if err != nil {
		panic(err)
	}
	var entries [][]string
	presented := false
	for _, v := range dataList {
		t, _ := v["title"].(string)
		t = head.DelTag(t)
		wikiType, _ := v["wikiType"].(string)
		twoSubjectName, _ := v["twoSubjectName"].(string)
		entries = append(entries, []string{t, wikiType, twoSubjectName})

		if t == q.Word {
			presented = true
		}
	}

	// 写入搜索结果数据
	infoPtr.Title = q.Word
	var num int
	err = GetTotalNum(byteData, &num)
	if err != nil {
		num = 0
	}
	infoPtr.Num = num
	infoPtr.Presented = presented
	if len(entries) == 0 {
		infoPtr.FirstPageReply = ""
	} else {
		infoPtr.FirstPageReply = head.UnescapeSerialize(entries)
	}
	// fmt.Printf("%+v\n", infoPtr)

	return infoPtr
}

// NewQuery is the construct function of type Query
func NewQuery(word string, accuracy string, field string, scope string, timeout int, sleep int) *Query {
	q := new(Query)
	q.Word = word
	q.Accuracy = accuracy
	q.Field = field
	q.Scope = scope
	q.TimeoutSec = timeout
	q.SleepMilliSec = sleep
	q.Headers = CachedConfig.Headers
	q.BaseToken = CachedConfig.BaseToken

	q.ApiURL = "https://www.zgbk.com/ecph/api/search"
	q.AdvanceSearchURL = fmt.Sprintf("https://www.zgbk.com/ecph/advanceSearch/result?SiteID=1&Alias=%s", scope)

	// Set FormData field
	q.SetFormData()

	return q
}

// TaskRun function conducts the total task:
// read word from excel, query to the API, collect the reponse information, write to excel finally.
func TaskRun(inpath string, outpath string, accuracy string, field string, scope string, timeout int, sleep int) {
	// Load data from excel
	entries, err := head.ReadExcel(inpath, "")
	if err != nil {
		panic(err)
	}
	columns := entries[0]
	entries = entries[1:]
	entryNum := len(entries)

	// result data variable
	result := [][]string{}

	// Query to the API
	for i, row := range entries {
		word := row[1]
		q := NewQuery(word, accuracy, field, scope, timeout, sleep)
		info := q.GetInfo()

		tempSli := []string{}
		for _, cell := range row {
			tempSli = append(tempSli, cell)
		}
		if info.Presented {
			tempSli = append(tempSli, "是")
		} else {
			tempSli = append(tempSli, "")
		}
		tempSli = append(tempSli, strconv.Itoa(info.Num))
		tempSli = append(tempSli, info.FirstPageReply)

		result = append(result, tempSli)

		// Processbar
		wordBar := word
		for i := len([]rune(word)); i < 32; i++ {
			wordBar += "　"
		}
		fmt.Printf("\r检索进度：%v / %v \t %v", i+1, entryNum, wordBar)
	}
	fmt.Println("")
	fmt.Println("Complete!")

	// Save the result
	columns = append(columns, "是否上线", "结果总数", "前8项")
	result = append([][]string{columns}, result...)
	head.WriteExcel(outpath, result)
}
