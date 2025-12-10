// 💡 시간 포맷팅 헬퍼 함수
export const formatDate = (dateString: string | undefined): string => {
  if (!dateString) return '미정';
  try {
    const date = new Date(dateString);
    return date.toLocaleString('ko-KR', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return '날짜 오류';
  }
};
