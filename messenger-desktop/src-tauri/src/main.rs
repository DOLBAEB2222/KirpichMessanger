#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod system_tray;
mod window;

use tauri::{CustomMenuItem, Manager, Menu, MenuItem, Submenu};

fn main() {
    let file_menu = Menu::new()
        .add_item(CustomMenuItem::new("settings".to_string(), "Settings"))
        .add_native_item(MenuItem::Separator)
        .add_item(CustomMenuItem::new("quit".to_string(), "Quit"));

    let view_menu = Menu::new()
        .add_item(CustomMenuItem::new("reload".to_string(), "Reload"))
        .add_item(CustomMenuItem::new("toggle_devtools".to_string(), "Toggle DevTools"));

    let menu = Menu::new()
        .add_submenu(Submenu::new("File", file_menu))
        .add_submenu(Submenu::new("View", view_menu));

    tauri::Builder::default()
        .system_tray(system_tray::build_tray())
        .on_system_tray_event(|app, event| system_tray::handle_tray_event(app, event))
        .menu(menu)
        .on_menu_event(|event| match event.menu_item_id() {
            "quit" => {
                event.window().app_handle().exit(0);
            }
            "settings" => {
                let _ = event.window().emit("open-settings", ());
            }
            "reload" => {
                let _ = event.window().reload();
            }
            "toggle_devtools" => {
                let window = event.window();
                if window.is_devtools_open() {
                    window.close_devtools();
                } else {
                    window.open_devtools();
                }
            }
            _ => {}
        })
        .invoke_handler(tauri::generate_handler![
            commands::login,
            commands::send_message,
            commands::upload_media,
            commands::get_chats
        ])
        .setup(|app| {
            window::create_main_window(app)?;
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
