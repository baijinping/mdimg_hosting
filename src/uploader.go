package src

import (
	"context"
	"sync"

	"github.com/baijinping/mdimg_hosting/src/hosting"
)

func uploader(wg *sync.WaitGroup, payload <-chan *ImageElement) {
	defer wg.Done()
	for elem := range payload {
		// 这里的检查是必要的
		if elem.AlreadyHosting() || !elem.IsFinishDownload() {
			continue
		}

		Logger.Infof("上传中: %s...", elem.HostingFileName())
		u, err := hosting.Instance.Upload(context.Background(), elem.HostingFileName(), elem.rawData)
		if err != nil {
			Logger.Fatalf("上传失败 %s: %s", elem.HostingFileName(), err)
		}
		Logger.Infof("上传成功: %s : %s", elem.HostingFileName(), u.String())
		elem.HostingURL = u
	}
}

func UploadImageToHosting(elems []*ImageElement) {
	const concurrency = 5
	var (
		wg      = new(sync.WaitGroup)
		payload = make(chan *ImageElement, concurrency)
	)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go uploader(wg, payload)
	}
	for _, elem := range elems {
		payload <- elem
	}
	close(payload)
	wg.Wait()
}
