//package main
//
//import (
//	"github.com/prometheus/client_golang/prometheus"
//	"time"
//)
//
////
//import "net/http"
//
//import (
//	"github.com/prometheus/client_golang/prometheus"
//	"github.com/prometheus/client_golang/prometheus/promhttp"
//)
//
//var (
//	requestTotal = prometheus.NewCounter(
//		prometheus.CounterOpts{
//			Name: "myapp_requests_total",
//			Help: "Total number of requests received",
//		},
//	)
//)
//
//func init() {
//	prometheus.MustRegister(requestTotal)
//}
//
//func main() {
//	http.Handle("/metrics", promhttp.Handler())
//	http.HandleFunc("/record", recordHandler)
//	http.ListenAndServe(":8080", nil)
//}
//
//func recordHandler(w http.ResponseWriter, r *http.Request) {
//	requestTotal.Inc()
//	w.Write([]byte("Metric recorded!\n"))
//}
//
////1. 创建 Prometheus Metric 数据项
////2. 注册定义好的 Metric
////3. 业务代码中埋点
////4. 提供 HTTP API 接口

package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/special_metrics_client_library"

	"time"
)

var (
	requestTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "myapp_requests_total",
			Help: "Total number of requests received",
		},
	)
)

func init() {
	prometheus.MustRegister(requestTotal)
}

func main() {
	// 模拟应用程序逻辑
	go func() {
		for {
			// 模拟增加请求数量
			requestTotal.Inc()

			// 发送 Metrics 到特定 Metric 服务
			err := special_metrics_client_library.SendMetricsToService("12.2.3.2:9321", requestTotal)
			if err != nil {
				// 处理错误
				// log.Println("Error sending metrics:", err)
			}

			time.Sleep(5 * time.Second) // 每5秒发送一次 Metrics
		}
	}()

	// 保持应用程序运行
	select {}
}
