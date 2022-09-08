package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/hailaz/areacodecn/data"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	statsURL = "www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm"
	maxLevel = 4
)

var aList []data.AreaCode
var mu = gmutex.New()
var DoMu = gmutex.New()
var DoneMu = gmutex.New()
var DataMapMu = gmutex.New()
var pool = grpool.New(50)
var wg = sync.WaitGroup{}
var gCurCookies []*http.Cookie
var gCurCookieJar *cookiejar.Jar
var UA = []string{
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.1 (KHTML, like Gecko) Chrome/14.0.835.163 Safari/535.1",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Opera/9.80 (Windows NT 6.1; U; zh-cn) Presto/2.9.168 Version/11.50",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 2.0.50727; SLCC2; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; InfoPath.3; .NET4.0C; Tablet PC 2.0; .NET4.0E)",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; ) AppleWebKit/534.12 (KHTML, like Gecko) Maxthon/3.0 Safari/534.12",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; InfoPath.3; .NET4.0C; .NET4.0E; SE 2.X MetaSr 1.0)",
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; GTB7.0)",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/534.3 (KHTML, like Gecko) Chrome/6.0.472.33 Safari/534.3 SE 2.X MetaSr 1.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.102 Safari/537.36 Edg/104.0.1293.70",
}

func init() {
	gCurCookies = nil
	//var err error;
	gCurCookieJar, _ = cookiejar.New(nil)

}

// ViewDataLen description
//
// createTime: 2022-09-08 10:22:16
//
// author: hailaz
func ViewDataLen() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		DataMapMu.RLockFunc(func() {
			log.Println(len(data.DataMap))
		})
	}
}

// main description
//
// createTime: 2022-08-30 12:56:57
//
// author: hailaz
func main() {
	go ViewDataLen()
	now := time.Now()
	// GetYearAreaCodeData(2021)
	RunDo()
	log.Println(time.Since(now))
}

// RunDo description
//
// createTime: 2022-09-07 23:34:26
//
// author: hailaz
func RunDo() {
	// "www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/2021/index.html": {Code: 100000000000, Level: 0},
	for k, v := range data.Do {
		GetAreaCode(path.Dir(k), path.Base(k), v.Code, v.Level)
	}
	wg.Wait()
	WriteRecord()
	WriteDataMap()
}

// GetYearAreacodeData description
//
// createTime: 2022-08-30 10:39:51
//
// author: hailaz
func GetYearAreaCodeData(year int) {
	// www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/2021/index.html
	GetAreaCode(GetYearSatasURL(year), "index.html", 100000000000, 0)
	// GetAreaCode("www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/2020/43/", "4301.html", 100000000000, 2)

	wg.Wait()
	// g.Dump(list)
	CreateDataFile(year, aList)
}

// GetYearSatasURL description
//
// createTime: 2022-08-30 10:38:00
//
// author: hailaz
func GetYearSatasURL(year int) string {
	return fmt.Sprintf("%s/%d/", statsURL, year)
}

// WriteRecord description
//
// createTime: 2022-09-07 23:18:59
//
// author: hailaz
func WriteRecord() {
	var tpl = `package data

var Do = map[string]AreaCode{
%s}
var Done = map[string]struct{}{
%s}

`

	var listDo = ""
	for k, v := range data.Do {
		listDo += fmt.Sprintf(`	"%s": {Code: %d, Level: %d},`+"\n", k, v.Code, v.Level)
	}
	var listDone = ""
	for k := range data.Done {
		listDone += fmt.Sprintf(`	"%s": {},`+"\n", k)
	}
	filePath := "data/record.go"
	err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filePath, []byte(fmt.Sprintf(tpl, listDo, listDone)), 0644)
	if err != nil {
		return
	}
}

// WriteDataMap description
//
// createTime: 2022-09-08 10:09:02
//
// author: hailaz
func WriteDataMap() {
	var tpl = `package data

var DataMap = map[string]AreaCode{
%s}

`
	var listDataMap = ""
	for k, v := range data.DataMap {
		listDataMap += fmt.Sprintf(`	"%s": {Code: %d, Name: "%s", Path: "%s", ParentCode: %d, Level: %d},`+"\n", k, v.Code, v.Name, v.Path, v.ParentCode, v.Level)
	}
	filePath := "data/data_map.go"
	err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filePath, []byte(fmt.Sprintf(tpl, listDataMap)), 0644)
	if err != nil {
		return
	}
}

// CreateDataFile description
//
// createTime: 2022-08-30 10:30:35
//
// author: hailaz
func CreateDataFile(year int, list []data.AreaCode) {
	if len(list) == 0 {
		return
	}
	var tpl = `package data

import "github.com/hailaz/areacodecn/data"

var AreaCodeList []data.AreaCode = []data.AreaCode{
%s}
`
	var listData = ""
	for _, v := range list {
		listData += fmt.Sprintf(`	{Code: %d, Name: "%s", Path: "%s", ParentCode: %d, Level: %d},`+"\n", v.Code, v.Name, v.Path, v.ParentCode, v.Level)
	}
	filePath := fmt.Sprintf("data/%d/data.go", year)
	err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filePath, []byte(fmt.Sprintf(tpl, listData)), 0644)
	if err != nil {
		return
	}
}

// GBK 转 UTF-8
func GBKToUTF8(src string, charSet string) (string, error) {
	if charSet == "GBK" {
		reader := transform.NewReader(bytes.NewReader([]byte(src)), simplifiedchinese.GBK.NewDecoder())
		d, e := ioutil.ReadAll(reader)
		if e != nil {
			return src, e
		}
		return string(d), nil
	}
	return src, nil
}

// GetDoc description
//
// createTime: 2022-08-30 13:14:04
//
// author: hailaz
func GetDoc(urlDir string, page string) (*goquery.Document, error) {
	reqUrl := path.Join(urlDir, page)
	client := &http.Client{
		Jar: gCurCookieJar,
	}

	req, err := http.NewRequest(http.MethodGet, "http://"+reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", UA[grand.N(0, len(UA)-1)])
	req.Header.Add("Host", "www.stats.gov.cn")
	req.Header.Add("Accept-Language", "zh-CN")
	req.Header.Add("Referer", reqUrl)
	// req.Header.Add("Cookie", "SF_cookie_1=15502425; wzws_cid=dc0ad41f5b219db56ca4cacfdf80c137140e2ee81c23c340f158e3f767d0b62044360d39204309e7491f6cf64eb213e14f0a4d777690756d5f5c730266e6ace4b4f530aeac69260a762932618431b7cd")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	// log.Println(resp.StatusCode, reqUrl)

	//全局保存
	// gCurCookies = gCurCookieJar.Cookies(req.URL)
	return doc, nil
}

// GetAreaCode description
//
// createTime: 2022-08-26 18:42:08
//
// author: hailaz
func GetAreaCode(urlDir string, page string, parentCode int, level int) []*data.AreaCodeTree {
	wg.Add(1)
	defer wg.Done()

	reqUrl := path.Join(urlDir, page)

	if IsDone(reqUrl) {
		// log.Println("已经处理过了", reqUrl)
		DeleteDo(reqUrl)
		return nil
	}
	// log.Println(reqUrl)
	// Load the HTML document
	doc, err := GetDoc(urlDir, page)
	if err != nil {
		log.Println(err)
		return nil
	}
	if strings.Contains(doc.Text(), "请开启JavaScript并刷新该页") {
		return nil
	} else {
		DeleteDo(reqUrl)
		AddDone(reqUrl)
	}
	// time.Sleep(time.Millisecond * 500)

	charSet := "UTF-8"

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if content, ok := s.Attr("content"); ok {
			if strings.Contains(content, "charset=gb2312") {
				// log.Println("gb2312")
				charSet = "GBK"
			}
		}
	})

	list := make([]*data.AreaCodeTree, 0)
	if level == 0 {
		doc.Find(".provincetr").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the title
			s.Find("a").Each(func(i int, s *goquery.Selection) {
				var areaCode = data.AreaCodeTree{ParentCode: parentCode, Level: level + 1}
				pathNext, ok := s.Attr("href")
				if ok {
					areaCode.Path = pathNext
				}
				areaCode.Name, _ = GBKToUTF8(s.Text(), charSet)
				code, err := strconv.Atoi(areaCode.Path[:2])
				if err == nil {
					areaCode.Code = code * 10000000000
				}
				// areaCode.Children = GetAreaCode(urlDir, areaCode.Path, areaCode.Code, level+1)
				tempUrl, tempPath, tempCode, tempLevel := urlDir, areaCode.Path, areaCode.Code, areaCode.Level
				AddDo(path.Join(tempUrl, tempPath), data.AreaCode{Code: tempCode, Level: tempLevel})
				pool.Add(context.Background(), func(ctx context.Context) {
					GetAreaCode(tempUrl, tempPath, tempCode, tempLevel)
				})

				areaCode.Path = "http://" + path.Join(path.Dir(reqUrl), areaCode.Path)
				// list = append(list, &areaCode)
				ListAppend(areaCode)
			})
		})
	} else {
		// Find the review items
		doc.Find("body > table > tbody > tr > td > table > tbody > tr > td > table > tbody > tr > td > table > tbody > tr").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the title
			var areaCode = data.AreaCodeTree{ParentCode: parentCode, Level: level + 1}
			s.Find("a").Each(func(i int, s *goquery.Selection) {
				pathNext, ok := s.Attr("href")
				if ok {
					areaCode.Path = pathNext
				}
				code, err := strconv.Atoi(s.Text())
				if err != nil {
					areaCode.Name, _ = GBKToUTF8(s.Text(), charSet)
				} else {
					areaCode.Code = code
				}
			})
			if areaCode.Name != "" {
				tempUrl, tempPath, tempCode, tempLevel := path.Dir(reqUrl), areaCode.Path, areaCode.Code, areaCode.Level
				// log.Println(tempUrl, tempPath, tempCode, tempLevel)
				AddDo(path.Join(tempUrl, tempPath), data.AreaCode{Code: tempCode, Level: tempLevel})
				if level < maxLevel {
					// areaCode.Children = GetAreaCode(urlDir, areaCode.Path, areaCode.Code, level+1)
					pool.Add(context.Background(), func(ctx context.Context) {
						GetAreaCode(tempUrl, tempPath, tempCode, tempLevel)
					})
				}
				// log.Println(level+1, urlDir, page, reqUrl, path.Dir(reqUrl), areaCode.Path)
				areaCode.Path = "http://" + path.Join(path.Dir(reqUrl), areaCode.Path)
				// list = append(list, &areaCode)
				ListAppend(areaCode)
			} else {
				s.Find(".villagetr td:nth-child(1)").Each(func(i int, s *goquery.Selection) {
					code, err := strconv.Atoi(s.Text())
					if err == nil {
						areaCode.Code = code
					}
				})
				s.Find(".villagetr td:nth-child(3)").Each(func(i int, s *goquery.Selection) {
					areaCode.Name, _ = GBKToUTF8(s.Text(), charSet)
				})
				if areaCode.Name != "" {
					// list = append(list, &areaCode)
					ListAppend(areaCode)
				}
			}

		})
	}
	return list
}

// ListAppend description
//
// createTime: 2022-08-30 13:04:25
//
// author: hailaz
func ListAppend(areaCode data.AreaCodeTree) {
	mu.LockFunc(func() {
		aList = append(aList, data.AreaCode{Code: areaCode.Code, Name: areaCode.Name, Path: areaCode.Path, ParentCode: areaCode.ParentCode, Level: areaCode.Level})
		// log.Println(areaCode.Level, len(aList))
	})
	AddDataMap(strconv.Itoa(areaCode.Code), data.AreaCode{Code: areaCode.Code, Name: areaCode.Name, Path: areaCode.Path, ParentCode: areaCode.ParentCode, Level: areaCode.Level})
}

// AddDo description
//
// createTime: 2022-09-07 23:08:21
//
// author: hailaz
func AddDo(doPath string, ac data.AreaCode) {
	DoMu.LockFunc(func() {
		data.Do[doPath] = ac
	})
}

// DeleteDo description
//
// createTime: 2022-09-07 23:08:21
//
// author: hailaz
func DeleteDo(doPath string) {
	DoMu.LockFunc(func() {
		delete(data.Do, doPath)
	})
}

// AddDone description
//
// createTime: 2022-09-07 23:08:21
//
// author: hailaz
func AddDone(doPath string) {
	DoneMu.LockFunc(func() {
		data.Done[doPath] = struct{}{}
	})
}

// DeleteDone description
//
// createTime: 2022-09-07 23:08:21
//
// author: hailaz
func DeleteDone(doPath string) {
	DoneMu.LockFunc(func() {
		delete(data.Done, doPath)
	})
}

// IsDone description
//
// createTime: 2022-09-08 09:38:34
//
// author: hailaz
func IsDone(doPath string) bool {
	DoneMu.Lock()
	defer DoneMu.Unlock()
	_, ok := data.Done[doPath]
	return ok
}

// AddDataMap description
//
// createTime: 2022-09-07 23:08:21
//
// author: hailaz
func AddDataMap(doPath string, ac data.AreaCode) {
	DataMapMu.LockFunc(func() {
		data.DataMap[doPath] = ac
	})
}

// DeleteDataMap description
//
// createTime: 2022-09-07 23:08:21
//
// author: hailaz
func DeleteDataMap(doPath string) {
	DataMapMu.LockFunc(func() {
		delete(data.DataMap, doPath)
	})
}
