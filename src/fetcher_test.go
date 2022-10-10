package src

import (
	"net/url"
	"os"
	"testing"
)

func TestSocks5Proxy(t *testing.T) {
	downloadImageFile = false
	const urlStr = `https://www.wikipedia.org/portal/wikipedia.org/assets/img/Wikipedia-logo-v2@1.5x.png`
	imgUrl, err := url.Parse(urlStr)
	if err != nil {
		t.Fatal(err.Error())
	}

	imgElem := &ImageElement{
		URL: imgUrl,
	}

	os.Setenv(`SOCKS5_PROXY`, "localhost:1080")

	err = BatchDownloadResource([]*ImageElement{imgElem})
	if err != nil {
		t.Fatal(err.Error())
	}
	if imgElem.rawData == nil {
		t.Fail()
	}
}
