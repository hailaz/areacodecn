package temp

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestSaveYearData description
//
// createTime: 2023-08-22 17:46:43
//
// author: hailaz
func TestSaveYearData(t *testing.T) {
	err := SaveYearData(2020)
	if err != nil {
		t.Error(err)
	}
}

// TestSaveYearDataAll description
//
// createTime: 2023-08-22 16:50:55
//
// author: hailaz
func TestSaveYearDataAll(t *testing.T) {
	for i := 1980; i <= 2021; i++ {
		err := SaveYearData(i)
		if err != nil {
			t.Error(err)
		}
		t.Log(i, "ok")
	}
}

// SaveYearData description
//
// createTime: 2023-08-22 17:40:53
//
// author: hailaz
func SaveYearData(year int) error {
	src := "src"
	out := "out"
	fileName := strconv.Itoa(year) + ".html"
	fileNameOut := strconv.Itoa(year) + ".txt"

	filePath := path.Join(src, fileName)
	filePathOut := path.Join(out, fileNameOut)
	// _, err := os.Stat(filePathOut)
	// if err == nil {
	// 	return nil
	// }
	doc, err := GetHTMLDocument(filePath)
	if err != nil {
		return err
	}
	dataOut := ""
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		str := s.Text()
		str = strings.TrimSpace(str)
		str = strings.ReplaceAll(str, "\u00a0", "")
		str = strings.ReplaceAll(str, "\n", "")
		str = strings.ReplaceAll(str, " ", "")
		// fmt.Println(code, name)
		if str != "" {
			// fmt.Println(str)
			regx := regexp.MustCompile(`(\d{6})(\W+)`)
			list := regx.FindStringSubmatch(str)
			if len(list) == 3 {
				dataOut += fmt.Sprintf("%s,%s\n", list[1], list[2])
			}
			// matched, err := regexp.MatchString("(\\d{6})(\\S+)", str)
			// fmt.Println(str, list)
		}

	})
	file, err := os.Create(filePathOut)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(dataOut)
	return nil
}

// TestNewFile description
//
// createTime: 2023-08-22 17:16:18
//
// author: hailaz
func TestNewFile(t *testing.T) {
	dir := "src"
	for i := 1980; i <= 2021; i++ {
		year := i
		fileName := strconv.Itoa(year) + ".html"
		filePath := path.Join(dir, fileName)
		_, err := os.Stat(filePath)
		if err == nil {
			continue
		}
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatal(err)
			return
		}
		defer file.Close()
	}
}
