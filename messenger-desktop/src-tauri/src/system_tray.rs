use tauri::{AppHandle, CustomMenuItem, Manager, SystemTray, SystemTrayEvent, SystemTrayMenu, SystemTrayMenuItem};

use crate::window;

pub fn build_tray() -> SystemTray {
    let show = CustomMenuItem::new("show".to_string(), "Show");
    let hide = CustomMenuItem::new("hide".to_string(), "Hide");
    let quit = CustomMenuItem::new("quit".to_string(), "Quit");

    let menu = SystemTrayMenu::new()
        .add_item(show)
        .add_item(hide)
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(quit);

    SystemTray::new().with_menu(menu)
}

pub fn handle_tray_event(app: &AppHandle, event: SystemTrayEvent) {
    match event {
        SystemTrayEvent::MenuItemClick { id, .. } => match id.as_str() {
            "show" => {
                if let Some(window) = window::get_main_window(app) {
                    let _ = window.show();
                    let _ = window.set_focus();
                } else {
                    let _ = window::create_main_window(app);
                }
            }
            "hide" => {
                if let Some(window) = window::get_main_window(app) {
                    let _ = window.hide();
                }
            }
            "quit" => {
                app.exit(0);
            }
            _ => {}
        },
        SystemTrayEvent::DoubleClick { .. } => {
            if let Some(window) = window::get_main_window(app) {
                let _ = window.show();
                let _ = window.set_focus();
            }
        }
        _ => {}
    }
}
