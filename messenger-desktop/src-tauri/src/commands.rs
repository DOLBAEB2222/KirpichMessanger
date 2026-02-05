use tauri::command;

#[command]
pub async fn login(_username: String, _password: String) -> Result<String, String> {
    // TODO: Implement API call to backend
    Ok("login_token".to_string())
}

#[command]
pub async fn send_message(_chat_id: String, _content: String) -> Result<(), String> {
    // TODO: Implement API call to backend
    Ok(())
}

#[command]
pub async fn get_chats() -> Result<String, String> {
    // TODO: Implement API call to backend
    Ok("[]".to_string())
}

#[command]
pub async fn upload_media(_file_path: String) -> Result<String, String> {
    // TODO: Implement file upload
    Ok("uploaded_file_url".to_string())
}
