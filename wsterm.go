package wsterm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"golang.org/x/net/websocket"
)

type Error struct {
	Code    int
	Message string
}

func NewError(err error) Error {
	return Error{-1, err.Error()}
}

var (
	ERR_OK                    = Error{0, ""}
	ERR_UNMARSHAL_DATA_FAILED = Error{9301, "unmarshal data failed"}
	ERR_PROCESS_IS_RUNNING    = Error{9302, "process is running"}
	ERR_PROCESS_IS_NOT_FOUND  = Error{9303, "process is not found"}
	ERR_UNKNOWN_MESSAGE_TYPE  = Error{9304, "unknown message type"}
	ERR_WRITE_FILE_FAILED     = Error{9305, "write file failed"}
	ERR_RECV                  = Error{9306, "recv message failed"}
)

type RawMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Output struct {
	io.Writer
	Message
}

func (o Output) Write(p []byte) (int, error) {
	o.Data = string(p)
	data, _ := json.Marshal(o.Message)
	n, err := o.Writer.Write(data)
	if err == nil {
		n = len(p)
	}
	return n, err
}

// Request Messages
//   {"type": "exec", "data": "command"}
//   {"type": "stop"}
//
// Response Messages
//   {"type": "stdout", "data": "stdout text"}
//   {"type": "stderr", "data": "stderr text"}
//   {"type": "exit",   "data": {"code": 0, "message", "error message"}}
func WSTerm(conn *websocket.Conn) {
	var reqs = make(chan RawMessage)
	var cmd *exec.Cmd
	var err error
	var name string
	var stdout, stderr *os.File
	var once sync.Once
	var mutex sync.Mutex
	var done = func(e Error) int {
		mutex.Lock()
		defer mutex.Unlock()
		once.Do(func() {
			var message string
			if err == nil {
				message = e.Message
			} else {
				message = err.Error()
			}
			websocket.JSON.Send(conn, Message{
				"exit",
				map[string]interface{}{
					"code":    e.Code,
					"message": message,
				},
			})
			conn.Close()

			if stdout != nil {
				stdout.Close()
			}
			if stderr != nil {
				stderr.Close()
			}
			if cmd != nil && cmd.Process != nil {
				cmd.Process.Signal(syscall.SIGTERM)
				cmd = nil
			}
			os.Remove(name)
			os.Remove(name + ".stdout")
			os.Remove(name + ".stderr")
			os.Remove(name + ".redirect")
		})
		return e.Code
	}
	defer done(Error{-1, "unexpected exit"})

	go func() {
		for {
			var req RawMessage
			err = websocket.JSON.Receive(conn, &req)
			if err != nil {
				done(ERR_RECV)
				close(reqs)
				return
			}
			reqs <- req
		}
	}()

	for req := range reqs {
		go func(req RawMessage) int {
			var data = req.Data
			log.Println(req.Type, string(data))
			switch req.Type {
			case "exec":
				var params string
				err = json.Unmarshal(data, &params)
				if err != nil {
					return done(ERR_UNMARSHAL_DATA_FAILED)
				}
				if cmd != nil {
					return done(ERR_PROCESS_IS_RUNNING)
				}
				file, err := ioutil.TempFile("/tmp", "termcmd")
				if err != nil {
					return done(ERR_WRITE_FILE_FAILED)
				}
				file.Write([]byte(params))
				file.Close()
				name = file.Name()
				syscall.Mkfifo(name+".stdout", 0644)
				syscall.Mkfifo(name+".stderr", 0644)
				bash := fmt.Sprintf(`
                    v=%s
                    bash $v 1>$v.stdout 2>$v.stderr
                `, name)
				redirect := name + ".redirect"
				ioutil.WriteFile(redirect, []byte(bash), 0644)
				// use expect to handle processes
				// kill all subprocesses when exit
				// There are several situations
				//   Nested subprocess
				//   Sudoed subprocess
				mutex.Lock()
				cmd = exec.Command("/usr/bin/expect")
				go func() {
					stdout, err = os.Open(name + ".stdout")
					if err != nil {
						done(NewError(err))
						return
					}
					stderr, err = os.Open(name + ".stderr")
					if err != nil {
						done(NewError(err))
						return
					}
					go io.Copy(Output{conn, Message{"stdout", ""}}, stdout)
					go io.Copy(Output{conn, Message{"stderr", ""}}, stderr)
				}()
				cmd.Stdin = bytes.NewBufferString(fmt.Sprintf(`
                    set timeout -1
                    spawn -noecho bash %s
                    expect eof
                    lassign [wait] pid spawnid os_error_flag value
                    exit $value
                `, redirect))
				err = cmd.Start()
				mutex.Unlock()
				if err != nil {
					return done(Error{-1, err.Error()})
				}
				err = cmd.Wait()
				switch e := err.(type) {
				case *exec.ExitError:
					return done(Error{e.Sys().(syscall.WaitStatus).ExitStatus(), err.Error()})
				default:
					return done(ERR_OK)
				}
			case "stop":
				mutex.Lock()
				defer mutex.Unlock()
				if cmd != nil && cmd.Process != nil {
					cmd.Process.Signal(syscall.SIGTERM)
					return -1
				} else {
					return done(ERR_PROCESS_IS_NOT_FOUND)
				}
			default:
				return done(ERR_UNKNOWN_MESSAGE_TYPE)
			}
		}(req)
	}
}

var Handler = websocket.Handler(WSTerm)
