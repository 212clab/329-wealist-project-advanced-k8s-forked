// src/components/layout/chat/ChatManager.tsx

import React, { useState, useEffect } from 'react';
import { ChatListDropdown } from './ChatListDropdown';
import { getMyChats } from '../../api/chatService';
import type { Chat } from '../../types/chat';
import { ChatPanel } from './ChatPanel';

interface ChatManagerProps {
  workspaceId: string;
  projectId?: string;
  isOpen: boolean;
  onClose: () => void;
}

export const ChatManager: React.FC<ChatManagerProps> = ({
  workspaceId,
  projectId,
  isOpen,
  onClose,
}) => {
  const [chats, setChats] = useState<Chat[]>([]);
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);
  const [showChatPanel, setShowChatPanel] = useState(false);

  useEffect(() => {
    if (isOpen) {
      loadChats();
    }
  }, [isOpen, workspaceId]);

  const loadChats = async () => {
    try {
      const chatList = await getMyChats();
      setChats(chatList);

      // 프로젝트 채팅방 자동 선택
      if (projectId && chatList.length > 0) {
        const projectChat = chatList.find((c) => c.projectId === projectId);
        if (projectChat) {
          setSelectedChatId(projectChat.chatId);
          setShowChatPanel(true);
        }
      }
    } catch (error) {
      console.error('채팅방 목록 로드 실패:', error);
    }
  };

  const handleChatSelect = (chatId: string) => {
    setSelectedChatId(chatId);
    setShowChatPanel(true);
  };

  const handleClosePanel = () => {
    setShowChatPanel(false);
    setSelectedChatId(null);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <>
      {/* 채팅방 목록 */}
      {!showChatPanel && (
        <ChatListDropdown chats={chats} onChatSelect={handleChatSelect} onClose={onClose} />
      )}

      {/* 채팅 패널 */}
      {showChatPanel && selectedChatId && (
        <ChatPanel chatId={selectedChatId} onClose={handleClosePanel} />
      )}
    </>
  );
};
