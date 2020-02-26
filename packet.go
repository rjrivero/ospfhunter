package main

import (
	"fmt"
	"strings"

	// Tengo que usar este fork hasta que arreglen los problemas con el parseo OSPF
	// Ver:
	// https://github.com/google/gopacket/pull/671
	// https://github.com/google/gopacket/pull/672
	"github.com/google/gopacket"
	"github.com/rjrivero/ring"
)

type packet struct {
	gopacket.Packet
	PacketNum int
}

type packetRing struct {
	Key   string
	Items []packet
	ring.Ring
}

func makePacketRing(key string, size int) packetRing {
	return packetRing{
		Key:   key,
		Items: make([]packet, size),
		Ring:  ring.New(size),
	}
}

// Forward the pointer to the next item of the ring
func (r *packetRing) Push(packetNum int, p gopacket.Packet) {
	r.Items[r.Ring.Push()] = packet{Packet: p, PacketNum: packetNum}
}

// String enumerates the frame numbers
func (r packetRing) String() string {
	if !r.Some() {
		return ""
	}
	// Primera linea: clave del grupo
	str := make([]string, 3)
	str[0] = fmt.Sprintf("Key: %s", r.Key)
	// Segunda linea: fechas
	iter := r.Ring
	start := r.Items[iter.PopFront()].Metadata().Timestamp
	stop := start
	if iter.Some() {
		stop = r.Items[iter.Pop()].Metadata().Timestamp
	}
	str[1] = fmt.Sprintf("Intervalo %s - %s", start.String(), stop.String())
	// Tercera linea: paquetes
	buf := make([]string, 0, r.Len())
	for iter := r.Ring; iter.Some(); {
		buf = append(buf, fmt.Sprintf("frame %d", r.Items[iter.PopFront()].PacketNum))
	}
	str[2] = strings.Join(buf, ", ")
	return strings.Join(str, "\n")
}
