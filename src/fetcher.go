package src

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
)

func BatchDownloadResource(imgElems []*ImageElement) error {
	_, err := os.Stat(imageDownloadCacheDir)
	dirExists := !os.IsNotExist(err)

	for _, elem := range imgElems {
		imgRaw := fetchImage(elem)
		if imgRaw == nil {
			continue
		}

		elem.HoldFile(imgRaw)

		if downloadImageFile { // 下载图片到本地
			if !dirExists {
				os.Mkdir(imageDownloadCacheDir, fs.ModeDir)
			}

			fullName := fmt.Sprintf("%s%s", imageDownloadCacheDir, elem.HostingFileName())
			file, err := os.Create(fullName)
			if err != nil {
				Logger.Errorf("创建文件失败: %v", err)
				continue
			}
			_, err = file.Write(imgRaw)
			if err != nil {
				Logger.Errorf("写入文件失败: %v", err)
				os.Remove(fullName)
				continue
			}
			file.Close()
		}
	}
	return nil
}

func fetchImage(elem *ImageElement) []byte {
	// 已在图床的图片无需处理
	if elem.AlreadyHosting() {
		Logger.Infof("已在图床: %s(%s)", elem.Name, elem.URL)
		return nil
	}

	Logger.Infof("正在下载: %s(%s)...", elem.Name, elem.URL)
	resp, err := http.Get(elem.URL.String())
	if err != nil {
		Logger.Infof("下载失败: %s(%s) err %s", elem.Name, elem.URL, err.Error())
		return nil
	}
	defer resp.Body.Close()

	imgRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	Logger.Infof("下载成功: [%s] %dB", elem.Name, len(imgRaw)/8)
	return imgRaw
}
