package main

import (
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/baijinping/mdimg_hosting/src"
	"github.com/baijinping/mdimg_hosting/src/hosting"
	"github.com/baijinping/mdimg_hosting/src/hosting/qcloud"
)

func main() {
	filename := "content/input.md"

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		src.Logger.Fatal(err)
	}

	elems := src.FindImageElement(content)

	// 初始化图床服务
	imgHost, err := qcloud.NewCOSHostingService(
		os.Getenv("COS_BUCKETNAME"),
		os.Getenv("COS_SECRETID"),
		os.Getenv("COS_SECRETKEY"),
	)
	if err != nil {
		src.Logger.Fatal(err)
	}
	hosting.Init(imgHost)

	// 逐个标记是否已在图床
	for _, elem := range elems {
		if hosting.Instance.IsHostingURL(elem.URL) {
			elem.SetHosting()
		}
	}

	// 批量下载
	err = src.BatchDownloadResource(elems)
	if err != nil {
		src.Logger.Fatal(err)
	}

	// 上传图床
	src.UploadImageToHosting(elems)

	// 重新渲染生成output
	newContent := src.ReRenderContent(content, elems)
	err = ioutil.WriteFile("content/output.md", newContent, fs.ModeExclusive)
	if err != nil {
		src.Logger.Fatal(err)
	}
}
