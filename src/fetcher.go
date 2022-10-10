package src

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
)

func BatchDownloadResource(imgElems []*ImageElement) error {
	_, err := os.Stat(imageDownloadCacheDir)
	dirExists := !os.IsNotExist(err)

	httpCli := newHttpClient()
	for _, elem := range imgElems {
		imgRaw := fetchImage(httpCli, elem)
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

func fetchImage(httpCli *http.Client, elem *ImageElement) []byte {
	// 已在图床的图片无需处理
	if elem.AlreadyHosting() {
		Logger.Infof("已在图床: %s(%s)", elem.Name, elem.URL)
		return nil
	}

	Logger.Infof("正在下载: %s(%s)...", elem.Name, elem.URL)
	resp, err := httpCli.Get(elem.URL.String())
	if err != nil {
		Logger.Infof("下载失败: %s(%s) err %s", elem.Name, elem.URL, err.Error())
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Logger.Infof("下载失败: %s(%s) err , 状态码 %d", elem.Name, elem.URL, resp.StatusCode)
		return nil
	}

	imgRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	Logger.Infof("下载成功: [%s] %dB", elem.Name, len(imgRaw)/8)
	return imgRaw
}

type DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

func newHttpClient() *http.Client {
	baseDialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	var dialContext DialContext
	if socks5Proxy := os.Getenv("SOCKS5_PROXY"); socks5Proxy != "" {
		var auth *proxy.Auth
		socks5User := os.Getenv("SOCKS5_USER")
		socks5Password := os.Getenv("SOCKS5_PASSWORD")
		if socks5User != "" {
			auth = &proxy.Auth{
				User:     socks5User,
				Password: socks5Password,
			}
		}

		dialSocksProxy, err := proxy.SOCKS5("tcp", socks5Proxy, auth, baseDialer)
		if err != nil {
			Logger.Error(errors.Wrap(err, "Error creating SOCKS5 proxy"))
			return http.DefaultClient
		}
		if contextDialer, ok := dialSocksProxy.(proxy.ContextDialer); ok {
			dialContext = contextDialer.DialContext
		} else {
			Logger.Error(errors.New("Failed type assertion to DialContext"))
			return http.DefaultClient
		}
		Logger.Debug("Using SOCKS5 proxy for http client:", "host", socks5Proxy)
	} else {
		dialContext = (baseDialer).DialContext
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
