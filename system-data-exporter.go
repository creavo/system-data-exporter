package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type SystemData struct {
	VirtualMemoryInfo *mem.VirtualMemoryStat `json:"virtual_memory_info"`
	DiskInfo          []disk.PartitionStat   `json:"disk_info"`
	UptimeInfo        uint64                 `json:"uptime_info"`
	HostInfo          *host.InfoStat         `json:"host_info"`
	CpuInfo           []cpu.InfoStat         `json:"cpu_info"`
	Processesinfo     []*process.Process     `json:"process_info"`
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
		printToStdout(sysData)
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

	disk, err := disk.Partitions(false)
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

	sysData := SystemData{
		VirtualMemoryInfo: v,
		DiskInfo:          disk,
		UptimeInfo:        uptime,
		HostInfo:          host,
		CpuInfo:           cpu,
		Processesinfo:     processes,
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

func printToStdout(sysData SystemData) {
	fmt.Fprintf(os.Stdout, fmt.Sprintf("%v\n", sysData))
}
