// src/components/storage/StorageHeader.tsx - Google Drive 스타일 헤더

import React, { useState, useCallback } from 'react';
import {
  Search,
  LayoutGrid,
  List,
  ChevronRight,
  Trash2,
  RotateCcw,
  ChevronDown,
  Info,
  X,
  SlidersHorizontal,
} from 'lucide-react';
import type { BreadcrumbItem, ViewMode, SortBy, SortDirection } from '../../types/storage';

interface StorageHeaderProps {
  breadcrumbs: BreadcrumbItem[];
  onBreadcrumbClick: (item: BreadcrumbItem) => void;
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
  sortBy: SortBy;
  sortDirection: SortDirection;
  onSort: (sortBy: SortBy) => void;
  searchQuery: string;
  onSearch: (query: string) => void;
  onSearchChange: (query: string) => void;
  selectedCount: number;
  onDelete: () => void;
  onRestore?: () => void;
  onEmptyTrash?: () => void;
  isTrash: boolean;
}

export const StorageHeader: React.FC<StorageHeaderProps> = ({
  breadcrumbs,
  onBreadcrumbClick,
  viewMode,
  onViewModeChange,
  sortBy,
  sortDirection,
  onSort,
  searchQuery,
  onSearch,
  onSearchChange,
  selectedCount,
  onDelete,
  onRestore,
  onEmptyTrash,
  isTrash,
}) => {
  const [showSortMenu, setShowSortMenu] = useState(false);
  const [isSearchFocused, setIsSearchFocused] = useState(false);

  const handleSearchSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      onSearch(searchQuery);
    },
    [searchQuery, onSearch]
  );

  const sortOptions: { value: SortBy; label: string }[] = [
    { value: 'name', label: '이름' },
    { value: 'modifiedAt', label: '마지막 수정일' },
    { value: 'size', label: '저장용량 사용량' },
    { value: 'type', label: '유형' },
  ];

  return (
    <div className="bg-[#f8fafd]">
      {/* 상단 검색 영역 - Google Drive 스타일 */}
      <div className="px-4 py-2 flex items-center gap-4">
        {/* 검색 바 - Google Drive 스타일 */}
        <form onSubmit={handleSearchSubmit} className="flex-1 max-w-3xl">
          <div
            className={`relative flex items-center transition-all duration-200 ${
              isSearchFocused
                ? 'bg-white shadow-lg rounded-lg'
                : 'bg-[#e9eef6] hover:bg-[#dde3ea] rounded-full'
            }`}
          >
            <div className="pl-4">
              <Search className="w-5 h-5 text-[#5f6368]" />
            </div>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => onSearchChange(e.target.value)}
              onFocus={() => setIsSearchFocused(true)}
              onBlur={() => setIsSearchFocused(false)}
              placeholder="드라이브에서 검색"
              className={`flex-1 bg-transparent px-4 py-3 text-[#1f1f1f] placeholder-[#5f6368] focus:outline-none text-base ${
                isSearchFocused ? '' : ''
              }`}
            />
            {searchQuery && (
              <button
                type="button"
                onClick={() => onSearchChange('')}
                className="p-2 mr-1 text-[#5f6368] hover:bg-gray-200 rounded-full"
              >
                <X className="w-5 h-5" />
              </button>
            )}
            <button
              type="button"
              className="p-2 mr-2 text-[#5f6368] hover:bg-gray-200 rounded-full"
              title="검색 옵션"
            >
              <SlidersHorizontal className="w-5 h-5" />
            </button>
          </div>
        </form>

        {/* 뷰 모드 & 정보 버튼 */}
        <div className="flex items-center gap-1">
          <button
            onClick={() => onViewModeChange('list')}
            className={`p-2 rounded-full transition-colors ${
              viewMode === 'list'
                ? 'bg-[#c2e7ff] text-[#001d35]'
                : 'text-[#5f6368] hover:bg-gray-200'
            }`}
            title="목록"
          >
            <List className="w-5 h-5" />
          </button>
          <button
            onClick={() => onViewModeChange('grid')}
            className={`p-2 rounded-full transition-colors ${
              viewMode === 'grid'
                ? 'bg-[#c2e7ff] text-[#001d35]'
                : 'text-[#5f6368] hover:bg-gray-200'
            }`}
            title="그리드"
          >
            <LayoutGrid className="w-5 h-5" />
          </button>
          <button
            className="p-2 rounded-full text-[#5f6368] hover:bg-gray-200"
            title="세부정보 보기"
          >
            <Info className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* 하단 브레드크럼 & 액션 영역 */}
      <div className="px-6 py-2 flex items-center justify-between border-b border-gray-200">
        {/* 브레드크럼 - Google Drive 스타일 */}
        <nav className="flex items-center gap-0.5">
          {breadcrumbs.map((item, index) => (
            <React.Fragment key={item.id ?? 'root'}>
              {index > 0 && <ChevronRight className="w-5 h-5 text-[#5f6368] mx-1" />}
              <button
                onClick={() => onBreadcrumbClick(item)}
                className={`px-2 py-1.5 rounded-lg transition-colors text-sm ${
                  index === breadcrumbs.length - 1
                    ? 'font-medium text-[#1f1f1f]'
                    : 'text-[#5f6368] hover:bg-gray-200'
                }`}
              >
                {item.name}
              </button>
              {index === breadcrumbs.length - 1 && (
                <ChevronDown className="w-4 h-4 text-[#5f6368] ml-0.5" />
              )}
            </React.Fragment>
          ))}
        </nav>

        {/* 액션 버튼들 */}
        <div className="flex items-center gap-2">
          {/* 선택된 항목이 있을 때 */}
          {selectedCount > 0 && (
            <div className="flex items-center gap-2 mr-4 px-3 py-1.5 bg-[#e8f0fe] rounded-full">
              <span className="text-sm text-[#1967d2] font-medium">
                {selectedCount}개 선택됨
              </span>

              {isTrash && onRestore && (
                <button
                  onClick={onRestore}
                  className="flex items-center gap-1.5 px-3 py-1 text-sm text-[#1967d2] hover:bg-[#d2e3fc] rounded-full transition"
                >
                  <RotateCcw className="w-4 h-4" />
                  복원
                </button>
              )}

              <button
                onClick={onDelete}
                className="flex items-center gap-1.5 px-3 py-1 text-sm text-[#c5221f] hover:bg-[#fce8e6] rounded-full transition"
              >
                <Trash2 className="w-4 h-4" />
                {isTrash ? '영구 삭제' : '삭제'}
              </button>
            </div>
          )}

          {/* 휴지통 비우기 */}
          {isTrash && onEmptyTrash && selectedCount === 0 && (
            <button
              onClick={onEmptyTrash}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-[#c5221f] hover:bg-[#fce8e6] rounded-lg transition"
            >
              <Trash2 className="w-4 h-4" />
              휴지통 비우기
            </button>
          )}

          {/* 정렬 - Google Drive 스타일 */}
          {!isTrash && (
            <div className="relative">
              <button
                onClick={() => setShowSortMenu(!showSortMenu)}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-[#5f6368] hover:bg-gray-200 rounded-lg transition"
              >
                이름
                <ChevronDown className="w-4 h-4" />
              </button>

              {showSortMenu && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setShowSortMenu(false)}
                  />
                  <div className="absolute right-0 top-full mt-1 w-48 bg-white rounded-lg shadow-xl border border-gray-200 py-1 z-20">
                    {sortOptions.map((option) => (
                      <button
                        key={option.value}
                        onClick={() => {
                          onSort(option.value);
                          setShowSortMenu(false);
                        }}
                        className={`w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center justify-between ${
                          sortBy === option.value ? 'text-[#1967d2] bg-[#e8f0fe]' : 'text-[#3c4043]'
                        }`}
                      >
                        {option.label}
                        {sortBy === option.value && (
                          <span className="text-xs text-[#5f6368]">
                            {sortDirection === 'asc' ? '↑' : '↓'}
                          </span>
                        )}
                      </button>
                    ))}
                  </div>
                </>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
