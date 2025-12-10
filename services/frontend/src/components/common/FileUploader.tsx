// src/components/common/FileUploader.tsx
import React, { useRef } from 'react';
import { UploadCloud, X, FileText } from 'lucide-react';

interface FileUploaderProps {
  selectedFile: File | null;
  previewUrl: string | null; // 로컬 미리보기 또는 서버 URL
  onFileSelect: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onRemoveFile: () => void;
  label?: string;
  disabled?: boolean;
  existingFileName?: string; // 이미 업로드된 파일명 표시용
}

export const FileUploader: React.FC<FileUploaderProps> = ({
  selectedFile,
  previewUrl,
  onFileSelect,
  onRemoveFile,
  label = '파일 첨부',
  disabled = false,
  existingFileName,
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 표시할 파일명 결정 (새로 선택한 파일 > 기존 파일명)
  const displayFileName = selectedFile?.name || existingFileName;

  // 미리보기 이미지가 있는지 여부 (URL이 있고, 진짜 이미지인지 체크는 간단히)
  const isImage =
    previewUrl?.match(/\.(jpeg|jpg|gif|png|webp)$/i) || selectedFile?.type.startsWith('image/');

  return (
    <div className="w-full">
      <label className="block text-sm font-semibold text-gray-700 mb-2">{label}</label>

      {/* 파일이 있거나 미리보기가 있는 경우 */}
      {selectedFile || previewUrl ? (
        <div className="relative flex items-center p-3 border border-gray-300 rounded-lg bg-gray-50">
          {/* 썸네일 또는 아이콘 */}
          <div className="w-10 h-10 mr-3 flex-shrink-0 bg-gray-200 rounded-md overflow-hidden flex items-center justify-center">
            {isImage && previewUrl ? (
              <img src={previewUrl} alt="Preview" className="w-full h-full object-cover" />
            ) : (
              <FileText className="w-6 h-6 text-gray-500" />
            )}
          </div>

          {/* 파일 정보 */}
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-gray-900 truncate">
              {displayFileName || 'Unknown File'}
            </p>
            <p className="text-xs text-gray-500">
              {selectedFile ? `${(selectedFile.size / 1024 / 1024).toFixed(2)} MB` : 'Uploaded'}
            </p>
          </div>

          {/* 삭제 버튼 */}
          {!disabled && (
            <button
              type="button"
              onClick={(e) => {
                e.preventDefault();
                if (fileInputRef.current) fileInputRef.current.value = '';
                onRemoveFile();
              }}
              className="ml-2 p-1 text-gray-400 hover:text-red-500 rounded-full hover:bg-gray-200 transition"
            >
              <X className="w-5 h-5" />
            </button>
          )}
        </div>
      ) : (
        /* 파일 업로드 영역 (비어있을 때) */
        <div
          onClick={() => !disabled && fileInputRef.current?.click()}
          className={`border-2 border-dashed border-gray-300 rounded-lg p-6 flex flex-col items-center justify-center text-center cursor-pointer transition ${
            disabled ? 'opacity-50 cursor-not-allowed' : 'hover:border-blue-500 hover:bg-blue-50'
          }`}
        >
          <UploadCloud className="w-8 h-8 text-gray-400 mb-2" />
          <p className="text-sm text-gray-600">클릭하여 파일을 업로드하세요</p>
          <p className="text-xs text-gray-400 mt-1">(최대 10MB, 모든 형식 지원)</p>
        </div>
      )}

      <input
        ref={fileInputRef}
        type="file"
        className="hidden"
        onChange={onFileSelect}
        disabled={disabled}
      />
    </div>
  );
};
