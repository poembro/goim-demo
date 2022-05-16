package comet

import (
	"goim-demo/internal/comet/conf"
	"goim-demo/pkg/bytes"
	"goim-demo/pkg/time"
)

// RoundOptions round options.
type RoundOptions struct {
	Timer        int
	TimerSize    int
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Round userd for connection round-robin get a reader/writer/timer for split big lock.
type Round struct {
	readers []bytes.Pool
	writers []bytes.Pool
	timers  []time.Timer
	options RoundOptions
}

// NewRound new a round struct.
func NewRound(c *conf.Config) (r *Round) {
	var i int
	r = &Round{
		options: RoundOptions{
			Reader:       c.TCP.Reader,      //32
			ReadBuf:      c.TCP.ReadBuf,     //1024
			ReadBufSize:  c.TCP.ReadBufSize, //8192
			Writer:       c.TCP.Writer,
			WriteBuf:     c.TCP.WriteBuf,
			WriteBufSize: c.TCP.WriteBufSize,
			Timer:        c.Protocol.Timer,     //32
			TimerSize:    c.Protocol.TimerSize, //2048
		}}
	// reader (注意 bytes.Pool 在goim/pkg/bytes/buffer.go)
	r.readers = make([]bytes.Pool, r.options.Reader) //r.options.Reader 为32
	for i = 0; i < r.options.Reader; i++ {
		//为每一个结构体 进行初始化内存buffer    r.options.ReadBuf为1024  r.options.ReadBufSize为1024
		r.readers[i].Init(r.options.ReadBuf, r.options.ReadBufSize)
	}
	// writer  (注意 bytes.Pool 在goim/pkg/bytes/buffer.go)
	r.writers = make([]bytes.Pool, r.options.Writer) //r.options.Writer 为32
	for i = 0; i < r.options.Writer; i++ {
		//同上
		r.writers[i].Init(r.options.WriteBuf, r.options.WriteBufSize)
	}
	// timer   (注意 time.Timer 在goim/pkg/time/timer.go)
	r.timers = make([]time.Timer, r.options.Timer)
	for i = 0; i < r.options.Timer; i++ {
		r.timers[i].Init(r.options.TimerSize)
	}
	return
}

// Timer get a timer.
func (r *Round) Timer(rn int) *time.Timer {
	return &(r.timers[rn%r.options.Timer])
}

// Reader get a reader memory buffer.
func (r *Round) Reader(rn int) *bytes.Pool {
	return &(r.readers[rn%r.options.Reader])
}

// Writer get a writer memory buffer pool.
func (r *Round) Writer(rn int) *bytes.Pool {
	return &(r.writers[rn%r.options.Writer])
}
