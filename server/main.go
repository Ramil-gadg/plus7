package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	socketio "github.com/googollee/go-socket.io"
)

var nat NAT

func hello() {
	fmt.Println("Hello Plus7!")
}

func listenLocal() {
	// Указание устройства, на котором будем слушать пакеты
	device := "lo0" // "any" для прослушивания всех интерфейсов на Linux

	// Тайм-аут для открытого устройства
	timeout := 1 * time.Millisecond

	// Открываем устройство для захвата пакетов
	handle, err := pcap.OpenLive(device, 1600, true, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Указываем фильтр для IPv4 пакетов
	var filter string = "ip and (tcp or udp)"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Начинаем захват пакетов")

	// Начинаем обработку захваченных пакетов
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Вывод информации о пакете
		fmt.Println("SPY >>>", packet)
	}
}

func parse() {
	data := []byte{
		0x45, 0x00, 0x00, 0x40, 0x00, 0x00, 0x40, 0x00, 0x40, 0x06, 0xeb, 0xe6,
		0x0a, 0x00, 0x00, 0x01, 0x95, 0x9a, 0xaf, 0x36, 0xf3, 0x34, 0x14, 0x66,
		0xe7, 0x47, 0x89, 0x2b, 0x00, 0x00, 0x00, 0x00, 0xb0, 0x02, 0xff, 0xff,
		0xb5, 0xb7, 0x00, 0x00, 0x02, 0x04, 0x05, 0xb4, 0x01, 0x03, 0x03, 0x04,
		0x01, 0x01, 0x08, 0x0a, 0x02, 0xe7, 0xb7, 0x7f, 0x00, 0x00, 0x00, 0x00,
		0x04, 0x02, 0x00, 0x00,
	}

	packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)

	fmt.Println(packet)

	// Извлеките IPv4 и TCP слои
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	tcpLayer := packet.Layer(layers.LayerTypeTCP)

	if ipv4Layer == nil || tcpLayer == nil {
		fmt.Println("Не удалось распознать слои IPv4 или TCP")
		return
	}

	ip, _ := ipv4Layer.(*layers.IPv4)
	tcp, _ := tcpLayer.(*layers.TCP)

	// Измените порт назначения
	newDstPort := layers.TCPPort(9090)
	tcp.DstPort = newDstPort

	// Обновите размер заголовка TCP
	tcp.SetNetworkLayerForChecksum(ip)

	// Создайте буфер для сериализации обновленного пакета
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}

	// Сериализуйте обновленные слои
	err := gopacket.SerializeLayers(buf, opts,
		ip,
		tcp,
	)

	if err != nil {
		fmt.Println("Ошибка сериализации:", err)
		return
	}

	// Получите обновленный пакет
	newPacketData := buf.Bytes()

	// Вывод информации об обновленном пакете
	fmt.Printf("Обновленный пакет:\n% x\n", newPacketData)
}

func main() {
	hello()

	// go listenLocal()

	nat = *NewNAT()

	server := socketio.NewServer(nil)

	server.OnConnect("/", func(client socketio.Conn) error {
		log.Println("Подключен клиент ID:", client.ID())

		toClientChannel := make(chan []byte)

		id, err := nat.AddClient(toClientChannel)

		if err != nil {
			return err
		}

		client.SetContext(id)

		go func() {
			for {
				data := <-toClientChannel
				client.Emit("in", data)
			}
		}()

		return nil
	})

	server.OnEvent("/", "out", func(client socketio.Conn, data []byte) {
		go func() {
			err := nat.WritePacket(client.Context().(int), data)

			if err != nil {
				log.Println("Ошибка при обработке пакета:", client.ID(), "Сообщение:", err)
			}
		}()
	})

	server.OnError("/", func(client socketio.Conn, err error) {
		log.Println("Ошибка клиента ID:", client.ID(), "Сообщение:", err)
	})

	server.OnDisconnect("/", func(client socketio.Conn, reason string) {
		log.Println("Соединение закрыто с клиентом ID:", client.ID(), "Причина:", reason)

		nat.RemoveClient(client.Context().(int))
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))

	log.Println("Сервер запущен на http://0.0.0.0:8000/socket.io/")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
