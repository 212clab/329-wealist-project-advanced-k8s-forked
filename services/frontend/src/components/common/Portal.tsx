import React from 'react';
import ReactDOM from 'react-dom';

interface PortalProps {
  children: React.ReactNode;
}

const Portal: React.FC<PortalProps> = ({ children }) => {
  // 💡 body에 모달을 삽입할 div 요소를 생성하거나 찾습니다.
  const el = document.getElementById('modal-root') || document.createElement('div');

  if (!document.getElementById('modal-root')) {
    el.id = 'modal-root';
    document.body.appendChild(el);
  }

  return ReactDOM.createPortal(children, el);
};

export default Portal;
