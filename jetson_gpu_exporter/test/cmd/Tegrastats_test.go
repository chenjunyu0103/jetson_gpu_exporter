package cmd

import (
	"fmt"
	"github.com/bearboy/jetson_prometheus_exporter/exporter"
	"github.com/robfig/cron/v3"
	"regexp"
	"testing"
	"time"
)

func TestGetSwap(t *testing.T) {
	var str = "RAM 1728/7763MB (lfb 1117x4MB) IRAM 1728/7763MB (lfb 1117MB) SWAP 0/3882MB (cached 0MB) CPU [5%@1190,1%@1190,off,off,off,off] EMC_FREQ 0% GR3D_FREQ 0% AO@35.5C GPU@35.5C PMIC@100C AUX@35.5C CPU@36C thermal@35.65C VDD_IN 3757/3757 VDD_CPU_GPU_CV 197/197 VDD_SOC 1066/1066 MTS fg 12% bg 13% GR3D 14%@36"
	matchRegxp := regexp.MustCompile("SWAP (\\d+)\\/(\\d+)(\\w)B( ?)\\(cached (\\d+)(\\w)B\\)")
	params := matchRegxp.FindStringSubmatch(str)
	for _, param := range params {
		fmt.Println(param)
	}
	fmt.Println(exporter.GetSwap(str))

	fmt.Println(exporter.GetRam(str))

	fmt.Println(exporter.GetCpu(str))
	fmt.Println(exporter.GetVdd(str))

	fmt.Println(exporter.GetTemp(str))
	fmt.Println(exporter.GetMTS(str))
	fmt.Println(exporter.GetGR3D(str))
	fmt.Println(exporter.GetIRam(str))
	//c := cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)), cron.WithLogger(
	//	cron.VerbosePrintfLogger(log.StandardLogger())))
	c := cron.New(cron.WithSeconds())
	i := 1
	spec := fmt.Sprintf("0 0 */%d * * *", 1)
	EntryID, err := c.AddFunc(spec, func() {
		fmt.Println(time.Now(), "每5s一次----------------", i)
		time.Sleep(time.Second * 60)
		i++
	})
	fmt.Println(time.Now(), EntryID, err)

	c.Start()
	time.Sleep(time.Second * 30)
}
