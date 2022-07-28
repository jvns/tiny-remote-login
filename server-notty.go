package main

// client: stty raw -echo && nc localhost 7777
// substantial parts of this are copied from https://github.com/creack/pty
import (
	"net"
	"os/exec"
)

func main() {
	sock, err := net.Listen("tcp", "localhost:7778")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := sock.Accept()
		if err != nil {
			panic(err)
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	tty, _ := conn.(*net.TCPConn).File()
	// start bash with tcp connection as stdin/stdout/stderr
	cmd := exec.Command("bash")
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.Start()
}
