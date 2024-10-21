//
//  plus7App.swift
//  plus7
//
//  Created by Рамил Гаджиев on 06.09.2024.
//

import SwiftUI
import SwiftData

@main
struct plus7App: App {
    
    var body: some Scene {
        WindowGroup {
            ContentView()
#if os(iOS)
                .hideKeyboardOnTap()
#endif
        }
    }
}
