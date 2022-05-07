package ping

import (
	"time"
)

type Statistic struct {
	IsLost bool
	Rtt    time.Duration
}

type Statistics struct {
	// PacketsRecv is the number of packets received.
	PacketsRecv int

	// PacketsSent is the number of packets sent.
	PacketsSent int

	// PacketsRecvDuplicates is the number of duplicate responses there were to a sent packet.
	PacketsRecvDuplicates int

	// PacketLoss is the percentage of packets lost.
	PacketLoss float64

	LocalIP string

	RemoteIP string

	// Rtts is all of the round-trip times sent via this pinger.
	Rtts []time.Duration

	// MinRtt is the minimum round-trip time sent via this pinger.
	MinRtt time.Duration

	// MaxRtt is the maximum round-trip time sent via this pinger.
	MaxRtt time.Duration

	// AvgRtt is the average round-trip time sent via this pinger.
	AvgRtt time.Duration

	// StdDevRtt is the standard deviation of the round-trip times sent via
	// this pinger.
	StdDevRtt time.Duration
}
