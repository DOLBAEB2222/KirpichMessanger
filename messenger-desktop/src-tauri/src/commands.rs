use serde::{Deserialize, Serialize};
use tauri::command;

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AuthRequest {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AuthResponse {
    pub token: String,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SendMessageRequest {
    pub chat_id: String,
    pub message: String,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SendMessageResponse {
    pub message_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UploadMediaRequest {
    pub chat_id: String,
    pub file_name: String,
    pub bytes: Vec<u8>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UploadMediaResponse {
    pub media_url: String,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChatSummary {
    pub id: String,
    pub title: String,
    pub last_message: Option<String>,
    pub unread_count: u32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct NotificationPayload {
    pub title: String,
    pub body: String,
}

#[command]
pub async fn login(payload: AuthRequest) -> Result<AuthResponse, String> {
    if payload.email.trim().is_empty() || payload.password.trim().is_empty() {
        return Err("Email and password are required".to_string());
    }

    Ok(AuthResponse {
        token: "dev-token".to_string(),
    })
}

#[command]
pub async fn send_message(payload: SendMessageRequest) -> Result<SendMessageResponse, String> {
    if payload.message.trim().is_empty() {
        return Err("Message cannot be empty".to_string());
    }

    Ok(SendMessageResponse {
        message_id: format!("msg-{}", payload.chat_id),
    })
}

#[command]
pub async fn upload_media(payload: UploadMediaRequest) -> Result<UploadMediaResponse, String> {
    if payload.bytes.is_empty() {
        return Err("File payload is empty".to_string());
    }

    Ok(UploadMediaResponse {
        media_url: format!("https://media.kirpich.app/{}/{}", payload.chat_id, payload.file_name),
    })
}

#[command]
pub async fn get_chats() -> Result<Vec<ChatSummary>, String> {
    Ok(vec![ChatSummary {
        id: "general".to_string(),
        title: "General".to_string(),
        last_message: Some("Welcome to KirpichMessanger".to_string()),
        unread_count: 0,
    }])
}

#[command]
pub async fn handle_notifications(payload: NotificationPayload) -> Result<(), String> {
    if payload.title.trim().is_empty() {
        return Err("Notification title is required".to_string());
    }

    Ok(())
}
