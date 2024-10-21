//
//  ContentViewModel.swift
//  plus7
//
//  Created by Рамил Гаджиев on 06.09.2024.
//

import Foundation
import NetworkExtension

class ContentViewModel: ObservableObject {
    
    @Published var vpnStatus: NEVPNStatus = .disconnected
    @Published var error: Error?
    private let manager: VPNManager
    
    init(manager: VPNManager = VPNManager()) {
        self.manager = manager
        configureVPN()
    }

    //MARK: - Interface
    
    var ipWebSocketAddress: String {
        valueFromAppGroupDefaults(for: "WebSocketIPAddress")
    }
    
    var ipIncludedAddress: String {
        valueFromAppGroupDefaults(for: "IncludedIPAddress")
    }
    
    func configureVPN() {
        Task {
            do {
                await updateError(nil)
                try await manager.configureVPN()
                let connection = manager.vpnManager?.connection
                await updateVPNStatus(connection?.status ?? .disconnected)
                observeToVPNManagerConnection(connection)
            } catch let error {
                await updateError(error)
            }
        }
    }

    func startVPN() {
        Task {
            do {
                await updateError(nil)
                try await manager.startVPN()
            } catch let error {
                await updateError(error)
            }
        }
    }

    func stopVPN() {
        manager.stopVPN()
    }
    
    func saveIPAddresses(_ ipWebSocketAddress: String, _ ipIncludedAddress: String) {
        guard let appGroupDefaults = UserDefaults(suiteName: "group.ramil.Gadzhiev.plus7") else {
            print("Не удалось получить доступ к App Group")
            return
        }
        
        let ipIncludedAddress = extractIPAddress(from: ipIncludedAddress)
        appGroupDefaults.set(ipWebSocketAddress, forKey: "WebSocketIPAddress")
        appGroupDefaults.set(ipIncludedAddress, forKey: "IncludedIPAddress")
        print("IP-адреса сохранены: \(ipWebSocketAddress) \(ipIncludedAddress)")
    }
    
    //MARK: - Private
    
    private func valueFromAppGroupDefaults(for key: String) -> String {
        guard let appGroupDefaults = UserDefaults(suiteName: "group.ramil.Gadzhiev.plus7"),
              let value = appGroupDefaults.string(forKey: key) else {
            return ""
        }
        return value
    }
    
    private func extractIPAddress(from urlString: String) -> String {
        let pattern = #"(?:(?<=://)|^)[\d\.]+(?=(:\d+|/|$))"#
        do {
            let regex = try NSRegularExpression(pattern: pattern)
            let range = NSRange(urlString.startIndex..<urlString.endIndex, in: urlString)
            if let match = regex.firstMatch(in: urlString, options: [], range: range) {
                if let range = Range(match.range, in: urlString) {
                    return String(urlString[range])
                }
            }
        } catch {
            print("Invalid regex: \(error)")
        }
        
        return urlString
    }
    
    private func observeToVPNManagerConnection(_ connection: NEVPNConnection?) {
        NotificationCenter.default.addObserver(forName: .NEVPNStatusDidChange, object: connection, queue: .main) { [weak self, weak connection] _ in
            self?.vpnStatus = connection?.status ?? .disconnected
        }
    }
    
    @MainActor private func updateError(_ error: Error?) {
        self.error = error
    }

    @MainActor private func updateVPNStatus(_ status: NEVPNStatus) {
        vpnStatus = status
    }
}
