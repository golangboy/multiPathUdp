package multiPathUDP

import (
	"bytes"
	"encoding/gob"
	mapset "github.com/deckarep/golang-set"
	"net"
	"sync"
	"time"
)

type LocalUdpConn struct {
	*net.UDPConn
}
type Client struct {
	msgId sync.Map
}
type Server struct {
	msgId sync.Map
}
type Message struct {
	ID    string
	msgID uint
	Data  []byte
}

func ConvertUDP(u *net.UDPConn) *LocalUdpConn {
	return &LocalUdpConn{u}
}
func ConvertAddr(addr net.Addr) net.UDPAddr {
	v := addr.(*net.UDPAddr)
	return *v
}

func (c *LocalUdpConn) WriteMessage(ID string, msgId uint, Data []byte) (int, error) {
	msg := Message{ID, msgId, Data}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(msg)
	if err != nil {
		return 0, err
	}
	return c.Write(buf.Bytes())
}
func (c *LocalUdpConn) ReadMessage() (*Message, error) {
	var msg Message
	var data [102400]byte
	read, err := c.Read(data[:])
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data[:read])
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
func (c *LocalUdpConn) ReadFromMessage() (*Message, net.UDPAddr, error) {
	var msg Message
	var data [102400]byte
	read, Addr, err := c.ReadFrom(data[:])
	udpAddr := ConvertAddr(Addr)
	if err != nil {
		return nil, udpAddr, err
	}
	buf := bytes.NewBuffer(data[:read])
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&msg)
	if err != nil {
		return nil, udpAddr, err
	}
	return &msg, udpAddr, nil
}
func (c *LocalUdpConn) WriteToMessage(ID string, msgId uint, addr *net.UDPAddr, Data []byte) (int, error) {
	msg := Message{ID, msgId, Data}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(msg)
	if err != nil {
		return 0, err
	}
	return c.WriteTo(buf.Bytes(), addr)
}
func (c *Client) connect2Middle(middleServers []string) []*net.UDPConn {
	var middle []*net.UDPConn
	for _, server := range middleServers {
		addr, err := net.ResolveUDPAddr("udp", server)
		if err != nil {
			panic(err)

		}
		dial, err := net.DialUDP("udp4", nil, addr)
		if err != nil {
			//fmt.Println(err)
			continue
		}
		middle = append(middle, dial)
	}
	return middle
}
func (c *Client) forward2Middle(conId string, rawData []byte, middle []*net.UDPConn) {
	msgId := c.getMsgID(conId)
	for _, m := range middle {
		m2 := ConvertUDP(m)
		m2.WriteMessage(conId, msgId, rawData)
		//fmt.Println(i)
	}
}
func (c *Client) listenMiddleAnd2client(udp *net.UDPConn, id2conId *sync.Map, middle []*net.UDPConn) {
	for _, m := range middle {
		m2 := ConvertUDP(m)
		go func() {
			for {
				//fmt.Println("listenMiddle0", m.LocalAddr(), m.RemoteAddr())
				message, err := m2.ReadMessage()
				//fmt.Println("listenMiddle1", m.LocalAddr(), m.RemoteAddr())
				if err != nil {
					panic(err)
					return
				}
				//fmt.Println("接收到Middle的", m.LocalAddr(), m.RemoteAddr(), string(message.Data))
				conID := message.ID
				v, ok := id2conId.Load(conID)
				if !ok {
					panic("err")
				}
				v2 := v.(net.Addr)
				udp.WriteTo(message.Data, v2)
			}
		}()
	}
}
func (c *Client) ListenRawAnd2Middle(port int, middleServers []string) {
	var buffer [10240]byte
	udp, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		panic(err)
		return
	}
	ms := c.connect2Middle(middleServers)
	if ms == nil || len(ms) == 0 {
		panic("connect failed")
		return
	}
	var id2conId sync.Map //string net.Addr

	c.listenMiddleAnd2client(udp, &id2conId, ms)
	// Read From Client
	for {
		//fmt.Println("来自raw client ka1")
		n, addr, err := udp.ReadFrom(buffer[:])
		//fmt.Println("来自raw client ka2", addr)
		if err != nil {
			return
		}
		// simple treat addr as ID
		conId := addr.String()
		id2conId.LoadOrStore(conId, addr)
		//fmt.Println("forward2Middle")
		c.forward2Middle(conId, buffer[:n], ms)
	}
}
func (c *Client) getMsgID(conID string) uint {
	v, ok := c.msgId.LoadOrStore(conID, 0)
	if !ok {
		panic("err")
	}
	c.msgId.Store(conID, v.(int)+1)
	return (v.(uint)) + 1
}
func (s *Server) forward2Target(udp *net.UDPConn, rawData []byte) {
	udp.Write(rawData)
}
func (s *Server) listenTargetAnd2middle(udpWithMiddle *LocalUdpConn, udpWithTarget *net.UDPConn, conId string, id2middle *sync.Map) {
	var buffer [10240]byte
	for {
		//fmt.Println("listen target111")
		udpWithTarget.SetReadDeadline(time.Now().Add(time.Minute * 3))
		read, err := udpWithTarget.Read(buffer[:])
		//fmt.Println("listen target222")
		if err != nil {
			udpWithTarget.Close()
			id2middle.Delete(conId)
			return
		}
		//fmt.Println("listen target", read, string(buffer[:read]))
		v, ok := id2middle.Load(conId)
		if !ok {
			panic("err")
		}
		v2 := v.(mapset.Set)
		v3 := v2.ToSlice()
		//fmt.Println("middle", len(v3))
		msgId := s.getMsgID(conId)
		for _, v4 := range v3 {

			addr := v4.(string)
			udpAddr, err := net.ResolveUDPAddr("udp4", addr)
			if err != nil {
				panic(err)
				return
			}
			//fmt.Println("xiexie", v4, addr, string(buffer[:read]), udpWithMiddle.RemoteAddr(), udpWithMiddle.LocalAddr())
			_, err = udpWithMiddle.WriteToMessage(conId, msgId, udpAddr, buffer[:read])
			if err != nil {
				panic(err)
				return
			}
		}
	}
}
func (s *Server) ListenMiddleAnd2Target(port int, targetServer string) {
	udp, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		panic(err)
		return
	}
	udp2 := ConvertUDP(udp)
	var id2target sync.Map // id,[]*net.UDPConn
	var id2middle sync.Map // id,[]string
	for {
		//fmt.Println("666")
		msg, addr, err := udp2.ReadFromMessage()
		if err != nil {
			//fmt.Println(err)
		}

		//fmt.Println(addr, "中间件收到消息", msg.ID, string(msg.Data))
		conID := msg.ID

		//save middle addr
		v, _ := id2middle.LoadOrStore(conID, mapset.NewSet())
		v.(mapset.Set).Add(addr.String())
		//fmt.Println(v)
		if _, ok := id2target.Load(conID); !ok {
			udpAddr, err := net.ResolveUDPAddr("udp4", targetServer)
			if err != nil {
				panic(err)
				return
			}
			dialUDP, err := net.DialUDP("udp4", nil, udpAddr)
			if err != nil {
				return
			}
			id2target.Store(conID, dialUDP)

			go s.listenTargetAnd2middle(udp2, dialUDP, conID, &id2middle)
		}
		v, ok := id2target.Load(conID)
		if !ok {
			panic("err")
		}
		v2 := v.(*net.UDPConn)
		//fmt.Println("发送消息到目标")
		s.forward2Target(v2, msg.Data)
	}
}
func (s *Server) getMsgID(conID string) uint {
	v, ok := s.msgId.LoadOrStore(conID, 0)
	if !ok {
		panic("err")
	}
	s.msgId.Store(conID, v.(int)+1)
	return (v.(uint)) + 1
}
