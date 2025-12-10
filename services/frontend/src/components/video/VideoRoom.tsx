import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  X,
  Mic,
  MicOff,
  Video,
  VideoOff,
  PhoneOff,
  Users,
  Monitor,
  MessageSquare,
  Copy,
  Check,
  Minimize2,
  Maximize2,
  GripHorizontal,
  Send,
  Captions,
  CaptionsOff,
  FlipHorizontal,
  Sparkles,
} from 'lucide-react';
import { VideoRoom as VideoRoomType, videoService } from '../../api/videoService';

// Display mode types
type DisplayMode = 'full' | 'mini';

interface UserProfile {
  id: string;
  nickName: string;
  profileImageUrl?: string | null;
}

interface VideoRoomProps {
  room: VideoRoomType;
  token: string;
  wsUrl: string;
  onLeave: () => void;
  userProfile?: UserProfile | null;
}

// LiveKit connection state types
type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'error';

interface Participant {
  identity: string;
  name?: string;
  isSpeaking: boolean;
  isMuted: boolean;
  isVideoEnabled: boolean;
  isLocal: boolean;
  videoTrack?: MediaStreamTrack;
  audioTrack?: MediaStreamTrack;
}

interface ChatMessage {
  id: string;
  sender: string;
  senderName: string;
  message: string;
  timestamp: Date;
  isLocal: boolean;
}

interface TranscriptLine {
  id: string;
  speaker: string;
  text: string;
  timestamp: Date;
  isFinal: boolean;
}

// Web Speech API type declarations
interface SpeechRecognitionEvent {
  results: SpeechRecognitionResultList;
  resultIndex: number;
}

interface SpeechRecognitionResultList {
  length: number;
  item(index: number): SpeechRecognitionResult;
  [index: number]: SpeechRecognitionResult;
}

interface SpeechRecognitionResult {
  isFinal: boolean;
  length: number;
  item(index: number): SpeechRecognitionAlternative;
  [index: number]: SpeechRecognitionAlternative;
}

interface SpeechRecognitionAlternative {
  transcript: string;
  confidence: number;
}

interface SpeechRecognition extends EventTarget {
  continuous: boolean;
  interimResults: boolean;
  lang: string;
  onresult: ((event: SpeechRecognitionEvent) => void) | null;
  onerror: ((event: Event) => void) | null;
  onend: (() => void) | null;
  start(): void;
  stop(): void;
  abort(): void;
}

declare global {
  interface Window {
    SpeechRecognition: new () => SpeechRecognition;
    webkitSpeechRecognition: new () => SpeechRecognition;
  }
}

export const VideoRoom: React.FC<VideoRoomProps> = ({
  room,
  token,
  wsUrl,
  onLeave,
  userProfile,
}) => {
  const [connectionState, setConnectionState] = useState<ConnectionState>('connecting');
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [isMuted, setIsMuted] = useState(false);
  const [isVideoEnabled, setIsVideoEnabled] = useState(false); // 기본적으로 카메라 OFF
  const [isScreenSharing, setIsScreenSharing] = useState(false);
  const [isMirrored, setIsMirrored] = useState(true); // 기본적으로 좌우 반전 활성화 (셀피 모드)
  const [isBlurEnabled, setIsBlurEnabled] = useState(false); // 배경 흐림 효과
  const [showParticipants, setShowParticipants] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, _setError] = useState<string | null>(null);

  // Chat state
  const [showChat, setShowChat] = useState(false);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatInput, setChatInput] = useState('');
  const chatEndRef = useRef<HTMLDivElement>(null);

  // Mini mode state
  const [displayMode, setDisplayMode] = useState<DisplayMode>('full');
  const [miniPosition, setMiniPosition] = useState({ x: 20, y: 20 }); // bottom-right offset
  const [isDragging, setIsDragging] = useState(false);
  const dragOffsetRef = useRef({ x: 0, y: 0 });

  // Subtitle/Transcript state
  const [isSubtitleEnabled, setIsSubtitleEnabled] = useState(false);
  const [currentSubtitle, setCurrentSubtitle] = useState('');
  const [transcript, setTranscript] = useState<TranscriptLine[]>([]);
  const recognitionRef = useRef<SpeechRecognition | null>(null);

  // Join notification state
  const [joinNotification, setJoinNotification] = useState<string | null>(null);
  const previousParticipantIds = useRef<Set<string>>(new Set());

  const localVideoRef = useRef<HTMLVideoElement>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const miniContainerRef = useRef<HTMLDivElement>(null);

  // Initialize local media
  useEffect(() => {
    const initLocalMedia = async () => {
      try {
        // 오디오만 먼저 요청 (카메라는 사용자가 켤 때 요청)
        const stream = await navigator.mediaDevices.getUserMedia({
          video: false,
          audio: true,
        });
        localStreamRef.current = stream;

        // Add local participant (카메라 OFF 상태로 시작)
        setParticipants([
          {
            identity: 'local',
            name: userProfile?.nickName || '나',
            isSpeaking: false,
            isMuted: false,
            isVideoEnabled: false,
            isLocal: true,
            videoTrack: undefined,
            audioTrack: stream.getAudioTracks()[0],
          },
        ]);

        // Connect to LiveKit (simplified - real implementation would use livekit-client SDK)
        connectToRoom();
      } catch (err) {
        console.error('Failed to get audio:', err);
        // 오디오 없이도 참여 가능하도록 함
        setParticipants([
          {
            identity: 'local',
            name: userProfile?.nickName || '나',
            isSpeaking: false,
            isMuted: true,
            isVideoEnabled: false,
            isLocal: true,
            videoTrack: undefined,
            audioTrack: undefined,
          },
        ]);
        setIsMuted(true);
        connectToRoom();
      }
    };

    initLocalMedia();

    return () => {
      // Cleanup
      if (localStreamRef.current) {
        localStreamRef.current.getTracks().forEach((track) => track.stop());
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // 참여자 입장 알림 표시
  const showJoinNotification = useCallback((name: string) => {
    setJoinNotification(`${name}님이 참여하셨습니다`);
    setTimeout(() => {
      setJoinNotification(null);
    }, 3000);
  }, []);

  // 참여자 목록 갱신
  const refreshParticipants = useCallback(async () => {
    try {
      const roomParticipants = await videoService.getParticipants(room.id);
      const activeParticipants = roomParticipants.filter((p) => p.isActive);

      // 새로운 참여자 확인 (로컬 제외)
      const currentIds = new Set(activeParticipants.map((p) => p.id));
      activeParticipants.forEach((p) => {
        if (!previousParticipantIds.current.has(p.id) && p.userId !== userProfile?.id) {
          // userId의 앞 6자리만 표시
          const shortName = `사용자 ${p.userId.slice(0, 6)}`;
          showJoinNotification(shortName);
        }
      });
      previousParticipantIds.current = currentIds;

      // 참여자 목록 업데이트 (로컬 참여자 유지)
      setParticipants((prev) => {
        const localParticipant = prev.find((p) => p.isLocal);
        const remoteParticipants: Participant[] = activeParticipants
          .filter((p) => p.userId !== userProfile?.id)
          .map((p) => ({
            identity: p.id,
            name: `참여자 ${p.userId.slice(0, 6)}`,
            isSpeaking: false,
            isMuted: false,
            isVideoEnabled: false,
            isLocal: false,
            videoTrack: undefined,
            audioTrack: undefined,
          }));

        return localParticipant ? [localParticipant, ...remoteParticipants] : remoteParticipants;
      });
    } catch (error) {
      console.error('Failed to refresh participants:', error);
    }
  }, [room.id, userProfile?.id, showJoinNotification]);

  const connectToRoom = useCallback(() => {
    setConnectionState('connecting');

    // 연결 성공 후 참여자 폴링 시작
    setTimeout(() => {
      setConnectionState('connected');
      // 초기 참여자 목록 로드
      refreshParticipants();
    }, 1500);
  }, [token, wsUrl, refreshParticipants]);

  // 참여자 목록 주기적 갱신 (3초마다)
  useEffect(() => {
    if (connectionState !== 'connected') return;

    const interval = setInterval(() => {
      refreshParticipants();
    }, 3000);

    return () => clearInterval(interval);
  }, [connectionState, refreshParticipants]);

  const toggleMute = () => {
    if (localStreamRef.current) {
      const audioTrack = localStreamRef.current.getAudioTracks()[0];
      if (audioTrack) {
        audioTrack.enabled = isMuted;
        setIsMuted(!isMuted);
        setParticipants((prev) => prev.map((p) => (p.isLocal ? { ...p, isMuted: !isMuted } : p)));
      }
    }
  };

  const toggleVideo = async () => {
    if (!isVideoEnabled) {
      // 카메라 켜기: 비디오 트랙이 없으면 새로 요청
      try {
        const videoStream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: false,
        });
        const videoTrack = videoStream.getVideoTracks()[0];

        // 기존 스트림에 비디오 트랙 추가
        if (localStreamRef.current) {
          localStreamRef.current.addTrack(videoTrack);
        } else {
          localStreamRef.current = videoStream;
        }

        // 비디오 엘리먼트에 스트림 연결
        if (localVideoRef.current) {
          localVideoRef.current.srcObject = localStreamRef.current;
        }

        setIsVideoEnabled(true);
        setParticipants((prev) =>
          prev.map((p) => (p.isLocal ? { ...p, isVideoEnabled: true, videoTrack } : p)),
        );
      } catch (err) {
        console.error('Failed to enable video:', err);
        alert('카메라 접근에 실패했습니다. 권한을 확인해주세요.');
      }
    } else {
      // 카메라 끄기
      if (localStreamRef.current) {
        const videoTrack = localStreamRef.current.getVideoTracks()[0];
        if (videoTrack) {
          videoTrack.stop();
          localStreamRef.current.removeTrack(videoTrack);
        }
      }
      setIsVideoEnabled(false);
      setParticipants((prev) =>
        prev.map((p) => (p.isLocal ? { ...p, isVideoEnabled: false, videoTrack: undefined } : p)),
      );
    }
  };

  const toggleScreenShare = async () => {
    if (!isScreenSharing) {
      try {
        const screenStream = await navigator.mediaDevices.getDisplayMedia({
          video: true,
        });
        // Handle screen share (simplified)
        setIsScreenSharing(true);
        screenStream.getVideoTracks()[0].onended = () => {
          setIsScreenSharing(false);
        };
      } catch (err) {
        console.error('Screen share failed:', err);
      }
    } else {
      setIsScreenSharing(false);
    }
  };

  const handleLeave = async () => {
    // Save transcript if there's content
    if (transcript.length > 0) {
      try {
        const transcriptText = getTranscriptText();
        await videoService.saveTranscript(room.id, transcriptText);
        console.log('Transcript saved successfully');
      } catch (error) {
        console.error('Failed to save transcript:', error);
        // Continue with leave even if transcript save fails
      }
    }

    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach((track) => track.stop());
    }
    onLeave();
  };

  const copyRoomLink = () => {
    navigator.clipboard.writeText(`${window.location.origin}/video/${room.id}`);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // Chat functions
  const sendChatMessage = () => {
    if (!chatInput.trim()) return;

    const newMessage: ChatMessage = {
      id: Date.now().toString(),
      sender: 'local',
      senderName: '나',
      message: chatInput.trim(),
      timestamp: new Date(),
      isLocal: true,
    };

    setChatMessages((prev) => [...prev, newMessage]);
    setChatInput('');

    // In real implementation, send via LiveKit DataChannel
  };

  const handleChatKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendChatMessage();
    }
  };

  // Scroll to bottom when new message arrives
  useEffect(() => {
    if (chatEndRef.current && showChat) {
      chatEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [chatMessages, showChat]);

  const formatChatTime = (date: Date) => {
    return date.toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit' });
  };

  // Speech Recognition (Subtitle) functions
  const startSpeechRecognition = useCallback(() => {
    const SpeechRecognitionAPI = window.SpeechRecognition || window.webkitSpeechRecognition;
    if (!SpeechRecognitionAPI) {
      console.warn('Speech Recognition not supported');
      return;
    }

    const recognition = new SpeechRecognitionAPI();
    recognition.continuous = true;
    recognition.interimResults = true;
    recognition.lang = 'ko-KR';

    recognition.onresult = (event: SpeechRecognitionEvent) => {
      let interimTranscript = '';
      let finalTranscript = '';

      for (let i = event.resultIndex; i < event.results.length; i++) {
        const result = event.results[i];
        if (result.isFinal) {
          finalTranscript += result[0].transcript;
        } else {
          interimTranscript += result[0].transcript;
        }
      }

      // Update current subtitle (shows interim results)
      setCurrentSubtitle(interimTranscript || finalTranscript);

      // Add final transcript to history
      if (finalTranscript) {
        const newLine: TranscriptLine = {
          id: Date.now().toString(),
          speaker: '나',
          text: finalTranscript,
          timestamp: new Date(),
          isFinal: true,
        };
        setTranscript((prev) => [...prev, newLine]);

        // Clear current subtitle after adding to transcript
        setTimeout(() => setCurrentSubtitle(''), 100);
      }
    };

    recognition.onerror = (event) => {
      console.error('Speech recognition error:', event);
    };

    recognition.onend = () => {
      // Restart if still enabled
      if (isSubtitleEnabled && recognitionRef.current) {
        try {
          recognitionRef.current.start();
        } catch (e) {
          console.error('Failed to restart recognition:', e);
        }
      }
    };

    recognitionRef.current = recognition;
    recognition.start();
  }, [isSubtitleEnabled]);

  const stopSpeechRecognition = useCallback(() => {
    if (recognitionRef.current) {
      recognitionRef.current.onend = null; // Prevent restart
      recognitionRef.current.stop();
      recognitionRef.current = null;
    }
    setCurrentSubtitle('');
  }, []);

  const toggleSubtitle = () => {
    if (isSubtitleEnabled) {
      stopSpeechRecognition();
    } else {
      startSpeechRecognition();
    }
    setIsSubtitleEnabled(!isSubtitleEnabled);
  };

  // Cleanup speech recognition on unmount
  useEffect(() => {
    return () => {
      if (recognitionRef.current) {
        recognitionRef.current.onend = null;
        recognitionRef.current.stop();
      }
    };
  }, []);

  // Get full transcript as text for download
  const getTranscriptText = () => {
    return transcript
      .map((line) => `[${formatChatTime(line.timestamp)}] ${line.speaker}: ${line.text}`)
      .join('\n');
  };

  const toggleDisplayMode = () => {
    setDisplayMode((prev) => (prev === 'full' ? 'mini' : 'full'));
  };

  // Drag handlers for mini mode
  const handleDragStart = (e: React.MouseEvent) => {
    if (displayMode !== 'mini') return;
    e.preventDefault();
    setIsDragging(true);
    const container = miniContainerRef.current;
    if (container) {
      const rect = container.getBoundingClientRect();
      dragOffsetRef.current = {
        x: e.clientX - rect.left,
        y: e.clientY - rect.top,
      };
    }
  };

  useEffect(() => {
    if (!isDragging) return;

    const handleMouseMove = (e: MouseEvent) => {
      const container = miniContainerRef.current;
      if (!container) return;

      const containerWidth = container.offsetWidth;
      const containerHeight = container.offsetHeight;

      // Calculate new position (from bottom-right)
      let newX = window.innerWidth - e.clientX - containerWidth + dragOffsetRef.current.x;
      let newY = window.innerHeight - e.clientY - containerHeight + dragOffsetRef.current.y;

      // Clamp to window bounds
      newX = Math.max(0, Math.min(newX, window.innerWidth - containerWidth));
      newY = Math.max(0, Math.min(newY, window.innerHeight - containerHeight));

      setMiniPosition({ x: newX, y: newY });
    };

    const handleMouseUp = () => {
      setIsDragging(false);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging]);

  if (connectionState === 'error') {
    return (
      <div className="fixed inset-0 bg-gray-900 flex items-center justify-center z-50">
        <div className="text-center text-white">
          <VideoOff className="w-16 h-16 mx-auto mb-4 opacity-50" />
          <h2 className="text-xl font-semibold mb-2">연결 오류</h2>
          <p className="text-gray-400 mb-4">{error}</p>
          <button
            onClick={onLeave}
            className="px-6 py-2 bg-red-500 hover:bg-red-600 rounded-lg transition"
          >
            나가기
          </button>
        </div>
      </div>
    );
  }

  // Mini mode rendering
  if (displayMode === 'mini') {
    return (
      <>
        {/* 참여자 입장 알림 (미니 모드) */}
        {joinNotification && (
          <div className="fixed top-4 right-4 z-[60] animate-fade-in">
            <div className="bg-green-500 text-white px-3 py-1.5 rounded-lg shadow-lg flex items-center gap-2">
              <Users className="w-3 h-3" />
              <span className="text-xs font-medium">{joinNotification}</span>
            </div>
          </div>
        )}
        <div
          ref={miniContainerRef}
          className="fixed z-50 bg-gray-900 rounded-xl shadow-2xl overflow-hidden"
          style={{
            width: '320px',
            height: '220px',
            right: `${miniPosition.x}px`,
            bottom: `${miniPosition.y}px`,
            cursor: isDragging ? 'grabbing' : 'default',
          }}
        >
          {/* Mini mode header - draggable */}
          <div
            className="flex items-center justify-between px-3 py-2 bg-gray-800 cursor-grab active:cursor-grabbing"
            onMouseDown={handleDragStart}
          >
            <div className="flex items-center gap-2 flex-1">
              <GripHorizontal className="w-4 h-4 text-gray-500" />
              <div
                className={`w-2 h-2 rounded-full ${
                  connectionState === 'connected'
                    ? 'bg-green-500'
                    : connectionState === 'connecting'
                    ? 'bg-yellow-500 animate-pulse'
                    : 'bg-red-500'
                }`}
              />
              <span className="text-white text-sm font-medium truncate flex-1">{room.name}</span>
            </div>
            <button
              onClick={toggleDisplayMode}
              className="p-1 rounded hover:bg-gray-700 text-white transition"
              title="전체 화면"
            >
              <Maximize2 className="w-4 h-4" />
            </button>
          </div>

          {/* Mini video view */}
          <div className="relative flex-1" style={{ height: 'calc(100% - 88px)' }}>
            <video
              ref={localVideoRef}
              autoPlay
              muted
              playsInline
              className={`w-full h-full object-cover ${!isVideoEnabled ? 'hidden' : ''}`}
              style={{
                transform: isMirrored ? 'scaleX(-1)' : 'none',
                filter: isBlurEnabled ? 'blur(0px)' : 'none',
              }}
            />
            {/* 배경 흐림 효과 오버레이 (간단 버전) */}
            {isBlurEnabled && isVideoEnabled && (
              <div
                className="absolute inset-0 pointer-events-none"
                style={{
                  backdropFilter: 'blur(8px)',
                  WebkitBackdropFilter: 'blur(8px)',
                  maskImage:
                    'radial-gradient(ellipse 40% 60% at 50% 40%, transparent 30%, black 70%)',
                  WebkitMaskImage:
                    'radial-gradient(ellipse 40% 60% at 50% 40%, transparent 30%, black 70%)',
                }}
              />
            )}
            {!isVideoEnabled && (
              <div className="absolute inset-0 flex items-center justify-center bg-gray-800">
                {userProfile?.profileImageUrl ? (
                  <img
                    src={userProfile.profileImageUrl}
                    alt={userProfile.nickName}
                    className="w-16 h-16 rounded-full object-cover border-2 border-gray-600"
                  />
                ) : (
                  <div className="w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-xl font-bold text-white border-2 border-gray-600">
                    {userProfile?.nickName?.[0]?.toUpperCase() || '나'}
                  </div>
                )}
              </div>
            )}
            {/* Participant count badge */}
            <div className="absolute top-2 right-2 px-2 py-1 bg-black/50 rounded-full flex items-center gap-1">
              <Users className="w-3 h-3 text-white" />
              <span className="text-white text-xs">{participants.length}</span>
            </div>
          </div>

          {/* Mini controls */}
          <div className="flex items-center justify-center gap-2 py-2 bg-gray-800">
            <button
              onClick={toggleMute}
              className={`p-2 rounded-full transition ${
                isMuted
                  ? 'bg-red-500 hover:bg-red-600 text-white'
                  : 'bg-gray-700 hover:bg-gray-600 text-white'
              }`}
              title={isMuted ? '마이크 켜기' : '마이크 끄기'}
            >
              {isMuted ? <MicOff className="w-4 h-4" /> : <Mic className="w-4 h-4" />}
            </button>

            <button
              onClick={toggleVideo}
              className={`p-2 rounded-full transition ${
                !isVideoEnabled
                  ? 'bg-red-500 hover:bg-red-600 text-white'
                  : 'bg-gray-700 hover:bg-gray-600 text-white'
              }`}
              title={isVideoEnabled ? '카메라 끄기' : '카메라 켜기'}
            >
              {isVideoEnabled ? <Video className="w-4 h-4" /> : <VideoOff className="w-4 h-4" />}
            </button>

            <button
              onClick={handleLeave}
              className="p-2 rounded-full bg-red-500 hover:bg-red-600 text-white transition"
              title="통화 종료"
            >
              <PhoneOff className="w-4 h-4" />
            </button>
          </div>
        </div>
      </>
    );
  }

  // Full mode rendering
  return (
    <div className="fixed inset-0 bg-gray-900 z-50 flex flex-col">
      {/* 참여자 입장 알림 */}
      {joinNotification && (
        <div className="absolute top-16 right-4 z-50 animate-fade-in">
          <div className="bg-green-500 text-white px-4 py-2 rounded-lg shadow-lg flex items-center gap-2">
            <Users className="w-4 h-4" />
            <span className="text-sm font-medium">{joinNotification}</span>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-gray-800">
        <div className="flex items-center gap-3">
          <div
            className={`w-3 h-3 rounded-full ${
              connectionState === 'connected'
                ? 'bg-green-500'
                : connectionState === 'connecting'
                ? 'bg-yellow-500 animate-pulse'
                : 'bg-red-500'
            }`}
          />
          <h1 className="text-white font-medium">{room.name}</h1>
          <span className="text-gray-400 text-sm">({participants.length}명 참여중)</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={toggleDisplayMode}
            className="p-2 rounded-lg bg-gray-700 hover:bg-gray-600 text-white transition"
            title="미니 모드"
          >
            <Minimize2 className="w-5 h-5" />
          </button>
          <button
            onClick={copyRoomLink}
            className="flex items-center gap-2 px-3 py-1.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-white text-sm transition"
          >
            {copied ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
            {copied ? '복사됨!' : '초대 링크'}
          </button>
          <button
            onClick={() => setShowParticipants(!showParticipants)}
            className={`p-2 rounded-lg transition ${
              showParticipants
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 hover:bg-gray-600 text-white'
            }`}
            title="참여자"
          >
            <Users className="w-5 h-5" />
          </button>
          <button
            onClick={() => setShowChat(!showChat)}
            className={`p-2 rounded-lg transition relative ${
              showChat ? 'bg-blue-600 text-white' : 'bg-gray-700 hover:bg-gray-600 text-white'
            }`}
            title="채팅"
          >
            <MessageSquare className="w-5 h-5" />
            {chatMessages.length > 0 && !showChat && (
              <div className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 rounded-full text-xs flex items-center justify-center">
                {chatMessages.length > 9 ? '9+' : chatMessages.length}
              </div>
            )}
          </button>
        </div>
      </div>

      {/* Video Grid */}
      <div className="flex-1 p-4 overflow-hidden flex gap-4">
        <div
          className={`flex-1 grid gap-4 ${
            participants.length === 1
              ? 'grid-cols-1'
              : participants.length <= 4
              ? 'grid-cols-2'
              : participants.length <= 9
              ? 'grid-cols-3'
              : 'grid-cols-4'
          }`}
        >
          {participants.map((participant) => (
            <div
              key={participant.identity}
              className={`relative bg-gray-800 rounded-xl overflow-hidden ${
                participant.isSpeaking ? 'ring-2 ring-green-500' : ''
              }`}
            >
              {participant.isLocal ? (
                <>
                  <video
                    ref={localVideoRef}
                    autoPlay
                    muted
                    playsInline
                    className={`w-full h-full object-cover ${
                      !participant.isVideoEnabled ? 'hidden' : ''
                    }`}
                    style={{ transform: isMirrored ? 'scaleX(-1)' : 'none' }}
                  />
                  {/* 배경 흐림 효과 오버레이 */}
                  {isBlurEnabled && participant.isVideoEnabled && (
                    <div
                      className="absolute inset-0 pointer-events-none"
                      style={{
                        backdropFilter: 'blur(10px)',
                        WebkitBackdropFilter: 'blur(10px)',
                        maskImage:
                          'radial-gradient(ellipse 35% 50% at 50% 35%, transparent 40%, black 80%)',
                        WebkitMaskImage:
                          'radial-gradient(ellipse 35% 50% at 50% 35%, transparent 40%, black 80%)',
                      }}
                    />
                  )}
                </>
              ) : (
                <div className="w-full h-full flex items-center justify-center">
                  {/* Remote video would go here */}
                  <div className="w-20 h-20 rounded-full bg-gray-700 flex items-center justify-center text-2xl font-bold text-white">
                    {participant.name?.[0]?.toUpperCase() || '?'}
                  </div>
                </div>
              )}

              {!participant.isVideoEnabled && (
                <div className="absolute inset-0 flex items-center justify-center bg-gray-800">
                  {participant.isLocal && userProfile?.profileImageUrl ? (
                    <img
                      src={userProfile.profileImageUrl}
                      alt={userProfile.nickName}
                      className="w-24 h-24 rounded-full object-cover border-4 border-gray-600"
                    />
                  ) : (
                    <div className="w-24 h-24 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-3xl font-bold text-white border-4 border-gray-600">
                      {participant.name?.[0]?.toUpperCase() || '?'}
                    </div>
                  )}
                </div>
              )}

              {/* Participant info overlay */}
              <div className="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-black/70 to-transparent">
                <div className="flex items-center justify-between">
                  <span className="text-white text-sm font-medium">
                    {participant.name || participant.identity}
                    {participant.isLocal && ' (나)'}
                  </span>
                  <div className="flex items-center gap-1">
                    {participant.isMuted && <MicOff className="w-4 h-4 text-red-500" />}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Participants Sidebar */}
        {showParticipants && (
          <div className="w-64 bg-gray-800 rounded-xl p-4">
            <h3 className="text-white font-medium mb-3">참여자</h3>
            <div className="space-y-2">
              {participants.map((p) => (
                <div
                  key={p.identity}
                  className="flex items-center gap-3 px-3 py-2 bg-gray-700/50 rounded-lg"
                >
                  <div className="w-8 h-8 rounded-full bg-blue-600 flex items-center justify-center text-white text-sm font-medium">
                    {p.name?.[0]?.toUpperCase() || '?'}
                  </div>
                  <span className="text-white text-sm flex-1">
                    {p.name || p.identity}
                    {p.isLocal && ' (나)'}
                  </span>
                  <div className="flex items-center gap-1">
                    {p.isMuted ? (
                      <MicOff className="w-4 h-4 text-red-500" />
                    ) : (
                      <Mic className="w-4 h-4 text-gray-400" />
                    )}
                    {p.isVideoEnabled ? (
                      <Video className="w-4 h-4 text-gray-400" />
                    ) : (
                      <VideoOff className="w-4 h-4 text-red-500" />
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Chat Sidebar */}
        {showChat && (
          <div className="w-80 bg-gray-800 rounded-xl flex flex-col">
            <div className="p-3 border-b border-gray-700 flex items-center justify-between">
              <h3 className="text-white font-medium">채팅</h3>
              <button
                onClick={() => setShowChat(false)}
                className="p-1 rounded hover:bg-gray-700 text-gray-400 transition"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-3 space-y-3">
              {chatMessages.length === 0 ? (
                <div className="text-center text-gray-500 text-sm py-8">
                  <MessageSquare className="w-8 h-8 mx-auto mb-2 opacity-50" />
                  <p>채팅을 시작해보세요</p>
                </div>
              ) : (
                chatMessages.map((msg) => (
                  <div
                    key={msg.id}
                    className={`flex flex-col ${msg.isLocal ? 'items-end' : 'items-start'}`}
                  >
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs text-gray-400">{msg.senderName}</span>
                      <span className="text-xs text-gray-500">{formatChatTime(msg.timestamp)}</span>
                    </div>
                    <div
                      className={`max-w-[80%] px-3 py-2 rounded-lg text-sm ${
                        msg.isLocal ? 'bg-blue-600 text-white' : 'bg-gray-700 text-white'
                      }`}
                    >
                      {msg.message}
                    </div>
                  </div>
                ))
              )}
              <div ref={chatEndRef} />
            </div>

            {/* Input */}
            <div className="p-3 border-t border-gray-700">
              <div className="flex gap-2">
                <input
                  type="text"
                  value={chatInput}
                  onChange={(e) => setChatInput(e.target.value)}
                  onKeyPress={handleChatKeyPress}
                  placeholder="메시지 입력..."
                  className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  onClick={sendChatMessage}
                  disabled={!chatInput.trim()}
                  className="p-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed rounded-lg text-white transition"
                >
                  <Send className="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Subtitle Overlay */}
      {currentSubtitle && (
        <div className="absolute bottom-24 left-1/2 transform -translate-x-1/2 max-w-[80%]">
          <div className="bg-black/80 text-white px-4 py-2 rounded-lg text-lg text-center">
            {currentSubtitle}
          </div>
        </div>
      )}

      {/* Controls */}
      <div className="flex items-center justify-center gap-4 py-4 bg-gray-800">
        <button
          onClick={toggleMute}
          className={`p-4 rounded-full transition ${
            isMuted
              ? 'bg-red-500 hover:bg-red-600 text-white'
              : 'bg-gray-700 hover:bg-gray-600 text-white'
          }`}
          title={isMuted ? '마이크 켜기' : '마이크 끄기'}
        >
          {isMuted ? <MicOff className="w-6 h-6" /> : <Mic className="w-6 h-6" />}
        </button>

        <button
          onClick={toggleVideo}
          className={`p-4 rounded-full transition ${
            !isVideoEnabled
              ? 'bg-red-500 hover:bg-red-600 text-white'
              : 'bg-gray-700 hover:bg-gray-600 text-white'
          }`}
          title={isVideoEnabled ? '카메라 끄기' : '카메라 켜기'}
        >
          {isVideoEnabled ? <Video className="w-6 h-6" /> : <VideoOff className="w-6 h-6" />}
        </button>

        <button
          onClick={toggleScreenShare}
          className={`p-4 rounded-full transition ${
            isScreenSharing
              ? 'bg-blue-500 hover:bg-blue-600 text-white'
              : 'bg-gray-700 hover:bg-gray-600 text-white'
          }`}
          title={isScreenSharing ? '화면 공유 중지' : '화면 공유'}
        >
          <Monitor className="w-6 h-6" />
        </button>

        <button
          onClick={toggleSubtitle}
          className={`p-4 rounded-full transition ${
            isSubtitleEnabled
              ? 'bg-yellow-500 hover:bg-yellow-600 text-white'
              : 'bg-gray-700 hover:bg-gray-600 text-white'
          }`}
          title={isSubtitleEnabled ? '자막 끄기' : '자막 켜기 (음성인식)'}
        >
          {isSubtitleEnabled ? (
            <Captions className="w-6 h-6" />
          ) : (
            <CaptionsOff className="w-6 h-6" />
          )}
        </button>

        {/* 좌우 반전 (미러) 버튼 */}
        <button
          onClick={() => setIsMirrored(!isMirrored)}
          className={`p-4 rounded-full transition ${
            isMirrored
              ? 'bg-purple-500 hover:bg-purple-600 text-white'
              : 'bg-gray-700 hover:bg-gray-600 text-white'
          }`}
          title={isMirrored ? '좌우 반전 끄기' : '좌우 반전 켜기'}
        >
          <FlipHorizontal className="w-6 h-6" />
        </button>

        {/* 배경 흐림 버튼 */}
        <button
          onClick={() => setIsBlurEnabled(!isBlurEnabled)}
          className={`p-4 rounded-full transition ${
            isBlurEnabled
              ? 'bg-pink-500 hover:bg-pink-600 text-white'
              : 'bg-gray-700 hover:bg-gray-600 text-white'
          }`}
          title={isBlurEnabled ? '배경 흐림 끄기' : '배경 흐림 켜기'}
        >
          <Sparkles className="w-6 h-6" />
        </button>

        <button
          onClick={handleLeave}
          className="p-4 rounded-full bg-red-500 hover:bg-red-600 text-white transition"
          title="통화 종료"
        >
          <PhoneOff className="w-6 h-6" />
        </button>
      </div>
    </div>
  );
};

export default VideoRoom;
