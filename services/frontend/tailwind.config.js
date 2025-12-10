/** @type {import('tailwindcss').Config} */
export default {
  // 1. Tailwind가 스캔할 파일 경로를 지정합니다. (중요!)
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}', // src 폴더 내부의 모든 JS/TS/JSX/TSX 파일
  ],
  theme: {
    extend: {},
  },
  plugins: [],
};
