package generate

import (
	"encoding/hex"
	"github.com/zc2638/gosl/htmlTemp"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

// 根据 规定目录读取文件，只读取目录下的example.go文件、extra目录、extend拓展目录
// extra目录需要读取下面所有文件夹和文件
// extend目录下不能有go文件，必须是目录，目录下必须有一个example.go文件和一个extra目录
// 自动写入 map[string]interface{}中，内容十六进制化
/*
	map结构
	map[string]interface{}{
		"/crypto/aes": map[string]interface{}{
			"example": "内容",
			"extra": map[string]interface{}{
				"filename1": "内容",
				"filename2": "内容",
				"dir": map[string]interface{} {
					"filename3": "内容",
				},
			},
			"extend": map[string]interface{} {
				// 只允许文件夹
				// 文件夹内部结构只包含example和extra
			},
		},
	}
*/
const (
	BasePath = "../../project/standard-library/"
	OutputDataPath = "htmlTemp/data.go"
)

func Build() {

	var str string
	for _, v := range htmlTemp.PackList {

		fd, err := ioutil.ReadDir(BasePath + v)
		if err != nil {
			log.Fatal(err)
		}

		dp := new(DirParse)
		dp.FieldName = v

		for _, f := range fd {

			ps := v + "/" + f.Name()
			var s string
			if !f.IsDir() {
				if f.Name() != "example.go" {
					continue
				}
				s, err = parseData(ps)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				if f.Name() == "extra" {
					s, err = parseDir(ps)
				}
				if f.Name() == "extend" {
					s, err = parseExtend(ps)
				}
				if err != nil {
					log.Fatal(err)
				}
			}
			dp.Content = append(dp.Content, s)
		}
		str += dp.ToString()
	}
	WriteFile(OutputDataPath, str)
}

func parseExtend(p string) (string, error) {

	fd, err := ioutil.ReadDir(BasePath + p)
	if err != nil {
		return "", err
	}

	for _, f := range fd {
		if !f.IsDir() {
			log.Fatal("extend文件夹只允许目录：", p)
		}
	}
	return parseDir(p)
}

func parseData(p string) (string, error) {

	var buf []byte
	b, err := ioutil.ReadFile(BasePath + p)
	if err != nil {
		return "", err
	}

	_, file := path.Split(p)

	buf = append(buf, []byte("```go\n")...)
	buf = append(buf, b...)
	buf = append(buf, []byte("\n```\n")...)

	str := `"` + file + `": "` + hex.EncodeToString(buf) + `",` + "\n"

	return str, nil
}

func parseDir(p string) (string, error) {

	fd, err := ioutil.ReadDir(BasePath + p)
	if err != nil {
		return "", err
	}

	dp := new(DirParse)
	dp.FieldName = path.Base(p)

	for _, f := range fd {

		ps := p + "/" + f.Name()
		if !f.IsDir() {
			s, err := parseData(ps)
			if err != nil {
				log.Fatal(err)
			}
			dp.Content = append(dp.Content, s)
			continue
		}

		sp, err := parseDir(ps)
		if err != nil {
			log.Fatal(err)
		}
		dp.Content = append(dp.Content, sp)
	}
	return dp.ToString(), nil
}

type DirParse struct {
	FieldName string
	Content   []string
}

func (d DirParse) ToString() string {

	header := `"` + d.FieldName + `": `
	header += "map[string]interface{}{\n"
	return header + strings.Join(d.Content, "") + "},\n"
}

func WriteFile(filename string, content string) {

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	data := []byte("package htmlTemp\n\nvar PackDataMap = map[string]interface{}{\n")
	data = append(data, []byte(content)...)
	data = append(data, []byte("}\n")...)

	if _, err := f.Write(data); err != nil {
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
