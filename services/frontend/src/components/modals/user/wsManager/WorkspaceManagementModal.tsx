/**
 * 워크스페이스 관리 모달 컴포넌트 (OWNER/ADMIN용)
 *
 * [메인 셸] 데이터 로드, 상태 관리, 탭 전환을 담당
 */

import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { X } from 'lucide-react';
import { useTheme } from '../../../../contexts/ThemeContext';
import { useAuth } from '../../../../contexts/AuthContext';
import { WorkspaceSettingsTab } from './tabs/WorkspaceSettingsTab'; // 💡 분리된 컴포넌트
import { WorkspaceMembersTab } from './tabs/WorkspaceMembersTab'; // 💡 분리된 컴포넌트

// 💡 API 함수 Import
import { getWorkspaceSettings, updateWorkspaceSettings, deleteWorkspace } from '../../../../api/userService';

import { WorkspaceSettingsResponse, UpdateWorkspaceSettingsRequest } from '../../../../types/user';

interface WorkspaceManagementModalProps {
  workspaceId: string;
  workspaceName: string;
  onClose: () => void;
}

const WorkspaceManagementModal: React.FC<WorkspaceManagementModalProps> = ({
  workspaceId,
  onClose,
}) => {
  const { theme } = useTheme();
  const { token } = useAuth(); // 인증 상태 확인용
  const navigate = useNavigate();

  // ========================================
  // 상태 관리
  // ========================================

  const [activeTab, setActiveTab] = useState<'settings' | 'members'>('settings');

  // 워크스페이스 설정 데이터 및 폼 상태
  const [settings, setSettings] = useState<WorkspaceSettingsResponse | null>(null);
  const [settingsForm, setSettingsForm] = useState<UpdateWorkspaceSettingsRequest>({
    workspaceName: '',
    workspaceDescription: '',
    isPublic: false,
    requiresApproval: false,
    onlyOwnerCanInvite: false,
  });

  // 로딩 및 에러 (전역 상태)
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 💡 [추가] 멤버 데이터가 갱신될 때 UI를 리프레시하기 위한 상태
  const [memberDataRefreshKey, setMemberDataRefreshKey] = useState(0);

  // ========================================
  // 초기 데이터 로드 (Settings Tab 전용)
  // ========================================

  const fetchWorkspaceSettings = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      // 💡 API 호출 시 token 인자 제거
      const settingsData = await getWorkspaceSettings(workspaceId);

      setSettings(settingsData);
      setSettingsForm({
        workspaceName: settingsData.workspaceName,
        workspaceDescription: settingsData.workspaceDescription,
        isPublic: settingsData.isPublic,
        requiresApproval: settingsData.requiresApproval,
        onlyOwnerCanInvite: settingsData.onlyOwnerCanInvite,
      });
    } catch (err: any) {
      console.error('[WorkspaceManagement] 설정 로드 실패:', err);
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`워크스페이스 정보를 불러오는데 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  }, [workspaceId]);

  useEffect(() => {
    if (!token) {
      setError('인증 토큰이 없습니다. 다시 로그인해 주세요.');
      return;
    }
    fetchWorkspaceSettings();
  }, [fetchWorkspaceSettings, token]);

  // ========================================
  // 설정 저장 핸들러 (Settings Tab 전용)
  // ========================================

  const handleSaveSettings = async () => {
    try {
      setLoading(true);
      setError(null);
      const updated = await updateWorkspaceSettings(workspaceId, settingsForm);
      setSettings(updated);
      console.log('워크스페이스 설정이 저장되었습니다.');
    } catch (err: any) {
      console.error('[WorkspaceManagement] 설정 저장 실패:', err);
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`설정 저장에 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 워크스페이스 삭제 핸들러
  // ========================================

  const handleDeleteWorkspace = async () => {
    const workspaceName = settings?.workspaceName || '워크스페이스';
    const confirmed = window.confirm(
      `정말로 "${workspaceName}" 워크스페이스를 삭제하시겠습니까?\n\n삭제된 워크스페이스는 복구할 수 없으며, 모든 프로젝트와 데이터가 함께 삭제됩니다.`
    );

    if (!confirmed) return;

    try {
      setLoading(true);
      setError(null);
      await deleteWorkspace(workspaceId);
      alert('워크스페이스가 삭제되었습니다.');
      onClose();
      navigate('/workspaces'); // 워크스페이스 선택 페이지로 이동
    } catch (err: any) {
      console.error('[WorkspaceManagement] 워크스페이스 삭제 실패:', err);
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`워크스페이스 삭제에 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 렌더링
  // ========================================

  if (!settings && loading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div className="bg-white p-8 rounded-xl shadow-lg">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-700">워크스페이스 정보를 불러오는 중...</p>
        </div>
      </div>
    );
  }

  // 초기 데이터 로드 실패 시
  if (error && !settings) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div className="bg-white p-8 rounded-xl shadow-lg">
          <p className="text-red-700 font-semibold mb-4">오류 발생</p>
          <p className="text-sm text-gray-700">{error}</p>
          <button
            onClick={onClose}
            className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            닫기
          </button>
        </div>
      </div>
    );
  }

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      onClick={onClose}
    >
      <div className="relative w-full max-w-2xl" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
        >
          {/* 헤더 */}
          <div className="flex items-center justify-between p-4 pb-3">
            <h2 className={`${theme.font.size.base} font-bold text-gray-800`}>
              워크스페이스 관리 ({settings?.workspaceName || '불러오는 중...'})
            </h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 rounded-lg transition"
              title="닫기"
            >
              <X className="w-4 h-4 text-gray-600" />
            </button>
          </div>

          {/* 탭 메뉴 */}
          <div className="flex border-b border-gray-200 px-6">
            <button
              onClick={() => setActiveTab('settings')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'settings' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              기본정보
              {activeTab === 'settings' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
              )}
            </button>
            <button
              onClick={() => setActiveTab('members')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'members' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              회원관리
              {activeTab === 'members' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
              )}
            </button>
          </div>

          {/* 탭 컨텐츠 */}
          <div className="p-6 space-y-5">
            {/* 에러 메시지 (전역 에러 상태 사용) */}
            {error && (
              <div className="p-3 bg-red-100 border border-red-400 text-red-700 rounded-md text-sm">
                {error}
              </div>
            )}

            {/* 기본정보 탭 (분리된 컴포넌트 사용) */}
            {activeTab === 'settings' && settings && (
              <WorkspaceSettingsTab
                settings={settings}
                settingsForm={settingsForm}
                setSettingsForm={setSettingsForm}
                handleSaveSettings={handleSaveSettings}
                handleDeleteWorkspace={handleDeleteWorkspace}
                loading={loading}
                error={error}
              />
            )}

            {/* 회원관리 탭 (분리된 컴포넌트 사용) */}
            {activeTab === 'members' && (
              <WorkspaceMembersTab
                key={workspaceId + memberDataRefreshKey} // 💡 [추가] 키를 사용하여 데이터 갱신 강제
                workspaceId={workspaceId}
                loadingGlobal={loading}
                errorGlobal={error}
                onDataRefreshed={() => setMemberDataRefreshKey((prev) => prev + 1)} // 💡 [추가] 갱신 콜백
              />
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkspaceManagementModal;
