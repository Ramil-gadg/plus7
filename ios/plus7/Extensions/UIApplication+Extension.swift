//
//  UIApplication+Extension.swift
//  plus7
//
//  Created by Рамил Гаджиев on 08.09.2024.
//

#if os(iOS)
import UIKit

extension UIApplication {
    func endEditing() {
        sendAction(#selector(UIResponder.resignFirstResponder), to: nil, from: nil, for: nil)
    }
}
#endif
