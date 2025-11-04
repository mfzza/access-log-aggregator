package main

// func NewAccessLog(time time.Time, host string, statusCode int, duration float64) *accessLog {
// 	return &accessLog{time, host, statusCode, duration}
// }

func main() {
	// TODO: implement multiple file read with goroutine
	// TODO: behaviour to default start read from tail (end of file), and read from beginning when have `-from-start` flag
	// TODO: tolerate common log rotation
	// TODO: -interval flag
	// TODO: tail -F like behaviour
	// TODO: handle graceful exit

	cfg := parseFlags()
	ss := processFiles(cfg.Files)
	ss.Print()
}
