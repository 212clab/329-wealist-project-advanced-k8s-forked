import React, { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTheme } from '../contexts/ThemeContext'; // UI 복구용 ThemeContext 재사용

const LoadingScreen = ({ msg = '인증 정보를 처리 중...' }) => {
  const { theme } = useTheme();
  return (
    <div className={`min-h-screen ${theme.colors.background} flex items-center justify-center p-4`}>
      <div className="p-8 bg-white rounded-xl shadow-lg">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
        <h1 className="text-xl font-medium text-gray-800">{msg}</h1>
      </div>
    </div>
  );
};

// 백엔드에서 받은 토큰을 처리하고 리다이렉트하는 컴포넌트
const OAuthRedirectPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // 1. URLSearchParams를 사용하여 쿼리 파라미터 파싱
    const params = new URLSearchParams(location.search);
    const accessToken = params.get('accessToken');
    const refreshToken = params.get('refreshToken');
    const nickName = params.get('nickName');
    const email = params.get('email'); // (필요하다면)
    const userId = params.get('userId'); // (필요하다면)

    if (accessToken && refreshToken) {
      // 2. 토큰과 유저 ID를 localStorage에 저장
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', refreshToken);
      if (userId) localStorage.setItem('userId', userId);
      if (nickName) localStorage.setItem('nickName', nickName);
      if (email) localStorage.setItem('userEmail', email); // 이메일도 저장

      console.log('✅ 토큰 저장 성공:', {
        accessToken: accessToken.substring(0, 10) + '...',
        refreshToken: refreshToken.substring(0, 10) + '...',
        nickName,
        email,
        userId,
      });

      // 3. 대시보드 페이지로 이동 (Protected Route 통과)
      navigate('/dashboard', { replace: true });
    } else {
      // 4. 필수 정보가 없는 경우 에러 처리 및 로그인 페이지로 리다이렉트
      console.error('❌ OAuth 콜백 필수 정보(토큰, nickName) 누락');
      setError('로그인 정보를 확인할 수 없습니다. 다시 시도해 주세요.');

      // 5. 에러가 발생해도 바로 리다이렉트되지 않도록 3초 대기 후 이동
      const timer = setTimeout(() => {
        navigate('/', { replace: true });
      }, 3000);

      return () => clearTimeout(timer); // 클린업 함수
    }
  }, [location.search, navigate]);

  return (
    <LoadingScreen
      msg={error ? `오류 발생: ${error}` : '로그인 처리 중입니다. 잠시만 기다려 주세요...'}
    />
  );
};

export default OAuthRedirectPage;
