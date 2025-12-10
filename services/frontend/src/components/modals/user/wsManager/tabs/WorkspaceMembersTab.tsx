// src/components/modals/WorkspaceMembersTab.tsx (수정됨)

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Search } from 'lucide-react';
import { useTheme } from '../../../../../contexts/ThemeContext';
import { useAuth } from '../../../../../contexts/AuthContext';
import {
  JoinRequestResponse,
  WorkspaceMemberResponse,
  WorkspaceMemberRole,
  // 💡 가상의 UserSearchResponse 타입 필요 (임시로 any 사용)
} from '../../../../../types/user';
import {
  getWorkspaceMembers,
  getPendingMembers,
  approveMember,
  rejectMember,
  updateMemberRole,
  removeMember,
  inviteUser,
} from '../../../../../api/userService';

interface WorkspaceMembersTabProps {
  workspaceId: string;
  onDataRefreshed: () => void;
  loadingGlobal: boolean;
  errorGlobal: string | null;
}

export const WorkspaceMembersTab: React.FC<WorkspaceMembersTabProps> = ({
  workspaceId,
  onDataRefreshed,
  loadingGlobal,
  errorGlobal,
}) => {
  const { theme } = useTheme();
  const { token } = useAuth();

  // 💡 [자체 상태 관리 시작]
  const [members, setMembers] = useState<WorkspaceMemberResponse[]>([]);
  const [pendingMembers, setPendingMembers] = useState<JoinRequestResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 💡 [추가된 상태] 초대할 사용자 검색 관련
  const [inviteSearchQuery, setInviteSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<any[]>([]); // 💡 UserSearchResponse[] 타입 가정
  const [selectedUserId, setSelectedUserId] = useState(''); // 드롭다운에서 선택된 사용자의 ID

  const [memberSearchQuery, setMemberSearchQuery] = useState('');

  // ========================================
  // 1. 데이터 로드 함수 (내부 구현)
  // ========================================
  const fetchWorkspaceData = useCallback(async () => {
    if (!token || !workspaceId) {
      setMembers([]);
      setPendingMembers([]);
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const [membersData, pendingData] = await Promise.all([
        getWorkspaceMembers(workspaceId),
        getPendingMembers(workspaceId),
      ]);

      console.log('✅ Pending Members:', pendingData); // 💡 디버깅용 로그 추가

      setMembers(membersData);
      setPendingMembers(pendingData);
    } catch (err: any) {
      console.error('[MembersTab] 데이터 로드 실패:', err);
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`회원 목록 로드 실패: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  }, [workspaceId, token]);

  useEffect(() => {
    fetchWorkspaceData();
  }, [fetchWorkspaceData]);

  // ========================================
  // 2. 회원 관리 핸들러 (내부 구현)
  // ========================================

  /**
   * 💡 [추가] 이메일/ID로 사용자 검색 (프론트엔드 목업 로직)
   * 💡 실제로는 API 호출 (예: searchUsersByEmail(inviteSearchQuery))이 필요합니다.
   */
  // const handleSearchUser = useCallback(() => {
  //   // ⚠️ 실제로는 여기서 searchUsersByEmail API를 호출해야 합니다.
  //   // 현재 API 스펙에 검색 기능이 없으므로, UI 구현을 위해 임시로 목업 데이터를 사용하거나,
  //   // 백엔드 개발자에게 해당 API를 요청해야 합니다.

  //   if (inviteSearchQuery.trim().length > 0) {
  //     // 임시 목업: 검색어와 ID가 같은 사용자를 선택된 것으로 간주
  //     setSelectedUserId(inviteSearchQuery.trim());
  //     setSearchResults([
  //       {
  //         id: inviteSearchQuery,
  //         userName: '검색된 사용자',
  //         userEmail: `${inviteSearchQuery}@example.com`,
  //       },
  //     ]);
  //   } else {
  //     setSearchResults([]);
  //     setSelectedUserId('');
  //   }
  // }, [inviteSearchQuery]);

  useEffect(() => {
    // 💡 검색어가 변경될 때마다 검색 로직 실행 (debounce 적용 권장)
    const delayDebounceFn = setTimeout(() => {
      // handleSearchUser(); // ⚠️ 실제 API 호출 시 주석 해제
    }, 500);

    return () => clearTimeout(delayDebounceFn);
  }, [inviteSearchQuery]);

  /**
   * 💡 사용자 이메일을 통해 최종 초대 실행
   */
  const handleInviteUser = async (email: string) => {
    if (!email.trim()) {
      setError('초대할 사용자의 이메일을 입력해주세요.');
      return;
    }
    setLoading(true);
    setError(null);
    try {
      await inviteUser(workspaceId, email);
      setInviteSearchQuery(''); // 검색창 초기화
      setSelectedUserId(''); // 선택된 사용자 초기화
      setSearchResults([]); // 검색 결과 초기화
      await fetchWorkspaceData();
      onDataRefreshed();
      console.log(`사용자 이메일 ${email}에 대한 초대가 완료되었습니다.`);
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`회원 초대에 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleApproveMember = async (userId: string) => {
    setLoading(true);
    setError(null);
    try {
      await approveMember(workspaceId, userId);

      // 💡 데이터 다시 불러오기 (순차 실행)
      await fetchWorkspaceData();

      // 💡 상위 컴포넌트에 알림
      onDataRefreshed();

      console.log('✅ 회원 승인이 완료되었습니다.');
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`회원 승인에 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleRejectMember = async (userId: string) => {
    setLoading(true);
    setError(null);
    try {
      await rejectMember(workspaceId, userId);

      // 💡 데이터 다시 불러오기 (순차 실행)
      await fetchWorkspaceData();

      // 💡 상위 컴포넌트에 알림
      onDataRefreshed();

      console.log('✅ 회원 요청을 거절했습니다.');
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`회원 거절에 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateRole = async (
    memberId: string,
    currentRole: WorkspaceMemberRole,
    newRole: 'ADMIN' | 'MEMBER',
  ) => {
    if (currentRole === newRole || currentRole === 'OWNER') return;

    setLoading(true);
    setError(null);
    try {
      await updateMemberRole(workspaceId, memberId, newRole);
      await fetchWorkspaceData();
      onDataRefreshed();
      console.log(`회원 역할이 ${newRole}로 변경되었습니다.`);
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`역할 변경에 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveMember = async (memberId: string, userName: string) => {
    console.warn(`[Confirmation] 사용자 ${userName} 퇴출을 진행합니다.`);
    setLoading(true);
    setError(null);
    try {
      await removeMember(workspaceId, memberId);
      await fetchWorkspaceData();
      onDataRefreshed();
      console.log(`${userName}님이 퇴출되었습니다.`);
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`회원 퇴출에 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 3. 회원 목록 필터링
  // ========================================

  const filteredMembers = useMemo(() => {
    if (!memberSearchQuery.trim()) return members;
    const query = memberSearchQuery.toLowerCase();
    return members.filter(
      (member) =>
        (member.nickName || '').toLowerCase().includes(query) ||
        (member.userEmail || '').toLowerCase().includes(query) ||
        (member.roleName || '').toLowerCase().includes(query),
    );
  }, [members, memberSearchQuery]);

  // ========================================
  // 4. 렌더링
  // ========================================

  const displayError = error || errorGlobal;
  const displayLoading = loading || loadingGlobal;

  return (
    <div className="space-y-6">
      {/* 에러 메시지 (상위/자체 에러 표시) */}
      {displayError && (
        <div className="p-3 bg-red-100 border border-red-400 text-red-700 rounded-md text-sm">
          {displayError}
        </div>
      )}

      {/* 💡 [수정/확장] 사용자 검색 및 초대 기능 */}
      <div className="relative">
        <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
          사용자 검색 후 초대 (이메일, ID 등):
        </label>
        <div className="flex gap-2">
          {/* 1. 검색 입력창 */}
          <input
            type="text"
            placeholder="사용자 이메일으로 검색"
            value={inviteSearchQuery}
            onChange={(e) => setInviteSearchQuery(e.target.value)}
            className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
            disabled={displayLoading}
          />
          {/* 2. 초대 버튼 (선택된 사용자가 있을 경우 활성화) */}
          <button
            // 💡 현재는 inviteSearchQuery가 userId 역할을 하도록 임시 설정
            onClick={() => handleInviteUser(inviteSearchQuery)}
            disabled={displayLoading || !inviteSearchQuery.trim()}
            className="px-4 py-2 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition disabled:opacity-50 font-semibold"
          >
            초대
          </button>
        </div>

        {/* 3. 검색 결과 드롭다운 (현재는 임시 UI) */}
        {searchResults.length > 0 && (
          <div
            className={`absolute z-10 w-full mt-1 ${theme.colors.card} border ${theme.colors.border} rounded-md shadow-lg max-h-40 overflow-y-auto`}
          >
            {searchResults.map((user) => (
              <div
                key={user.userId}
                onClick={() => {
                  setSelectedUserId(user.userId);
                  setInviteSearchQuery(user.userId); // 선택된 ID로 검색창 업데이트 (선택 완료 시)
                  setSearchResults([]); // 드롭다운 닫기
                }}
                className={`p-3 cursor-pointer hover:bg-gray-100 ${
                  selectedUserId === user.userId ? 'bg-blue-50' : ''
                }`}
              >
                <p className="text-sm font-medium">{user.nickName}</p>
                <p className="text-xs text-gray-500">{user.userEmail}</p>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 승인 대기 목록 (기존 로직 유지) */}
      {pendingMembers?.length > 0 && (
        <div className="border border-gray-200 rounded-lg p-4">
          <p className="text-sm font-semibold text-gray-700 mb-3">
            승인 대기 목록 ({pendingMembers?.length}명)
          </p>
          <div className="space-y-2">
            {pendingMembers?.map((member) => (
              <div
                key={member.joinRequestId}
                className="flex items-center justify-between bg-gray-50 p-3 rounded border border-gray-200"
              >
                <div>
                  <p className="text-sm font-medium text-gray-700">{member.nickName}</p>
                  <p className="text-xs text-gray-500">{member.userEmail}</p>
                </div>
                <div className="flex gap-2">
                  {/* API 연동: 승인 */}
                  <button
                    onClick={() => handleApproveMember(member.userId)}
                    disabled={displayLoading}
                    className="px-3 py-1 bg-green-500 text-white text-xs rounded hover:bg-green-600 transition disabled:opacity-50"
                  >
                    승인
                  </button>
                  {/* API 연동: 거절 */}
                  <button
                    onClick={() => handleRejectMember(member.userId)}
                    disabled={displayLoading}
                    className="px-3 py-1 bg-red-500 text-white text-xs rounded hover:bg-red-600 transition disabled:opacity-50"
                  >
                    거절
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 조직 회원 목록 (기존 로직 유지) */}
      <div className="border border-gray-200 rounded-lg p-4">
        <div className="flex items-center justify-between mb-3 gap-3">
          <p className="text-sm font-semibold text-gray-700 whitespace-nowrap">
            조직 회원 목록 ({members?.length}명)
          </p>
          {/* 조직 회원 검색 */}
          <div className="relative flex-1 max-w-xs">
            <input
              type="text"
              placeholder="이름, 이메일, 역할 검색"
              value={memberSearchQuery}
              onChange={(e) => setMemberSearchQuery(e.target.value)}
              className={`w-full px-3 pl-8 py-1.5 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} text-xs ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
            />
            <Search className="absolute left-2.5 top-1/2 transform -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
          </div>
        </div>
        <div className="space-y-2">
          {filteredMembers?.map((member) => (
            <div
              key={member?.workspaceMemberId}
              className="flex items-center justify-between bg-gray-50 p-3 rounded border border-gray-200"
            >
              <div>
                <p className="text-sm font-medium text-gray-700">
                  {member?.nickName}{' '}
                  <span className="text-xs text-blue-600 font-semibold">({member?.roleName})</span>
                </p>
                <p className="text-xs text-gray-500">{member?.userEmail}</p>
              </div>
              {/* API 연동: 역할 변경 / 퇴출 */}
              {member?.roleName !== 'OWNER' && (
                <div className="flex gap-2">
                  {member?.roleName === 'MEMBER' && (
                    <button
                      onClick={() => handleUpdateRole(member?.workspaceMemberId, member?.roleName, 'ADMIN')}
                      disabled={displayLoading}
                      className="px-3 py-1 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition disabled:opacity-50"
                    >
                      ADMIN
                    </button>
                  )}
                  {member?.roleName === 'ADMIN' && (
                    <button
                      onClick={() => handleUpdateRole(member?.workspaceMemberId, member?.roleName, 'MEMBER')}
                      disabled={displayLoading}
                      className="px-3 py-1 bg-gray-500 text-white text-xs rounded hover:bg-gray-600 transition disabled:opacity-50"
                    >
                      MEMBER
                    </button>
                  )}
                  <button
                    onClick={() => handleRemoveMember(member?.workspaceMemberId, member?.nickName)}
                    disabled={displayLoading}
                    className="px-3 py-1 bg-red-500 text-white text-xs rounded hover:bg-red-600 transition disabled:opacity-50"
                  >
                    퇴출
                  </button>
                </div>
              )}
            </div>
          ))}
          {filteredMembers?.length === 0 && (
            <p className="text-center text-sm text-gray-500 py-4">
              {memberSearchQuery.trim() ? '검색 결과가 없습니다.' : '조직 회원이 없습니다.'}
            </p>
          )}
        </div>
      </div>
    </div>
  );
};
