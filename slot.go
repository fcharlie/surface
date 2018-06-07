package surface

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type slotAppender struct {
	MaxFileSize int64
	path        string
	file        *os.File
	mu          sync.Mutex
	written     int64
	buf         []byte
	pid         int
}

func (a *slotAppender) open() {
	var err error
	a.file, err = os.OpenFile(a.path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		a.file = nil
	}
}

func (a *slotAppender) rotate() {
	if a.file != nil {
		a.file.Close()
		a.file = nil
	}
	dir := filepath.Dir(a.path)
	base := filepath.Base(a.path)
	ext := filepath.Ext(base)
	name := base[0 : len(base)-len(ext)]
	xdir := filepath.Join(dir, "old-log")
	if _, err := os.Stat(xdir); err != nil {
		err = os.Mkdir(xdir, 0777)
		if err != nil {
			fmt.Printf("mkdir failed %s\n", err)
		}
	}
	now := time.Now()
	nf := fmt.Sprintf("%s%c%s.%04d%02d%02d-%02d%02d-%02d%s", xdir, os.PathSeparator, name, now.Year(), now.Month(),
		now.Day(), now.Hour(),
		now.Minute(), now.Second(), ext)
	err := os.Rename(a.path, nf)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func (a *slotAppender) write(s string) {
	if a.file == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.written >= a.MaxFileSize {
		// TODO write a
		a.rotate()
		a.written = 0
	}
	a.written += int64(len(s))
	a.file.WriteString(s)
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// formatHeader writes log header to buf in following order:
func (a *slotAppender) formatHeader(buf *[]byte, t time.Time, prefix string, pid int) {
	*buf = append(*buf, '[')
	if len(prefix) != 0 {
		*buf = append(*buf, prefix...)
		*buf = append(*buf, "] ["...)
	}
	itoa(buf, pid, -1)
	*buf = append(*buf, "] "...)
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '-')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '-')
	itoa(buf, day, 2)
	*buf = append(*buf, ' ')
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	*buf = append(*buf, ' ')
}

func (a *slotAppender) writev(prefix string, s string) error {
	if a.file == nil {
		return nil
	}
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buf = a.buf[:0]
	a.formatHeader(&a.buf, now, prefix, a.pid)
	a.buf = append(a.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		a.buf = append(a.buf, '\n')
	}
	if a.written >= a.MaxFileSize {
		// TODO write a
		a.rotate()
		a.written = 0
	}
	a.written += int64(len(a.buf))
	_, err := a.file.Write(a.buf)
	return err
}

func (a *slotAppender) changeSize(sz int64) {
	a.mu.Lock()
	a.MaxFileSize = sz
	a.mu.Unlock()
}

func newSlotAppender(f string) *slotAppender {
	a := &slotAppender{
		MaxFileSize: 104857600,
		written:     0,
		path:        f,
		pid:         os.Getpid(),
	}
	if s, err := os.Stat(f); err == nil {
		if s.Size() >= a.MaxFileSize {
			a.rotate()
			return a
		}
	}
	a.open()
	return a
}

// Slot todo
type Slot struct {
	log *slotAppender
	bus *slotAppender
}

// RolateSize to
func (l *Slot) RolateSize(sz int64) {
	if l.log != nil {
		l.log.changeSize(sz)
	}
	if l.bus != nil {
		l.bus.changeSize(sz)
	}
}

// Output logger
func (l *Slot) Output(prefix string, s string) error {
	//now := time.Now()
	if l.log != nil {
		return l.log.writev(prefix, s)
	}
	return nil
}

// DEBUG logger out
func (l *Slot) DEBUG(format string, v ...interface{}) {
	l.Output("DEBUG", fmt.Sprintf(format, v...))
}

// INFO logger out
func (l *Slot) INFO(format string, v ...interface{}) {
	l.Output("INFO", fmt.Sprintf(format, v...))
}

// ERROR logger out
func (l *Slot) ERROR(format string, v ...interface{}) {
	l.Output("ERROR", fmt.Sprintf(format, v...))
}

// FATAL logger out
func (l *Slot) FATAL(format string, v ...interface{}) {
	l.Output("FATAL", fmt.Sprintf(format, v...))
}

// Access logger out
func (l *Slot) Access(format string, v ...interface{}) {
	if l.bus != nil {
		l.bus.writev("", fmt.Sprintf(format, v...))
	}
}

// Initialize TODO
func (l *Slot) Initialize(af string, ef string) error {
	l.log = newSlotAppender(ef)
	l.bus = newSlotAppender(af)
	return nil
}
