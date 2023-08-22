package temp

import (
	"log"
	"os"

	"github.com/PuerkitoBio/goquery"
)

// GetHTMLDocument description
//
// createTime: 2023-08-22 16:49:34
//
// author: hailaz
func GetHTMLDocument(filePath string) (*goquery.Document, error) {
	// 打开HTML文件
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 使用goquery从文件中解析HTML
	return goquery.NewDocumentFromReader(file)
}
