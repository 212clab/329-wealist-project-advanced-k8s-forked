// src/hooks/usePresence.ts
// 🔥 Global Presence Hook - 앱 접속 시 온라인 상태 자동 등록

import { useEffect, useRef } from 'react';
import {
  connectPresenceWebSocket,
  disconnectPresenceWebSocket,
  isPresenceConnected,
} from '../utils/presenceWebsocket';

interface UsePresenceOptions {
  onStatusChange?: (data: { type: string; userId: string; payload?: { status: string } }) => void;
}

/**
 * Global Presence Hook
 *
 * 앱에 로그인한 사용자를 자동으로 온라인 상태로 등록합니다.
 * MainLayout 또는 App 수준에서 사용하세요.
 *
 * @example
 * // MainLayout.tsx
 * const { isConnected } = usePresence({
 *   onStatusChange: (data) => {
 *     if (data.type === 'USER_STATUS') {
 *       console.log(`User ${data.userId} is now ${data.payload?.status}`);
 *     }
 *   }
 * });
 */
export const usePresence = (options?: UsePresenceOptions) => {
  const connectedRef = useRef(false);
  const onStatusChangeRef = useRef(options?.onStatusChange);

  // 최신 콜백 유지
  useEffect(() => {
    onStatusChangeRef.current = options?.onStatusChange;
  }, [options?.onStatusChange]);

  // Presence WebSocket 연결
  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (!token) {
      console.log('⚠️ [usePresence] 토큰 없음 - 연결 건너뜀');
      return;
    }

    // 1번만 연결
    if (!connectedRef.current) {
      connectedRef.current = true;
      console.log('🟢 [usePresence] Presence 연결 시작');

      connectPresenceWebSocket((data) => {
        onStatusChangeRef.current?.(data);
      });
    }

    // 클린업: 언마운트 시에만 연결 해제
    return () => {
      if (connectedRef.current) {
        console.log('🔌 [usePresence] 컴포넌트 언마운트 - 연결 해제');
        disconnectPresenceWebSocket();
        connectedRef.current = false;
      }
    };
  }, []);

  return {
    isConnected: isPresenceConnected(),
  };
};

export default usePresence;
