//
//  NEVPNStatus+Extension.swift
//  plus7
//
//  Created by Рамил Гаджиев on 08.09.2024.
//

import NetworkExtension

extension NEVPNStatus {
    var description: String {
        switch self {
        case .invalid:
            return "Invalid"
        case .disconnected:
            return "Disconnected"
        case .connecting:
            return "Connecting"
        case .connected:
            return "Connected"
        case .reasserting:
            return "Reasserting"
        case .disconnecting:
            return "Disconnecting"
        @unknown default:
            return "Unknown"
        }
    }
}
