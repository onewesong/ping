package ping

import (
	"bytes"
	"log"
	"math"
	"net"
	"os"
	"sync"
	"time"
)

type Pinger struct {
	laddr *net.IPAddr
	raddr *net.IPAddr

	// Count tells pinger to stop after sending (and receiving) Count echo
	// packets. If this option is not specified, pinger will operate until
	// interrupted.
	Count int

	// Interval is the wait time between each packet send. Default is 1s.
	Interval time.Duration

	// Timeout specifies a timeout before ping exits, regardless of how many
	// packets have been received.
	Timeout time.Duration

	// Verbose output each ping detail.
	Verbose bool

	// Number of packets sent
	PacketsSent int

	// Number of packets received
	PacketsRecv int

	// Number of duplicate packets received
	PacketsRecvDuplicates int

	// Round trip time statistics
	minRtt    time.Duration
	maxRtt    time.Duration
	avgRtt    time.Duration
	stdDevRtt time.Duration
	stddevm2  time.Duration
	statsMu   sync.RWMutex

	// rtts is all of the Rtts
	rtts []time.Duration

	// is finished
	finished bool

	// OnSetup is called when Pinger has finished setting up the listening socket
	OnSetup func()

	// OnSend is called when Pinger sends a packet
	OnSend func(*Packet)

	// OnLost is called when Pinger lost a packet
	OnLost func(*Packet)

	// OnRecv is called when Pinger receives and processes a packet
	OnRecv func(*Packet)

	// OnFinish is called when Pinger exits
	OnFinish func(*Statistics)
}

func (p *Pinger) updateStatistics(pkt *Packet) {
	p.statsMu.Lock()
	defer p.statsMu.Unlock()

	p.PacketsRecv++
	p.rtts = append(p.rtts, pkt.Rtt)

	if p.PacketsRecv == 1 || pkt.Rtt < p.minRtt {
		p.minRtt = pkt.Rtt
	}

	if pkt.Rtt > p.maxRtt {
		p.maxRtt = pkt.Rtt
	}

	pktCount := time.Duration(p.PacketsRecv)
	// welford's online method for stddev
	// https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Welford's_online_algorithm
	delta := pkt.Rtt - p.avgRtt
	p.avgRtt += delta / pktCount
	delta2 := pkt.Rtt - p.avgRtt
	p.stddevm2 += delta * delta2

	p.stdDevRtt = time.Duration(math.Sqrt(float64(p.stddevm2 / pktCount)))
}

func (p *Pinger) Statistics() *Statistics {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()
	sent := p.PacketsSent
	loss := float64(sent-p.PacketsRecv) / float64(sent) * 100
	s := Statistics{
		PacketsSent:           sent,
		PacketsRecv:           p.PacketsRecv,
		PacketsRecvDuplicates: p.PacketsRecvDuplicates,
		PacketLoss:            loss,
		Rtts:                  p.rtts,
		LocalIP:               p.laddr.String(),
		RemoteIP:              p.raddr.String(),
		MaxRtt:                p.maxRtt,
		MinRtt:                p.minRtt,
		AvgRtt:                p.avgRtt,
		StdDevRtt:             p.stdDevRtt,
	}
	return &s
}

func NewPinger(localIP, remoteIP string, timeout time.Duration, count int) *Pinger {
	laddr := net.IPAddr{IP: net.ParseIP(localIP)}
	raddr := net.IPAddr{IP: net.ParseIP(remoteIP)}
	return &Pinger{
		Interval: 1 * time.Second,

		laddr:   &laddr,
		raddr:   &raddr,
		Timeout: timeout,
		Count:   count,
	}
}

func (p *Pinger) Run() {
	if p.finished {
		return
	}
	defer p.Finish()
	ping := func(seq int) {
		var isLost = false
		err, packet := p.Ping(seq)
		if err != nil {
			isLost = true
			handler := p.OnLost
			if handler != nil {
				handler(&packet)
			}
		} else {
			handler := p.OnRecv
			if handler != nil {
				handler(&packet)
			}
			p.updateStatistics(&packet)
		}
		if p.Verbose {
			if isLost {
				log.Printf("lost seq=%d timeout=%ds", p.PacketsSent, p.Timeout.Milliseconds())
			} else {
				log.Printf("pong seq=%d time=%dms ttl=%v size=%dbyte", p.PacketsSent, packet.Rtt.Milliseconds(), packet.TTL, packet.Nbytes)
			}
		}
		p.PacketsSent++
	}
	for count := p.Count; count != 0; {
		if count > 0 {
			count--
		}
		ping(p.PacketsSent)
		time.Sleep(p.Interval)
	}
	return
}

func (p *Pinger) Ping(seq int) (err error, packet Packet) {
	packet.Seq = seq
	start := time.Now()
	c, err := net.DialIP("ip4:icmp", p.laddr, p.raddr)
	if err != nil {
		return
	}
	c.SetDeadline(time.Now().Add(p.Timeout))
	defer c.Close()

	typ := icmpv4EchoRequest
	xid, xseq := os.Getpid()&0xffff, 1
	wb, err := (&icmpMessage{
		Type: typ, Code: 0, SequenceNum: seq & 0xffff,
		Body: &icmpEcho{
			ID: xid, Seq: xseq,
			Data: bytes.Repeat([]byte("Ping"), 3),
		},
	}).Marshal()
	if err != nil {
		return
	}
	if _, err = c.Write(wb); err != nil {
		return
	}
	var m *icmpMessage
	rb := make([]byte, 20+len(wb))
	for {
		if _, err = c.Read(rb); err != nil {
			return
		}
		packet.TTL = int(rb[8])
		rb = ipv4Payload(rb)
		packet.Nbytes = len(rb)
		if m, err = parseICMPMessage(rb); err != nil {
			return
		}
		switch m.Type {
		case icmpv4EchoRequest, icmpv6EchoRequest:
			continue
		}
		packet.Rtt = time.Since(start)
		break
	}

	return
}

func ipv4Payload(b []byte) []byte {
	if len(b) < 20 {
		return b
	}
	hdrlen := int(b[0]&0x0f) << 2
	return b[hdrlen:]
}

var finishOnce sync.Once

func (p *Pinger) Finish() {
	finishOnce.Do(func() {
		p.finished = true
		handler := p.OnFinish
		if handler != nil {
			s := p.Statistics()
			handler(s)
		}
	})
}
