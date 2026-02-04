export interface User {
  id: string;
  username: string;
  avatarUrl?: string;
}

export interface Message {
  id: string;
  content: string;
  senderId: string;
  timestamp: number;
}

export interface Chat {
  id: string;
  name: string;
  participants: User[];
  lastMessage?: Message;
}
