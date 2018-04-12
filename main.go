package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/morrah77/go-challenge-007/proc"
)

type Conf struct {
	KeyTtl     time.Duration
	ListenAddr string
}

var conf Conf

func init() {
	flag.DurationVar(&conf.KeyTtl, `key-ttl`, time.Second*10, `Storage keys TTL in nanoseconds`)
	flag.StringVar(&conf.ListenAddr, `listen-addr`, `localhost:12345`, `Address to listen`)
	flag.Parse()
}

func main() {
	p := proc.NewChannelProcessor(conf.KeyTtl)
	println(`Created storage with TTL`, conf.KeyTtl, `ns`)
	p.Start()
	defer p.Stop()
	ls, err := net.Listen(`tcp`, conf.ListenAddr)
	if err != nil {
		println(err.Error())
		return

	}
	println(`Listening on`, conf.ListenAddr)
	defer ls.Close()
	for {
		conn, err := ls.Accept()
		if err != nil {
			println(err.Error())
			continue

		}
		println(`Accepted connection from`, conn.RemoteAddr().String())
		go processConnection(conn, p)
	}
}

func processConnection(conn net.Conn, proc *proc.ChannelProcessor) {
	netReader := bufio.NewReader(conn)
	for {
		b, err := netReader.ReadBytes('\n')
		if err != nil {
			conn.Write([]byte(err.Error() + "\n"))
			conn.Close()
			break
		}
		cmd, err := parseCommandJson(b)
		if err != nil {
			cmd, err = parseCommandString(b)
		}
		if err != nil {
			conn.Write([]byte(err.Error() + "\n"))
			conn.Close()
			break
		}
		println(`Process command`, cmd.Command, cmd.Key, cmd.Value, `from`, conn.RemoteAddr().String())
		switch cmd.Command {
		case `Close`:
			println(`Close`)
			conn.Close()
		case `Create`:
			println(`Create`)
			err := proc.Create(cmd.Key, cmd.Value)
			if err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				conn.Write([]byte("OK\n"))
			}
		case `Update`:
			println(`Update`)
			err := proc.Update(cmd.Key, cmd.Value)
			if err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				conn.Write([]byte("OK\n"))
			}
		case `TTL`:
			println(`TTL`)
			err := proc.SetTtl(cmd.Key, cmd.Value)
			if err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				conn.Write([]byte("OK\n"))
			}
		case `Remove`:
			println(`Remove`)
			err := proc.Remove(cmd.Key)
			if err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				conn.Write([]byte("OK\n"))
			}
		case `Get`:
			println(`Get`)
			value, err := proc.Get(cmd.Key)
			if err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				resp := fmt.Sprintf("%#v\n", value)
				println(resp)
				conn.Write([]byte(resp))
			}
		case `List`:
			println(`List`)
			value, err := proc.KeyList()
			fmt.Printf("List main value %#v\n", value)
			if err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				resp := fmt.Sprintf("%#v\n", value)
				println(resp)
				conn.Write([]byte(resp))
			}
		default:
			println(`Unknown command`, cmd.Command)
			conn.Write([]byte("Unknown command " + cmd.Command + "\n"))
		}

	}
}

type Command struct {
	Command string `json:"command"`
	Key     string `json:"key,omitempty"`
	Value   string `json:"value,omitempty"`
}

func parseCommandJson(b []byte) (cmd *Command, err error) {
	cmd = &Command{}
	err = json.Unmarshal(b, cmd)
	return cmd, err
}

func parseCommandString(b []byte) (cmd *Command, err error) {
	cmd = &Command{}
	bl := len(b) - 1
	if b[bl] == '\n' {
		b = b[:bl]
	}
	parts := strings.Split(string(b), ` `)
	pl := len(parts)
	if pl < 1 {
		return cmd, errors.New(`Invalid command`)
	}
	for i, part := range parts {
		ppl := len(part) - 1
		if part[ppl] == ' ' {
			parts[i] = parts[i][:ppl]
		}
	}
	cmd.Command = parts[0]
	if pl > 1 {
		cmd.Key = parts[1]
	}
	if pl > 2 {
		cmd.Value = strings.Join(parts[2:], ` `)
	}
	return cmd, err
}
