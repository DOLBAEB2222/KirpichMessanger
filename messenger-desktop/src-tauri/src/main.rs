// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod system_tray;

use commands::*;

fn main() {
    tauri::Builder::default()
        .system_tray(system_tray::build_tray())
        .on_system_tray_event(|app, event| system_tray::handle_tray_event(app, event))
        .invoke_handler(tauri::generate_handler![
            login,
            send_message,
            get_chats,
            upload_media
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
