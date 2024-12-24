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
	"github.com/shirou/gopsutil/process"
)

type StorageInfo struct {
	Bytes uint64 `json:"bytes"`
	KiB   uint64 `json:"KiB"`
	MiB   uint64 `json:"MiB"`
	GiB   uint64 `json:"GiB"`
}

type CPUInfo struct {
	ModelName   string    `json:"modelName"`
	Cores       []float64 `json:"cores"`
	UsedPercent float64   `json:"usedPercent"`
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

type ProcessInfo struct {
	PID         int32       `json:"pid"`
	Name        string      `json:"name"`
	CPUPercent  float64     `json:"cpuPercent"`
	MemoryUsage StorageInfo `json:"memoryUsage"`
}

type SysInfo struct {
	Memory    MemoryInfo    `json:"memory"`
	CPU       CPUInfo       `json:"cpu"`
	Disk      DiskInfo      `json:"disk"`
	Processes []ProcessInfo `json:"processes"`
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

func calculateAverageUsagePercent(usagePercents []float64) float64 {
	if len(usagePercents) == 0 {
		return 0.0
	}

	var total float64
	for _, percent := range usagePercents {
		total += percent
	}

	return roundToThreeDecimalPlaces(total / float64(len(usagePercents)))
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
			const KiBDivisor = 1024
			const MiBDivisor = 1024 * 1024
			const GiBDivisor = 1024 * 1024 * 1024

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

			var processInfos []ProcessInfo
			procs, err := process.Processes()
			if err != nil {
				log.Printf("Unable to get procs stats: %s", err.Error())
				return
			}

			for _, proc := range procs {
				name, err := proc.Name()
				if err != nil {
					log.Printf("Unable to get proc name: %s", err.Error())
				}
				cpuPercent, err := proc.CPUPercent()
				if err != nil {
					log.Printf("Unable to get cpu percent: %s", err.Error())
				}
				memoryUsage, err := proc.MemoryInfo()
				if err != nil {
					log.Printf("Unable to get memory info: %s", err.Error())
				}

				processInfos = append(processInfos, ProcessInfo{
					PID:        proc.Pid,
					Name:       name,
					CPUPercent: roundToThreeDecimalPlaces(cpuPercent),
					MemoryUsage: StorageInfo{
						Bytes: memoryUsage.RSS,
						KiB:   memoryUsage.RSS / KiBDivisor,
						MiB:   memoryUsage.RSS / MiBDivisor,
						GiB:   memoryUsage.RSS / GiBDivisor,
					},
				})
			}

			data := SysInfo{
				Memory: MemoryInfo{
					Capacity: StorageInfo{
						Bytes: memStat.Total,
						KiB:   memStat.Total / KiBDivisor,
						MiB:   memStat.Total / MiBDivisor,
						GiB:   memStat.Total / GiBDivisor,
					},
					Usage: StorageInfo{
						Bytes: memStat.Used,
						KiB:   memStat.Used / KiBDivisor,
						MiB:   memStat.Used / MiBDivisor,
						GiB:   memStat.Used / GiBDivisor,
					},
					Availability: StorageInfo{
						Bytes: memStat.Free,
						KiB:   memStat.Free / KiBDivisor,
						MiB:   memStat.Free / MiBDivisor,
						GiB:   memStat.Free / GiBDivisor,
					},
					UsedPercent: roundToThreeDecimalPlaces(memStat.UsedPercent),
				},
				CPU: CPUInfo{
					ModelName:   cpuStat[0].ModelName,
					Cores:       formattedCoreUsagePercents,
					UsedPercent: calculateAverageUsagePercent(formattedCoreUsagePercents),
				},
				Disk: DiskInfo{
					Capacity: StorageInfo{
						Bytes: diskStat.Total,
						KiB:   diskStat.Total / KiBDivisor,
						MiB:   diskStat.Total / MiBDivisor,
						GiB:   diskStat.Total / GiBDivisor,
					},
					Usage: StorageInfo{
						Bytes: diskStat.Used,
						KiB:   diskStat.Used / KiBDivisor,
						MiB:   diskStat.Used / MiBDivisor,
						GiB:   diskStat.Used / GiBDivisor,
					},
					Availability: StorageInfo{
						Bytes: diskStat.Free,
						KiB:   diskStat.Free / KiBDivisor,
						MiB:   diskStat.Free / MiBDivisor,
						GiB:   diskStat.Free / GiBDivisor,
					},
					UsedPercent: roundToThreeDecimalPlaces(diskStat.UsedPercent),
				},
				Processes: processInfos,
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
