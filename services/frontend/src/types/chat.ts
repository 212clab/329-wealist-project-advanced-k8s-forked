// src/types/chat.ts

/**
 * @summary Chat 관련 타입 정의
 */

// =======================================================
// Chat Types
// =======================================================

export type ChatType = 'DM' | 'GROUP' | 'PROJECT';
export type MessageType = 'TEXT' | 'IMAGE' | 'FILE';

/**
 * @summary 채팅방 응답
 */
export interface Chat {
  chatId: string;
  workspaceId: string;
  projectId?: string;
  chatType: ChatType;
  chatName?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  participants?: ChatParticipant[];
  unreadCount?: number;
}

/**
 * @summary 채팅방 참여자
 */
export interface ChatParticipant {
  participantId: string;
  chatId: string;
  userId: string;
  joinedAt: string;
  lastReadAt: string;
  isActive: boolean;
}

/**
 * @summary 메시지 응답
 */
export interface Message {
  messageId: string;
  chatId: string;
  userId: string;
  userName?: string;
  userProfileImage?: string;
  content: string;
  messageType: MessageType;
  fileUrl?: string;
  fileName?: string;
  fileSize?: number;
  createdAt: string;
  updatedAt: string;
  reads?: MessageRead[];
  isMine?: boolean; // 프론트엔드 전용
}

/**
 * @summary 메시지 읽음 처리
 */
export interface MessageRead {
  readId: string;
  messageId: string;
  userId: string;
  readAt: string;
}

/**
 * @summary 채팅방 생성 요청
 */
export interface CreateChatRequest {
  workspaceId: string;
  projectId?: string;
  chatType: ChatType;
  chatName: string;
  participants: string[];
}

/**
 * @summary 메시지 전송 요청
 */
export interface SendMessageRequest {
  content: string;
  messageType?: MessageType;
  fileUrl?: string;
  fileName?: string;
  fileSize?: number;
}

/**
 * @summary WebSocket 메시지
 */
export interface WSChatMessage {
  type:
    | 'MESSAGE'
    | 'TYPING_START'
    | 'TYPING_STOP'
    | 'READ_MESSAGE'
    | 'USER_JOINED'
    | 'USER_LEFT'
    | 'MESSAGE_RECEIVED'
    | 'USER_TYPING'
    | 'MESSAGE_READ';
  chatId?: string;
  userId?: string;
  userName?: string;
  content?: string;
  messageType?: MessageType;
  fileUrl?: string;
  fileName?: string;
  fileSize?: number;
  messageId?: string;
  timestamp?: string;
  payload?: any;
}
