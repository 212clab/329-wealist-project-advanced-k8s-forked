import React from 'react';
import { Trash2 } from 'lucide-react';
import { WorkspaceSettingsResponse, UpdateWorkspaceSettingsRequest } from '../../../../../types/user';
import { useTheme } from '../../../../../contexts/ThemeContext';

interface WorkspaceSettingsTabProps {
  settings: WorkspaceSettingsResponse;
  settingsForm: UpdateWorkspaceSettingsRequest;
  setSettingsForm: React.Dispatch<React.SetStateAction<UpdateWorkspaceSettingsRequest>>;
  handleSaveSettings: () => Promise<void>;
  handleDeleteWorkspace: () => Promise<void>;
  loading: boolean;
  error: string | null;
}

export const WorkspaceSettingsTab: React.FC<WorkspaceSettingsTabProps> = ({
  settingsForm,
  setSettingsForm,
  handleSaveSettings,
  handleDeleteWorkspace,
  loading,
}) => {
  const { theme } = useTheme();

  return (
    <div className="space-y-4">
      <div>
        <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
          워크스페이스 이름:
        </label>
        <input
          type="text"
          value={settingsForm.workspaceName || ''}
          onChange={(e) => setSettingsForm((prev) => ({ ...prev, workspaceName: e.target.value }))}
          className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
          placeholder="워크스페이스 이름"
        />
      </div>

      <div>
        <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
          설명:
        </label>
        <textarea
          value={settingsForm.workspaceDescription || ''}
          onChange={(e) =>
            setSettingsForm((prev) => ({ ...prev, workspaceDescription: e.target.value }))
          }
          rows={3}
          className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
          placeholder="워크스페이스 설명"
        />
      </div>

      <div className="space-y-3">
        {/* 공개/비공개 토글 */}
        <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
          <div>
            <p className="text-sm font-medium text-gray-700">공개 워크스페이스</p>
            <p className="text-xs text-gray-500">
              공개 시 다른 사용자가 검색하여 가입 신청할 수 있습니다
            </p>
          </div>
          <button
            onClick={() => setSettingsForm((prev) => ({ ...prev, isPublic: !prev.isPublic }))}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              settingsForm.isPublic ? 'bg-blue-600' : 'bg-gray-300'
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                settingsForm.isPublic ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
        </div>

        {/* 승인제/비승인제 토글 */}
        <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
          <div>
            <p className="text-sm font-medium text-gray-700">가입 승인제</p>
            <p className="text-xs text-gray-500">활성화 시 가입 신청을 승인해야 회원이 됩니다</p>
          </div>
          <button
            onClick={() =>
              setSettingsForm((prev) => ({ ...prev, requiresApproval: !prev.requiresApproval }))
            }
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              settingsForm.requiresApproval ? 'bg-blue-600' : 'bg-gray-300'
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                settingsForm.requiresApproval ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
        </div>

        {/* OWNER 전용 초대 토글 */}
        <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
          <div>
            <p className="text-sm font-medium text-gray-700">OWNER만 초대 가능</p>
            <p className="text-xs text-gray-500">활성화 시 OWNER만 새 회원을 초대할 수 있습니다</p>
          </div>
          <button
            onClick={() =>
              setSettingsForm((prev) => ({ ...prev, onlyOwnerCanInvite: !prev.onlyOwnerCanInvite }))
            }
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              settingsForm.onlyOwnerCanInvite ? 'bg-blue-600' : 'bg-gray-300'
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                settingsForm.onlyOwnerCanInvite ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
        </div>
      </div>

      <button
        onClick={handleSaveSettings}
        disabled={loading}
        className={`w-full ${theme.colors.primary} text-white py-3 ${
          theme.effects.borderRadius
        } font-semibold transition ${
          loading ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-90'
        }`}
      >
        {loading ? '저장 중...' : '설정 저장'}
      </button>

      {/* 위험 구역 - 워크스페이스 삭제 */}
      <div className="mt-6 pt-4 border-t border-gray-200">
        <div className="flex items-center justify-between p-3 bg-red-50 rounded-lg">
          <div>
            <p className="text-sm font-medium text-red-600">워크스페이스 삭제</p>
            <p className="text-xs text-gray-500">
              모든 프로젝트, 보드, 회원 정보가 함께 삭제됩니다
            </p>
          </div>
          <button
            onClick={handleDeleteWorkspace}
            disabled={loading}
            className={`p-2 rounded-lg text-red-500 hover:bg-red-100 transition ${
              loading ? 'opacity-50 cursor-not-allowed' : ''
            }`}
            title="워크스페이스 삭제"
          >
            <Trash2 className="w-5 h-5" />
          </button>
        </div>
      </div>
    </div>
  );
};
