package src

import (
	"flag"

	"go.uber.org/zap"
)

var (
	// ReserveSourceImageUrl 是否保留图片源地址
	ReserveSourceImageUrl bool

	// 是否保存下载图片到本地
	downloadImageFile = false
	// 图片缓存目录
	imageDownloadCacheDir = "image/"
)

var (
	Logger *zap.SugaredLogger
)

func init() {
	flag.BoolVar(&ReserveSourceImageUrl, "reserve_src_url", true, "是否保留图片源地址,避免丢失图片来源")
	flag.BoolVar(&downloadImageFile, "save_image", true, "是否保存下载的图片到本地")

	zaplog, _ := zap.NewDevelopment()
	Logger = zaplog.Sugar()
}
