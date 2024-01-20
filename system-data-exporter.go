package main

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	v, _ := mem.VirtualMemory()

	fmt.Println(v)

	uptime, _ := host.Uptime() // uptime in seconds
	fmt.Println(uptime)

	host, _ := host.Info()
	fmt.Println(host)

	disk, _ := disk.Partitions(false)

	fmt.Println(disk)
}
