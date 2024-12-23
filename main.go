package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type StorageInfo struct {
	Bytes uint64 `json:"bytes"`
	MiB   uint64 `json:"MiB"`
	GiB   uint64 `json:"GiB"`
}

type CPUInfo struct {
	ModelName string    `json:"modelName"`
	Cores     []float64 `json:"cores"`
}

type MemoryInfo struct {
	Capacity     StorageInfo `json:"capacity"`
	Usage        StorageInfo `json:"usage"`
	Availability StorageInfo `json:"availability"`
	UsedPercent  float64     `json:"usedPercent"`
}

type DiskInfo struct {
	Capacity     StorageInfo `json:"capacity"`
	Usage        StorageInfo `json:"usage"`
	Availability StorageInfo `json:"availability"`
	UsedPercent  float64     `json:"usedPercent"`
}

type SysInfo struct {
	Memory MemoryInfo `json:"memory"`
	CPU    CPUInfo    `json:"cpu"`
	Disk   DiskInfo   `json:"disk"`
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./ui")))
	http.HandleFunc("/sys-info", sysInfoHandler)

	fmt.Printf("Application started.\nLink: http://localhost:%d\n", 8000)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("ERROR: %s", err.Error())
	}
}

func roundToThreeDecimalPlaces(value float64) float64 {
	roundedValue, err := strconv.ParseFloat(fmt.Sprintf("%.3f", value), 64)
	if err != nil {
		log.Printf("Error rounding value: %s", err.Error())
	}

	return roundedValue
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

			coreUsagePercents, err := cpu.Percent(0, true)
			if err != nil {
				log.Printf("Unable to get core usage percents: %s", err.Error())
				return
			}

			formattedCoreUsagePercents := make([]float64, len(coreUsagePercents))
			for i, percent := range coreUsagePercents {
				formattedCoreUsagePercents[i] = roundToThreeDecimalPlaces(percent)
			}

			diskStat, err := disk.Usage("/")
			if err != nil {
				log.Printf("Unable to get disk stats: %s", err.Error())
				return
			}

			const MiBDivisor = 1024 * 1024
			const GiBDivisor = 1024 * 1024 * 1024

			data := SysInfo{
				Memory: MemoryInfo{
					Capacity: StorageInfo{
						Bytes: memStat.Total,
						MiB:   memStat.Total / MiBDivisor,
						GiB:   memStat.Total / GiBDivisor,
					},
					Usage: StorageInfo{
						Bytes: memStat.Used,
						MiB:   memStat.Used / MiBDivisor,
						GiB:   memStat.Used / GiBDivisor,
					},
					Availability: StorageInfo{
						Bytes: memStat.Free,
						MiB:   memStat.Free / MiBDivisor,
						GiB:   memStat.Free / GiBDivisor,
					},
					UsedPercent: roundToThreeDecimalPlaces(memStat.UsedPercent),
				},
				CPU: CPUInfo{
					ModelName: cpuStat[0].ModelName,
					Cores:     formattedCoreUsagePercents,
				},
				Disk: DiskInfo{
					Capacity: StorageInfo{
						Bytes: diskStat.Total,
						MiB:   diskStat.Total / MiBDivisor,
						GiB:   diskStat.Total / GiBDivisor,
					},
					Usage: StorageInfo{
						Bytes: diskStat.Used,
						MiB:   diskStat.Used / MiBDivisor,
						GiB:   diskStat.Used / GiBDivisor,
					},
					Availability: StorageInfo{
						Bytes: diskStat.Free,
						MiB:   diskStat.Free / MiBDivisor,
						GiB:   diskStat.Free / GiBDivisor,
					},
					UsedPercent: roundToThreeDecimalPlaces(diskStat.UsedPercent),
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
