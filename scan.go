package main

import (
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
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		return empty, err
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetNum := 0
	groupMap := make(map[string]*group)
	for packet := range packetSource.Packets() {
		packetNum++
		if err := packet.ErrorLayer(); err != nil {
			return empty, xerrors.Errorf("Error decodificando paquete #%d: %w", packetNum, err)
		}
		key, err := groupBy(packet)
		if err != nil {
			return empty, xerrors.Errorf("Error calculando key del paquete #%d: %w", packetNum, err)
		}
		if key == "" {
			continue
		}
		atSecond := packet.Metadata().Timestamp.Unix()
		current, ok := groupMap[key]
		if !ok {
			currsc := group{
				slidingCount: makeSlidingCount(interval, count),
				packetRing:   makePacketRing(count),
			}
			current = &currsc
			groupMap[key] = current
		}
		current.Push(packetNum, packet)
		if burst := current.Inc(atSecond); burst >= count {
			return current.packetRing, nil
		}
	}
	return empty, nil
}
