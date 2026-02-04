import { invoke } from '@tauri-apps/api/tauri';

export type AuthPayload = {
  email: string;
  password: string;
};

export type AuthResponse = {
  token: string;
};

export type SendMessagePayload = {
  chatId: string;
  message: string;
};

export type UploadMediaPayload = {
  chatId: string;
  fileName: string;
  bytes: number[];
};

export type ChatSummary = {
  id: string;
  title: string;
  lastMessage?: string;
  unreadCount: number;
};

export const login = (payload: AuthPayload) =>
  invoke<AuthResponse>('login', { payload });

export const sendMessage = (payload: SendMessagePayload) =>
  invoke<{ messageId: string }>('send_message', { payload });

export const uploadMedia = (payload: UploadMediaPayload) =>
  invoke<{ mediaUrl: string }>('upload_media', { payload });

export const getChats = () => invoke<ChatSummary[]>('get_chats');

export const sendNotification = (title: string, body: string) =>
  invoke<void>('handle_notifications', { payload: { title, body } });

export class RealtimeClient {
  private socket: WebSocket | null = null;
  constructor(private url: string) {}

  connect(onMessage: (event: MessageEvent) => void) {
    this.socket = new WebSocket(this.url);
    this.socket.onmessage = onMessage;
  }

  send(data: string) {
    this.socket?.send(data);
  }

  close() {
    this.socket?.close();
    this.socket = null;
  }
}
