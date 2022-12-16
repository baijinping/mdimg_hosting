package src

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// ImageElement is markdown图片元素
// ![Name](URL)
type ImageElement struct {
	// 原数据中的起止索引
	StartIdx, EndIdx int
	// 原数据中的名称和url
	Name string
	URL  *url.URL

	// 图片数据
	rawData []byte
	// 本地文件名
	localFileName string

	// 是否已是图床地址
	onHosting bool
	// 图床URL
	HostingURL *url.URL
}

func (ie *ImageElement) SetHosting() {
	ie.onHosting = true
}

func (ie *ImageElement) AlreadyHosting() bool {
	return ie.onHosting
}

func (ie *ImageElement) IsFinishDownload() bool {
	return len(ie.rawData) > 0
}

func (ie *ImageElement) HoldFile(rawData []byte) {
	ie.rawData = rawData
}

// HostingFileName 构造图床文件名
func (ie *ImageElement) HostingFileName() string {
	if ie.localFileName != "" {
		return ie.localFileName
	}

	rule := map[string]string{
		" ":  "_",
		"\t": "_",
		"/":  "-",
		"\\": "-",
	}
	markName := ie.Name
	for old, new := range rule {
		markName = strings.ReplaceAll(markName, old, new)
	}

	pathSplit := strings.Split(ie.URL.Path, "/")
	urlPathName := pathSplit[len(pathSplit)-1]

	if urlPathName == markName {
		return urlPathName
	}
	ie.localFileName = fmt.Sprintf("%s-%s", markName, urlPathName)
	return ie.localFileName
}

var reImageElement = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

// FindImageElement 在文档中解析并查找image标签
func FindImageElement(content []byte) []*ImageElement {
	matchs := reImageElement.FindAllSubmatchIndex(content, -1)
	elems := make([]*ImageElement, 0, len(matchs))
	for _, match := range matchs {
		elem := &ImageElement{
			StartIdx: match[0],
			EndIdx:   match[1],
			Name:     string(content[match[2]:match[3]]),
		}

		{ // 解析url
			urlStr := string(content[match[4]:match[5]])
			imgUrl, err := url.Parse(urlStr)
			if err != nil {
				continue
			}
			elem.URL = imgUrl
		}

		elem.HostingFileName()

		elems = append(elems, elem)
	}
	return elems
}

// ReRenderContent 渲染最终的markdown文档,主要是替换image标签的url
func ReRenderContent(content []byte, elems []*ImageElement) []byte {
	if len(elems) == 0 {
		cpy := make([]byte, 0, len(content))
		copy(cpy, content)
		return cpy
	}

	// 确保image标签按顺序排列
	sort.Slice(elems, func(i, j int) bool {
		return elems[i].StartIdx < elems[j].StartIdx
	})

	buf := bytes.Buffer{}
	buf.Grow(len(content))

	startIdx := 0
	for _, elem := range elems {
		buf.Write(content[startIdx:elem.StartIdx])
		buf.WriteString(renderImageMark(elem))
		startIdx = elem.EndIdx
	}
	buf.Write(content[startIdx:])

	return buf.Bytes()
}

func renderImageMark(elem *ImageElement) string {
	imgUrl := elem.URL.String()
	if elem.HostingURL != nil {
		imgUrl = elem.HostingURL.String()
	}
	imgUrl, _ = url.QueryUnescape(imgUrl)

	imageMark := fmt.Sprintf("![%s](%s)", elem.Name, imgUrl)

	// 决定是否保留源地址
	if ReserveSourceImageUrl && elem.IsFinishDownload() {
		return fmt.Sprintf("%s{from %s}", imageMark, elem.URL.String())
	}
	return imageMark
}
