// src/hooks/useUserLookup.ts
import { useMemo } from 'react';

export interface Member {
  userId: string;
  nickName?: string;
  userEmail?: string;
  profileImageUrl?: string | null; // null이 올 수도 있으니 타입에 포함
}

export const useUserLookup = (members: Member[] = []) => {
  // 1. Lookup Table 생성 (ID -> Member 객체)
  const userMap = useMemo(() => {
    return members.reduce((acc, member) => {
      acc[member.userId] = member;
      return acc;
    }, {} as Record<string, Member>);
  }, [members]);

  // 2. 닉네임 가져오기
  const getNickname = (userId: string) => {
    const member = userMap[userId];
    return member?.nickName || member?.userEmail || userId;
  };

  // 3. 유저 객체 가져오기
  const getUser = (userId: string) => {
    return userMap[userId];
  };

  // 4. [추가됨] 프로필 이미지 URL 가져오기
  const getProfileUrl = (userId: string) => {
    return userMap[userId]?.profileImageUrl || null;
  };

  // 리턴 객체에 getProfileUrl 추가
  return { getNickname, getUser, getProfileUrl };
};
