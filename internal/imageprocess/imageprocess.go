package imageprocess

import (
	"encoding/json"
	"fmt"
	"net/http"
	"projects/platToolv2/internal/entrypage"
	"projects/platToolv2/internal/head"
	"projects/platToolv2/internal/querypage"
	"regexp"
	"strconv"
)

// 调用querypage的SingleRun函数
func querypageSR(word string) [][]string {
	accuracy := "1"
	field := "title"
	scope := "all"
	timeout := 5
	sleep := 500
	info := querypage.SingleRun(word, accuracy, field, scope, timeout, sleep)

	var entries [][]string
	presented := info.Presented
	// 只处理名称精确相同情况
	if presented {
		replyStr := info.FirstPageReply
		var replies [][]string
		json.Unmarshal([]byte(replyStr), &replies)

		// 取出名称精确相同的词条
		for _, v := range replies {
			if v[0] == word {
				entries = append(entries, v)
			}
		}
	}

	return entries
}

// 调用entrypage的SingleRun函数
func entrypageSR(url string) []string {
	timeout := 5
	sleep := 500
	res := entrypage.SingleRun(url, timeout, sleep)
	return res
}

type EntryImgInfo struct {
	EntryURL      string
	ThumbnailURLs []string
	ResourceURLs  []string
	Num           int
	MissMatch     bool
	NotHealth     bool
	Remark        string
}

// 从条目文本收集图片url
func collectImgURL(text string, imgInfo *EntryImgInfo) {
	pat := regexp.MustCompile("<img .*?>")
	finds := pat.FindAllString(text, -1)
	for _, find := range finds {
		patThumbnail := regexp.MustCompile("src=\"(.*?)\"")
		matches := patThumbnail.FindAllStringSubmatch(find, -1)
		if matches == nil || len(matches) > 1 {
			imgInfo.MissMatch = true
		} else {
			imgInfo.ThumbnailURLs = append(imgInfo.ThumbnailURLs, matches[0][1])
		}
		patResource := regexp.MustCompile(("imgResource=\"(.*?)\""))
		matches = patResource.FindAllStringSubmatch(find, -1)
		if matches == nil || len(matches) > 1 {
			imgInfo.MissMatch = true
		} else {
			imgInfo.ResourceURLs = append(imgInfo.ResourceURLs, matches[0][1])
		}
	}

	imgInfo.Num = len(imgInfo.ThumbnailURLs)
	return
}

// 检查图片url
func CheckImgURL(imgInfo *EntryImgInfo) {
	urls := imgInfo.ThumbnailURLs
	urls = append(urls, imgInfo.ResourceURLs...)
	client := &http.Client{}
	for _, url := range urls {
		req, _ := http.NewRequest("HEAD", url, nil)
		req.Header.Set("Referer", imgInfo.EntryURL)
		resp, err := client.Do(req)
		if err != nil {
			imgInfo.Remark += "Head请求img失败；"
			imgInfo.NotHealth = true
			continue
		}

		v, ok := resp.Header["Content-Length"]
		if !ok {
			imgInfo.Remark += "响应头Content-Length不存在；"
			imgInfo.NotHealth = true
			continue
		}
		if len(v) == 0 {
			imgInfo.Remark += "响应头Content-Length异常；"
			imgInfo.NotHealth = true
			continue
		}
		contentlength, err := strconv.Atoi(v[0])
		if err != nil || contentlength == 0 {
			imgInfo.Remark += "img长度异常；"
			imgInfo.NotHealth = true
		}
	}
}

// 完整的图片检查功能
func TaskRun(inpath, outpath, suffix string) {
	// Load data from excel
	entries, err := head.ReadExcel(inpath, "")
	if err != nil {
		panic(err)
	}
	// columns := entries[0]
	entries = entries[1:]
	entryNum := len(entries)
	// entryNum := len(entries)

	// result data variable
	result := [][]string{}
	columns := []string{"条目名称", "版块", "学科", "条目ID", "图片数量", "图片是否正常", "备注"}
	result = append(result, columns)

	// Query to the API
	for i, row := range entries {
		word := row[1]
		queryRes := querypageSR(word)
		for _, item := range queryRes {
			entryID := item[3]
			domain := item[1]
			url := fmt.Sprintf("https://www.zgbk.com/ecph/words?SiteID=1&ID=%s&Type=%s", entryID, domain)
			entryRes := entrypageSR(url)
			text := entryRes[1] + entryRes[2]
			var imgInfo EntryImgInfo
			imgInfo.EntryURL = url
			collectImgURL(text, &imgInfo)
			CheckImgURL(&imgInfo)

			var newItem []string
			if imgInfo.NotHealth {
				newItem = append(item, strconv.Itoa(imgInfo.Num), "否", imgInfo.Remark)
			} else {
				newItem = append(item, strconv.Itoa(imgInfo.Num), "是", imgInfo.Remark)
			}
			result = append(result, newItem)
		}

		// Processbar
		wordBar := word
		for i := len([]rune(word)); i < 32; i++ {
			wordBar += "　"
		}
		fmt.Printf("\r检查进度：%v / %v \t %v", i+1, entryNum, wordBar)

		// 保存阶段性结果
		if i%50 == 0 {
			err = head.WriteExcel(outpath, result)
			if err != nil {
				panic(err)
			}
		}
	}

	// save to excel file
	err = head.WriteExcel(outpath, result)
	if err != nil {
		panic(err)
	}
}
