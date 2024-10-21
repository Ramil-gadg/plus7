//
//  WriteListeningPacketsService.swift
//  plus7
//
//  Created by Рамил Гаджиев on 07.09.2024.
//

import Foundation

enum WriteListeningPacketsService {
    
    static func logText() {
        if let appGroupDirectory = FileManager.default.containerURL(forSecurityApplicationGroupIdentifier: "group.ramil.Gadzhiev.plus7"),
           let documentsDirectory = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first {
            
            let appGroupFileURL = appGroupDirectory.appendingPathComponent("packets.txt")
            let documentsFileURL = documentsDirectory.appendingPathComponent("packets.txt")
            
            do {
                // Копируем файл из App Group в Documents
                try FileManager.default.copyItem(at: appGroupFileURL, to: documentsFileURL)
                print("Файл скопирован в Documents: \(documentsFileURL)")
            } catch {
                print("Ошибка копирования файла: \(error)")
            }
        }
    }
    
    static func copyFileToDocuments() {
        if let appGroupDirectory = FileManager.default.containerURL(forSecurityApplicationGroupIdentifier: "group.ramil.Gadzhiev.plus7"),
           let documentsDirectory = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first {
            
            let appGroupFileURL = appGroupDirectory.appendingPathComponent("packets.txt")
            let documentsFileURL = documentsDirectory.appendingPathComponent("packets.txt")
            
            do {
                // Копируем файл из App Group в Documents
                try FileManager.default.copyItem(at: appGroupFileURL, to: documentsFileURL)
                print("Файл скопирован в Documents: \(documentsFileURL)")
            } catch {
                print("Ошибка копирования файла: \(error)")
            }
        }
    }
    
    static func deleteFiles() {
        // Удаление файла из App Group
        if let appGroupDirectory = FileManager.default.containerURL(forSecurityApplicationGroupIdentifier: "group.com.yourcompany.vpn") {
            let appGroupFileURL = appGroupDirectory.appendingPathComponent("packets.txt")
            
            if FileManager.default.fileExists(atPath: appGroupFileURL.path) {
                do {
                    try FileManager.default.removeItem(at: appGroupFileURL)
                    print("Файл удален из App Group: \(appGroupFileURL.path)")
                } catch {
                    print("Ошибка удаления файла из App Group: \(error)")
                }
            } else {
                print("Файл не найден в App Group")
            }
        }
        
        // Удаление файла из Documents
//        if let documentsDirectory = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first {
//            let documentsFileURL = documentsDirectory.appendingPathComponent("packets.txt")
//            
//            if FileManager.default.fileExists(atPath: documentsFileURL.path) {
//                do {
//                    try FileManager.default.removeItem(at: documentsFileURL)
//                    print("Файл удален из Documents: \(documentsFileURL.path)")
//                } catch {
//                    print("Ошибка удаления файла из Documents: \(error)")
//                }
//            } else {
//                print("Файл не найден в Documents")
//            }
//        }
    }
}
