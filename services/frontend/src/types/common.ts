export type IROLES =
  | 'OWNER'
  | 'ADMIN'
  | 'MEMBER'
  | 'ORGANIZER' // 아마 기존에 있던 역할
  | 'PENDING' // 💡 [추가] 워크스페이스 참여 요청 대기 상태
  | 'GUEST'; // 💡 [추가] 읽기 전용 또는 외부 게스트 역할
export type IFieldDefaultType = 'stages' | 'roles' | 'importances';
export type IFieldOption = {
  key: IFieldDefaultType;
  value: string;
};
