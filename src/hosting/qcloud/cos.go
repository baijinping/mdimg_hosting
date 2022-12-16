package qcloud

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"
	"go.uber.org/zap"

	"github.com/baijinping/mdimg_hosting/src/hosting/internal"
)

// COSHostingService 腾讯云对象存储图床服务
// Go SDK文档 https://cloud.tencent.com/document/product/436/31215
type COSHostingService struct {
	client *cos.Client

	// bucket公共读url
	bucketUrl *url.URL
}

func NewCOSHostingService(bucketName, secretId, secretKey string) (*COSHostingService, error) {
	bucketUrl := fmt.Sprintf("https://%s.cos.ap-nanjing.myqcloud.com", bucketName)
	u, err := url.Parse(bucketUrl)
	if err != nil {
		return nil, err
	}
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		},
	})
	return &COSHostingService{
		client:    client,
		bucketUrl: u,
	}, nil
}

func (svc *COSHostingService) IsExist(ctx context.Context, filename string) (isExists bool, err error) {
	ok, err := svc.client.Object.IsExist(ctx, filename)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func calMD5Digest(msg []byte) []byte {
	m := md5.New()
	m.Write(msg)
	return m.Sum(nil)
}

func (svc *COSHostingService) isRepeat(ctx context.Context, key string, raw []byte) (bool, error) {
	resp, err := svc.client.Object.Head(ctx, key, nil)
	if err != nil {
		return false, err
	}

	bs1 := calMD5Digest(raw)
	md5str := fmt.Sprintf(`"%x"`, bs1)
	if resp.Header.Get("ETag") != md5str {
		return false, nil
	}
	return true, nil
}

func (svc *COSHostingService) Upload(ctx context.Context, filename string, raw []byte) (u *url.URL, err error) {
	isExists, err := svc.IsExist(ctx, filename)
	if err != nil {
		return nil, err
	}
	if isExists {
		isRepeat, err := svc.isRepeat(ctx, filename, raw)
		if err != nil {
			return nil, err
		}
		if isRepeat {
			Logger.Infof("检测为重复文件: %s  (size=%dB)", filename, len(raw))
			return svc.getObjectURL(filename), nil
		}

		// 存在同名且不相同文件,需重新命名文件
		filename = internal.GenNewName(filename)
	}

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: "image/jpeg",
		},
	}
	_, err = svc.client.Object.Put(ctx, filename, bytes.NewReader(raw), opt)
	if err != nil {
		panic(err)
	}
	return svc.getObjectURL(filename), nil
}

func (svc *COSHostingService) getObjectURL(key string) *url.URL {
	return svc.client.Object.GetObjectURL(key)
}

func (svc *COSHostingService) IsHostingURL(imgUrl *url.URL) bool {
	return imgUrl.Host == svc.bucketUrl.Host
}

var (
	Logger *zap.SugaredLogger
)

func init() {
	zaplog, _ := zap.NewDevelopment()
	Logger = zaplog.Sugar()
}
