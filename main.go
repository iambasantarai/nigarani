package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/mem"
)

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

	rc := http.NewResponseController(w)

	for {
		select {
		case <-clientDisconnected:
			fmt.Println("Client has been disconnected.")
		case <-memT.C:
			m, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("Unable to get mem: %s", err.Error())
				return
			}

			total := m.Total
			totalMiB := total / (1024 * 1024)
			totalGiB := total / (1024 * 1024 * 1024)

			used := m.Used
			usedMiB := used / (1024 * 1024)
			usedGiB := used / (1024 * 1024 * 1024)

			free := m.Free
			freeMiB := free / (1024 * 1024)
			freeGiB := free / (1024 * 1024 * 1024)

			if _, err := fmt.Fprintf(w, "event:mem\ndata:total:%d,totalMiB:%d,totalGiB:%d,used:%d,usedMiB:%d,usedGiB:%d,free:%d,freeMiB:%d,freeGiB:%d\n\n", total, totalMiB, totalGiB, used, usedMiB, usedGiB, free, freeMiB, freeGiB); err != nil {
				log.Printf("Unable to get mem: %s", err.Error())
				return
			}

			rc.Flush()
		}
	}
}
