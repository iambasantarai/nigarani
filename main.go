package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type MemoryInfo struct {
	Bytes uint64 `json:"bytes"`
	MiB   uint64 `json:"MiB"`
	GiB   uint64 `json:"GiB"`
}

type CPUInfo struct {
	ModelName string    `json:"modelName"`
	Cores     []float64 `json:"cores"`
}

type DiskInfo struct {
	Bytes uint64 `json:"bytes"`
	MiB   uint64 `json:"MiB"`
	GiB   uint64 `json:"GiB"`
}

type MemoryData struct {
	Capacity     MemoryInfo `json:"capacity"`
	Usage        MemoryInfo `json:"usage"`
	Availability MemoryInfo `json:"availability"`
}

type DiskData struct {
	Capacity     DiskInfo `json:"capacity"`
	Usage        DiskInfo `json:"usage"`
	Availability DiskInfo `json:"availability"`
}

type SysInfo struct {
	Memory MemoryData `json:"memory"`
	CPU    CPUInfo    `json:"cpu"`
	Disk   DiskData   `json:"disk"`
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./ui")))
	http.HandleFunc("/sys-info", sysInfoHandler)

	fmt.Printf("Application started.\nLink: http://localhost:%d\n", 8000)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("ERROR: %s", err.Error())
	}
}

func sysInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	tickerT := time.NewTicker(time.Second)
	defer tickerT.Stop()

	clientDisconnected := r.Context().Done()

	for {
		select {
		case <-clientDisconnected:
			fmt.Println("Client has been disconnected.")
			return
		case <-tickerT.C:
			memStat, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("Unable to get memory stats: %s", err.Error())
				return
			}

			cpuStat, err := cpu.Info()
			if err != nil {
				log.Printf("Unable to get cpu stats: %s", err.Error())
				return
			}

			usedCPUPercentage, err := cpu.Percent(0, true)
			if err != nil {
				log.Printf("Unable to get used cpu stats: %s", err.Error())
				return
			}

			diskStat, err := disk.Usage("/")
			if err != nil {
				log.Printf("Unable to get disk stats: %s", err.Error())
				return
			}

			data := SysInfo{
				Memory: MemoryData{
					Capacity: MemoryInfo{
						Bytes: memStat.Total,
						MiB:   memStat.Total / (1024 * 1024),
						GiB:   memStat.Total / (1024 * 1024 * 1024),
					},
					Usage: MemoryInfo{
						Bytes: memStat.Used,
						MiB:   memStat.Used / (1024 * 1024),
						GiB:   memStat.Used / (1024 * 1024 * 1024),
					},
					Availability: MemoryInfo{
						Bytes: memStat.Free,
						MiB:   memStat.Free / (1024 * 1024),
						GiB:   memStat.Free / (1024 * 1024 * 1024),
					},
				},
				CPU: CPUInfo{
					ModelName: cpuStat[0].ModelName,
					Cores:     usedCPUPercentage,
				},
				Disk: DiskData{
					Capacity: DiskInfo{
						Bytes: diskStat.Total,
						MiB:   diskStat.Total / (1024 * 1024),
						GiB:   diskStat.Total / (1024 * 1024 * 1024),
					},
					Usage: DiskInfo{
						Bytes: diskStat.Used,
						MiB:   diskStat.Used / (1024 * 1024),
						GiB:   diskStat.Used / (1024 * 1024 * 1024),
					},
					Availability: DiskInfo{
						Bytes: diskStat.Free,
						MiB:   diskStat.Free / (1024 * 1024),
						GiB:   diskStat.Free / (1024 * 1024 * 1024),
					},
				},
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Printf("Unable to marshal data: %s", err.Error())
				return
			}

			if _, err := fmt.Fprintf(w, "event:sysInfo\ndata:%s\n\n", jsonData); err != nil {
				log.Printf("Unable to write data: %s", err.Error())
				return
			}

			w.(http.Flusher).Flush()
		}
	}
}
