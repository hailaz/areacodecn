package temp

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
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

// TestReadFile description
//
// createTime: 2023-08-23 14:31:55
//
// author: hailaz
func TestReadFile(t *testing.T) {
	dir := "out"
	codeMap := make(map[string]string)
	for i := 1980; i <= 2021; i++ {
		year := i
		fileName := strconv.Itoa(year) + ".txt"
		filePath := path.Join(dir, fileName)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
			return
		}
		// fmt.Println(string(data))
		for _, line := range strings.Split(string(data), "\n") {
			if len(line) == 0 {
				continue
			}
			// fmt.Println(line)
			code := line[:6]
			name := line[7:]
			if codeUp := code[:4] + "00"; code != codeUp {
				if nameUp, ok := codeMap[codeUp]; ok {
					name = nameUp + "/" + name
				}

				if codeUp := code[:2] + "0000"; code != codeUp {
					if nameUp, ok := codeMap[codeUp]; ok {
						name = nameUp + "/" + name
					}
				}
			}
			// fmt.Println(code, name)
			if nameOld, ok := codeMap[code]; ok {
				if nameOld != name {
					t.Logf("%d年，区域代码[%s]改名，[%s]=>[%s]", year, code, nameOld, name)
					codeMap[code] = name
				}
			} else {
				codeMap[code] = name
				if year > 1980 {
					t.Logf("%d年，新增区域代码[%s]，[%s]", year, code, name)
				}
			}
		}
		t.Logf("%d年，共%d个区域", year, len(codeMap))
	}

	file, err := os.Create(path.Join(dir, "code.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	listCode := make([]string, 0, len(codeMap))
	for code, name := range codeMap {
		listCode = append(listCode, fmt.Sprintf("%s,%s\n", code, name))
	}

	sort.Strings(listCode)
	for _, line := range listCode {
		file.WriteString(line)
	}
	fileJson, err := os.Create(path.Join(dir, "code.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer fileJson.Close()
	b, _ := json.MarshalIndent(codeMap, "", "  ")
	fileJson.Write(b)

}
