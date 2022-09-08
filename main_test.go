package main

import (
	"strings"
	"testing"
	"time"
)

// TestGetYearAreaCodeData description
//
// createTime: 2022-08-26 18:35:12
//
// author: hailaz
func TestGetYearAreaCodeData(t *testing.T) {
	now := time.Now()
	GetYearAreaCodeData(2020)
	t.Log(time.Since(now))
}

// TestGetDoc description
//
// createTime: 2022-08-30 22:41:59
//
// author: hailaz
func TestGetDoc(t *testing.T) {
	doc, err := GetDoc(GetYearSatasURL(2021), "index.html")
	if err != nil {
		t.Fatal(err)
	}
	// doc.Find("script").Each(func(i int, s *goquery.Selection) {
	// 	// t.Log(s.Text())
	// 	vm := goja.New()
	// 	v, err := vm.RunString(s.Text())
	// 	if err != nil {
	// 		fmt.Println("JS代码有问题！", err)
	// 		return
	// 	}

	// 	t.Log(v.String())
	// })
	// t.Log(doc.Html())
	t.Log(strings.Contains(doc.Text(), "请开启JavaScript并刷新该页"))

}

// TestWriteDataMap description
//
// createTime: 2022-09-08 13:07:29
//
// author: hailaz
func TestWriteDataMap(t *testing.T) {
	WriteDataMap()
}
