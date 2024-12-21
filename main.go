package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/mem"
)

type MemoryInfo struct {
	Bytes uint64 `json:"bytes"`
	MiB   uint64 `json:"MiB"`
	GiB   uint64 `json:"GiB"`
}

type MemoryData struct {
	Capacity     MemoryInfo `json:"capacity"`
	Usage        MemoryInfo `json:"usage"`
	Availability MemoryInfo `json:"availability"`
}

type SysInfo struct {
	Memory MemoryData `json:"memory"`
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

	memT := time.NewTicker(time.Second)
	defer memT.Stop()

	clientDisconnected := r.Context().Done()

	for {
		select {
		case <-clientDisconnected:
			fmt.Println("Client has been disconnected.")
			return
		case <-memT.C:
			m, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("Unable to get memory stats: %s", err.Error())
				return
			}

			data := SysInfo{
				Memory: MemoryData{
					Capacity: MemoryInfo{
						Bytes: m.Total,
						MiB:   m.Total / (1024 * 1024),
						GiB:   m.Total / (1024 * 1024 * 1024),
					},
					Usage: MemoryInfo{
						Bytes: m.Used,
						MiB:   m.Used / (1024 * 1024),
						GiB:   m.Used / (1024 * 1024 * 1024),
					},
					Availability: MemoryInfo{
						Bytes: m.Free,
						MiB:   m.Free / (1024 * 1024),
						GiB:   m.Free / (1024 * 1024 * 1024),
					},
				},
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Printf("Unable to marshal memory data: %s", err.Error())
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
