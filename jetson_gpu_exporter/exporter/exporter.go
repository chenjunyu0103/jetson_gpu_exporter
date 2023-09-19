package exporter

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	namespace = "nvidia_jetson"
)

type Collector struct {
	sync.Mutex
	text      string
	cpuGauge  *prometheus.GaugeVec
	gr3dFreq  *prometheus.GaugeVec
	emcFreq   *prometheus.GaugeVec
	vddGauge  *prometheus.GaugeVec
	ramGauge  *prometheus.GaugeVec
	swapGauge *prometheus.GaugeVec
	iRamGauge *prometheus.GaugeVec
	tempGauge *prometheus.GaugeVec
	mtsGauge  *prometheus.GaugeVec
	gr3dGauge *prometheus.GaugeVec
}

type Exporter struct {
	Interval          int
	Path              string
	CleanFileInterval int
	Tegrastats        *Tegrastats
	Collector         *Collector
}

func (e *Exporter) InitPrometheus() {
	prometheus.MustRegister(e.Collector)
}
func NewExporter(interval int, path string, cleanFileInterval int, tegrastats *Tegrastats) *Exporter {
	return &Exporter{
		Interval:          interval,
		Path:              filepath.Clean(path),
		CleanFileInterval: cleanFileInterval,
		Tegrastats:        tegrastats,
		Collector: &Collector{
			cpuGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "cpu",
					Help:      "cpu statistics from tegrastats",
				},
				[]string{"number", "statistic"},
			),
			gr3dFreq: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "grd3_freq",
					Help:      "grd3_freq statistics from tegrastats",
				},
				[]string{"statistic"},
			),
			emcFreq: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "emc_freq",
					Help:      "emc_freq statistics from tegrastats",
				},
				[]string{"statistic"},
			),
			vddGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "vdd",
					Help:      "vdd statistics from tegrastats",
				},
				[]string{"label", "statistic"},
			),
			ramGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "ram",
					Help:      "ram statistics from tegrastats",
				},
				[]string{"statistic"},
			),
			swapGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "swap",
					Help:      "swap statistics from tegrastats",
				},
				[]string{"statistic"},
			),
			iRamGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "iram",
					Help:      "iram statistics from tegrastats",
				},
				[]string{"statistic"},
			),
			tempGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "temp",
					Help:      "temp statistics from tegrastats",
				},
				[]string{"statistic"},
			),
			mtsGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "mts",
					Help:      "mts statistics from tegrastats",
				},
				[]string{"statistic"},
			),
			gr3dGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "gr3d",
					Help:      "gr3d statistics from tegrastats",
				},
				[]string{"statistic"},
			),
		},
	}
}
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.cpuGauge.Describe(ch)
	c.gr3dFreq.Describe(ch)
	c.emcFreq.Describe(ch)
	c.ramGauge.Describe(ch)
	c.swapGauge.Describe(ch)
	c.vddGauge.Describe(ch)
	c.iRamGauge.Describe(ch)
	c.tempGauge.Describe(ch)
	c.mtsGauge.Describe(ch)
	c.gr3dGauge.Describe(ch)
}
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Only one Collect call in progress at a time.
	c.Lock()
	defer c.Unlock()
	c.cpuGauge.Reset()
	c.gr3dFreq.Reset()
	c.emcFreq.Reset()
	c.ramGauge.Reset()
	c.swapGauge.Reset()
	c.vddGauge.Reset()
	c.iRamGauge.Reset()
	c.tempGauge.Reset()
	c.mtsGauge.Reset()
	c.gr3dGauge.Reset()
	RamMap := GetRam(c.text)
	if RamMap != nil {
		c.ramGauge.WithLabelValues("used").Set(StringToFloat64(RamMap["use"]))
		c.ramGauge.WithLabelValues("tot").Set(StringToFloat64(RamMap["tot"]))
		c.ramGauge.WithLabelValues("lfb_size").Set(StringToFloat64(RamMap["size"]))
		c.ramGauge.WithLabelValues("lfb_noblock").Set(StringToFloat64(RamMap["nblock"]))
	}
	CpuMap := GetCpu(c.text)
	if CpuMap != nil {
		for key, cpuInfo := range CpuMap {
			c.cpuGauge.WithLabelValues(key, "status").Set(float64(cpuInfo.status))
			c.cpuGauge.WithLabelValues(key, "load").Set(float64(cpuInfo.load))
			c.cpuGauge.WithLabelValues(key, "Frequency (MHz)").Set(float64(cpuInfo.freq))
			c.cpuGauge.WithLabelValues(key, "governor").Set(CpuFreqPowerString(cpuInfo.governor))
		}
	}
	swapMap := GetSwap(c.text)
	if swapMap != nil {
		c.swapGauge.WithLabelValues("used").Set(StringToFloat64(swapMap["use"]))
		c.swapGauge.WithLabelValues("tot").Set(StringToFloat64(swapMap["tot"]))
		c.swapGauge.WithLabelValues("cached").Set(StringToFloat64(swapMap["cached"]))
	}
	iRamMap := GetIRam(c.text)
	if iRamMap != nil {
		c.iRamGauge.WithLabelValues("used").Set(StringToFloat64(iRamMap["use"]))
		c.iRamGauge.WithLabelValues("tot").Set(StringToFloat64(iRamMap["tot"]))
		c.iRamGauge.WithLabelValues("lfb").Set(StringToFloat64(iRamMap["lfb"]))
	}

	gr3dFreqMap := GetGr3dFreq(c.text)
	if gr3dFreqMap != nil {
		c.gr3dFreq.WithLabelValues("utilization_percentage").Set(StringToFloat64(gr3dFreqMap["use"]))
		c.gr3dFreq.WithLabelValues("frequency").Set(StringToFloat64(gr3dFreqMap["frequency"]))
	}
	emcMap := GetEmcFreq(c.text)
	if emcMap != nil {
		c.emcFreq.WithLabelValues("utilization_percentage").Set(StringToFloat64(emcMap["use"]))
		c.emcFreq.WithLabelValues("frequency").Set(StringToFloat64(emcMap["frequency"]))
	}
	vddMap := GetVdd(c.text)
	if vddMap != nil {
		for _, vddInfo := range vddMap {
			c.vddGauge.WithLabelValues(vddInfo.label, "current").Set(float64(vddInfo.current))
			c.vddGauge.WithLabelValues(vddInfo.label, "average").Set(float64(vddInfo.average))
		}
	}
	tempMap := GetTemp(c.text)
	if tempMap != nil {
		for key, value := range tempMap {
			c.tempGauge.WithLabelValues(key).Set(value)
		}
	}
	mtsMap := GetMTS(c.text)
	if mtsMap != nil {
		c.mtsGauge.WithLabelValues("fg").Set(float64(mtsMap["fg"]))
		c.mtsGauge.WithLabelValues("bg").Set(float64(mtsMap["bg"]))
	}
	gr3dMap := GetGR3D(c.text)
	if gr3dMap != nil {
		c.gr3dGauge.WithLabelValues("use").Set(float64(gr3dMap["use"]))
		c.gr3dGauge.WithLabelValues("freq").Set(float64(gr3dMap["freq"]))
	}
	c.cpuGauge.Collect(ch)
	c.gr3dFreq.Collect(ch)
	c.emcFreq.Collect(ch)
	c.vddGauge.Collect(ch)
	c.ramGauge.Collect(ch)
	c.swapGauge.Collect(ch)
	c.iRamGauge.Collect(ch)
	c.tempGauge.Collect(ch)
	c.mtsGauge.Collect(ch)
	c.gr3dGauge.Collect(ch)
}
func CpuFreqPowerString(cpuFreq string)  (ans float64) {
	/*
	performance         运行于最大频率   ---- 1

	powersave         运行于最小频率  ---- 2

	userspace         运行于用户指定的频率  ---3

	ondemand         按需快速动态调整CPU频率， 一有cpu计算量的任务，就会立即达到最大频率运行，空闲时间增加就降低频率 ---4

	conservative         按需快速动态调整CPU频率， 比 ondemand 的调整更保守 ---5

	schedutil         基于调度程序调整 CPU 频率   ---- 0
	
	smartass    聪明模式，是I和C模式的升级，该模式在比i模式不差的响应的前提下会做到了更加省电 ---6
	       流畅度： 最高，流畅

	Hotplug    类似于ondemand, 但是cpu会在关屏下尝试关掉一个cpu，并且带有deep sleep，比较省电。  ---7
	       流畅度：一般，流畅
	 */
	switch {
	case cpuFreq == "schedutil":
		return float64(0)
	case cpuFreq == "performance":
		return float64(1)
	case cpuFreq=="powersave":
		return float64(2)
	case cpuFreq=="userspace":
		return float64(3)
	case cpuFreq=="ondemand":
		return float64(4)
	case cpuFreq=="conservative":
		return float64(5)
	case cpuFreq=="smartass":
		return float64(6)
	case cpuFreq=="Hotplug":
		return float64(7)
	}
    return 
}

func (e *Exporter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.runAnalysis()
	promhttp.Handler().ServeHTTP(w, req)
}
func (e *Exporter) runAnalysis() {
	text := e.Tegrastats.Read()
	e.Collector.text = text
	log.Info("Analysis done")
}
func (e *Exporter) RunServer(addr string) {

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(ServeIndex))
	router.Handle("/metrics", e)
	server := http.Server{
		Addr:    addr,
		Handler: router,
	}
	log.Printf("Providing metrics at http://%s/metrics", addr)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()
	log.Println("start job to clean up logfile ......  ")
	cleanJob := cron.New(cron.WithSeconds())
	spec := fmt.Sprintf("0 0 */%d * * *", e.CleanFileInterval)
	_, err := cleanJob.AddFunc(spec, func() {
		e.Tegrastats.cleanUpFile()
	})
	if err != nil {
		log.Fatalf("start job to clean up logfile fail error: %s", err)
	}
	cleanJob.Start()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ... ")
	e.Tegrastats.Stop()
	cleanJob.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

// ServeIndex serves index page
func ServeIndex(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-type", "text/html")
	res := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
	<meta name="viewport" content="width=device-width">
	<title>jetson Prometheus Exporter</title>
</head>
<body>
<h1>Disk Usage Prometheus Exporter</h1>
<p>
	<a href="/metrics">Metrics</a>
</p>
<p>
	<a href="#">Homepage</a>
</p>
</body>
</html>
`
	fmt.Fprint(w, res)
}
func StringToFloat64(fValue string) float64 {
	s, _ := strconv.ParseFloat(fValue, 64)
	return s
}
