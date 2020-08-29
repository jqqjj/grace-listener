package graceful

import (
	"errors"
	"flag"
	"github.com/mitchellh/go-ps"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var (
	Logger       *log.Logger
	ArgDaemonKey = "daemon"

	ForkTimeout    = time.Second * 10
	ForkTimeoutErr = errors.New("forking timeout")

	SigGraceFinish  = syscall.Signal(12)
	SigGraceRestart = syscall.SIGHUP
)

type GraceListener struct {
	net.Listener

	addr     string
	wg       *sync.WaitGroup
	exitChan chan error
}

func NewGraceListener(addr string) (*GraceListener, error) {
	l := &GraceListener{
		addr:     addr,
		wg:       new(sync.WaitGroup),
		exitChan: make(chan error),
	}
	if err := l.init(); err != nil {
		l.log(err)
		return nil, err
	}
	return l, nil
}

func (l *GraceListener) init() error {
	if ln, err := l.getListener(); err != nil {
		return err
	} else {
		l.Listener = ln
	}

	if l.isDaemonRequest() {
		/** run as a daemon **/
		if err := l.forkAndStart(); err != nil {
			return err
		}
		os.Exit(0)
	} else {
		//tell parent the child is ready
		if p, err := ps.FindProcess(os.Getppid()); err == nil {
			if p.Executable() == filepath.Base(os.Args[0]) {
				_ = l.sendSignal(os.Getppid(), SigGraceFinish) //send signal to parent
			}
		}
		go l.loopEvents()
	}

	return nil
}

func (l *GraceListener) loopEvents() {
	sig := make(chan os.Signal)
	signal.Notify(sig, SigGraceRestart)

LOOP:
	for {
		select {
		case <-sig:
			if err := l.forkAndStart(); err == nil {
				break LOOP
			} else {
				l.log(err)
			}
			_ = l.Close()
		case <-l.exitChan:
			break LOOP
		}
	}
	l.wg.Wait()
	os.Exit(0)
}

func (l *GraceListener) Accept() (net.Conn, error) {
	if c, err := l.Listener.Accept(); err != nil {
		if ne, ok := err.(net.Error); !ok || !ne.Temporary() {
			l.exitChan <- err
		}
		return nil, err
	} else {
		return NewGraceConnection(l.wg, c), nil
	}
}
func (l *GraceListener) Close() error {
	return l.Listener.Close()
}

func (l *GraceListener) getListener() (net.Listener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", l.addr)
	if err != nil {
		return nil, err
	}

	file := os.NewFile(3, "graceListener")
	ln, err := net.FileListener(file)
	if err == nil && ln.Addr().(*net.TCPAddr).Port == tcpAddr.Port {
		if tcpAddr.IP == nil || tcpAddr.IP.Equal(ln.Addr().(*net.TCPAddr).IP) {
			return ln.(*net.TCPListener), nil
		}
	}
	return net.ListenTCP("tcp", tcpAddr)
}
func (l *GraceListener) isDaemonRequest() bool {
	daemon := flag.Bool(ArgDaemonKey, false, "[option] Run as a daemon.")
	flag.Parse()
	return *daemon
}
func (l *GraceListener) forkAndStart() error {
	lnFile, err := l.Listener.(*net.TCPListener).File()
	if err != nil {
		return err
	}

	//do not exit until receive signal from child or timeout
	childErrChan := make(chan error)
	sig := make(chan os.Signal)
	signal.Notify(sig, SigGraceFinish)
	defer signal.Stop(sig)
	go func() {
		select {
		case <-sig:
			childErrChan <- nil
		case <-time.After(ForkTimeout):
			childErrChan <- ForkTimeoutErr
		}
	}()

	cmd := exec.Command(os.Args[0])
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	cmd.ExtraFiles = []*os.File{lnFile}
	if err := cmd.Start(); err != nil {
		return err
	}

	return <-childErrChan
}
func (l *GraceListener) log(err ...interface{}) {
	if Logger != nil {
		Logger.Println(err...)
	}
}
func (l *GraceListener) sendSignal(pid int, sig os.Signal) error {
	if p, err := os.FindProcess(pid); err != nil {
		return err
	} else {
		return p.Signal(sig)
	}
}
