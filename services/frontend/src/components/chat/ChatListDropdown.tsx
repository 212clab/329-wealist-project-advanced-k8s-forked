// src/components/layout/chat/ChatListDropdown.tsx

import React, { useRef, useEffect } from 'react';
import { X } from 'lucide-react';
import type { Chat } from '../../types/chat';

interface ChatListDropdownProps {
  chats: Chat[];
  onChatSelect: (chatId: string) => void;
  onClose: () => void;
}

export const ChatListDropdown: React.FC<ChatListDropdownProps> = ({
  chats,
  onChatSelect,
  onClose,
}) => {
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [onClose]);

  return (
    <div
      ref={dropdownRef}
      className="fixed left-20 top-20 w-72 bg-white shadow-2xl rounded-lg z-50 border border-gray-200 max-h-96 overflow-y-auto"
    >
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h3 className="font-bold text-gray-900">채팅방</h3>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700">
            <X className="w-5 h-5" />
          </button>
        </div>
      </div>

      <div className="p-2">
        {chats.length === 0 ? (
          <div className="text-center py-8 text-gray-500 text-sm">채팅방이 없습니다</div>
        ) : (
          chats.map((chat) => (
            <button
              key={chat.chatId}
              onClick={() => onChatSelect(chat.chatId)}
              className="w-full text-left p-3 hover:bg-gray-100 rounded transition"
            >
              <div className="font-semibold text-sm text-gray-900">{chat.chatName || '채팅방'}</div>
              <div className="text-xs text-gray-500 mt-1">
                {chat.chatType === 'PROJECT' && '프로젝트 채팅'}
                {chat.chatType === 'GROUP' && '그룹 채팅'}
                {chat.chatType === 'DM' && 'DM'}
              </div>
            </button>
          ))
        )}
      </div>
    </div>
  );
};
