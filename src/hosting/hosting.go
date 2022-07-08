package hosting

import (
	"context"
	"net/url"
)

// Service 图床服务
type Service interface {
	// IsExist 检查对象是否已存在
	IsExist(ctx context.Context, filename string) (isExists bool, err error)
	// Upload 上传文件, 返回对象访问URL
	Upload(ctx context.Context, filename string, raw []byte) (u *url.URL, err error)
	// IsHostingURL 图片地址是否已是图床地址
	IsHostingURL(imgUrl *url.URL) bool
}

var Instance Service

func Init(svc Service) {
	Instance = svc
}
