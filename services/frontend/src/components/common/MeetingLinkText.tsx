// src/components/common/MeetingLinkText.tsx
// 회의 기록 링크를 감지하여 클릭 가능하게 만드는 컴포넌트

import React, { useMemo } from 'react';
import { Video } from 'lucide-react';

interface MeetingLinkTextProps {
  text: string;
  onMeetingClick?: (meetingId: string) => void;
  className?: string;
}

// 회의기록 링크 패턴: [회의기록:uuid]
const MEETING_LINK_PATTERN = /\[회의기록:([a-f0-9-]{36})\]/gi;

interface TextPart {
  type: 'text' | 'meeting';
  content: string;
  meetingId?: string;
}

export const MeetingLinkText: React.FC<MeetingLinkTextProps> = ({
  text,
  onMeetingClick,
  className = '',
}) => {
  const parts = useMemo(() => {
    const result: TextPart[] = [];
    let lastIndex = 0;
    let match;

    const regex = new RegExp(MEETING_LINK_PATTERN);

    while ((match = regex.exec(text)) !== null) {
      // 매치 전의 텍스트
      if (match.index > lastIndex) {
        result.push({
          type: 'text',
          content: text.slice(lastIndex, match.index),
        });
      }

      // 회의 링크
      result.push({
        type: 'meeting',
        content: match[0],
        meetingId: match[1],
      });

      lastIndex = match.index + match[0].length;
    }

    // 나머지 텍스트
    if (lastIndex < text.length) {
      result.push({
        type: 'text',
        content: text.slice(lastIndex),
      });
    }

    return result;
  }, [text]);

  // 회의 링크가 없으면 그냥 텍스트 반환
  if (parts.length === 1 && parts[0].type === 'text') {
    return <span className={className}>{text}</span>;
  }

  return (
    <span className={className}>
      {parts.map((part, index) => {
        if (part.type === 'text') {
          return <span key={index}>{part.content}</span>;
        }

        return (
          <button
            key={index}
            onClick={() => part.meetingId && onMeetingClick?.(part.meetingId)}
            className="inline-flex items-center gap-1 px-2 py-0.5 mx-0.5 bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 rounded-md hover:bg-blue-200 dark:hover:bg-blue-900/60 transition text-sm font-medium"
            title="클릭하여 회의 기록 보기"
          >
            <Video className="w-3 h-3" />
            회의 기록
          </button>
        );
      })}
    </span>
  );
};

export default MeetingLinkText;
