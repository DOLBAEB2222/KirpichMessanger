use tauri::{command, State};

#[command]
pub async fn login(username: String, password: String) -> Result<String, String> {
    // TODO: Implement API call to backend
    Ok("login_token".to_string())
}

#[command]
pub async fn send_message(chat_id: String, content: String) -> Result<(), String> {
    // TODO: Implement API call to backend
    Ok(())
}

#[command]
pub async fn get_chats() -> Result<String, String> {
    // TODO: Implement API call to backend
    Ok("[]".to_string())
}

#[command]
pub async fn upload_media(file_path: String) -> Result<String, String> {
    // TODO: Implement file upload
    Ok("uploaded_file_url".to_string())
}
