package exporter

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
        "errors"
)
type runCmd func(cmd *exec.Cmd) error

var command runCmd

type CpuInfo struct {
	index  int
	status int
	load   int
	freq   int
        governor string
}
type VddInfo struct {
	index   int
	label   string
	current int
	average int
}
type Tegrastats struct {
	Interval int
	LogPath  string
	LogFile  string
}

//Start
// Call Tegrastats to generate the file to be monitored
func (e *Tegrastats) Start(interval int, path string) {
	getTegrastatsBin()
	e.LogFile = filepath.Join(path, "tegrastats.log")
	e.Interval = interval
	cmdStr := fmt.Sprintf("tegrastats --interval %d --logfile %s &", interval, e.LogFile)
	log.Println("exec cmd " + cmdStr)
	runCommand(cmdStr)
}

//Stop
//exec tegrastats --stop cmd
func (e *Tegrastats) Stop() {
	cmdStr := "tegrastats --stop"
	log.Println("exec cmd " + cmdStr)
	log.Println("Shutdown tegrastats Server start ")
	runCommand(cmdStr)
	log.Println("Shutdown tegrastats Server end ")
}

//Read
// read tegrastats.log
func (e *Tegrastats) Read() string {
	cmdStr := "tail -1 " + e.LogFile
	cmd := exec.Command("sh", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("combined out:%s\n", string(out))
		log.Errorf("cmd.Run() failed with %s\n", err)
		return ""
	} else {
		log.Printf("combined out:%s\n", string(out))
		return string(out)
	}
}

//cleanUpFile
// clean up tegrastats.log
func (e *Tegrastats) cleanUpFile() {
	cmdStr := "cat /dev/null > " + e.LogFile
	runCommand(cmdStr)
	log.Println("cleanUpFile end ......")
}
func runCommand(cmdStr string) {
	cmd := exec.Command("sh", "-c", cmdStr)
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() %s failed with %s\n", cmdStr, err)
	}
}

func getTegrastatsBin() bool {
	var binPaths = []string{"/usr/bin/tegrastats", "/home/nvidia/tegrastats"}
	for _, path := range binPaths {
		_, err := os.Stat(path)
		if err == nil {
			return true
		}
	}
	//panic("tegrastats not found in (/usr/bin/tegrastats,/home/nvidia/tegrastats)")
	return false
}

//GetSwap
//SWAP X/Y (cached Z)
//X = Amount of SWAP in use in megabytes.
//Y = Total amount of SWAP available for applications.
//Z = Amount of SWAP cached in megabytes.
func GetSwap(text string) map[string]string {
	var swapMap = make(map[string]string)
	matchRegxp := regexp.MustCompile("SWAP (\\d+)\\/(\\d+)(\\w)B( ?)\\(cached (\\d+)(\\w)B\\)")
	params := matchRegxp.FindStringSubmatch(text)
	if params != nil {
		swapMap["use"] = params[1]
		swapMap["tot"] = params[2]
		swapMap["unit"] = params[3]
		swapMap["cached"] = params[5] + params[6]
	}
	return swapMap
}

// GetIRam
//IRAM X/Y (lfb Z)
//IRAM is memory local to the video hardware engine.
//X = Amount of IRAM memory in use, in kilobytes.
//Y = Total amount of IRAM memory available.
//Z = Size of the largest free block.
func GetIRam(text string) map[string]string {
	var ramMap = make(map[string]string)
	matchRegxp := regexp.MustCompile("IRAM (\\d+)\\/(\\d+)(\\w)B( ?)\\(lfb (\\d+)(\\w)B\\)")
	params := matchRegxp.FindStringSubmatch(text)
	if params != nil {
		ramMap["use"] = params[1]
		ramMap["tot"] = params[2]
		ramMap["unit"] = params[3]
		ramMap["lfb"] = params[5]
	}
	return ramMap
}

//GetRam
//RAM X/Y (lfb NxZ)
//Largest Free Block (lfb) is a statistic about the memory allocator.
//It refers to the largest contiguous block of physical memory
//that can currently be allocated: at most 4 MB.
//It can become smaller with memory fragmentation.
//The physical allocations in virtual memory can be bigger.
//X = Amount of RAM in use in MB.
//Y = Total amount of RAM available for applications.
//N = The number of free blocks of this size.
//Z = is the size of the largest free block.
func GetRam(text string) map[string]string {
	var ramMap = make(map[string]string)
	matchRegxp := regexp.MustCompile("RAM (\\d+)\\/(\\d+)(\\w)B( ?)\\(lfb (\\d+)x(\\d+)(\\w)B\\)")
	params := matchRegxp.FindStringSubmatch(text)
	if params != nil {
		ramMap["use"] = params[1]
		ramMap["tot"] = params[2]
		ramMap["unit"] = params[3]
		ramMap["nblock"] = params[5]
		ramMap["size"] = params[6]
	}
	return ramMap
}

//GetCpu
//CPU [X%,Y%, , ]@Z or CPU [X%@Z, Y%@Z,...]
//X and Y are rough approximations based on time spent
//in the system idle process as reported by the Linux kernel in /proc/stat.
//X = Load statistics for each of the CPU cores relative to the
//current running frequency Z, or 'off' in case a core is currently powered down.
//Y = Load statistics for each of the CPU cores relative to the
//current running frequency Z, or 'off' in case a core is currently powered down.
//Z = CPU frequency in megahertz. Goes up or down dynamically depending on the CPU workload.

func ExecCommand(cmd string, command runCmd) (ans string) {
	//cmdAndArgs := strings.Fields("")
	//cmdAndArgs = append(cmdAndArgs, cmd)
	log.Infoln("11111111111111111111111")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmdstr := exec.Command("sh", "-c",cmd)
	log.Infoln("the cmdStr: ",cmdstr)
	cmdstr.Stdout = &stdout
	cmdstr.Stderr = &stderr
	//err := command(cmdstr)
	err := cmdstr.Run()
	log.Infoln("the 11111111----1-1111")
	//log.Infoln("the cmd.Stdout: ", stdout.String())
	if err != nil {
		exitCode := 1
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		}
		log.Infoln("%w: command filed . code :%d | command: %s | stdout: %s | stderr: %s", err, exitCode, cmd, stdout.String(), stderr.String())
		return ""
	}
	ans = stdout.String()
	log.Infoln("the ans.....",ans)
	return ans
}
func IsFile(path string) bool {
	return !IsDir(path)
}
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}
func GetCpu(text string) map[string]CpuInfo {
	var cpuMap = make(map[string]CpuInfo)
	matchRegxp := regexp.MustCompile("CPU \\[(.*)\\]")
	params := matchRegxp.FindStringSubmatch(text)
	if params != nil {
		cpus := strings.Split(params[1], ",")
		for i, value := range cpus {
			cpuInfo := new(CpuInfo)
			cpuInfo.index = i
			if value == "off" {
				cpuInfo.status = 0
			} else {
				cpuInfo.status = 1
				//Frequency
				matchFreq := regexp.MustCompile("\\b(\\d+)%@(\\d+)")
				paramsFreq := matchFreq.FindStringSubmatch(value)
				if paramsFreq != nil {
					val, _ := strconv.Atoi(paramsFreq[1])
					cpuInfo.load = val
					val, _ = strconv.Atoi(paramsFreq[2])
					cpuInfo.freq = val
				}
				governName := "/sys/devices/system/cpu/cpu" + strconv.Itoa(i) + "/cpufreq/scaling_governor"
				log.Infoln("the governName: ",governName)
				if IsFile(governName) {
					log.Infoln("cat cpu scaling_governor========")
					governCmd := "cat /sys/devices/system/cpu/cpu" + strconv.Itoa(i) + "/cpufreq/scaling_governor"
					govern :=ExecCommand(governCmd,command)
					if govern != "" {
						cpuInfo.governor = govern
					}else {
						cpuInfo.governor=""
					}
				} else {
					return cpuMap
				}
			}
			cpuMap[strconv.Itoa(i)] = *cpuInfo
		}
	}
	return cpuMap
}
func GetGr3dFreq(text string) map[string]string {
	var gr3dFreqMap = make(map[string]string)
	matchRegxp := regexp.MustCompile("GR3D_FREQ ([0-9]*)%@?([0-9]*)?")
	params := matchRegxp.FindStringSubmatch(text)
	if params != nil {
		gr3dFreqMap["use"] = params[1]
		gr3dFreqMap["frequency"] = params[2]
	}
	return gr3dFreqMap
}
func GetEmcFreq(text string) map[string]string {
	var emcMap = make(map[string]string)
	matchRegxp := regexp.MustCompile("EMC_FREQ ([0-9]*)%@?([0-9]*)?")
	params := matchRegxp.FindStringSubmatch(text)
	if params != nil {
		emcMap["use"] = params[1]
		emcMap["frequency"] = params[2]
	}
	return emcMap
}
func GetVdd(text string) map[string]VddInfo {
	var vddMap = make(map[string]VddInfo)
	matchRegxp := regexp.MustCompile("VDD_([A-Za-z0-9_]*) ([0-9]*)\\/([0-9]*)")
	params := matchRegxp.FindAllStringSubmatch(text, -1)
	if params != nil {
		for i, value := range params {
			vddInfo := VddInfo{}
			vddInfo.index = i
			vddInfo.label = value[1]
			val, _ := strconv.Atoi(value[2])
			vddInfo.current = val
			val, _ = strconv.Atoi(value[3])
			vddInfo.average = val
			vddMap[value[1]] = vddInfo
		}
	}
	return vddMap
}
func GetTemp(text string) map[string]float64 {
	var temp = make(map[string]float64)
	matchRegxp := regexp.MustCompile("\\b(\\w+)@(-?[0-9.]+)C\\b")
	params := matchRegxp.FindAllStringSubmatch(text, -1)
	if params != nil {
		for _, value := range params {
			val, _ := strconv.ParseFloat(value[2], 64)
			temp[value[1]] = val
		}
	}
	return temp
}
func GetMTS(text string) map[string]int {
	var mtsMap = make(map[string]int)
	matchRegxp := regexp.MustCompile("MTS fg (\\d+)% bg (\\d+)%")
	params := matchRegxp.FindStringSubmatch(text)
	if params != nil {
		val, _ := strconv.Atoi(params[1])
		mtsMap["fg"] = val
		val, _ = strconv.Atoi(params[2])
		mtsMap["bg"] = val
	}
	return mtsMap
}
func GetGR3D(text string) map[string]int {
	var gr3dMap = make(map[string]int)
	matchRegxp := regexp.MustCompile("GR3D ([0-9]*)%@([0-9]*)?")
	params := matchRegxp.FindStringSubmatch(text)
	if params != nil {
		val, _ := strconv.Atoi(params[1])
		gr3dMap["use"] = val
		val, _ = strconv.Atoi(params[2])
		gr3dMap["freq"] = val
	}
	return gr3dMap
}
