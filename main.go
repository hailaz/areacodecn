package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/hailaz/areacodecn/data"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	statsURL = "www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm"
	maxLevel = 1
)

var aList []data.AreaCode
var mu = gmutex.New()
var pool = grpool.New(1)
var wg = sync.WaitGroup{}

// main description
//
// createTime: 2022-08-30 12:56:57
//
// author: hailaz
func main() {
	now := time.Now()
	GetYearAreaCodeData(2021)

	log.Println(time.Since(now))
}

// GetYearAreacodeData description
//
// createTime: 2022-08-30 10:39:51
//
// author: hailaz
func GetYearAreaCodeData(year int) {
	GetAreaCode(GetYearSatasURL(year), "index.html", 100000000000, 0)
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

// CreateDataFile description
//
// createTime: 2022-08-30 10:30:35
//
// author: hailaz
func CreateDataFile(year int, list []data.AreaCode) {
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
	resp, err := http.Get("http://" + reqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
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
	log.Println(reqUrl)
	// Load the HTML document
	doc, err := GetDoc(urlDir, page)
	if err != nil {
		log.Println(err)
		return nil
	}

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
				var areaCode = data.AreaCodeTree{ParentCode: parentCode, Level: level}
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
				tempUrl, tempPath, tempCode, tempLevel := urlDir, areaCode.Path, areaCode.Code, level+1
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
			var areaCode = data.AreaCodeTree{ParentCode: parentCode, Level: level}
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
				if level < maxLevel {
					// areaCode.Children = GetAreaCode(urlDir, areaCode.Path, areaCode.Code, level+1)
					tempUrl, tempPath, tempCode, tempLevel := urlDir, areaCode.Path, areaCode.Code, level+1
					pool.Add(context.Background(), func(ctx context.Context) {
						GetAreaCode(tempUrl, tempPath, tempCode, tempLevel)
					})
				}
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
		log.Println(len(aList))
	})
}
