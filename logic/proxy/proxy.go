package proxy

import (
	"net"
	"bufio"
	log "github.com/thinkboy/log4go"
	"time"
	m "mproxy/logic/memcache"
	"strconv"
	"strings"
	"sync"
	"mproxy"
	"bytes"
)

var (
	RESULT_STORED = []byte("STORED\r\n")
	RESULT_DELETED = []byte("DELETED\r\n")
)

var (
	mcs = map[string]*m.Client{}

	mip = &sync.Pool{New:func() interface{} {
		return m.Item{}
	}}
	brp = &sync.Pool{New:func() interface{} {
		return &bufio.Reader{}
	}}
	wgp = &sync.Pool{New:func() interface{} {
		return sync.WaitGroup{}
	}}
)

func Run(conf mproxy.Conf) {
	servers := conf.Servers
	m.MAX_IDLE_CONNS = conf.MemcacheMaxIdleConns
	for _, v := range servers {
		mcs[v] = m.New(v)
	}

	l, err := net.Listen("tcp", ":" + conf.Port)
	if err != nil {
		log.Error("net listen error ", err)
		panic(err)
	}

	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Error("accept error ", err)
			continue
		}
		go relay(c)

	}
}

func relay(c net.Conn) {
	i, _ := mip.Get().(m.Item)
	r, _ := brp.Get().(*bufio.Reader)

	defer func() {
		mip.Put(i)
		brp.Put(r)
		c.Close()
	}()

	err := c.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		log.Error("set dead line error ", err)
		return
	}

	r = bufio.NewReader(c)

	var (
		value []byte
		b []byte
		rl []byte
		s string
		f []string
	)

	for {
		rl, err = r.ReadBytes('\n')
		if err != nil {
			return
		}
		s = string(rl)
		f = strings.Fields(s)

		if len(f) < 1 {
			log.Error("params error ", f)
			return
		}

		oc := f[0]
		switch oc {
		case "set":
			bytesLen, _ := strconv.Atoi(f[4])
			value, err = r.ReadBytes('\n')
			if err != nil {
				return
			}
			if len(value) < bytesLen {
				for {
					b, err = r.ReadBytes('\n')
					if err != nil {
						return
					}
					value = append(value, b...)
					if len(value) < bytesLen {
						continue
					}

					break
				}
				value = bytes.TrimRight(value, "\r\n")
			}
			flag, _ := strconv.Atoi(f[2])
			exp, _ := strconv.Atoi(f[3])

			i.Key = f[1]
			i.Flags = uint32(flag)
			i.Expiration = int32(exp)
			i.Value = value
			set(c, &i)
			break
		case "delete":
			del(c, f[1])
			break
		case "quit":
			c.Close()
			return
		default:
			c.Write([]byte("CLIENT_ERROR UNSUPPORTED\r\n"))
		}

	}
}

func del(c net.Conn, k string) {
	wg := wgp.Get().(sync.WaitGroup)

	defer func(c net.Conn) {
		c.Write(RESULT_DELETED)
		wgp.Put(wg)
	}(c)

	for s, mc := range mcs {
		wg.Add(1)
		go func(mc *m.Client, s string) {
			defer wg.Done()
			err := mc.Delete(k)
			if err != nil {
				log.Info("del error ", s, err)
				return
			}
		}(mc, s)
	}
	wg.Wait()
}

func set(c net.Conn, i *m.Item) {
	wg := wgp.Get().(sync.WaitGroup)

	defer func(c net.Conn) {
		c.Write(RESULT_STORED)
		wgp.Put(wg)
	}(c)

	for s, mc := range mcs {
		wg.Add(1)
		go func(mc *m.Client, s string) {
			defer wg.Done()
			err := mc.Set(i)
			if err != nil {
				log.Info("set error ", s, err)
				return
			}
		}(mc, s)
	}
	wg.Wait()
}