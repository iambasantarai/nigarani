# Nigarānī (निगरानी)

Nigarānī (Nepali for "monitoring") is a simple web application designed to monitor and display system performance metrics in real-time. It leverages the [gopsutil](https://github.com/shirou/gopsutil) library to gather information about the system's CPU, memory, and disk usage, and uses [Server-Sent Events (SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) to push the data to the frontend UI for live updates.
