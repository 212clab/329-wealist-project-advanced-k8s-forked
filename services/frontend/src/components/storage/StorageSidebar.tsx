// src/components/storage/StorageSidebar.tsx - Google Drive 스타일 사이드바 (접기 기능 추가)

import React, { useState, useRef, useEffect } from 'react';
import {
  Plus,
  Upload,
  FolderPlus,
  FileUp,
  HardDrive,
  Users,
  Clock,
  Star,
  Trash2,
  Cloud,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { StorageUsage, formatFileSize } from '../../types/storage';

type NavigationSection = 'my-drive' | 'shared' | 'recent' | 'starred' | 'trash';

interface StorageSidebarProps {
  activeSection: NavigationSection;
  onSectionChange: (section: NavigationSection) => void;
  onNewFolder: () => void;
  onUpload: () => void;
  storageUsage: StorageUsage | null;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
}

interface NavItem {
  id: NavigationSection;
  label: string;
  icon: React.ReactNode;
}

export const StorageSidebar: React.FC<StorageSidebarProps> = ({
  activeSection,
  onSectionChange,
  onNewFolder,
  onUpload,
  storageUsage,
  isCollapsed,
  onToggleCollapse,
}) => {
  const [showNewMenu, setShowNewMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // 외부 클릭 시 메뉴 닫기
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setShowNewMenu(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const navItems: NavItem[] = [
    { id: 'my-drive', label: '내 드라이브', icon: <HardDrive className="w-5 h-5" /> },
    { id: 'shared', label: '공유 문서함', icon: <Users className="w-5 h-5" /> },
    { id: 'recent', label: '최근 문서함', icon: <Clock className="w-5 h-5" /> },
    { id: 'starred', label: '중요 문서함', icon: <Star className="w-5 h-5" /> },
    { id: 'trash', label: '휴지통', icon: <Trash2 className="w-5 h-5" /> },
  ];

  // 스토리지 사용량 퍼센트 계산 (기본 15GB 제한 가정)
  const storageLimit = 15 * 1024 * 1024 * 1024; // 15GB in bytes
  const usedPercent = storageUsage
    ? Math.min((storageUsage.totalSize / storageLimit) * 100, 100)
    : 0;

  return (
    <div
      className={`h-full flex flex-col bg-[#f8fafd] border-r border-gray-200 transition-all duration-300 flex-shrink-0 relative ${
        isCollapsed ? 'w-[72px]' : 'w-[240px]'
      }`}
    >
      {/* 접기/펼치기 버튼 */}
      <button
        onClick={onToggleCollapse}
        className="absolute -right-3 top-6 w-6 h-6 bg-white border border-gray-300 rounded-full shadow-md flex items-center justify-center hover:bg-gray-100 transition-colors z-30"
        title={isCollapsed ? '사이드바 펼치기' : '사이드바 접기'}
      >
        {isCollapsed ? (
          <ChevronRight className="w-4 h-4 text-gray-600" />
        ) : (
          <ChevronLeft className="w-4 h-4 text-gray-600" />
        )}
      </button>

      {/* 새로 만들기 버튼 - Google Drive 스타일 */}
      <div className="p-4 pt-5" ref={menuRef}>
        <button
          onClick={() => setShowNewMenu(!showNewMenu)}
          className={`inline-flex items-center gap-3 bg-white rounded-2xl shadow-md hover:shadow-lg hover:bg-[#edf2fc] transition-all duration-200 border border-gray-200 ${
            isCollapsed ? 'p-3 justify-center' : 'pl-4 pr-6 py-4'
          }`}
          title="새로 만들기"
        >
          <Plus className="w-6 h-6 text-gray-800" />
          {!isCollapsed && <span className="text-sm font-medium text-gray-800">새로 만들기</span>}
        </button>

        {/* 드롭다운 메뉴 */}
        {showNewMenu && (
          <div
            className={`absolute ${
              isCollapsed ? 'left-[72px]' : 'left-4'
            } top-[76px] w-[280px] bg-white rounded-lg shadow-xl border border-gray-200 py-2 z-30`}
          >
            {/* 새 폴더 */}
            <button
              onClick={() => {
                onNewFolder();
                setShowNewMenu(false);
              }}
              className="w-full flex items-center gap-4 px-4 py-3 hover:bg-gray-100 transition-colors"
            >
              <div className="w-9 h-9 flex items-center justify-center">
                <FolderPlus className="w-6 h-6 text-gray-600" />
              </div>
              <span className="text-sm text-gray-800">새 폴더</span>
            </button>

            <div className="h-px bg-gray-200 my-1 mx-4" />

            {/* 파일 업로드 */}
            <button
              onClick={() => {
                onUpload();
                setShowNewMenu(false);
              }}
              className="w-full flex items-center gap-4 px-4 py-3 hover:bg-gray-100 transition-colors"
            >
              <div className="w-9 h-9 flex items-center justify-center">
                <FileUp className="w-6 h-6 text-gray-600" />
              </div>
              <span className="text-sm text-gray-800">파일 업로드</span>
            </button>

            {/* 폴더 업로드 */}
            <button
              onClick={() => {
                onUpload();
                setShowNewMenu(false);
              }}
              className="w-full flex items-center gap-4 px-4 py-3 hover:bg-gray-100 transition-colors"
            >
              <div className="w-9 h-9 flex items-center justify-center">
                <Upload className="w-6 h-6 text-gray-600" />
              </div>
              <span className="text-sm text-gray-800">폴더 업로드</span>
            </button>
          </div>
        )}
      </div>

      {/* 네비게이션 메뉴 - Google Drive 스타일 */}
      <nav className="flex-1 px-3 py-2 overflow-y-auto">
        <ul className="space-y-0.5">
          {navItems.map((item) => (
            <li key={item.id}>
              <button
                onClick={() => onSectionChange(item.id)}
                className={`w-full flex items-center gap-3 transition-all text-sm ${
                  isCollapsed ? 'justify-center p-3 rounded-xl' : 'pl-3 pr-4 py-2 rounded-full'
                } ${
                  activeSection === item.id
                    ? 'bg-[#c2e7ff] text-[#001d35] font-medium'
                    : 'text-[#444746] hover:bg-[#e4e4e4]'
                }`}
                title={isCollapsed ? item.label : undefined}
              >
                <span className={activeSection === item.id ? 'text-[#001d35]' : 'text-[#444746]'}>
                  {item.icon}
                </span>
                {!isCollapsed && <span>{item.label}</span>}
              </button>
            </li>
          ))}
        </ul>
      </nav>

      {/* 스토리지 사용량 - Google Drive 스타일 */}
      <div className={`p-4 border-t border-gray-200 ${isCollapsed ? 'hidden' : ''}`}>
        <button className="w-full flex items-center gap-3 pl-3 pr-4 py-2 rounded-full text-sm text-[#444746] hover:bg-[#e4e4e4] transition-colors mb-3">
          <Cloud className="w-5 h-5" />
          <span>저장용량</span>
        </button>

        <div className="pl-3 pr-4">
          <div className="w-full h-1 bg-gray-300 rounded-full overflow-hidden mb-2">
            <div
              className={`h-full rounded-full transition-all ${
                usedPercent > 90
                  ? 'bg-red-500'
                  : usedPercent > 70
                  ? 'bg-yellow-500'
                  : 'bg-[#1a73e8]'
              }`}
              style={{ width: `${Math.max(usedPercent, 1)}%` }}
            />
          </div>
          <p className="text-xs text-[#5f6368]">
            15GB 중 {storageUsage ? formatFileSize(storageUsage.totalSize) : '0 B'} 사용
          </p>
        </div>
      </div>

      {/* 접힌 상태에서 스토리지 아이콘만 표시 */}
      {isCollapsed && (
        <div className="p-4 border-t border-gray-200 flex justify-center">
          <div className="relative" title={`15GB 중 ${storageUsage ? formatFileSize(storageUsage.totalSize) : '0 B'} 사용`}>
            <Cloud className="w-5 h-5 text-[#444746]" />
            {usedPercent > 70 && (
              <div
                className={`absolute -top-1 -right-1 w-2 h-2 rounded-full ${
                  usedPercent > 90 ? 'bg-red-500' : 'bg-yellow-500'
                }`}
              />
            )}
          </div>
        </div>
      )}
    </div>
  );
};
