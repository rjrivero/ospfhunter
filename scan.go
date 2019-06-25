package main

import (
	"io"
	"log"
	"os"

	// Tengo que usar este fork hasta que arreglen los problemas con el parseo OSPF
	// Ver:
	// https://github.com/google/gopacket/pull/671
	// https://github.com/google/gopacket/pull/672
	"github.com/rjrivero/gopacket"
	"github.com/rjrivero/gopacket/pcap"
	"golang.org/x/xerrors"
)

// keyFunc returns a group key from the given packet.
type keyFunc func(gopacket.Packet) (string, error)

type group struct {
	slidingCount
	packetRing
}

type packetIterator struct {
	Source     *gopacket.PacketSource
	packetNum  int
	lastPacket gopacket.Packet
	lastError  error
}

func (p *packetIterator) Next() bool {
	packet, err := p.Source.NextPacket()
	if err != nil {
		if !xerrors.Is(err, io.EOF) {
			p.lastError = err
		}
		return false
	}
	if packet == nil {
		return false
	}
	p.lastPacket = packet
	p.packetNum++
	return true
}

func (p *packetIterator) Value() (int, gopacket.Packet) {
	return p.packetNum, p.lastPacket
}

func (p *packetIterator) Err() error {
	return p.lastError
}

// Scan the file for a burst of frames with the same key.
func scan(filename string, interval, count int, groupBy keyFunc) (empty packetRing, err error) {
	// Me salto directorios
	stat, err := os.Stat(filename)
	if err != nil {
		return empty, err
	}
	if stat.IsDir() {
		log.Println("Ignorando ", filename, " por ser un directorio")
		return empty, nil
	}
	// Abro la captura y la envuelvo en un iterador
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		return empty, err
	}
	defer handle.Close()
	packetSource := packetIterator{
		Source: gopacket.NewPacketSource(handle, handle.LinkType()),
	}
	// Y ahora, voy agrupando ráfagas
	groupMap := make(map[string]*group)
	for packetSource.Next() {
		packetNum, nextPacket := packetSource.Value()
		if err := nextPacket.ErrorLayer(); err != nil {
			return empty, xerrors.Errorf("Error decodificando paquete #%d: %w", packetNum, err)
		}
		key, err := groupBy(nextPacket)
		if err != nil {
			return empty, xerrors.Errorf("Error calculando key del paquete #%d: %w", packetNum, err)
		}
		if key == "" {
			continue
		}
		atSecond := nextPacket.Metadata().Timestamp.Unix()
		current, ok := groupMap[key]
		if !ok {
			current = &group{
				slidingCount: makeSlidingCount(interval, count),
				packetRing:   makePacketRing(count),
			}
			groupMap[key] = current
		}
		current.Push(packetNum, nextPacket)
		if burst := current.Inc(atSecond); burst >= count {
			return current.packetRing, nil
		}
	}
	if err := packetSource.Err(); err != nil {
		return empty, xerrors.Errorf("Error iterando paquetes: %w", err)
	}
	return empty, nil
}
