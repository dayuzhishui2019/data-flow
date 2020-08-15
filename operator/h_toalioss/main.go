package h_toalioss

import (
	"bytes"
	"crypto/tls"
	"dyzs/data-flow/concurrent"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
	"dyzs/data-flow/util/aes"
	"dyzs/data-flow/util/uuid"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

func init() {
	stream.RegistHandler("toalioss", func() stream.Handler {
		return &OssUploader{}
	})
}

type OssUploader struct {
	accessId   string
	accessKey  string
	endpoint   string
	bucketName string

	aesKey []byte

	ossClient *oss.Client
	bucket    *oss.Bucket

	executor        *concurrent.Executor
	client          *http.Client
	imageServerAddr string
}

func (h *OssUploader) Init(config interface{}) error {
	capacity := 20
	logger.LOG_WARN("------------------ toalioss config ------------------")
	logger.LOG_WARN("toalioss_accessId", context.GetString("toalioss_accessId"))
	logger.LOG_WARN("toalioss_accessKey", context.GetString("toalioss_accessKey"))
	logger.LOG_WARN("toalioss_endpoint", context.GetString("toalioss_endpoint"))
	logger.LOG_WARN("toalioss_bucketName", context.GetString("toalioss_bucketName"))
	logger.LOG_WARN("toalioss_aeskey", context.GetString("toalioss_aeskey"))
	logger.LOG_WARN("------------------------------------------------------")
	h.accessId = context.GetString("toalioss_accessId")
	h.accessKey = context.GetString("toalioss_accessKey")
	h.endpoint = context.GetString("toalioss_endpoint")
	h.bucketName = context.GetString("toalioss_bucketName")
	h.aesKey = []byte(context.GetString("toalioss_aeskey"))
	h.executor = concurrent.NewExecutor(capacity)

	return h.initOssClient()
}

func (iu *OssUploader) initOssClient() error {
	client, err := oss.New(iu.endpoint, iu.accessId, iu.accessKey, oss.HTTPClient(&http.Client{Transport: newTransport()}))
	if err != nil {
		return err
	}
	iu.ossClient = client
	//bucket
	// 创建存储空间。
	ok, err := client.IsBucketExist(iu.bucketName)
	if err != nil {
		return err
	}
	if !ok {
		err = client.CreateBucket(iu.bucketName, oss.ACL(oss.ACLPublicRead))
	}
	if err != nil {
		return err
	}
	iu.bucket, err = client.Bucket(iu.bucketName)
	if err != nil {
		return err
	}
	logger.LOG_WARN("oss 初始化完成，bucket:", iu.bucketName)
	return nil
}

func (iu *OssUploader) Handle(data interface{}, next func(interface{}) error) error {
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
	paths := make([][]string, 0)

	for _, wrap := range wraps {
		tasks = append(tasks, func() {
			path, e := iu.uploadOss(wrap)
			if e != nil {
				uploadErr = e
			} else {
				lock.Lock()
				paths = append(paths, path)
				lock.Unlock()
			}
		})
	}
	err := iu.executor.SubmitSyncBatch(tasks)
	if err != nil {
		logger.LOG_ERROR("上传文件oss失败：", err)
		return errors.New("上传文件oss失败：" + err.Error())
	}
	if uploadErr != nil {
		logger.LOG_ERROR("上传文件oss失败：", uploadErr)
		return errors.New("上传文件oss失败：" + uploadErr.Error())
	}
	return next(paths)
}

func (iu *OssUploader) uploadOss(wrap *gat1400.Gat1400Wrap) (res []string, err error) {
	imageBytes, err := wrap.BuildToJson()
	if err != nil {
		logger.LOG_INFO("数据json序列化失败", err)
		return nil, errors.New("数据json序列化失败")
	}
	//aes加密
	path := "alioss/huihai/gat1400/" + strings.ReplaceAll(uuid.UUID(), "-", "")
	if len(iu.aesKey) > 0 {
		logger.LOG_INFO("oss-aes加密前长度：", len(imageBytes))
		imageBytes, err = aes.EncryptAES(imageBytes, iu.aesKey)
		if err != nil {
			return nil, errors.New("oss-aes加密失败，" + err.Error() + ",aes-key:" + string(iu.aesKey))
		}
		logger.LOG_INFO("oss-aes加密后长度：", len(imageBytes))
	}

	err = util.Retry(func() error {
		start := time.Now()
		err := iu.bucket.PutObject(path, bytes.NewReader(imageBytes))
		logger.LOG_WARN("oss upload 耗时：" + time.Since(start).String())
		if err != nil {
			return err
		}
		logger.LOG_INFO("oss upload success：", path)
		return nil
	}, 3, 100*time.Millisecond)

	if err != nil {
		logger.LOG_WARN("上传文件失败", err)
		return nil, err
	}
	return []string{wrap.DataType, path}, err
}

func (iu *OssUploader) Close() error {
	if iu.executor != nil {
		iu.executor.Close()
		iu.executor = nil
	}
	return nil
}

func newTransport() *http.Transport {
	// New Transport
	transport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			d := net.Dialer{
				Timeout:   time.Second * 30,
				KeepAlive: 30 * time.Second,
			}
			conn, err := d.Dial(netw, addr)
			if err != nil {
				return nil, err
			}
			return newTimeoutConn(conn, time.Second*60, time.Second*300), nil
		},
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       time.Second * 50,
		ResponseHeaderTimeout: time.Second * 60,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	return transport
}

// timeoutConn handles HTTP timeout
type timeoutConn struct {
	conn        net.Conn
	timeout     time.Duration
	longTimeout time.Duration
}

func newTimeoutConn(conn net.Conn, timeout time.Duration, longTimeout time.Duration) *timeoutConn {
	conn.SetReadDeadline(time.Now().Add(longTimeout))
	return &timeoutConn{
		conn:        conn,
		timeout:     timeout,
		longTimeout: longTimeout,
	}
}

func (c *timeoutConn) Read(b []byte) (n int, err error) {
	c.SetReadDeadline(time.Now().Add(c.timeout))
	n, err = c.conn.Read(b)
	c.SetReadDeadline(time.Now().Add(c.longTimeout))
	return n, err
}

func (c *timeoutConn) Write(b []byte) (n int, err error) {
	c.SetWriteDeadline(time.Now().Add(c.timeout))
	n, err = c.conn.Write(b)
	c.SetReadDeadline(time.Now().Add(c.longTimeout))
	return n, err
}

func (c *timeoutConn) Close() error {
	return c.conn.Close()
}

func (c *timeoutConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *timeoutConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *timeoutConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *timeoutConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *timeoutConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
