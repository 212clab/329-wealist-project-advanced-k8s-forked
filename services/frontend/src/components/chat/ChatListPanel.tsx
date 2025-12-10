// src/components/chat/ChatListPanel.tsx

import React, { useState, useEffect } from 'react';
import { X, Search, MessageCircle, Users, Plus, Check, ArrowLeft } from 'lucide-react';
import { getMyChats, createChat } from '../../api/chatService';
import { getWorkspaceMembers } from '../../api/userService';
import type { Chat } from '../../types/chat';
import type { WorkspaceMemberResponse } from '../../types/user';

interface ChatListPanelProps {
  workspaceId: string;
  onChatSelect: (chatId: string) => void;
  onClose: () => void;
  onChatCreated?: () => void; // 🔥 채팅방 생성 후 콜백
  onUnreadCountChange?: (count: number) => void; // 🔥 읽지 않은 메시지 수 변경 콜백
}

// 🔥 색상 헬퍼
const getColorByIndex = (index: number) => {
  const colors = ['bg-indigo-500', 'bg-pink-500', 'bg-green-500', 'bg-purple-500', 'bg-yellow-500'];
  return colors[index % colors.length];
};

export const ChatListPanel: React.FC<ChatListPanelProps> = ({
  workspaceId,
  onChatSelect,
  onClose,
  onChatCreated,
  onUnreadCountChange,
}) => {
  const [chats, setChats] = useState<Chat[]>([]);
  const [members, setMembers] = useState<WorkspaceMemberResponse[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  // 🔥 그룹 채팅 생성 모드
  const [isCreateMode, setIsCreateMode] = useState(false);
  const [selectedMembers, setSelectedMembers] = useState<string[]>([]);
  const [groupName, setGroupName] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  const currentUserId = localStorage.getItem('userId');

  // 채팅방 목록 & 멤버 로드
  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true);
      try {
        // 병렬로 로드
        const [allChats, workspaceMembers] = await Promise.all([
          getMyChats(),
          getWorkspaceMembers(workspaceId),
        ]);

        console.log('📋 [ChatList] 전체 채팅방:', allChats);
        console.log('📋 [ChatList] 멤버:', workspaceMembers.length);

        // 워크스페이스 필터링
        const filteredChats = allChats.filter(
          (chat) => String(chat.workspaceId) === String(workspaceId)
        );

        setChats(filteredChats);
        setMembers(workspaceMembers);

        // 🔥 총 읽지 않은 메시지 수 계산 후 부모에게 알림
        const totalUnread = filteredChats.reduce((sum, chat) => sum + (chat.unreadCount || 0), 0);
        onUnreadCountChange?.(totalUnread);
      } catch (error) {
        console.error('Failed to load chats:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, [workspaceId]);

  // 검색 필터링
  const filteredChats = chats.filter((chat) => {
    if (!searchQuery) return true;
    const chatName = chat.chatName || '';
    return chatName.toLowerCase().includes(searchQuery.toLowerCase());
  });

  // 🔥 참여자에서 상대방 찾기 (DM용)
  const getOtherParticipant = (chat: Chat): WorkspaceMemberResponse | undefined => {
    if (!chat.participants) return undefined;
    const otherUserId = chat.participants.find((p) => p.userId !== currentUserId)?.userId;
    return members.find((m) => m.userId === otherUserId);
  };

  // 🔥 참여자 목록 (본인 제외)
  const getOtherParticipants = (chat: Chat): WorkspaceMemberResponse[] => {
    if (!chat.participants) return [];
    const otherUserIds = chat.participants
      .filter((p) => p.userId !== currentUserId)
      .map((p) => p.userId);
    return members.filter((m) => otherUserIds.includes(m.userId));
  };

  // 🔥 채팅방 이름 (DM은 상대방 이름, 그룹은 chatName 또는 참여자 이름)
  const getChatDisplayName = (chat: Chat): string => {
    if (chat.chatType === 'DM') {
      const other = getOtherParticipant(chat);
      return other?.nickName || other?.userEmail || '알 수 없음';
    }
    if (chat.chatName) return chat.chatName;
    const others = getOtherParticipants(chat);
    if (others.length === 0) return '그룹 채팅';
    if (others.length <= 3) return others.map((m) => m.nickName || m.userEmail || 'Unknown').join(', ');
    return `${others.slice(0, 2).map((m) => m.nickName || m.userEmail || 'Unknown').join(', ')} 외 ${others.length - 2}명`;
  };

  // 마지막 메시지 시간 포맷
  const formatTime = (date: string) => {
    const now = new Date();
    const messageDate = new Date(date);
    const diffMs = now.getTime() - messageDate.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return '방금 전';
    if (diffMins < 60) return `${diffMins}분 전`;
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)}시간 전`;
    return messageDate.toLocaleDateString('ko-KR', { month: 'short', day: 'numeric' });
  };

  // 🔥 멤버 선택 토글
  const toggleMemberSelection = (userId: string) => {
    setSelectedMembers((prev) =>
      prev.includes(userId)
        ? prev.filter((id) => id !== userId)
        : [...prev, userId]
    );
  };

  // 🔥 그룹 채팅 생성
  const handleCreateGroupChat = async () => {
    if (selectedMembers.length === 0) {
      alert('최소 1명 이상 선택해주세요.');
      return;
    }

    setIsCreating(true);
    try {
      const chatType = selectedMembers.length === 1 ? 'DM' : 'GROUP';
      const chat = await createChat({
        workspaceId,
        chatType,
        chatName: groupName || (chatType === 'DM' ? 'DM' : 'Group Chat'),
        participants: selectedMembers,
      });

      console.log('✅ 채팅방 생성 완료:', chat);

      // 채팅 목록 새로고침
      const allChats = await getMyChats();
      const filteredChats = allChats.filter(
        (c) => String(c.workspaceId) === String(workspaceId)
      );
      setChats(filteredChats);

      // 초기화
      setIsCreateMode(false);
      setSelectedMembers([]);
      setGroupName('');

      // 생성된 채팅방으로 이동
      onChatSelect(chat.chatId);
      onChatCreated?.();
    } catch (error) {
      console.error('❌ 채팅방 생성 실패:', error);
      alert('채팅방 생성에 실패했습니다.');
    } finally {
      setIsCreating(false);
    }
  };

  // 🔥 본인 제외한 멤버 목록
  const otherMembers = members.filter((m) => m.userId !== currentUserId);

  // 🔥 아바타 렌더링
  const renderAvatar = (chat: Chat) => {
    if (chat.chatType === 'DM') {
      // DM: 상대방 프로필 사진
      const other = getOtherParticipant(chat);
      return (
        <div className="relative flex-shrink-0">
          {other?.profileImageUrl ? (
            <img
              src={other.profileImageUrl}
              alt={other?.nickName || other?.userEmail}
              className="w-12 h-12 rounded-full object-cover"
            />
          ) : (
            <div
              className={`w-12 h-12 rounded-full flex items-center justify-center text-white font-bold ${getColorByIndex(0)}`}
            >
              {other?.nickName?.[0] || other?.userEmail?.[0] || '?'}
            </div>
          )}
          {/* 읽지 않은 메시지 빨간 점 */}
          {(chat.unreadCount ?? 0) > 0 && (
            <div className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center font-bold">
              {(chat.unreadCount ?? 0) > 9 ? '9+' : chat.unreadCount}
            </div>
          )}
        </div>
      );
    }

    // 그룹/프로젝트: Avatar Stack (컴팩트하게)
    const others = getOtherParticipants(chat).slice(0, 3);
    return (
      <div className="relative flex-shrink-0 w-12 h-12">
        <div className="flex -space-x-4">
          {others.length === 0 ? (
            <div className="w-12 h-12 rounded-full bg-gray-300 flex items-center justify-center">
              <Users className="w-6 h-6 text-gray-500" />
            </div>
          ) : (
            others.map((member, index) => (
              <div
                key={member.userId}
                className="w-8 h-8 rounded-full ring-2 ring-white overflow-hidden"
                style={{ zIndex: others.length - index }}
              >
                {member.profileImageUrl ? (
                  <img
                    src={member.profileImageUrl}
                    alt={member.nickName || member.userEmail}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div
                    className={`w-full h-full flex items-center justify-center text-white text-xs font-bold ${getColorByIndex(index)}`}
                  >
                    {member.nickName?.[0] || member.userEmail?.[0] || '?'}
                  </div>
                )}
              </div>
            ))
          )}
        </div>
        {/* 읽지 않은 메시지 빨간 점 */}
        {(chat.unreadCount ?? 0) > 0 && (
          <div className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center font-bold z-10">
            {(chat.unreadCount ?? 0) > 9 ? '9+' : chat.unreadCount}
          </div>
        )}
      </div>
    );
  };

  // 🔥 그룹 채팅 생성 모드 UI
  if (isCreateMode) {
    return (
      <div className="h-full w-full bg-white flex flex-col">
        {/* 헤더 */}
        <div className="p-4 border-b bg-gradient-to-r from-blue-600 to-blue-700 text-white">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <button
                onClick={() => {
                  setIsCreateMode(false);
                  setSelectedMembers([]);
                  setGroupName('');
                }}
                className="p-1 hover:bg-white/20 rounded transition"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
              <h2 className="font-bold text-lg">새 채팅</h2>
            </div>
            <button
              onClick={onClose}
              className="p-1 hover:bg-white/20 rounded transition"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* 그룹 이름 입력 (2명 이상 선택 시) */}
          {selectedMembers.length > 1 && (
            <input
              type="text"
              placeholder="그룹 이름 (선택사항)"
              value={groupName}
              onChange={(e) => setGroupName(e.target.value)}
              className="w-full px-4 py-2 bg-white/20 text-white placeholder-white/60 rounded-lg focus:outline-none focus:ring-2 focus:ring-white/50"
            />
          )}
        </div>

        {/* 선택된 멤버 표시 */}
        {selectedMembers.length > 0 && (
          <div className="p-3 bg-blue-50 border-b flex items-center gap-2 flex-wrap">
            <span className="text-xs text-blue-600 font-medium">선택됨:</span>
            {selectedMembers.map((userId) => {
              const member = members.find((m) => m.userId === userId);
              return (
                <span
                  key={userId}
                  className="px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded-full flex items-center gap-1"
                >
                  {member?.nickName || member?.userEmail || '알 수 없음'}
                  <button
                    onClick={() => toggleMemberSelection(userId)}
                    className="hover:text-blue-900"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </span>
              );
            })}
          </div>
        )}

        {/* 멤버 목록 */}
        <div className="flex-1 overflow-y-auto">
          <div className="divide-y">
            {otherMembers.map((member, index) => {
              const isSelected = selectedMembers.includes(member.userId);
              return (
                <button
                  key={member.userId}
                  onClick={() => toggleMemberSelection(member.userId)}
                  className={`w-full p-4 transition text-left flex items-center gap-3 ${
                    isSelected ? 'bg-blue-50' : 'hover:bg-gray-50'
                  }`}
                >
                  {/* 아바타 */}
                  <div className="relative flex-shrink-0">
                    {member.profileImageUrl ? (
                      <img
                        src={member.profileImageUrl}
                        alt={member.nickName || member.userEmail}
                        className="w-10 h-10 rounded-full object-cover"
                      />
                    ) : (
                      <div
                        className={`w-10 h-10 rounded-full flex items-center justify-center text-white font-bold ${getColorByIndex(index)}`}
                      >
                        {member.nickName?.[0] || member.userEmail?.[0] || '?'}
                      </div>
                    )}
                  </div>

                  {/* 이름 */}
                  <div className="flex-1">
                    <p className="font-medium text-sm text-gray-900">{member.nickName || member.userEmail || 'Unknown'}</p>
                    <p className="text-xs text-gray-500">{member.userEmail}</p>
                  </div>

                  {/* 체크박스 */}
                  <div
                    className={`w-6 h-6 rounded-full border-2 flex items-center justify-center transition ${
                      isSelected
                        ? 'bg-blue-500 border-blue-500 text-white'
                        : 'border-gray-300'
                    }`}
                  >
                    {isSelected && <Check className="w-4 h-4" />}
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* 생성 버튼 */}
        <div className="p-4 border-t bg-gray-50">
          <button
            onClick={handleCreateGroupChat}
            disabled={selectedMembers.length === 0 || isCreating}
            className="w-full py-3 bg-blue-500 text-white font-medium rounded-lg hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed transition flex items-center justify-center gap-2"
          >
            {isCreating ? (
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white" />
            ) : (
              <>
                <MessageCircle className="w-5 h-5" />
                {selectedMembers.length === 1 ? '1:1 대화 시작' : `그룹 채팅 시작 (${selectedMembers.length}명)`}
              </>
            )}
          </button>
        </div>
      </div>
    );
  }

  // 🔥 기본 채팅 리스트 UI
  return (
    <div className="h-full w-full bg-white flex flex-col">
      {/* 헤더 */}
      <div className="p-4 border-b bg-gradient-to-r from-blue-600 to-blue-700 text-white">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <MessageCircle className="w-5 h-5" />
            <h2 className="font-bold text-lg">채팅</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 hover:bg-white/20 rounded transition"
            title="닫기"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 검색바 */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="채팅방 검색..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-white/20 text-white placeholder-white/60 rounded-lg focus:outline-none focus:ring-2 focus:ring-white/50"
          />
        </div>
      </div>

      {/* 채팅 리스트 */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
          </div>
        ) : filteredChats.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-gray-400">
            <MessageCircle className="w-12 h-12 mb-3 opacity-50" />
            <p className="text-sm">{searchQuery ? '검색 결과가 없습니다' : '채팅방이 없습니다'}</p>
            <p className="text-xs mt-1">아래 + 버튼으로 채팅을 시작하세요</p>
          </div>
        ) : (
          <div className="divide-y">
            {filteredChats.map((chat) => (
              <button
                key={chat.chatId}
                onClick={() => onChatSelect(chat.chatId)}
                className="w-full p-4 hover:bg-gray-50 transition text-left"
              >
                <div className="flex items-center gap-3">
                  {/* 🔥 프로필 아바타 */}
                  {renderAvatar(chat)}

                  {/* 내용 */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between mb-1">
                      <h3 className="font-semibold text-sm text-gray-900 truncate">
                        {getChatDisplayName(chat)}
                      </h3>
                      <span className="text-xs text-gray-400 flex-shrink-0 ml-2">
                        {formatTime(chat.updatedAt)}
                      </span>
                    </div>

                    <p className="text-xs text-gray-500 truncate">
                      {chat.chatType === 'DM' && '1:1 대화'}
                      {chat.chatType === 'GROUP' && `그룹 채팅 · ${getOtherParticipants(chat).length + 1}명`}
                      {chat.chatType === 'PROJECT' && `프로젝트 채팅`}
                    </p>
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* 🔥 새 채팅 버튼 */}
      <div className="p-4 border-t bg-gray-50">
        <button
          onClick={() => setIsCreateMode(true)}
          className="w-full py-3 bg-blue-500 text-white font-medium rounded-lg hover:bg-blue-600 transition flex items-center justify-center gap-2"
        >
          <Plus className="w-5 h-5" />
          새 채팅 시작
        </button>
      </div>
    </div>
  );
};
