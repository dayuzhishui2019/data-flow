package main

import (
	"crypto/tls"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model"
	_ "dyzs/data-flow/operator"
	"dyzs/data-flow/proxy"
	"dyzs/data-flow/stream"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/json-iterator/go/extra"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var TASK_FLOW = map[string][]string{
	//data
	"1400server":     []string{"1400server", "1400filter", "uploadimage", "1400tokafkamsg", "kafkaproducer"},
	"1":              []string{"1400server", "1400filter", "uploadimage", "1400tokafkamsg", "kafkaproducer"},
	"1400client":     []string{"kafkaconsumer", "kafkamsgto1400", "1400filter", "downloadimage", "1400client", "kafkaproducer"},
	"2":              []string{"kafkaconsumer", "kafkamsgto1400", "1400filter", "downloadimage", "1400client", "kafkaproducer"},
	"sendtoalicloud": []string{"kafkaconsumer", "kafkamsgto1400", "1400filter", "downloadimage", "toalioss", "toalidatahub"},
	"statistics":     []string{"kafkaconsumer", "1400digesttoredis"},
	"1400servertest": []string{"1400server", "1400filter", "uploadimage", "1400tokafkamsg"},
	//video
	"onvif": []string{"onvif"},
	"101":   []string{"onvif"},
}

func main() {
	//res,err := http.Get("https://www.sina.com")
	//if err!=nil{
	//	fmt.Println(err)
	//	return
	//}
	//body,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(body))

	//fmt.Println("___"+time.Now().UTC().Format(time.RFC3339Nano)+"___")
	//
	//resource := &model.Resource{
	//	ID:           "1232131312",
	//	GbID:         "",
	//	ParentId:     "",
	//	AreaNumber:   "",
	//	DominionCode: "",
	//	Type:         "",
	//	Func:         "",
	//	MvcIP:        "192.168.1.15",
	//	MvcPort:      "80",
	//	MvcUsername:  "admin",
	//	MvcPassword:  "abc123abc123",
	//	MvcChannels:  "",
	//	Name:         "",
	//}
	////
	//rs ,err := e_onvif.LoadResourceChannels(resource)
	//
	//fmt.Println(rs,err)

	//fmt.Println("通道个数：",rs[0].Token)

	//if len(rs)>0{
	//	//rtsp
	//	//rtsp,err := e_onvif.LoadChannelRTSP(resource,string(rs[0].Token))
	//	//fmt.Println(rtsp,err)
	//
	//	//ptz
	//err = e_onvif.ControlPTZ(resource, string(rs[0].Token), "DOWN", 0.5)
	//fmt.Println(err)
	//
	//	time.Sleep(2*time.Second)
	//	_ = e_onvif.ControlPTZ(resource,string(rs[0].Token),"STOP",0)
	//}

	//return

	context.Set("$manage_port", os.Getenv("MANAGE_PORT"))
	context.Set("$host", os.Getenv("HOST"))
	context.Set("$logLevel", os.Getenv("LOG_LEVEL"))

	logger.Init()

	context.Init()

	//context.Set("CENTER_IP", "106.13.71.247")
	//context.Set("CENTER_PORT", "18080")

	//json模糊匹配
	extra.RegisterFuzzyDecoders()

	//启动组件管理服务代理
	proxy.StartManagerProxy(context.GetString("$manage_port"))

	context.Set("$task", &model.Task{
		ID:         "test_onvif",
		AccessType: "onvif",
	})

	var currentStream *stream.Stream
	context.WatchConfig(func() {
		if currentStream != nil {
			currentStream.Close()
		}

		task, err := context.GetTask()
		if err != nil {
			logger.LOG_WARN("未定义任务")
			return
		}
		flow, ok := TASK_FLOW[task.AccessType]
		if !ok {
			logger.LOG_ERROR("未定义的taskType:", task.AccessType)
			return
		}

		myStream, err := stream.Build(flow)
		if err != nil {
			logger.LOG_WARN("流程初始化失败：", err)
			return
		}
		err = myStream.Init()
		if err != nil {
			logger.LOG_WARN("流程初始化失败：", err)
			myStream.Close()
			return
		}
		currentStream = myStream
		myStream.Run()
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case <-c:
		break
	}
}

func init1400server() {
	context.Set("$task", &model.Task{
		AccessType: "1400servertest",
	})
	context.Set("1400server_viewLibId", "11223344556677889900")
	context.Set("1400server_serverPort", "14000")
	context.Set("1400server_openAuth", false)
	context.Set("1400server_username", "")
	context.Set("1400server_password", "")
}



func test(){
	//datahub

	//account := datahub.NewAliyunAccount("tSNTJeWLf0ebHNAm", "GMubqWiPxf4Pq4LMTT39SPnkpPWmeG")
	//config := &datahub.Config{
	//	EnableBinary: false,
	//	HttpClient: &http.Client{
	//		Transport: &http.Transport{
	//			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//		},
	//	},
	//}
	//dh := datahub.NewClientWithConfig("https://101.89.99.42:443", config, account)
	//
	//topic,err := dh.GetTopic("sj_dwly_sjjr","huihai1400")
	//if err!=nil{
	//	fmt.Println(err)
	//	return
	//}
	//
	//gr,err := dh.GetCursor("sj_dwly_sjjr","huihai1400","1",datahub.LATEST)
	//if err!=nil{
	//	fmt.Println(err)
	//	return
	//}
	//ggr,err  := dh.GetTupleRecords("sj_dwly_sjjr","huihai1400","1",gr.Cursor,1,topic.RecordSchema)
	//if err!=nil{
	//	fmt.Println(err)
	//	return
	//}
	//
	//for _,record:=range ggr.Records{
	//	data,ok := record.(*datahub.TupleRecord)
	//	if ok{
	//		fmt.Println(data.Values)
	//	}else{
	//		fmt.Println("no tuple")
	//	}
	//}


	//oss
	filepath := "alioss/dayu/gat1400/b94afa60e19e4ea9942f00100fe49dbc"

	client, err := oss.New("https://101.89.99.157:443", "tSNTJeWLf0ebHNAm", "GMubqWiPxf4Pq4LMTT39SPnkpPWmeG", oss.HTTPClient(&http.Client{Transport: newTransport()}))
	if err != nil {
		fmt.Println(err)
		return
	}

	bucket, err := client.Bucket("sj-dwly-sjjr")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = bucket.GetObjectToFile(filepath,"./test.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
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
