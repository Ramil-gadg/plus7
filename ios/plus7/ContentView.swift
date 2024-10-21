//
//  ContentView.swift
//  plus7
//
//  Created by Рамил Гаджиев on 06.09.2024.
//

import SwiftUI

struct ContentView: View {
    @StateObject var viewModel = ContentViewModel()
    
    @State private var ipWebSocketAddress = ""
    @State private var ipIncludedAddress = ""
    @State private var isPresentedLogs = false
    
    var isConnected: Bool {
        viewModel.vpnStatus == .connected || 
        viewModel.vpnStatus == .disconnecting
    }
    
    var inProcess: Bool {
        viewModel.vpnStatus == .connecting ||
        viewModel.vpnStatus == .disconnecting
    }

   var body: some View {
       VStack {
           if let error = viewModel.error {
               Text("Произошла ошибка")
                   .font(.title2)
                   .foregroundStyle(.red)
               Text(error.localizedDescription)
                   .foregroundStyle(.red)
               Button(action: {
                       viewModel.configureVPN()
               }) {
                   Text("Попробовать еще раз")
               }
               .padding()
               .background(Color.red)
               .foregroundColor(.white)
               .cornerRadius(10)
           } else {
               VStack(spacing: 60) {
                   VStack {
                       Group {
                           TextField("Введите IP-адрес для WebSocket", text: $ipWebSocketAddress)
                           TextField("Введите Included IP-адрес", text: $ipIncludedAddress)
                       }
                           .padding(.bottom)
                           .textFieldStyle(RoundedBorderTextFieldStyle())
#if os(iOS)
                           .keyboardType(.numbersAndPunctuation)
#endif
                       Button(action: {
                           viewModel.saveIPAddresses(ipWebSocketAddress, ipIncludedAddress)
                       }) {
                           Text("Сохранить")
                       }
                       .padding()
                       .background(Color.blue)
                       .foregroundColor(.white)
                       .cornerRadius(10)
                   }
                   
                   VStack {
                       Text("VPN Status: \(viewModel.vpnStatus.description)")
                       
                       Button(action: {
                           if viewModel.vpnStatus != .connected {
                               viewModel.startVPN()
                           } else {
                               viewModel.stopVPN()
                           }
                       }) {
                           Text(isConnected ? "Stop VPN" : "Start VPN")
                       }
                       .padding()
                       .background(isConnected ? Color.red : Color.green)
                       .foregroundColor(.white)
                       .opacity(inProcess ? 0.5 : 1)
                       .cornerRadius(10)
                   }
                   
                   Button {
                       isPresentedLogs.toggle()
                   } label: {
                       Text("Посмотреть логи")
                   }

               }
           }
       }
       .onAppear {
       ipWebSocketAddress = viewModel.ipWebSocketAddress
       ipIncludedAddress = viewModel.ipIncludedAddress
       }
       .sheet(isPresented: $isPresentedLogs) {
           LogsView()
       }
       .padding()
   }
}
