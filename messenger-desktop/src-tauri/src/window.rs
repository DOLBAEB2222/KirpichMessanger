use tauri::{AppHandle, Manager, WindowBuilder, WindowUrl};

pub fn create_main_window(app: &AppHandle) -> tauri::Result<()> {
    let window = WindowBuilder::new(app, "main", WindowUrl::App("/".into()))
        .title("KirpichMessanger")
        .inner_size(1200.0, 800.0)
        .min_inner_size(960.0, 640.0)
        .build()?;

    window.show()?;
    Ok(())
}

pub fn get_main_window(app: &AppHandle) -> Option<tauri::Window> {
    app.get_window("main")
}
