import axios from 'axios';

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost';

export interface VideoRoom {
  id: string;
  name: string;
  workspaceId: string;
  creatorId: string;
  maxParticipants: number;
  isActive: boolean;
  participantCount: number;
  participants: VideoParticipant[];
  createdAt: string;
  updatedAt: string;
}

export interface VideoParticipant {
  id: string;
  userId: string;
  joinedAt: string;
  leftAt?: string;
  isActive: boolean;
}

export interface CreateRoomRequest {
  name: string;
  workspaceId: string;
  maxParticipants?: number;
}

export interface JoinRoomResponse {
  room: VideoRoom;
  token: string;
  wsUrl: string;
}

export interface CallHistoryParticipant {
  userId: string;
  joinedAt: string;
  leftAt: string;
  durationSeconds: number;
}

export interface CallHistory {
  id: string;
  roomName: string;
  workspaceId: string;
  creatorId: string;
  startedAt: string;
  endedAt: string;
  durationSeconds: number;
  totalParticipants: number;
  participants: CallHistoryParticipant[];
}

export interface CallHistoryResponse {
  success: boolean;
  data: CallHistory[];
  total: number;
  limit: number;
  offset: number;
}

export interface Transcript {
  id: string;
  callHistoryId: string;
  roomId: string;
  content: string;
  createdAt: string;
}

interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: {
    code: string;
    message: string;
  };
}

const getAuthHeader = () => {
  const token = localStorage.getItem('accessToken');
  return token ? { Authorization: `Bearer ${token}` } : {};
};

export const videoService = {
  // Create a new video room
  async createRoom(request: CreateRoomRequest): Promise<VideoRoom> {
    const response = await axios.post<ApiResponse<VideoRoom>>(
      `${API_BASE}/api/video/rooms`,
      request,
      { headers: getAuthHeader() }
    );
    return response.data.data;
  },

  // Get rooms for a workspace
  async getWorkspaceRooms(workspaceId: string, activeOnly: boolean = true): Promise<VideoRoom[]> {
    const response = await axios.get<ApiResponse<VideoRoom[]>>(
      `${API_BASE}/api/video/rooms/workspace/${workspaceId}`,
      {
        params: { active: activeOnly },
        headers: getAuthHeader(),
      }
    );
    return response.data.data || [];
  },

  // Get room details
  async getRoom(roomId: string): Promise<VideoRoom> {
    const response = await axios.get<ApiResponse<VideoRoom>>(
      `${API_BASE}/api/video/rooms/${roomId}`,
      { headers: getAuthHeader() }
    );
    return response.data.data;
  },

  // Join a video room
  async joinRoom(roomId: string, userName?: string): Promise<JoinRoomResponse> {
    const response = await axios.post<ApiResponse<JoinRoomResponse>>(
      `${API_BASE}/api/video/rooms/${roomId}/join`,
      {},
      {
        params: userName ? { userName } : {},
        headers: getAuthHeader(),
      }
    );
    return response.data.data;
  },

  // Leave a video room
  async leaveRoom(roomId: string): Promise<void> {
    await axios.post(
      `${API_BASE}/api/video/rooms/${roomId}/leave`,
      {},
      { headers: getAuthHeader() }
    );
  },

  // End a video room (creator only)
  async endRoom(roomId: string): Promise<void> {
    await axios.post(
      `${API_BASE}/api/video/rooms/${roomId}/end`,
      {},
      { headers: getAuthHeader() }
    );
  },

  // Get room participants
  async getParticipants(roomId: string): Promise<VideoParticipant[]> {
    const response = await axios.get<ApiResponse<VideoParticipant[]>>(
      `${API_BASE}/api/video/rooms/${roomId}/participants`,
      { headers: getAuthHeader() }
    );
    return response.data.data || [];
  },

  // Get call history for a workspace
  async getWorkspaceCallHistory(
    workspaceId: string,
    limit: number = 20,
    offset: number = 0
  ): Promise<CallHistoryResponse> {
    const response = await axios.get<CallHistoryResponse>(
      `${API_BASE}/api/video/history/workspace/${workspaceId}`,
      {
        params: { limit, offset },
        headers: getAuthHeader(),
      }
    );
    return response.data;
  },

  // Get current user's call history
  async getMyCallHistory(
    limit: number = 20,
    offset: number = 0
  ): Promise<CallHistoryResponse> {
    const response = await axios.get<CallHistoryResponse>(
      `${API_BASE}/api/video/history/me`,
      {
        params: { limit, offset },
        headers: getAuthHeader(),
      }
    );
    return response.data;
  },

  // Get single call history by ID
  async getCallHistory(historyId: string): Promise<CallHistory | null> {
    try {
      const response = await axios.get<ApiResponse<CallHistory>>(
        `${API_BASE}/api/video/history/${historyId}`,
        { headers: getAuthHeader() }
      );
      return response.data.data;
    } catch (error) {
      console.error('Failed to fetch call history:', error);
      return null;
    }
  },

  // Save transcript for a room
  async saveTranscript(roomId: string, content: string): Promise<Transcript> {
    const response = await axios.post<ApiResponse<Transcript>>(
      `${API_BASE}/api/video/rooms/${roomId}/transcript`,
      { content },
      { headers: getAuthHeader() }
    );
    return response.data.data;
  },

  // Get transcript for a call history
  async getTranscript(historyId: string): Promise<Transcript | null> {
    const response = await axios.get<ApiResponse<Transcript | null>>(
      `${API_BASE}/api/video/history/${historyId}/transcript`,
      { headers: getAuthHeader() }
    );
    return response.data.data;
  },
};

export default videoService;
