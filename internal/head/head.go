package head

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// Config is struct containing headers and token info
type Config struct {
	Headers   map[string]string `json:"headers"`
	BaseToken string            `json:"base_token"`
}

// ReadFile reads []byte from file
func ReadFile(p string) []byte {
	filebytes, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	return filebytes
}

// loadJson load init info for request
func LoadConfig(p string) Config {
	filebytes := ReadFile(p)
	var config Config
	err := json.Unmarshal(filebytes, &config)
	if err != nil {
		fmt.Println("failed to load configuration.")
		panic(err)
	}
	return config
}

// GetMD5 gets md5 string from input string
func GetMD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// GetPageToken gets the crypto token referring to each page
func GetPageToken(baseToken string, httpURL string) string {
	var seg string
	if strings.Index(httpURL, "?") != -1 {
		if strings.Index(httpURL, "ecph_manager") != -1 {
			seg = httpURL[strings.Index(httpURL, "ecph_manager")+len("ecph_manager") : strings.Index(httpURL, "?")]
		} else {
			seg = httpURL[strings.Index(httpURL, "/ecph")+len("/ecph") : strings.Index(httpURL, "?")]
		}
	} else {
		if strings.Index(httpURL, "ecph_manager") != -1 {
			seg = httpURL[strings.Index(httpURL, "ecph_manager")+len("ecph_manager"):]
		} else {
			seg = httpURL[strings.Index(httpURL, "/ecph")+len("/ecph"):]
		}
	}
	fullStr := baseToken + seg
	result := GetMD5(fullStr)
	return result
}

// HeadGet starts a request by GET method with headers
func HeadGet(url string, headers map[string]string, timeoutSec int) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * time.Duration(timeoutSec),
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// PostForm is 1 of 2 POST types
func PostForm(PostURL string, headers map[string]string, data url.Values, timeoutSec int) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * time.Duration(timeoutSec),
	}

	req, err := http.NewRequest("POST", PostURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	// fmt.Println(req.Header)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Serialize func serializes object by json
func UnescapeSerialize(v interface{}) string {
	buff := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buff)
	encoder.SetEscapeHTML(false)
	encoder.Encode(v)
	return string(buff.Bytes())
}

// DelTag gets the pure content without xml or html tag
func DelTag(s string) string {
	reg := regexp.MustCompile("<.*?>")
	r := reg.ReplaceAllString(s, "")
	return r
}

// ReadExcel function reads the excel file and return 2 dimension slice
func ReadExcel(path string, sheetName string) ([][]string, error) {
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	var r [][]string
	fd, err := excelize.OpenFile(path)
	if err != nil {
		fmt.Println(err.Error())
		return r, err
	}
	rows := fd.GetRows(sheetName)
	r = rows
	return r, err
}

// WriteExcelSheet function write 2 dimension slice to the 1st sheet of a new .xlsx file
func WriteExcel(path string, interf interface{}) (err error) {
	fd := excelize.NewFile()
	fd.NewSheet("Sheet1")

	data := [][]interface{}{}

	v := reflect.ValueOf(interf)
	if v.Kind() != reflect.Slice {
		err = errors.New("Wrong input data type.")
		return
	}
	for i := 0; i < v.Len(); i++ {
		itemV := v.Index(i)
		if itemV.Kind() != reflect.Slice {
			err = errors.New("Wrong input data type.")
			return
		}
		Seri := []interface{}{}
		for j := 0; j < itemV.Len(); j++ {
			cellV := itemV.Index(j)
			Seri = append(Seri, cellV)
		}
		data = append(data, Seri)
	}

	rowNum := len(data)
	for i := 0; i < rowNum; i++ {
		axis := "A" + strconv.Itoa(i+1)
		fd.SetSheetRow("Sheet1", axis, &data[i])
	}

	// Write to file
	if err = fd.SaveAs(path); err != nil {
		fmt.Println(err.Error())
		return
	}
	return err
}

// 产生一个用来区分文件命名的后缀
func TimeSuffix() string {
	now := time.Now()
	y, mth, d := now.Date()
	// year
	yStr := strconv.Itoa(y)
	// month
	mthStr := strconv.Itoa(int(mth))
	// day
	dStr := strconv.Itoa(d)
	// hour
	h := now.Hour()
	hStr := strconv.Itoa(h)
	// minute
	m := now.Minute()
	mStr := strconv.Itoa(m)
	// second
	s := now.Second()
	sStr := strconv.Itoa(s)

	suffix := yStr + mthStr + dStr + "-" + hStr + "-" + mStr + "-" + sStr
	return suffix
}
