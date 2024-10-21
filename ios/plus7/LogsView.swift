//
//  LogsView.swift
//  plus7
//
//  Created by Рамил Гаджиев on 02.10.2024.
//

import SwiftUI

struct LogsView: View {
    
    @AppStorage("receivedBytes", store: UserDefaults(suiteName: "group.ramil.Gadzhiev.plus7")) var logs: String = ""
    @Environment(\.dismiss) var dismiss
    
    private func deleteLogs() {
        logs = ""
    }
    
    var body: some View {
        VStack {
            HStack {
                Spacer()
                Button {
                    dismiss()
                } label: {
                    Image(systemName: "xmark.circle")
                }
            }
            Button {
                deleteLogs()
            } label: {
                Text("Delete logs")
            }
            ScrollView(.vertical) {
                Text(logs.isEmpty ? "Пусто" : logs)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .padding()
    }
}

