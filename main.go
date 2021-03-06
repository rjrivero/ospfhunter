package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	// Se debe utilizar una versión de gopacket que incluya estos fixes:
	// https://github.com/google/gopacket/pull/671
	// https://github.com/google/gopacket/pull/672
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/xerrors"
)

// MinInterval is the minimum time window length
const MinInterval = 10

// MaxInterval is the maximum time window length
const MaxInterval = 1000

// MinCount is the minimum number of hits to find in the time window
const MinCount = 2

func main() {

	var interval, count int
	var test bool
	flag.IntVar(&interval, "i", 60, "Intervalo de tiempo en el que contar paquetes (segundos)")
	flag.IntVar(&count, "c", 10, "Número de paquetes que deben coincidir en el intervalo")
	flag.BoolVar(&test, "t", false, "HACK: Usar como pipe de salida de tcpdump -xx -r <pcap>")
	flag.Parse()
	filenameList := flag.Args()

	// El modo test es solo para ayudar a generar PRs para gopacket
	if test {
		formatTCPDump()
		return
	}

	if interval < MinInterval || interval > MaxInterval {
		log.Fatalln("El intervalo debe estar entre ", MinInterval, " y ", MaxInterval)
	}
	if count < MinCount {
		log.Fatalln("El número de paquetes debe ser al menos ", MinCount)
	}
	if len(filenameList) <= 0 {
		log.Fatalln("Debe especificar al menos un fichero pcap")
	}
	log.Println("Buscando ráfaga de ", count, " paquetes en ", interval, " segundos")

	// Lanzar todos los ficheros en paralelo
	wg := sync.WaitGroup{}
	wg.Add(len(filenameList))

	for _, filename := range filenameList {
		go func(filename string) {
			defer wg.Done()
			scanner, err := newScanner(filename, interval, count, unicastKey)
			if err != nil {
				// Log is concurrency-safe, we can run from the goroutine
				log.Printf("Error creando scanner para fichero %s: %+v\n", filename, err)
				return
			}
			defer scanner.Close()
			for scanner.Next() {
				log.Printf("Ráfaga encontrada en fichero %s:\n%s\n\n", filename, scanner.Burst().String())
			}
			if err := scanner.Err(); err != nil {
				log.Printf("Error procesando fichero %s: %+v\n", filename, err)
			}
		}(filename)
	}
	wg.Wait()
}

// unicastKey returns src and dst IP if the packet is a MaxAge unicast LSA
func unicastKey(p gopacket.Packet) (string, error) {
	ok, err := isMaxAge(p)
	if err != nil || !ok {
		return "", err
	}
	return unicastAddresses(p)
}

// isMaxAge returns true if the packet is a MaxAge OSPF LSA frame
func isMaxAge(p gopacket.Packet) (bool, error) {
	if ospfLayer := p.Layer(layers.LayerTypeOSPF); ospfLayer != nil {
		ospf, ok := ospfLayer.(*layers.OSPFv2)
		if !ok {
			return false, xerrors.Errorf("Payload incorrecto en OSPFLayer: %+v", ospfLayer)
		}
		if ospf.Type == layers.OSPFLinkStateUpdate {
			lsUpdate, ok := ospf.Content.(layers.LSUpdate)
			if !ok {
				return false, xerrors.Errorf("Payload incorrecto en LSUpdate: %+v", ospf.Content)
			}
			for _, lsa := range lsUpdate.LSAs {
				if lsa.LSAge >= 3600 {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// unicastAddresses returns the src and dest IP if both are unicast. Otherwise returns ""
func unicastAddresses(p gopacket.Packet) (string, error) {
	if ipLayer := p.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip, ok := ipLayer.(*layers.IPv4)
		if !ok {
			return "", xerrors.Errorf("Payload incorrecto en IPLayer: %+v", ipLayer)
		}
		if !ip.SrcIP.IsMulticast() && !ip.DstIP.IsMulticast() {
			return strings.Join([]string{ip.SrcIP.String(), ip.DstIP.String()}, "-"), nil
		}
	}
	return "", nil
}

func formatTCPDump() {
	// Helper func para convertir la salida de tcpdump en un formato "amigable" para
	// construir test-cases de github.com/google/layers/ospf_test.go.
	//
	// Lo he tenido que usar un par de veces para hacer PRs.
	//
	// Usar con: tcpdump -xx -r <pcap file> | ospfhunter -t
	for b := bufio.NewScanner(os.Stdin); b.Scan(); {
		line := strings.TrimSpace(b.Text())
		if strings.HasPrefix(line, "0x") {
			parts := strings.Fields(line)
			top := 9
			if top > len(parts) {
				top = len(parts)
			}
			bytes := make([]string, 0, 16)
			for _, part := range parts[1:top] {
				bytes = append(bytes, fmt.Sprintf("0x%s", part[0:2]))
				bytes = append(bytes, fmt.Sprintf("0x%s", part[2:4]))
			}
			fmt.Println(strings.Join(bytes, ", "))
		}
	}
}
