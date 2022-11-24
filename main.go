package main

// out file
// go build -o main.go server.go
func main() {
	svr := NewServer("127.0.0.1", 8888)
	svr.Start()
}
