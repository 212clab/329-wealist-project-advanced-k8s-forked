import React, { useState, useEffect, useMemo } from 'react';
import {
  X,
  Plus,
  Video,
  Users,
  Phone,
  Clock,
  Calendar,
  LogOut,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { videoService, VideoRoom, CallHistory } from '../../api/videoService';
import { getWorkspaceMembers } from '../../api/userService';
import { WorkspaceMemberResponse } from '../../types/user';
import { CallHistoryDetailModal } from './CallHistoryDetailModal';

interface VideoCallPanelProps {
  workspaceId: string;
  userProfile: { id: string; nickName: string } | null;
  onClose: () => void;
  onJoinRoom: (room: VideoRoom, token: string, wsUrl: string) => void;
  currentRoomId?: string; // 현재 열린 VideoRoom의 ID
  onLeaveCurrentRoom?: () => void; // VideoRoom 닫기 콜백
}

type TabType = 'active' | 'history';

const formatDuration = (seconds: number): string => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}시간 ${minutes}분`;
  } else if (minutes > 0) {
    return `${minutes}분 ${secs}초`;
  }
  return `${secs}초`;
};

const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  const now = new Date();
  const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));

  if (diffDays === 0) {
    return `오늘 ${date.toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit' })}`;
  } else if (diffDays === 1) {
    return `어제 ${date.toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit' })}`;
  } else if (diffDays < 7) {
    return `${diffDays}일 전`;
  }
  return date.toLocaleDateString('ko-KR', { month: 'short', day: 'numeric' });
};

export const VideoCallPanel: React.FC<VideoCallPanelProps> = ({
  workspaceId,
  userProfile,
  onClose,
  onJoinRoom,
  currentRoomId,
  onLeaveCurrentRoom,
}) => {
  const { theme } = useTheme();
  const [activeTab, setActiveTab] = useState<TabType>('active');
  const [rooms, setRooms] = useState<VideoRoom[]>([]);
  const [callHistory, setCallHistory] = useState<CallHistory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isHistoryLoading, setIsHistoryLoading] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newRoomName, setNewRoomName] = useState('');
  const [maxParticipants, setMaxParticipants] = useState(10);
  const [isCreating, setIsCreating] = useState(false);
  const [isJoining, setIsJoining] = useState<string | null>(null);
  const [isLeaving, setIsLeaving] = useState<string | null>(null);
  const [expandedHistoryId, setExpandedHistoryId] = useState<string | null>(null);
  const [selectedHistory, setSelectedHistory] = useState<CallHistory | null>(null);
  const [members, setMembers] = useState<WorkspaceMemberResponse[]>([]);

  // userId -> 멤버 정보 맵
  const memberMap = useMemo(() => {
    const map = new Map<string, WorkspaceMemberResponse>();
    members.forEach((member) => {
      map.set(member.userId, member);
    });
    return map;
  }, [members]);

  // userId로 멤버 이름 가져오기
  const getMemberName = (userId: string): string => {
    const member = memberMap.get(userId);
    return member?.nickName || `사용자 ${userId.slice(0, 6)}`;
  };

  // userId로 멤버 프로필 이미지 가져오기
  const getMemberProfileImage = (userId: string): string | null => {
    const member = memberMap.get(userId);
    return member?.profileImageUrl || null;
  };

  // 현재 사용자가 방에 참여 중인지 확인
  const isUserInRoom = (room: VideoRoom): boolean => {
    if (!userProfile?.id) return false;
    return room.participants?.some((p) => p.userId === userProfile.id && p.isActive) || false;
  };

  useEffect(() => {
    loadRooms();
    loadMembers();
    // Poll for room updates
    const interval = setInterval(loadRooms, 5000);
    return () => clearInterval(interval);
  }, [workspaceId]);

  const loadMembers = async () => {
    try {
      const data = await getWorkspaceMembers(workspaceId);
      setMembers(data);
    } catch (error) {
      console.error('Failed to load members:', error);
    }
  };

  useEffect(() => {
    if (activeTab === 'history') {
      loadCallHistory();
    }
  }, [activeTab, workspaceId]);

  const loadRooms = async () => {
    try {
      const data = await videoService.getWorkspaceRooms(workspaceId, true);
      setRooms(data);
    } catch (error) {
      console.error('Failed to load rooms:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const loadCallHistory = async () => {
    setIsHistoryLoading(true);
    try {
      const response = await videoService.getWorkspaceCallHistory(workspaceId);
      setCallHistory(response.data || []);
    } catch (error) {
      console.error('Failed to load call history:', error);
    } finally {
      setIsHistoryLoading(false);
    }
  };

  const handleCreateRoom = async () => {
    if (!newRoomName.trim()) return;

    setIsCreating(true);
    try {
      const room = await videoService.createRoom({
        name: newRoomName.trim(),
        workspaceId,
        maxParticipants,
      });
      setRooms((prev) => [room, ...prev]);
      setShowCreateModal(false);
      setNewRoomName('');
      setMaxParticipants(10);

      // Auto-join the created room
      handleJoinRoom(room.id);
    } catch (error) {
      console.error('Failed to create room:', error);
      alert('방 생성에 실패했습니다.');
    } finally {
      setIsCreating(false);
    }
  };

  const handleJoinRoom = async (roomId: string) => {
    setIsJoining(roomId);
    try {
      const response = await videoService.joinRoom(roomId, userProfile?.nickName);
      onJoinRoom(response.room, response.token, response.wsUrl);
    } catch (error: any) {
      console.error('Failed to join room:', error);
      const errorCode = error.response?.data?.error?.code;
      if (errorCode === 'ROOM_FULL') {
        alert('방이 가득 찼습니다.');
      } else if (errorCode === 'ALREADY_IN_ROOM') {
        alert('이미 방에 참여 중입니다.');
      } else if (errorCode === 'ROOM_ENDED') {
        alert('종료된 방입니다.');
        loadRooms();
      } else {
        alert('방 참여에 실패했습니다.');
      }
    } finally {
      setIsJoining(null);
    }
  };

  const handleLeaveRoom = async (roomId: string) => {
    setIsLeaving(roomId);
    try {
      await videoService.leaveRoom(roomId);
      // 현재 열린 VideoRoom이면 닫기
      if (roomId === currentRoomId && onLeaveCurrentRoom) {
        onLeaveCurrentRoom();
      }
      // 방 목록 갱신
      loadRooms();
    } catch (error) {
      console.error('Failed to leave room:', error);
      alert('방 나가기에 실패했습니다.');
    } finally {
      setIsLeaving(null);
    }
  };

  return (
    <div
      className={`fixed top-0 right-0 w-80 h-full ${theme.colors.card} shadow-2xl z-50 flex flex-col border-l ${theme.colors.border}`}
    >
      {/* Header */}
      <div className={`p-4 border-b ${theme.colors.border} flex items-center justify-between`}>
        <div className="flex items-center gap-2">
          <Video className="w-5 h-5 text-blue-500" />
          <h2 className={`font-semibold ${theme.colors.text}`}>영상통화</h2>
        </div>
        <button
          onClick={onClose}
          className={`p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition`}
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Tabs */}
      <div className={`flex border-b ${theme.colors.border}`}>
        <button
          onClick={() => setActiveTab('active')}
          className={`flex-1 px-4 py-3 text-sm font-medium transition ${
            activeTab === 'active'
              ? 'text-blue-500 border-b-2 border-blue-500'
              : `${theme.colors.textSecondary} hover:text-blue-500`
          }`}
        >
          <Phone className="w-4 h-4 inline-block mr-1" />
          진행중
        </button>
        <button
          onClick={() => setActiveTab('history')}
          className={`flex-1 px-4 py-3 text-sm font-medium transition ${
            activeTab === 'history'
              ? 'text-blue-500 border-b-2 border-blue-500'
              : `${theme.colors.textSecondary} hover:text-blue-500`
          }`}
        >
          <Clock className="w-4 h-4 inline-block mr-1" />
          히스토리
        </button>
      </div>

      {/* Create Room Button - only show in active tab */}
      {activeTab === 'active' && (
        <div className="p-4">
          <button
            onClick={() => setShowCreateModal(true)}
            className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition font-medium"
          >
            <Plus className="w-5 h-5" />새 통화방 만들기
          </button>
        </div>
      )}

      {/* Content Area */}
      <div className="flex-1 overflow-y-auto px-4 pb-4">
        {activeTab === 'active' ? (
          /* Active Rooms */
          isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            </div>
          ) : rooms.length === 0 ? (
            <div className={`text-center py-8 ${theme.colors.textSecondary}`}>
              <Video className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p>활성화된 통화방이 없습니다</p>
              <p className="text-sm mt-1">새 통화방을 만들어보세요!</p>
            </div>
          ) : (
            <div className="space-y-3">
              {rooms.map((room) => {
                const inRoom = isUserInRoom(room);
                return (
                  <div
                    key={room.id}
                    className={`p-4 rounded-lg border ${
                      inRoom
                        ? 'border-green-500 bg-green-50 dark:bg-green-900/20'
                        : theme.colors.border
                    } ${theme.colors.card} hover:border-blue-500 transition`}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <h3 className={`font-medium ${theme.colors.text}`}>{room.name}</h3>
                      <div className="flex items-center gap-1">
                        {inRoom && (
                          <span className="px-2 py-0.5 text-xs rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
                            참여중
                          </span>
                        )}
                        <span
                          className={`px-2 py-0.5 text-xs rounded-full ${
                            room.isActive
                              ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                              : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                          }`}
                        >
                          {room.isActive ? '진행중' : '종료됨'}
                        </span>
                      </div>
                    </div>
                    <div
                      className={`flex items-center gap-2 text-sm ${theme.colors.textSecondary} mb-3`}
                    >
                      <Users className="w-4 h-4" />
                      <span>
                        {room.participantCount} / {room.maxParticipants}명
                      </span>
                    </div>

                    {/* 버튼 영역 */}
                    <div className="flex gap-2">
                      {/* 참여하기 / 다시 참여 버튼 */}
                      <button
                        onClick={() => handleJoinRoom(room.id)}
                        disabled={
                          isJoining === room.id ||
                          !room.isActive ||
                          (!inRoom && room.participantCount >= room.maxParticipants)
                        }
                        className={`flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-lg transition font-medium text-sm ${
                          room.isActive
                            ? 'bg-green-500 hover:bg-green-600 text-white'
                            : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                        }`}
                      >
                        {isJoining === room.id ? (
                          <>
                            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                            참여 중...
                          </>
                        ) : (
                          <>
                            <Phone className="w-4 h-4" />
                            {inRoom ? '다시 참여' : '참여하기'}
                          </>
                        )}
                      </button>

                      {/* 나가기 버튼 (참여 중일 때만) */}
                      {inRoom && (
                        <button
                          onClick={() => handleLeaveRoom(room.id)}
                          disabled={isLeaving === room.id}
                          className="px-3 py-2 rounded-lg bg-red-100 hover:bg-red-200 text-red-600 transition font-medium text-sm flex items-center gap-1"
                          title="방 나가기"
                        >
                          {isLeaving === room.id ? (
                            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-red-600"></div>
                          ) : (
                            <LogOut className="w-4 h-4" />
                          )}
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )
        ) : /* Call History */
        isHistoryLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          </div>
        ) : callHistory.length === 0 ? (
          <div className={`text-center py-8 ${theme.colors.textSecondary}`}>
            <Clock className="w-12 h-12 mx-auto mb-3 opacity-50" />
            <p>통화 기록이 없습니다</p>
            <p className="text-sm mt-1">통화가 종료되면 여기에 기록됩니다</p>
          </div>
        ) : (
          <div className="space-y-3 mt-4">
            {callHistory.map((history) => {
              const isExpanded = expandedHistoryId === history.id;
              return (
                <div
                  key={history.id}
                  className={`rounded-lg border ${theme.colors.border} ${theme.colors.card} overflow-hidden`}
                >
                  {/* 헤더 (클릭 가능) */}
                  <button
                    onClick={() => setExpandedHistoryId(isExpanded ? null : history.id)}
                    className="w-full p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-800 transition"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <h3 className={`font-medium ${theme.colors.text}`}>{history.roomName}</h3>
                      {isExpanded ? (
                        <ChevronUp className="w-4 h-4 text-gray-400" />
                      ) : (
                        <ChevronDown className="w-4 h-4 text-gray-400" />
                      )}
                    </div>
                    <div className={`space-y-1 text-sm ${theme.colors.textSecondary}`}>
                      <div className="flex items-center gap-2">
                        <Calendar className="w-4 h-4" />
                        <span>{formatDate(history.endedAt)}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Clock className="w-4 h-4" />
                        <span>{formatDuration(history.durationSeconds)}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Users className="w-4 h-4" />
                        <span>{history.totalParticipants}명 참여</span>
                      </div>
                    </div>
                  </button>

                  {/* 확장 시 참여자 + 상세보기 버튼 */}
                  {isExpanded && (
                    <div
                      className={`px-4 pb-4 pt-2 border-t ${theme.colors.border} bg-gray-50 dark:bg-gray-800/50`}
                    >
                      {history.participants && history.participants.length > 0 && (
                        <>
                          <p className={`text-xs font-medium ${theme.colors.textSecondary} mb-2`}>
                            참여자 목록
                          </p>
                          <div className="space-y-2 mb-3">
                            {history.participants.map((participant, idx) => {
                              const profileImage = getMemberProfileImage(participant.userId);
                              const memberName = getMemberName(participant.userId);
                              return (
                                <div
                                  key={`${participant.userId}-${idx}`}
                                  className={`flex items-center justify-between text-sm ${theme.colors.text}`}
                                >
                                  <div className="flex items-center gap-2">
                                    {profileImage ? (
                                      <img
                                        src={profileImage}
                                        alt={memberName}
                                        className="w-6 h-6 rounded-full object-cover"
                                      />
                                    ) : (
                                      <div className="w-6 h-6 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center text-xs font-medium text-blue-600 dark:text-blue-300">
                                        {memberName.charAt(0).toUpperCase()}
                                      </div>
                                    )}
                                    <span className="text-xs text-gray-700 dark:text-gray-300 truncate max-w-[120px]">
                                      {memberName}
                                    </span>
                                  </div>
                                  <span className={`text-xs ${theme.colors.textSecondary}`}>
                                    {formatDuration(participant.durationSeconds)}
                                  </span>
                                </div>
                              );
                            })}
                          </div>
                        </>
                      )}
                      {/* 상세보기 버튼 */}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedHistory(history);
                        }}
                        className="w-full py-2 px-3 rounded-lg bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium transition"
                      >
                        상세보기
                      </button>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Create Room Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-60">
          <div className={`${theme.colors.card} rounded-xl p-6 w-96 max-w-[90vw] shadow-2xl`}>
            <h3 className={`text-lg font-semibold mb-4 ${theme.colors.text}`}>새 통화방 만들기</h3>
            <div className="space-y-4">
              <div>
                <label className={`block text-sm font-medium mb-1 ${theme.colors.textSecondary}`}>
                  방 이름
                </label>
                <input
                  type="text"
                  value={newRoomName}
                  onChange={(e) => setNewRoomName(e.target.value)}
                  placeholder="예: 팀 미팅"
                  className={`w-full px-3 py-2 rounded-lg border ${theme.colors.border} ${theme.colors.card} ${theme.colors.text} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                />
              </div>
              <div>
                <label className={`block text-sm font-medium mb-1 ${theme.colors.textSecondary}`}>
                  최대 참여자 수
                </label>
                <select
                  value={maxParticipants}
                  onChange={(e) => setMaxParticipants(Number(e.target.value))}
                  className={`w-full px-3 py-2 rounded-lg border ${theme.colors.border} ${theme.colors.card} ${theme.colors.text} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                >
                  <option value={2}>2명</option>
                  <option value={5}>5명</option>
                  <option value={10}>10명</option>
                  <option value={20}>20명</option>
                  <option value={50}>50명</option>
                </select>
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setShowCreateModal(false)}
                className={`flex-1 px-4 py-2 rounded-lg border ${theme.colors.border} ${theme.colors.text} hover:bg-gray-100 dark:hover:bg-gray-800 transition`}
              >
                취소
              </button>
              <button
                onClick={handleCreateRoom}
                disabled={!newRoomName.trim() || isCreating}
                className="flex-1 px-4 py-2 rounded-lg bg-blue-500 hover:bg-blue-600 text-white transition disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isCreating ? '생성 중...' : '만들기'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 히스토리 상세 모달 */}
      {selectedHistory && (
        <CallHistoryDetailModal
          history={selectedHistory}
          onClose={() => setSelectedHistory(null)}
          memberMap={memberMap}
        />
      )}
    </div>
  );
};

export default VideoCallPanel;
