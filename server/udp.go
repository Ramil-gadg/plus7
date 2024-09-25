package main

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func CreateTranslationUDP(clientIP net.IP, clientPort uint16, ip net.IP, port uint16, toClientChannel chan []byte) (*Translation, error) {
	newConn, err := net.Dial("udp", fmt.Sprint(ip.String(), ":", port)) // TODO: format??

	if err != nil {
		return nil, errors.New("Не удалось открыть UPD соединение: " + err.Error())
	}

	fromClientChannel := make(chan gopacket.Packet)

	go func() error {
		for {
			packet := <-fromClientChannel

			udp, _ := packet.Layer(layers.LayerTypeUDP).(*layers.UDP)

			_, err := newConn.Write(udp.Payload)

			if err != nil {
				fmt.Println("Ошибка при записи клиентского UDP")
			}
		}
	}()

	translation := Translation{
		port:    uint16(clientPort),
		ttl:     time.Now().Add(30 * time.Second), // FIXME: Реализовать TTL
		conn:    &newConn,
		channel: fromClientChannel,
	}

	go (func() {
		defer newConn.Close()

		for {
			buf := make([]byte, 4096)
			bufLen, err := newConn.Read(buf)

			if err != nil {
				fmt.Println("Ошибка при чтении клиентского UDP")
				return
			}

			if bufLen == 0 {
				continue
			}

			ipLayer := &layers.IPv4{
				SrcIP:    ip,
				DstIP:    clientIP,
				Version:  4,
				IHL:      5,
				TTL:      64,
				Protocol: layers.IPProtocolUDP,
			}

			udpLayer := &layers.UDP{
				SrcPort: layers.UDPPort(port),
				DstPort: layers.UDPPort(clientPort),
			}
			udpLayer.Payload = buf[:bufLen]

			udpLayer.SetNetworkLayerForChecksum(ipLayer)

			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}

			packetBuf := gopacket.NewSerializeBuffer()
			err = gopacket.SerializeLayers(packetBuf, opts, ipLayer, udpLayer, gopacket.Payload(udpLayer.Payload))

			if err != nil {
				fmt.Println("Ошибка сериализации:", err)
				return
			}

			packetData := packetBuf.Bytes()

			toClientChannel <- packetData
		}
	})()

	return &translation, nil
}
