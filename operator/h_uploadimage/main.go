package h_uploadimage

import (
	"bytes"
	"dyzs/data-flow/concurrent"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/gat1400/base"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
	"dyzs/data-flow/util/base64"
	"dyzs/data-flow/util/uuid"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"
)

func init() {
	stream.RegistHandler("uploadimage", func() stream.Handler {
		return &ImageUploader{}
	})
}

type ImageUploader struct {
	executor        *concurrent.Executor
	client          *http.Client
	imageServerAddr string
}

func (iu *ImageUploader) Init(config interface{}) error {
	capacity := 20
	configCapacity := context.GetInt("uploadimage_capacity")
	if configCapacity > 0 {
		capacity = configCapacity
	}
	host := context.GetString("$host")
	if host == "" {
		host = "gofastdfs"
	}
	imageServerAddr := host + ":8080"
	logger.LOG_WARN("------------------ imagedeal config ------------------")
	logger.LOG_WARN("uploadimage_capacity : " + strconv.Itoa(capacity))
	logger.LOG_WARN("uploadimage_imageServerAddr : " + imageServerAddr)
	logger.LOG_WARN("------------------------------------------------------")
	if imageServerAddr == "" {
		return errors.New("Handle [uploadimage]:uploadimage_imageServerAddr 不能为空")
	}
	iu.imageServerAddr = imageServerAddr
	iu.client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false, //false 长链接 true 短连接
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        capacity * 5, //client对与所有host最大空闲连接数总和
			MaxConnsPerHost:     capacity,
			MaxIdleConnsPerHost: capacity,         //连接池对每个host的最大连接数量,当超出这个范围时，客户端会主动关闭到连接
			IdleConnTimeout:     60 * time.Second, //空闲连接在连接池中的超时时间
		},
		Timeout: 5 * time.Second,
	}
	iu.executor = concurrent.NewExecutor(capacity)
	return nil
}

func (iu *ImageUploader) Handle(data interface{}, next func(interface{}) error) error {
	wraps, ok := data.([]*gat1400.Gat1400Wrap)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [uploadimage] 数据格式错误，need []*daghub.StandardModelWrap , get %T", reflect.TypeOf(data)))
	}
	if len(wraps) == 0 {
		return nil
	}
	tasks := make([]func(), 0)
	var uploadErr error
	var lock sync.Mutex
	for _, wrap := range wraps {
		for _, item := range wrap.GetSubImageInfos() {
			func(img *base.SubImageInfo) {
				tasks = append(tasks, func() {
					e := iu.uploadImage(img)
					if e != nil {
						lock.Lock()
						uploadErr = e
						lock.Unlock()
					}
				})
			}(item)
		}
	}
	err := iu.executor.SubmitSyncBatch(tasks)
	if err != nil {
		logger.LOG_ERROR("上传图片失败：", err)
		return errors.New("上传图片失败：" + err.Error())
	}
	if uploadErr != nil {
		logger.LOG_ERROR("上传图片失败：", uploadErr)
		return errors.New("上传图片失败：" + uploadErr.Error())
	}
	return next(wraps)
}

func (iu *ImageUploader) uploadImage(image *base.SubImageInfo) error {
	imageData := image.Data
	if imageData == "" {
		logger.LOG_INFO("图片无base64数据")
		return nil
	}
	imageBytes, err := base64.Decode(imageData)
	if err != nil {
		logger.LOG_INFO("图片base64解码失败")
		return errors.New("图片base64解码失败")
	}
	err = util.Retry(func() error {
		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)
		fileWriter, _ := bodyWriter.CreateFormFile("file", uuid.UUIDShort())
		_, err = fileWriter.Write(imageBytes)
		if err != nil {
			return err
		}
		contentType := bodyWriter.FormDataContentType()
		err = bodyWriter.Close()
		if err != nil {
			return err
		}
		start := time.Now()
		resp, err := iu.client.Post("http://"+iu.imageServerAddr+"/upload", contentType, bodyBuffer)
		logger.LOG_WARN("upload 耗时：" + time.Since(start).String())
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		resBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return errors.New("上传图片异常:" + string(resBytes))
		}
		image.Data = string(resBytes)
		return nil
	}, 3, 100*time.Millisecond)

	if err != nil {
		logger.LOG_WARN("上传图片失败", err)
	}
	return err
}

func (iu *ImageUploader) Close() error {
	if iu.executor != nil {
		iu.executor.Close()
		iu.executor = nil
	}
	return nil
}
