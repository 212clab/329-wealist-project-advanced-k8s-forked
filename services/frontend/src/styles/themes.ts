export interface Theme {
  name: string;
  cssClass: string;
  colors: {
    primary: string;
    primaryHover: string;
    primaryDark: string;
    secondary: string;
    background: string;
    card: string;
    border: string;
    text: string;
    subText: string;
    textSecondary: string; // 💡 추가된 색상: 상태 및 액션 색상
    success: string; // 성공/생성 버튼 (예: 새 조직 생성)
    successHover: string; // 성공 버튼 호버
    info: string; // 정보/링크 (예: 조직 검색하기)
    danger: string; // 오류/위험 (예: 에러 메시지)
  };
  effects: {
    shadow: string;
    headerShadow: string;
    borderWidth: string;
    cardBorderWidth: string;
    borderRadius: string;
  };
  font: {
    family: string;
    size: {
      xs: string;
      sm: string;
      base: string;
      lg: string;
      xl: string;
      xxl: string;
      '3xl': string;
    };
  };
}

// 모던 깔끔한 테마
export const modernTheme: Theme = {
  name: 'modern',
  cssClass: 'theme-modern',
  colors: {
    primary: 'bg-blue-600',
    primaryHover: 'hover:bg-blue-700',
    primaryDark: 'bg-blue-800',
    secondary: 'bg-white',
    background: 'bg-gray-50',
    card: 'bg-white',
    border: 'border-gray-200',
    text: 'text-gray-900',
    subText: 'text-gray-500',
    textSecondary: 'text-gray-500',
    success: 'bg-green-600',
    successHover: 'hover:bg-green-700',
    info: 'text-blue-600',
    danger: 'text-red-500',
  },
  effects: {
    shadow: '0 1px 3px rgba(0,0,0,0.1)',
    headerShadow: '0 2px 4px rgba(0,0,0,0.1)',
    borderWidth: 'border',
    cardBorderWidth: 'border',
    borderRadius: 'rounded-lg',
  },
  font: {
    family:
      "'Pretendard Variable', 'Pretendard', -apple-system, BlinkMacSystemFont, 'Noto Sans KR', 'Malgun Gothic', '맑은 고딕', 'Apple SD Gothic Neo', sans-serif",
    size: {
      xs: 'text-xs font-semibold',
      sm: 'text-sm font-semibold',
      base: 'text-base font-semibold',
      lg: 'text-lg font-bold',
      xl: 'text-2xl font-bold',
      xxl: 'text-3xl font-bold',
      '3xl': 'text-4xl font-extrabold',
    },
  },
};

// 모든 테마 export
export const themes = {
  modern: modernTheme,
};

export type ThemeName = keyof typeof themes;
