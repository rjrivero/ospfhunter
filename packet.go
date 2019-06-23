package main

import (
	"fmt"
	"strings"

	// Tengo que usar este fork hasta que arreglen los problemas con el parseo OSPF
	// Ver:
	// https://github.com/google/gopacket/pull/671
	// https://github.com/google/gopacket/pull/672
	"github.com/rjrivero/gopacket"
)

type packet struct {
	gopacket.Packet
	PacketNum int
}

type packetRing struct {
	Items []packet
	Ring
}

func makePacketRing(size int) packetRing {
	return packetRing{
		Items: make([]packet, size),
		Ring:  Ring{Size: size},
	}
}

// Forward the pointer to the next item of the ring
func (r *packetRing) Push(packetNum int, p gopacket.Packet) {
	r.Items[r.HeadNext()] = packet{Packet: p, PacketNum: packetNum}
}

// String enumerates the frame numbers
func (r packetRing) String() string {
	buf := make([]string, 0, r.Size)
	for iter := r.Each(); iter.Next(); {
		buf = append(buf, fmt.Sprintf("frame %d", r.Items[iter.At].PacketNum))
	}
	return strings.Join(buf, ", ")
}
