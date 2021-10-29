package entrypage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"projects/platToolv2/internal/head"
	"projects/platToolv2/internal/office"
	"reflect"
	"regexp"
	"time"
)

// CachedConfig is the global variable of config
var CachedConfig head.Config

// Init function
func init() {
	configPath := "./config/header.json"
	CachedConfig = head.LoadConfig(configPath)
}

// Request is the struct including query info
type Req struct {
	FrameURL      string
	ApiURL        string
	AuthorURL     string
	TimeoutSec    int
	SleepMilliSec int
	Headers       map[string]string
	BaseToken     string
	FrameToken    string
	ApiToken      string
	AuthorToken   string
}

func (r *Req) GetEntryID() string {
	frameURL := r.FrameURL

	pat := regexp.MustCompile("&ID=(.*?)&")
	finds := pat.FindStringSubmatch(frameURL)
	if len(finds) == 0 {
		return ""
	}
	id := finds[1]
	return id
}

func (r *Req) GetApiURL() string {
	id := r.GetEntryID()
	apiURL := fmt.Sprintf("https://www.zgbk.com/ecph/api/words/%s?Type=Extend&SiteID=1&Preview=false", id)
	return apiURL
}

func (r *Req) GetAuthorURL() string {
	id := r.GetEntryID()
	authorURL := fmt.Sprintf("https://www.zgbk.com/ecph/api/author?ID=%s", id)
	return authorURL
}

// a struct store the parsed metas
type Meta struct {
	title  string
	subc   string
	body   string
	ref    string
	author string
	remark string
}

// ParseFrameHTML parses the frame html
func ParseFrameHTML(html string, meta *Meta) {
	// title
	pat1 := regexp.MustCompile("<h2 class=\"fl\">(.*?)</h2>")
	finds := pat1.FindStringSubmatch(html)
	var title string
	if len(finds) == 0 {
		title = ""
	} else {
		title = finds[1]
	}
	meta.title = title

	//
	pat2 := regexp.MustCompile("<div class=\"summary fontsize\">([\\s\\S]*?)</div>")
	finds = pat2.FindStringSubmatch(html)
	var subc string
	if len(finds) == 0 {
		subc = ""
	} else {
		subc = finds[1]
	}
	meta.subc = subc
}

// ParseApiBytes parses the json data responsed from api
func ParseApiBytes(data []byte, meta *Meta) {
	var root map[string]interface{}
	json.Unmarshal(data, &root)

	_, ok := root["data"]
	if !ok {
		meta.body = ""
		meta.ref = ""
		return
	}

	contents := ""
	contentinfo, ok := root["data"].(map[string]interface{})["contentinfo"]
	if ok {
		if contentinfo != nil {
			// contents = contentinfo.([]interface{})[0].(map[string]interface{})["Content"].(string)
			slis := contentinfo.([]interface{})
			for _, sli := range slis {
				innerSli := sli.(map[string]interface{})
				content := innerSli["Content"].(string)
				name := innerSli["Name"]
				if name != nil {
					content = "<p>" + name.(string) + "</p>" + content
				}
				contents += content
			}
		}
	}
	meta.body = contents

	ref := ""
	extendinfo, ok := root["data"].(map[string]interface{})["extendinfo"]
	if ok {
		if extendinfo != nil {
			extendContent := extendinfo.([]interface{})[0].(map[string]interface{})["Content"].([]interface{})
			for _, v := range extendContent {
				ref += "<p>" + v.(string) + "</p>"
			}
		}
	}
	meta.ref = ref
}

//
func ParseAuthorBytes(data []byte, meta *Meta) {
	var root map[string]interface{}
	json.Unmarshal(data, &root)

	dt, ok := root["data"]
	if !ok || dt == nil {
		meta.author = ""
		return
	}

	itf, ok := root["data"].(map[string]interface{})
	if !ok {
		meta.author = ""
		return
	}
	_, ok = itf["datalist"]
	if !ok {
		meta.author = ""
		return
	}
	datalist := itf["datalist"].([]interface{})
	authors := ""
	for _, itf := range datalist {
		authorMeta := itf.(map[string]interface{})
		name := authorMeta["name"].(string)
		authors += name + ";"
	}

	meta.author = authors
}

func (r *Req) GetMeta() *Meta {
	meta := new(Meta)

	// frame page
	headers := r.Headers
	headers["token"] = r.FrameToken
	resp, err := head.HeadGet(r.FrameURL, headers, r.TimeoutSec)
	time.Sleep(time.Duration(r.SleepMilliSec) * time.Millisecond)
	if err != nil {
		meta.remark += "Failed to access frame page;"
	}
	defer resp.Body.Close()

	htmlBytes, _ := ioutil.ReadAll(resp.Body)
	ParseFrameHTML(string(htmlBytes), meta)

	// api json data
	headers["token"] = r.ApiToken
	headers["Referer"] = r.FrameURL
	resp, err = head.HeadGet(r.ApiURL, headers, r.TimeoutSec)
	time.Sleep(time.Duration(r.SleepMilliSec) * time.Millisecond)
	if err != nil {
		meta.remark += "Failed to access body api;"
	}
	defer resp.Body.Close()

	jsonBytes, _ := ioutil.ReadAll(resp.Body)
	ParseApiBytes(jsonBytes, meta)

	// author json data
	headers["token"] = r.AuthorToken
	resp, err = head.HeadGet(r.AuthorURL, headers, r.TimeoutSec)
	time.Sleep(time.Duration(r.SleepMilliSec) * time.Millisecond)
	if err != nil {
		meta.remark += "Failed to access author api;"
	}
	defer resp.Body.Close()
	authorBytes, _ := ioutil.ReadAll(resp.Body)
	ParseAuthorBytes(authorBytes, meta)

	return meta
}

// NewRequest generates a new Req type
func NewRequest(frameURL string, timeoutSec int, sleepMilliSec int) *Req {
	req := new(Req)

	req.FrameURL = frameURL
	req.TimeoutSec = timeoutSec
	req.SleepMilliSec = sleepMilliSec

	apiURL := req.GetApiURL()
	req.ApiURL = apiURL

	authorURL := req.GetAuthorURL()
	req.AuthorURL = authorURL

	req.Headers = CachedConfig.Headers
	req.BaseToken = CachedConfig.BaseToken

	frameToken := head.GetPageToken(CachedConfig.BaseToken, frameURL)
	req.FrameToken = frameToken

	apiToken := head.GetPageToken(CachedConfig.BaseToken, apiURL)
	req.ApiToken = apiToken

	authorToken := head.GetPageToken(CachedConfig.BaseToken, authorURL)
	req.AuthorToken = authorToken

	return req
}

func pTag(s string) []string {
	pat := regexp.MustCompile("<p[\\s\\S]*?</p>")
	ps := pat.FindAllString(s, -1)

	if len(ps) == 0 {
		return []string{head.DelTag(s)}
	}

	// delete tags
	for i := 0; i < len(ps); i++ {
		ps[i] = head.DelTag(ps[i])
	}

	return ps
}

func BlankLine() office.Para {
	return office.Para{Typ: "text", Text: "\n"}
}

func CollectPara(entries [][]string) []office.Para {
	var paras []office.Para
	for _, entry := range entries[1:] {
		title := "条目名称：" + entry[2]
		para := office.Para{Typ: "text", Text: title}
		paras = append(paras, para)

		url := "网址：" + entry[1]
		para = office.Para{Typ: "text", Text: url}
		paras = append(paras, para)

		subc := entry[3]
		ps := pTag(subc)
		for _, item := range ps {
			para = office.Para{Typ: "text", Text: item}
			paras = append(paras, para)
		}

		body := entry[4]
		ps = pTag(body)
		for _, item := range ps {
			para = office.Para{Typ: "text", Text: item}
			paras = append(paras, para)
		}

		ref := entry[5]
		ps = pTag(ref)
		for _, item := range ps {
			para = office.Para{Typ: "text", Text: item}
			paras = append(paras, para)
		}

		author := "作者：" + entry[6]
		para = office.Para{Typ: "text", Text: author}
		paras = append(paras, para)

		paras = append(paras, BlankLine())
		paras = append(paras, BlankLine())
	}

	return paras
}

// TaskRun function conducts the total task
func TaskRun(inpath string, outpath string, docxpath string, timeout int, sleep int) {
	entries, err := head.ReadExcel(inpath, "Sheet1")
	if err != nil {
		panic(err)
	}
	colNames := entries[0]
	entries = entries[1:]
	entryNum := len(entries)

	var keys []string
	var probe Meta
	typ := reflect.TypeOf(probe)
	for i := 0; i < typ.NumField(); i++ {
		keys = append(keys, typ.Field(i).Name)
	}

	var res [][]string
	res = append(res, append(colNames, keys...))
	for i, item := range entries {
		url := item[1]
		req := NewRequest(url, timeout, sleep)
		meta := req.GetMeta()

		v := reflect.ValueOf(*meta)
		for j := 0; j < v.NumField(); j++ {
			item = append(item, v.Field(j).String())
		}

		res = append(res, item)

		fmt.Printf("\r完成进度：%v / %v", i+1, entryNum)
	}

	// save to excel file
	err = head.WriteExcel(outpath, res)
	if err != nil {
		panic(err)
	}

	// save to docx file
	paras := CollectPara(res)
	office.WriteDocxFile(docxpath, paras)
}

// a single run on the given url
func SingleRun(url string, timeout int, sleep int) []string {
	var res []string
	req := NewRequest(url, timeout, sleep)
	meta := req.GetMeta()

	v := reflect.ValueOf(*meta)
	for i := 0; i < v.NumField(); i++ {
		res = append(res, v.Field(i).String())
	}
	return res
}
