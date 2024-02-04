package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type DeviceDiskUsageInfo struct {
	DeviceName string          `json:"device_name"`
	Mountpoint string          `json:"mountpoint"`
	DiskUsage  *disk.UsageStat `json:"disk_usage"`
}

type SystemData struct {
	GoOs              string                 `json:"go_os"`
	GoArch            string                 `json:"go_arch"`
	CpuPercent        float64                `json:"cpu_percent"`
	VirtualMemoryInfo *mem.VirtualMemoryStat `json:"virtual_memory_info"`
	DiskInfo          []disk.PartitionStat   `json:"disk_info"`
	HostInfo          *host.InfoStat         `json:"host_info"`
	CpuInfo           []cpu.InfoStat         `json:"cpu_info"`
	DiskUsageInfo     []DeviceDiskUsageInfo  `json:"disk_usage_info"`
	CPULoadAvgInfo    *load.AvgStat          `json:"cpu_load_averge_info"`
	NetworkInterfaces net.InterfaceStatList  `json:"network_interfaces_info"`
}

func main() {
	urlEndpoint := flag.String("url", "-", "URL to send data to")
	flag.Parse()

	sysData, err := initializeSystemData()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if *urlEndpoint == "-" {
		if err := printToStdout(sysData); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	} else {
		url, err := url.ParseRequestURI(*urlEndpoint)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		if err := sendToURL(sysData, url.String()); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

}

func initializeSystemData() (SystemData, error) {

	v, err := mem.VirtualMemory()
	if err != nil {
		return SystemData{}, err
	}

	host, err := host.Info()
	if err != nil {
		return SystemData{}, err
	}

	diskInfo, err := disk.Partitions(false)
	if err != nil {
		return SystemData{}, err
	}

	diskUsageInfo := []DeviceDiskUsageInfo{}

	for _, device := range diskInfo {
		if _, err := os.Stat(device.Mountpoint); err == nil {
			diskUsage, err := disk.Usage(device.Mountpoint)
			if err != nil {
				return SystemData{}, err
			}

			deviceDiskUsageInfo := DeviceDiskUsageInfo{
				DeviceName: device.Device,
				Mountpoint: device.Mountpoint,
				DiskUsage:  diskUsage,
			}

			diskUsageInfo = append(diskUsageInfo, deviceDiskUsageInfo)
		}
	}

	// calculates cpu-usage within 5 seconds
	cpuPercent, err := cpu.Percent(5000000000, false)
	if err != nil {
		return SystemData{}, err
	}

	cpu, err := cpu.Info()
	if err != nil {
		return SystemData{}, err
	}

	loadAvg, err := load.Avg()
	if err != nil {
		return SystemData{}, err
	}

	netInterfaces, err := net.Interfaces()
	if err != nil {
		return SystemData{}, err
	}

	sysData := SystemData{
		GoOs:              runtime.GOOS,
		GoArch:            runtime.GOARCH,
		CpuPercent:        cpuPercent[0],
		VirtualMemoryInfo: v,
		DiskInfo:          diskInfo,
		HostInfo:          host,
		CpuInfo:           cpu,
		DiskUsageInfo:     diskUsageInfo,
		CPULoadAvgInfo:    loadAvg,
		NetworkInterfaces: netInterfaces,
	}

	return sysData, nil
}

func sendToURL(sysData SystemData, url string) error {
	jsonBytes, err := json.Marshal(sysData)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	fmt.Println("Response-Status:", response.Status)

	b, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Response-Content:", string(b))
	return nil
}

func printToStdout(sysData SystemData) error {
	jsonBytes, err := json.Marshal(sysData)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, fmt.Sprintf("%v\n", string(jsonBytes)))
	return nil
}
