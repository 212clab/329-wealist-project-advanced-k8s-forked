// src/components/modals/CustomFieldManagerModal.tsx

import React, { useState, useEffect, useCallback } from 'react';
import { X, Plus, Trash2, Tag, CheckSquare, AlertCircle, Menu } from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';

import { FieldOptionResponse } from '../../../types/board';
import { getFieldOptions, deleteFieldOption } from '../../../api/boardService';
import { LoadingSpinner } from '../../common/LoadingSpinner';

interface CustomFieldManagerModalProps {
  projectId: string;
  onClose: () => void;
  onFieldsUpdated: () => void;
}

// 💡 필드 정보를 간단하게 표현하는 로컬 타입
interface LocalFieldInfo {
  fieldType: 'stage' | 'role' | 'importance';
  name: string;
  icon: React.ReactNode;
}

export const CustomFieldManagerModal: React.FC<CustomFieldManagerModalProps> = ({
  // projectId,
  onClose,
  onFieldsUpdated,
}) => {
  const { theme } = useTheme();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 💡 [수정] 고정된 필드 목록 (stage, role, importance만 관리)
  const fields: LocalFieldInfo[] = [
    {
      fieldType: 'stage',
      name: '진행 단계',
      icon: <CheckSquare className="w-4 h-4 text-gray-500" />,
    },
    { fieldType: 'role', name: '역할', icon: <Tag className="w-4 h-4 text-gray-500" /> },
    {
      fieldType: 'importance',
      name: '중요도',
      icon: <AlertCircle className="w-4 h-4 text-gray-500" />,
    },
  ];

  const [selectedField, setSelectedField] = useState<LocalFieldInfo>(fields[0]);
  const [fieldOptions, setFieldOptions] = useState<FieldOptionResponse[]>([]);

  // ========================================
  // 1. 데이터 로드
  // ========================================

  const fetchOptions = useCallback(async (fieldType: 'stage' | 'role' | 'importance') => {
    setLoading(true);
    setError(null);
    try {
      // 💡 [API] 필드 타입별 옵션 로드
      const options = await getFieldOptions(fieldType);
      setFieldOptions(options.sort((a, b) => a.displayOrder - b.displayOrder));
    } catch (err: any) {
      setError(`옵션 로드 실패: ${err.message}`);
      setFieldOptions([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (selectedField) {
      fetchOptions(selectedField.fieldType);
    }
  }, [selectedField, fetchOptions]);

  // ========================================
  // 2. 옵션 삭제 핸들러
  // ========================================

  const handleDeleteOption = async (optionId: string, label: string) => {
    if (!window.confirm(`정말 옵션 '${label}'을(를) 삭제하시겠습니까? (복구 불가)`)) return;

    setLoading(true);
    try {
      await deleteFieldOption(optionId);
      await fetchOptions(selectedField.fieldType);
      onFieldsUpdated();
    } catch (err: any) {
      setError(`옵션 삭제 실패: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 3. 렌더링 헬퍼
  // ========================================

  // const renderFieldIcon = (fieldType: string) => {
  //   switch (fieldType) {
  //     case 'stage':
  //       return <CheckSquare className="w-4 h-4 text-gray-500" />;
  //     case 'role':
  //       return <Tag className="w-4 h-4 text-gray-500" />;
  //     case 'importance':
  //       return <AlertCircle className="w-4 h-4 text-gray-500" />;
  //     case 'date':
  //       return <Calendar className="w-4 h-4 text-gray-500" />;
  //     case 'single_user':
  //       return <User className="w-4 h-4 text-gray-500" />;
  //     default:
  //       return <CheckSquare className="w-4 h-4 text-gray-500" />;
  //   }
  // };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[120] p-4">
      <div
        className={`${theme.colors.card} rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col`}
      >
        {/* Header */}
        <div
          className={`p-6 border-b ${theme.colors.border} flex justify-between items-center flex-shrink-0`}
        >
          <h2 className={`text-xl font-bold ${theme.colors.text}`}>커스텀 필드 관리</h2>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700 transition-colors">
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Main Content Area (2 Columns) */}
        <div className="flex flex-1 min-h-0">
          {/* Left Column: Field List */}
          <div
            className={`w-64 border-r ${theme.colors.border} p-4 flex flex-col flex-shrink-0 overflow-y-auto`}
          >
            <h3 className="font-semibold text-sm mb-3">프로젝트 필드</h3>

            {fields.map((field) => (
              <div
                key={field.fieldType}
                onClick={() => setSelectedField(field)}
                className={`flex items-center gap-2 p-2 rounded-md cursor-pointer transition ${
                  selectedField?.fieldType === field.fieldType
                    ? 'bg-blue-100 text-blue-700 font-medium'
                    : 'hover:bg-gray-50 text-gray-700'
                }`}
              >
                {field.icon}
                <span className="text-sm truncate">{field.name}</span>
                <span className="text-[10px] bg-gray-200 px-1 rounded">시스템</span>
              </div>
            ))}

            {/* 새 필드 추가 버튼 */}
            <button
              onClick={() => alert('새 필드 추가는 보드 생성/수정 모달에서 접근해주세요.')}
              className="mt-4 w-full py-2 border-2 border-dashed border-gray-300 rounded-md text-gray-600 hover:border-blue-500 transition-colors flex items-center justify-center gap-2 text-sm"
            >
              <Plus className="w-4 h-4" /> 새 필드 추가
            </button>
          </div>

          {/* Right Column: Field Detail & Options */}
          <div className="flex-1 p-6 overflow-y-auto">
            {selectedField ? (
              <div className="space-y-6">
                {/* Field Header */}
                <div className="border-b pb-4">
                  <div className="flex justify-between items-start">
                    <h3 className="text-xl font-bold">{selectedField.name}</h3>
                  </div>
                  <p className="text-sm text-gray-500">유형: {selectedField.fieldType}</p>
                  <p className="text-xs text-gray-400 mt-1">
                    시스템 필드는 이름을 변경하거나 삭제할 수 없습니다.
                  </p>
                </div>

                {/* Options Management */}
                <div className="space-y-4 pt-4 border-t">
                  <div className="flex items-center justify-between">
                    <h4 className="font-semibold">옵션 목록 ({fieldOptions.length}개)</h4>
                  </div>

                  {loading && (
                    <div className="flex justify-center py-8">
                      <LoadingSpinner message="옵션 로드 중..." />
                    </div>
                  )}

                  {error && (
                    <div className="p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
                      {error}
                    </div>
                  )}

                  {/* Options List */}
                  {!loading && !error && (
                    <div className="flex flex-col gap-2 max-h-96 overflow-y-auto">
                      {fieldOptions.length === 0 ? (
                        <div className="text-center py-8 text-gray-500">
                          <p>옵션이 없습니다.</p>
                          <p className="text-sm mt-2">보드 생성 모달에서 옵션을 추가해보세요.</p>
                        </div>
                      ) : (
                        fieldOptions.map((option) => (
                          <div
                            key={option.optionId}
                            className="flex justify-between items-center p-3 bg-white border rounded-md hover:border-gray-300 transition"
                          >
                            <div className="flex items-center gap-3">
                              <Menu className="w-4 h-4 text-gray-400 cursor-move" />
                              <span
                                className="w-4 h-4 rounded-full flex-shrink-0"
                                style={{ backgroundColor: option.color }}
                              ></span>
                              <div className="flex flex-col">
                                <span className="text-sm font-medium">{option.label}</span>
                                <span className="text-xs text-gray-500">값: {option.value}</span>
                              </div>
                            </div>
                            <div className="flex gap-2 items-center">
                              <span className="text-xs text-gray-400">
                                순서: {option.displayOrder}
                              </span>
                              {!option.isSystemDefault && (
                                <button
                                  onClick={() => handleDeleteOption(option.optionId, option.label)}
                                  className="p-1 hover:bg-red-100 rounded-full transition"
                                  title="삭제"
                                >
                                  <Trash2 className="w-4 h-4 text-red-600" />
                                </button>
                              )}
                              {option.isSystemDefault && (
                                <span className="text-[10px] bg-gray-200 px-2 py-1 rounded">
                                  기본
                                </span>
                              )}
                            </div>
                          </div>
                        ))
                      )}
                    </div>
                  )}

                  {/* Info Message */}
                  <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
                    <p className="text-sm text-blue-800">
                      💡 <strong>옵션 추가 방법:</strong> 보드 생성/수정 모달에서 필드 드롭다운
                      하단의 "필드 관리" 버튼을 클릭하여 새 옵션을 추가할 수 있습니다.
                    </p>
                  </div>
                </div>
              </div>
            ) : (
              <p className="text-gray-500">필드를 선택해주세요.</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
