package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type DeviceDiskUsageInfo struct {
	DeviceName string          `json:"device_name"`
	DiskUsage  *disk.UsageStat `json:"disk_usage"`
}

type SystemData struct {
	VirtualMemoryInfo *mem.VirtualMemoryStat `json:"virtual_memory_info"`
	DiskInfo          []disk.PartitionStat   `json:"disk_info"`
	UptimeInfo        uint64                 `json:"uptime_info"`
	HostInfo          *host.InfoStat         `json:"host_info"`
	CpuInfo           []cpu.InfoStat         `json:"cpu_info"`
	Processesinfo     []*process.Process     `json:"process_info"`
	DiskUsageInfo     []DeviceDiskUsageInfo  `json:"disk_usage_info"`
	CPULoadAvgInfo       *load.AvgStat          `json:"cpu_load_averge_info"`
	NetworkInterfaces []net.Interface        `json:"network_interfaces_info"`
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

	uptime, err := host.Uptime()
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

	cpu, err := cpu.Info()
	if err != nil {
		return SystemData{}, err
	}

	processes, err := process.Processes()
	if err != nil {
		return SystemData{}, err
	}

	diskUsageInfo := []DeviceDiskUsageInfo{}

	for _, device := range diskInfo {
		diskUsage, err := disk.Usage(device.Device)
		if err != nil {
			return SystemData{}, err
		}

		deviceDiskUsageInfo := DeviceDiskUsageInfo{
			DeviceName: device.Device,
			DiskUsage:  diskUsage,
		}

		diskUsageInfo = append(diskUsageInfo, deviceDiskUsageInfo)
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
		VirtualMemoryInfo: v,
		DiskInfo:          diskInfo,
		UptimeInfo:        uptime,
		HostInfo:          host,
		CpuInfo:           cpu,
		Processesinfo:     processes,
		DiskUsageInfo:     diskUsageInfo,
		CPULoadAvgInfo:       loadAvg,
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

	fmt.Println("Response Status:", response.Status)
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
