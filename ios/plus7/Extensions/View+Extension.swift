//
//  View+Extension.swift
//  plus7
//
//  Created by Рамил Гаджиев on 08.09.2024.
//

#if os(iOS)
import SwiftUI

//MARK: - hideKeyboardOnTap
extension View {
    func hideKeyboardOnTap() -> some View {
        self.modifier(HideKeyboardOnTapModifier())
    }
}

struct HideKeyboardOnTapModifier: ViewModifier {
    func body(content: Content) -> some View {
        ZStack {
            Color.clear
                .contentShape(Rectangle())
                .onTapGesture {
                    UIApplication.shared.endEditing()
                }
            content
        }
    }
}
#endif
