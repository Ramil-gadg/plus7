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
}

type Client struct {
	id              int
	udpTranslations []*Translation
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
				handle:          handle,
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

func CreateTranslationUDP(clientIP net.IP, clientPort uint16, ip net.IP, port uint16, handle func([]byte)) (*Translation, error) {
	newConn, err := net.Dial("udp", fmt.Sprint("%s:%x", ip.String(), port)) // TODO: format??

	if err != nil {
		return nil, errors.New("Не удалось открыть UPD соединение")
	}

	translation := Translation{
		port: uint16(clientPort),
		ttl:  time.Now().Add(30 * time.Second),
		conn: &newConn,
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

			// Создание TCP и IP слоев
			ipLayer := &layers.IPv4{
				SrcIP:    ip,
				DstIP:    clientIP,
				Version:  4,
				IHL:      5,
				TTL:      64,
				Protocol: layers.IPProtocolTCP,
			}

			udpLayer := &layers.UDP{
				SrcPort: layers.UDPPort(port),
				DstPort: layers.UDPPort(clientPort),
			}

			udpLayer.Payload = buf[:bufLen]

			packetBuf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{}

			err = gopacket.SerializeLayers(packetBuf, opts, ipLayer, udpLayer)

			if err != nil {
				fmt.Println("Ошибка сериализации:", err)
				return
			}

			// Получение готового пакета
			packetData := packetBuf.Bytes()

			handle(packetData)
		}
	})()

	return &translation, nil
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

		_, err := (*translation.conn).Write(udp.Payload)

		if err != nil {
			fmt.Println("Ошибка при записи серверного UDP")
		}
	} else if ip.Protocol == layers.IPProtocolTCP {
		// TODO: TCP
	} else {
		return errors.New("Неподдерживаемый IP-протокол")
	}

	// TODO:

	return nil
}
