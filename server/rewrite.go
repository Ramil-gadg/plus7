package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func RewriteTCP(packet gopacket.Packet, newIP net.IP, newPort layers.TCPPort) ([]byte, error) {
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	tcpLayer := packet.Layer(layers.LayerTypeTCP)

	if ipv4Layer == nil || tcpLayer == nil {
		return nil, errors.New("Не удалось распознать слои IPv4 или TCP")
	}

	ip, _ := ipv4Layer.(*layers.IPv4)
	tcp, _ := tcpLayer.(*layers.TCP)

	ip.DstIP = newIP
	tcp.DstPort = newPort

	tcp.SetNetworkLayerForChecksum(ip) // FIXME: sure need for update? it will on serialize

	buf := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{}, ip, tcp)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Ошибка сборки пакета: %x", err)) // FIXME: check format
	}

	newPacketData := buf.Bytes()

	return newPacketData, nil
}
