import React, { useState, useEffect } from 'react';
import { X, Download, Clock, Calendar, Users, Loader2, Check, Link } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { CallHistory, videoService } from '../../api/videoService';
import { WorkspaceMemberResponse } from '../../types/user';

interface CallHistoryDetailModalProps {
  history: CallHistory;
  onClose: () => void;
  memberMap: Map<string, WorkspaceMemberResponse>;
}

const formatDuration = (seconds: number): string => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}시간 ${minutes}분 ${secs}초`;
  } else if (minutes > 0) {
    return `${minutes}분 ${secs}초`;
  }
  return `${secs}초`;
};

const formatDateTime = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString('ko-KR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

export const CallHistoryDetailModal: React.FC<CallHistoryDetailModalProps> = ({
  history,
  onClose,
  memberMap,
}) => {
  const { theme } = useTheme();
  const [transcript, setTranscript] = useState<string | null>(null);
  const [isLoadingTranscript, setIsLoadingTranscript] = useState(true);
  const [copied, setCopied] = useState(false);

  // 회의 ID 복사 (보드에 붙여넣기용)
  const copyMeetingLink = () => {
    const meetingLink = `[회의기록:${history.id}]`;
    navigator.clipboard.writeText(meetingLink);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

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

  // Fetch transcript when modal opens
  useEffect(() => {
    const fetchTranscript = async () => {
      try {
        setIsLoadingTranscript(true);
        const transcriptData = await videoService.getTranscript(history.id);
        setTranscript(transcriptData?.content || null);
      } catch (error) {
        console.error('Failed to fetch transcript:', error);
        setTranscript(null);
      } finally {
        setIsLoadingTranscript(false);
      }
    };

    fetchTranscript();
  }, [history.id]);

  const handleDownload = () => {
    // 회의록 텍스트 생성
    let content = `📹 회의 기록\n`;
    content += `${'='.repeat(50)}\n\n`;
    content += `📌 회의명: ${history.roomName}\n`;
    content += `📅 시작: ${formatDateTime(history.startedAt)}\n`;
    content += `📅 종료: ${formatDateTime(history.endedAt)}\n`;
    content += `⏱️ 총 시간: ${formatDuration(history.durationSeconds)}\n`;
    content += `👥 참여자 수: ${history.totalParticipants}명\n\n`;

    content += `${'─'.repeat(50)}\n`;
    content += `👥 참여자 목록\n`;
    content += `${'─'.repeat(50)}\n`;

    if (history.participants && history.participants.length > 0) {
      history.participants.forEach((p, idx) => {
        content += `${idx + 1}. ${getMemberName(p.userId)}\n`;
        content += `   - 참여 시간: ${formatDuration(p.durationSeconds)}\n`;
      });
    }

    if (transcript) {
      content += `\n${'─'.repeat(50)}\n`;
      content += `📝 회의 내용\n`;
      content += `${'─'.repeat(50)}\n`;
      content += transcript;
    }

    // 파일 다운로드
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `회의록_${history.roomName}_${
      new Date(history.endedAt).toISOString().split('T')[0]
    }.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]"
      onClick={onClose}
    >
      <div
        className={`${theme.colors.card} rounded-xl w-[480px] max-w-[95vw] max-h-[85vh] shadow-2xl flex flex-col`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* 헤더 */}
        <div className={`flex items-center justify-between p-4 border-b ${theme.colors.border}`}>
          <h2 className={`text-lg font-semibold ${theme.colors.text}`}>회의 상세 정보</h2>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 내용 */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {/* 회의명 */}
          <div
            className={`p-4 rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800`}
          >
            <h3 className="text-xl font-bold text-blue-700 dark:text-blue-300">
              {history.roomName}
            </h3>
            {/* 회의 ID 복사 */}
            <div className="mt-2 flex items-center gap-2">
              <span className={`text-xs ${theme.colors.textSecondary}`}>
                ID: {history.id.slice(0, 8)}...
              </span>
              <button
                onClick={copyMeetingLink}
                className={`flex items-center gap-1 px-2 py-1 text-xs rounded ${
                  copied
                    ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
                } transition`}
                title="보드/프로젝트에 붙여넣기용 링크 복사"
              >
                {copied ? (
                  <>
                    <Check className="w-3 h-3" />
                    복사됨
                  </>
                ) : (
                  <>
                    <Link className="w-3 h-3" />
                    링크 복사
                  </>
                )}
              </button>
            </div>
          </div>

          {/* 시간 정보 */}
          <div className={`p-4 rounded-lg border ${theme.colors.border} space-y-3`}>
            <h4 className={`font-medium ${theme.colors.text} flex items-center gap-2`}>
              <Clock className="w-4 h-4" />
              시간 정보
            </h4>
            <div className={`space-y-2 text-sm ${theme.colors.textSecondary}`}>
              <div className="flex items-center gap-2">
                <Calendar className="w-4 h-4 text-green-500" />
                <span className="font-medium">시작:</span>
                <span>{formatDateTime(history.startedAt)}</span>
              </div>
              <div className="flex items-center gap-2">
                <Calendar className="w-4 h-4 text-red-500" />
                <span className="font-medium">종료:</span>
                <span>{formatDateTime(history.endedAt)}</span>
              </div>
              <div className="flex items-center gap-2">
                <Clock className="w-4 h-4 text-blue-500" />
                <span className="font-medium">총 시간:</span>
                <span className="font-bold text-blue-600 dark:text-blue-400">
                  {formatDuration(history.durationSeconds)}
                </span>
              </div>
            </div>
          </div>

          {/* 참여자 목록 */}
          <div className={`p-4 rounded-lg border ${theme.colors.border}`}>
            <h4 className={`font-medium ${theme.colors.text} flex items-center gap-2 mb-3`}>
              <Users className="w-4 h-4" />
              참여자 ({history.totalParticipants}명)
            </h4>
            <div className="space-y-2">
              {history.participants && history.participants.length > 0 ? (
                history.participants.map((participant, idx) => {
                  const profileImage = getMemberProfileImage(participant.userId);
                  const memberName = getMemberName(participant.userId);
                  return (
                    <div
                      key={`${participant.userId}-${idx}`}
                      className={`flex items-center justify-between p-2 rounded-lg bg-gray-50 dark:bg-gray-800`}
                    >
                      <div className="flex items-center gap-3">
                        {profileImage ? (
                          <img
                            src={profileImage}
                            alt={memberName}
                            className="w-8 h-8 rounded-full object-cover"
                          />
                        ) : (
                          <div className="w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center text-sm font-medium text-blue-600 dark:text-blue-300">
                            {memberName.charAt(0).toUpperCase()}
                          </div>
                        )}
                        <span className={`text-sm font-medium ${theme.colors.text}`}>
                          {memberName}
                        </span>
                      </div>
                      <span className={`text-sm ${theme.colors.textSecondary}`}>
                        {formatDuration(participant.durationSeconds)}
                      </span>
                    </div>
                  );
                })
              ) : (
                <p className={`text-sm ${theme.colors.textSecondary}`}>참여자 정보가 없습니다.</p>
              )}
            </div>
          </div>

          {/* 회의 내용 (자막) */}
          {isLoadingTranscript ? (
            <div className={`p-4 rounded-lg border ${theme.colors.border}`}>
              <div className="flex items-center justify-center gap-2 py-4">
                <Loader2 className="w-5 h-5 animate-spin text-blue-500" />
                <span className={`text-sm ${theme.colors.textSecondary}`}>
                  회의 자막 불러오는 중...
                </span>
              </div>
            </div>
          ) : transcript ? (
            <div className={`p-4 rounded-lg border ${theme.colors.border}`}>
              <h4 className={`font-medium ${theme.colors.text} mb-3`}>📝 회의 내용</h4>
              <div
                className={`p-3 rounded bg-gray-50 dark:bg-gray-800 text-sm ${theme.colors.text} whitespace-pre-wrap max-h-48 overflow-y-auto`}
              >
                {transcript}
              </div>
            </div>
          ) : (
            <div className={`p-4 rounded-lg border ${theme.colors.border} border-dashed`}>
              <p className={`text-sm ${theme.colors.textSecondary} text-center`}>
                📝 회의 자막 기록이 없습니다.
              </p>
            </div>
          )}
        </div>

        {/* 푸터 */}
        <div className={`p-4 border-t ${theme.colors.border} flex justify-end gap-2`}>
          <button
            onClick={onClose}
            className={`px-4 py-2 rounded-lg border ${theme.colors.border} ${theme.colors.text} hover:bg-gray-100 dark:hover:bg-gray-800 transition`}
          >
            닫기
          </button>
          <button
            onClick={handleDownload}
            className="px-4 py-2 rounded-lg bg-blue-500 hover:bg-blue-600 text-white transition flex items-center gap-2"
          >
            <Download className="w-4 h-4" />
            회의록 다운로드
          </button>
        </div>
      </div>
    </div>
  );
};

export default CallHistoryDetailModal;
