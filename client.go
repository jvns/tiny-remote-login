package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"unsafe"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: client <host> <port>")
		os.Exit(1)
	}
	hostname := os.Args[1]
	port := os.Args[2]

	conn, _ := net.Dial("tcp", hostname+":"+port)

	fd := os.Stdin.Fd()
	oldState := MakeRaw(fd)

	go func() {
		io.Copy(conn, os.Stdin)
	}()
	io.Copy(os.Stdout, conn)
	Restore(fd, oldState)
}

func MakeRaw(fd uintptr) syscall.Termios {
	// from https://github.com/getlantern/lantern/blob/devel/archive/src/golang.org/x/crypto/ssh/terminal/util.go
	var oldState syscall.Termios
	ioctl(fd, syscall.TCGETS, uintptr(unsafe.Pointer(&oldState)))

	newState := oldState
	newState.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON | syscall.IXOFF
	newState.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	ioctl(fd, syscall.TCSETS, uintptr(unsafe.Pointer(&newState)))
	return oldState
}

func Restore(fd uintptr, oldState syscall.Termios) {
	ioctl(fd, syscall.TCSETS, uintptr(unsafe.Pointer(&oldState)))
}

func ioctl(fd, cmd, ptr uintptr) {
	syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
}
