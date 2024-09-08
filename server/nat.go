package main

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Translation struct {
	port uint16
	conn *net.Conn
	ttl  time.Time

	Write func(packet gopacket.Packet) error
}

type Client struct {
	id              int
	udpTranslations []*Translation
	tcpTranslations []*Translation
	handle          func([]byte)
}

type NAT struct {
	clients []*Client
}

func NewNAT() *NAT {
	nat := NAT{}

	nat.clients = make([]*Client, 1024)

	return &nat
}

func (this *NAT) AddClient(handle func(data []byte)) (int, error) {
	for id, client := range this.clients {
		if client == nil {
			newClient := Client{
				udpTranslations: make([]*Translation, 65536),
				tcpTranslations: make([]*Translation, 65536),
				handle: func(data []byte) {
					fmt.Println(gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default))

					handle(data)
				},
			}

			this.clients[id] = &newClient
			return id, nil
		}
	}

	return 0, errors.New("Все свободные слоты для клиентов заняты")
}

func (this *NAT) RemoveClient(id int) error {
	client := this.clients[id]

	if client == nil {
		return errors.New("Клиент не инициализирован")
	}

	this.clients[id] = nil

	// FIXME: close all socket

	return nil
}

func (this *NAT) WritePacket(id int, packetData []byte) error {
	client := this.clients[id]

	if client == nil {
		return errors.New("Клиент не инициализирован")
	}

	packet := gopacket.NewPacket(packetData, layers.LayerTypeIPv4, gopacket.Default)

	fmt.Println(packet)

	ip, _ := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)

	if ip.Protocol == layers.IPProtocolUDP {
		udp, _ := packet.Layer(layers.LayerTypeUDP).(*layers.UDP)

		translation := client.udpTranslations[udp.SrcPort]
		if translation == nil {
			newTranslation, err := CreateTranslationUDP(ip.SrcIP, uint16(udp.SrcPort), ip.DstIP, uint16(udp.DstPort), client.handle)

			if err != nil {
				return err
			}

			client.udpTranslations[udp.SrcPort] = newTranslation
			translation = newTranslation
		}

		err := translation.Write(packet)

		if err != nil {
			fmt.Println("Ошибка при записи серверного UDP")
		}
	} else if ip.Protocol == layers.IPProtocolTCP {
		tcp, _ := packet.Layer(layers.LayerTypeTCP).(*layers.TCP)

		translation := client.tcpTranslations[tcp.SrcPort]
		if translation == nil {
			newTranslation, err := CreateTranslationTCP(ip.SrcIP, uint16(tcp.SrcPort), ip.DstIP, uint16(tcp.DstPort), client.handle)

			if err != nil {
				return err
			}

			client.tcpTranslations[tcp.SrcPort] = newTranslation
			translation = newTranslation
		}

		err := translation.Write(packet)

		if err != nil {
			fmt.Println("Ошибка при записи серверного TCP")
		}
	} else {
		return errors.New("Неподдерживаемый IP-протокол")
	}

	// TODO:

	return nil
}
