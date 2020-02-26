package main

import (
	"io"
	"log"
	"os"

	// Tengo que usar este fork hasta que arreglen los problemas con el parseo OSPF
	// Ver:
	// https://github.com/google/gopacket/pull/671
	// https://github.com/google/gopacket/pull/672
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"golang.org/x/xerrors"
)

// keyFunc returns a group key from the given packet.
type keyFunc func(gopacket.Packet) (string, error)

type group struct {
	slidingCount
	packetRing
	inBurst bool
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

// scanner scans pcap file looking for bursts
type scanner struct {
	iterator        packetIterator
	handle          *pcap.Handle
	groupMap        map[string]*group
	groupBy         keyFunc
	interval, count int
	lastBurst       packetRing
	lastErr         error
}

// NewScanner allocates a new Burst scanner
func newScanner(filename string, interval, count int, groupBy keyFunc) (*scanner, error) {
	// Me salto directorios
	stat, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		log.Println("Ignorando ", filename, " por ser un directorio")
		return nil, nil
	}
	// Abro la captura y la envuelvo en un iterador
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		return nil, err
	}
	return &scanner{
		iterator: packetIterator{
			Source: gopacket.NewPacketSource(handle, handle.LinkType()),
		},
		handle:   handle,
		groupMap: make(map[string]*group),
		groupBy:  groupBy,
		interval: interval,
		count:    count,
	}, nil
}

// Close the scanner
func (s *scanner) Close() {
	if s != nil && s.handle != nil {
		s.handle.Close()
		s.handle = nil
	}
}

// Next returns true if there is a burst in the pcap file
func (s *scanner) Next() bool {
	if s == nil || s.handle == nil || s.lastErr != nil {
		return false
	}
	for s.iterator.Next() {
		packetNum, nextPacket := s.iterator.Value()
		if err := nextPacket.ErrorLayer(); err != nil {
			s.lastErr = xerrors.Errorf("Error decodificando paquete #%d: %w", packetNum, err)
			return false
		}
		key, err := s.groupBy(nextPacket)
		if err != nil {
			s.lastErr = xerrors.Errorf("Error calculando key del paquete #%d: %w", packetNum, err)
			return false
		}
		if key == "" {
			continue
		}
		atSecond := nextPacket.Metadata().Timestamp.Unix()
		current, ok := s.groupMap[key]
		if !ok {
			current = &group{
				slidingCount: makeSlidingCount(s.interval, s.count),
				// Permito recoger ráfagas largas, hasta 10 veces más paquetes
				packetRing: makePacketRing(key, 10*s.count),
			}
			s.groupMap[key] = current
		}
		current.Push(packetNum, nextPacket)
		burst := current.Inc(atSecond)
		switch {
		case !current.inBurst && burst >= s.count:
			current.inBurst = true
		case current.inBurst && burst < s.count:
			delete(s.groupMap, key)
			s.lastBurst = current.packetRing
			return true
		}
	}
	// Loop may end while there are still some bursts. In that case,
	// yield them now.
	for key, group := range s.groupMap {
		delete(s.groupMap, key)
		if group.inBurst {
			s.lastBurst = group.packetRing
			return true
		}
	}
	if err := s.iterator.Err(); err != nil {
		s.lastErr = xerrors.Errorf("Error iterando paquetes: %w", err)
	}
	// If the file is exhausted, release early
	s.handle.Close()
	s.handle = nil
	return false
}

// Burst returns the current burst
func (s *scanner) Burst() packetRing {
	return s.lastBurst
}

// Err returns the last error in this Scanner
func (s *scanner) Err() error {
	if s == nil {
		return nil
	}
	return s.lastErr
}
