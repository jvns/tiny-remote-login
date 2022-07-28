package main

// usage: go run server.go bash
// substantial parts of this are copied from https://github.com/creack/pty
import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"unsafe"
)

type Args struct {
	Public bool
	Cmd    []string
}

func parseArgs() Args {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: ./server [--public] <command> [<arg>...]")
		os.Exit(1)
	}
	public := false
	if args[0] == "--public" {
		public = true
		args = args[1:]
	}
	return Args{
		Public: public,
		Cmd:    args,
	}
}

func main() {
	args := parseArgs()

	var addr string
	if args.Public {
		addr = ":7777"
	} else {
		addr = "127.0.0.1:7777"
	}
	sock, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Listening on", addr)

	for {
		conn, err := sock.Accept()
		if err != nil {
			panic(err)
		}
		go handle(conn, args.Cmd)
	}
}

func handle(conn net.Conn, command []string) {
	fmt.Println("Receiving connection from", conn.RemoteAddr())
	defer conn.Close()
	pty, tty, err := open()
	if err != nil {
		panic(err)
	}
	defer func() {
		pty.Close()
		tty.Close()
	}()
	Setsize(tty, &Winsize{
		Cols: 80,
		Rows: 24,
	})
	// start bash with tty as stdin/stdout/stderr, use bubblewrap for isolation

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Start()
	go func() {
		cmd.Wait()
		conn.Close()
	}()

	// copy input from tcp connection to pty and pty to tcp connection
	go func() {
		io.Copy(pty, conn)
	}()
	io.Copy(conn, pty)
}

/* create the pty and tty */

func open() (pty, tty *os.File, err error) {
	p, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	// In case of error after this point, make sure we close the ptmx fd.
	defer func() {
		if err != nil {
			_ = p.Close() // Best effort.
		}
	}()

	sname := ptsname(p)
	unlockpt(p)

	t, _ := os.OpenFile(sname, os.O_RDWR|syscall.O_NOCTTY, 0)
	return p, t, nil
}

func ioctl(fd, cmd, ptr uintptr) {
	syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
}

func ptsname(f *os.File) string {
	var n uint32
	ioctl(f.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n)))
	return "/dev/pts/" + strconv.Itoa(int(n))
}

func unlockpt(f *os.File) {
	var u int32
	// use TIOCSPTLCK with a pointer to zero to clear the lock
	ioctl(f.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&u)))
}

/* window size */

type Winsize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels
	Y    uint16 // ws_ypixel: Height in pixels
}

// Setsize resizes t to s.
func Setsize(t *os.File, ws *Winsize) {
	ioctl(t.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(ws)))
}
