package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	TCPState_INIT        = 0
	TCPState_SYN_SENT    = 1
	TCPState_ESTABLISHED = 2
	TCPState_FIN_WAIT_1  = 3
	TCPState_FIN_WAIT_2  = 4
	TCPState_CLOSE_WAIT  = 5
	TCPState_CLOSING     = 6
	TCPState_LAST_ACK    = 7
	TCPState_TIME_WAIT   = 8
	TCPState_CLOSED      = 9
)

func CreatePacketDataTCP(srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16, tcpLayer *layers.TCP) ([]byte, error) {
	ipLayer := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Version:  4,
		IHL:      5,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}

	tcpLayer.SetNetworkLayerForChecksum(ipLayer)

	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	packetBuf := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(packetBuf, opts, ipLayer, tcpLayer, gopacket.Payload(tcpLayer.Payload))

	if err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}

func CreateTranslationTCP(clientIP net.IP, clientPort uint16, ip net.IP, port uint16, toClientChannel chan []byte) (*Translation, error) {
	newConn, err := net.Dial("tcp", fmt.Sprint(ip, ":", port))

	if err != nil {
		return nil, errors.New("Не удалось открыть TCP соединение: " + err.Error())
	}

	state := TCPState_INIT

	var a uint32 = 0
	var b uint32 = rand.Uint32() & 0xFFFFFF

	mu := sync.Mutex{}

	fromClientChannel := make(chan gopacket.Packet)

	go func() {
		for {
			packet := <-fromClientChannel
			{
				mu.Lock()
				defer mu.Unlock()

				tcp, _ := packet.Layer(layers.LayerTypeTCP).(*layers.TCP)

				/*

					1. Я получаю SYN пакет с A=Seq
						- Устанавливаю целевое соединение. Если оно установлено
					2. Я передаю клиенту SYN-ACK. Ack=A+1, Seq=B=random()
					3. Я получаю Seq=A+1, Ack=B+1

				*/

				if state == TCPState_INIT && tcp.SYN {
					state = TCPState_SYN_SENT

					a = tcp.Seq + 1

					packetData, err := CreatePacketDataTCP(ip, port, clientIP, clientPort, &layers.TCP{
						SrcPort: layers.TCPPort(port),
						DstPort: layers.TCPPort(clientPort),
						Seq:     b,
						Ack:     a,
						SYN:     true,
						ACK:     true,
						Window:  65535,
					})

					if err != nil {
						fmt.Println("Ошибка TCP:", err.Error())
						return
					}

					toClientChannel <- packetData
					b += 1
				} else if state == TCPState_SYN_SENT && tcp.ACK {
					state = TCPState_ESTABLISHED

					packetData, err := CreatePacketDataTCP(ip, port, clientIP, clientPort, &layers.TCP{
						SrcPort: layers.TCPPort(port),
						DstPort: layers.TCPPort(clientPort),
						Seq:     b,
						Ack:     a,
						ACK:     true,
						Window:  65535,
					})

					if err != nil {
						fmt.Println("Ошибка TCP:", err.Error())
						return
					}

					toClientChannel <- packetData
				} else if state == TCPState_ESTABLISHED && tcp.PSH {
					a = tcp.Seq
					a += uint32(len(tcp.Payload))

					newConn.Write(tcp.Payload)
					fmt.Println("WRITE OUT:", string(tcp.Payload))

					packetData, err := CreatePacketDataTCP(ip, port, clientIP, clientPort, &layers.TCP{
						SrcPort: layers.TCPPort(port),
						DstPort: layers.TCPPort(clientPort),
						Seq:     b,
						Ack:     a,
						ACK:     true,
						Window:  65535,
					})

					if err != nil {
						fmt.Println("Ошибка TCP:", err.Error())
						return
					}

					toClientChannel <- packetData
				} else if state == TCPState_ESTABLISHED && tcp.FIN {
					// TODO: connOpened ? fin-wait-1 : fin-wait-2
				}

				return
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
			buf := make([]byte, 64000)
			bufLen, err := newConn.Read(buf)

			if err != nil {
				fmt.Println("Ошибка при чтении установленного TCP:", err.Error())
				return
			}

			fmt.Println("READ FROM OUT:", string(buf[:bufLen]))

			if bufLen == 0 {
				continue
			}

			mu.Lock()
			defer mu.Unlock()

			tcpLayer := &layers.TCP{
				SrcPort: layers.TCPPort(port),
				DstPort: layers.TCPPort(clientPort),
				Seq:     b,
				Ack:     a,
				PSH:     true,
				ACK:     true,
				Window:  65535,
			}
			tcpLayer.Payload = buf[:bufLen]

			packetData, err := CreatePacketDataTCP(ip, port, clientIP, clientPort, tcpLayer)

			if err != nil {
				fmt.Println("Ошибка:", err)
				return
			}

			toClientChannel <- packetData
			b += uint32(bufLen)
		}

		//TODO: fin
	})()

	return &translation, nil
}
