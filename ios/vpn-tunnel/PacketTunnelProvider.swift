//
//  PacketTunnelProvider.swift
//  vpn-tunnel
//
//  Created by Рамил Гаджиев on 06.09.2024.
//

import NetworkExtension
import SocketIO

class PacketTunnelProvider: NEPacketTunnelProvider {
    
    var socketManager: SocketManager?
    var socket: SocketIOClient?
    
    override func startTunnel(options: [String : NSObject]? = nil) async throws {
        let ipWebSocketAddress = try ipWebSocketAddress()
        let ipv4Settings = NEIPv4Settings(addresses: ["10.0.0.1"], subnetMasks: ["255.255.255.0"])
        if let ipv4IncludedAddress = ipv4IncludedAddress() {
            ipv4Settings.includedRoutes = [ipv4IncludedAddress]
        }
        let ipv6Settings = NEIPv6Settings(addresses: ["fd66:f83a:c650::1"], networkPrefixLengths: [128])
        if let ipv6IncludedAddress = ipv6IncludedAddress() {
            ipv6Settings.includedRoutes = [ipv6IncludedAddress]
        }
        
        let networkSettings = NEPacketTunnelNetworkSettings(tunnelRemoteAddress: "10.0.0.1")
        networkSettings.ipv4Settings = ipv4Settings
        networkSettings.ipv6Settings = ipv6Settings
        
        try await setTunnelNetworkSettings(networkSettings)
        setupSocketConnection(ipWebSocketAddress)
        startPacketProcessing()
    }
    
    override func stopTunnel(with reason: NEProviderStopReason) async {
        socket?.leaveNamespace()
        socket?.disconnect()
        socket = nil
        socketManager = nil
    }
    
    override func cancelTunnelWithError(_ error: Error?) {
        socket?.leaveNamespace()
        socket?.disconnect()
        socket = nil
        socketManager = nil
    }
    
    private func ipWebSocketAddress() throws -> String {
        guard let appGroupDefaults = UserDefaults(suiteName: "group.ramil.Gadzhiev.plus7"),
              let ipAddress = appGroupDefaults.string(forKey: "WebSocketIPAddress") else {
            print("Не удалось получить IP-адрес")
            throw NSError(domain: "MyVPNErrorDomain", code: 1, userInfo: nil)
        }
        print("IP-вебсокет адрес получен в расширении: \(ipAddress)")
        return ipAddress
    }
    
    private func ipv4IncludedAddress() -> NEIPv4Route? {
        guard let appGroupDefaults = UserDefaults(suiteName: "group.ramil.Gadzhiev.plus7"),
              let ipAddress = appGroupDefaults.string(forKey: "IncludedIPAddress") else {
            return nil
        }
        var sin = sockaddr_in()
        if ipAddress.withCString({ cstring in inet_pton(AF_INET, cstring, &sin.sin_addr) }) == 1 {
            // IPv4 peer.
            return NEIPv4Route(destinationAddress: ipAddress, subnetMask: "255.255.255.0")
        }
        print("Ошибка преобразования IP-адреса")
        return nil
//        if let _ = IPv4Address(ipAddress) {
//            let route = NEIPv4Route(destinationAddress: ipAddress, subnetMask: "255.255.255.255")
//            return route
//        } else {
//            print("Ошибка преобразования IP-адреса")
//            return nil
//        }
    }
    
    private func ipv6IncludedAddress() -> NEIPv6Route? {
        guard let appGroupDefaults = UserDefaults(suiteName: "group.ramil.Gadzhiev.plus7"),
              let ipAddress = appGroupDefaults.string(forKey: "IncludedIPAddress") else {
            return nil
        }
        var sin6 = sockaddr_in6()
        if ipAddress.withCString({ cstring in inet_pton(AF_INET6, cstring, &sin6.sin6_addr) }) == 1 {
            // IPv6 peer.
            return NEIPv6Route(destinationAddress: ipAddress, networkPrefixLength: 128)
        }
        print("Ошибка преобразования IP-адреса")
        return nil
//        if let _ = IPv6Address(ipAddress) {
//            let route = NEIPv6Route(destinationAddress: ipAddress, networkPrefixLength: 128)
//            return route
//        } else {
//            print("Ошибка преобразования IP-адреса")
//            return nil
//        }
    }
    
    private func setupSocketConnection(_ address: String) {
        if let url = URL(string: address) {
            socketManager = SocketManager(socketURL: url, config: [.log(true), .compress])
            socket = socketManager?.defaultSocket
            
            socket?.on("in") { [weak self] dataArray, ack in
                guard let self = self else { return }
                if let dataStr = dataArray.first as? String {
                    if let data = NSData(base64Encoded: dataStr, options: []) as? Data {
                        self.writePacketToDevice(data)
                    }
                } else if let data = dataArray.first as? Data {
                    self.writePacketToDevice(data)
                }
            }
            socket?.connect()
        }
    }
    
    private func startPacketProcessing() {
        packetFlow.readPackets { [weak self] packets, protocols in
            guard let self = self else { return }
            
            for packet in packets {
                writePacketToFile(packet)
                 self.socket?.emit("out", [UInt8](packet))
            }
            self.startPacketProcessing()
        }
    }
    
    private func writePacketToDevice(_ packet: Data) {
        writePacketToFile(packet)
         packetFlow.writePackets([packet], withProtocols: [AF_INET as NSNumber])
    }
    
    private func writePacketToFile(_ packet: Data) {
        
        if let userDefaults = UserDefaults(suiteName: "group.ramil.Gadzhiev.plus7") {
            let byteArray = [UInt8](packet)
            if byteArray.count >= 12 {
                let sequenceNumber = byteArray[4...7].withUnsafeBytes {
                    $0.load(as: UInt32.self).bigEndian  // Конвертируем из Big-Endian в правильный формат
                }
                
                let acknowledgementNumber = packet[8...11].withUnsafeBytes {
                    $0.load(as: UInt32.self).bigEndian  // Конвертируем из Big-Endian в правильный формат
                }
                
                let text = "Sequence Number: \(sequenceNumber)" + ("\n") + "Acknowledgement Number: \(acknowledgementNumber)"
                
                let existingLogs = (userDefaults.value(forKey: "receivedBytes") as? String) ?? ""
                let log = existingLogs + (existingLogs.isEmpty ? "" : "\n\n") + text
                userDefaults.set(log, forKey: "receivedBytes")
                userDefaults.synchronize()
            }
            
            
//            let packetString = byteArray.map { String(format: "%02x", $0) }.joined(separator: " ")
//            let existingLogs = (userDefaults.value(forKey: "receivedBytes") as? String) ?? ""
//            let log = existingLogs + (existingLogs.isEmpty ? "" : "\n\n") + packetString
//            userDefaults.set(log, forKey: "receivedBytes")
//            userDefaults.synchronize()  // Необязательно, но можно вызвать для немедленного сохранения
        }
//        if let appGroupDirectory = FileManager.default.containerURL(forSecurityApplicationGroupIdentifier: "group.ramil.Gadzhiev.plus7") {
//            let fileURL = appGroupDirectory.appendingPathComponent("packets.txt")
//
//            if FileManager.default.fileExists(atPath: fileURL.path()) {
//                let fileHandle = try? FileHandle(forWritingTo: fileURL)
//                fileHandle?.seekToEndOfFile()
//                if let data = (packetString + "\n\n").data(using: .utf8) {
//                    fileHandle?.write(data)
//                }
//                fileHandle?.closeFile()
//            } else {
//                try? packetString.write(to: fileURL, atomically: true, encoding: .utf8)
//            }
//        }
    }
    
}
