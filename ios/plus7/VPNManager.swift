//
//  VPNManager.swift
//  plus7
//
//  Created by Рамил Гаджиев on 06.09.2024.
//

import NetworkExtension

class VPNManager {
    private (set) var vpnManager: NETunnelProviderManager?
//    private let writeListeningPacketsService: WriteListeningPacketsService
//    
//    init(writeListeningPacketsService: WriteListeningPacketsService = WriteListeningPacketsService()) {
//        self.writeListeningPacketsService = writeListeningPacketsService
//    }
    
    //MARK: - Interface
    
    func configureVPN() async throws {
        vpnManager = try await manager()
        print("VPN configured")
    }
    
    func startVPN() async throws {
        try await configureVPN()
//        writeListeningPacketsService.deleteFiles()
        try vpnManager?.connection.startVPNTunnel()
        print("VPN started successfully")
    }
    
    func stopVPN() {
        vpnManager?.connection.stopVPNTunnel()
//        writeListeningPacketsService.copyFileToDocuments()
        print("VPN stopped")
    }
    
    //MARK: - Private
    
    private func manager() async throws -> NETunnelProviderManager {
        let managers = try await NETunnelProviderManager.loadAllFromPreferences()
        if let manager = managers.first {
            return manager
        }
        let manager = try await configureManager()
        try await manager.loadFromPreferences()
        return manager
    }

    private func configureManager() async throws -> NETunnelProviderManager {
        let manager = NETunnelProviderManager()
        let tunnelProtocol = NETunnelProviderProtocol()
        tunnelProtocol.providerBundleIdentifier = "ramil.Gadzhiev.plus7.vpn-tunnel"
        tunnelProtocol.serverAddress = "10.0.0.1"
        tunnelProtocol.disconnectOnSleep = false
        
        manager.protocolConfiguration = tunnelProtocol
        manager.localizedDescription = "VPN Plus7"
        manager.isEnabled = true
        try await manager.saveToPreferences()
        return manager
    }
}
