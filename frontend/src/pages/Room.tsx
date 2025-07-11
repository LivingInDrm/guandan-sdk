import { useEffect, useState } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { ArrowLeft, Users, Wifi, WifiOff, AlertCircle } from 'lucide-react';

import Table from '../components/Table';
import Hand from '../components/Hand';
import PlayerInfo from '../components/PlayerInfo';
import ControlBar from '../components/ControlBar';

import { 
  useConnectionState, 
  useGameState, 
  usePlayerState, 
  useErrorState,
  initializePlayerSeat
} from '../store';
import { SeatID } from '../types';

const Room: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  
  const { isConnected, connectionStatus, connect, disconnect } = useConnectionState();
  const { players, status, currentDeal } = useGameState();
  const { mySeat, myHand } = usePlayerState();
  const { } = useErrorState();

  const [isInitialized, setIsInitialized] = useState(false);
  const [showConnectionError, setShowConnectionError] = useState(false);

  // Initialize player seat and connection
  useEffect(() => {
    if (!roomId) {
      navigate('/');
      return;
    }

    const seatParam = searchParams.get('seat');
    if (seatParam === null) {
      navigate('/');
      return;
    }

    const seatNumber = parseInt(seatParam);
    if (isNaN(seatNumber) || seatNumber < 0 || seatNumber > 3) {
      navigate('/');
      return;
    }

    // Convert seat number to SeatID
    const seatMap: SeatID[] = ['east', 'south', 'west', 'north'];
    const seat = seatMap[seatNumber];
    
    // Initialize player seat
    initializePlayerSeat(seat);
    
    // Connect to WebSocket
    const wsUrl = `ws://${window.location.host}/api/room/${roomId}/ws?seat=${seatNumber}`;
    connect(wsUrl);
    
    setIsInitialized(true);

    // Cleanup on unmount
    return () => {
      disconnect();
    };
  }, [roomId, searchParams, navigate, connect, disconnect]);

  // Handle connection errors
  useEffect(() => {
    if (connectionStatus === 'error') {
      setShowConnectionError(true);
    } else {
      setShowConnectionError(false);
    }
  }, [connectionStatus]);

  // Handle back button
  const handleBack = () => {
    disconnect();
    navigate('/');
  };

  // Get player by seat
  const getPlayerBySeat = (seat: SeatID) => {
    return players.find(p => p.seat === seat);
  };

  // Get current player
  const getCurrentPlayer = () => {
    if (!currentDeal) return null;
    return getPlayerBySeat(currentDeal.currentTurn);
  };

  // Arrange players for display
  const arrangePlayersForDisplay = () => {
    if (!mySeat) return { top: null, left: null, right: null, bottom: null };

    const seatOrder: SeatID[] = ['east', 'south', 'west', 'north'];
    const myIndex = seatOrder.indexOf(mySeat);
    
    // Arrange players relative to my position
    const arrangement = {
      bottom: getPlayerBySeat(mySeat), // Me at bottom
      right: getPlayerBySeat(seatOrder[(myIndex + 1) % 4]), // Next player clockwise
      top: getPlayerBySeat(seatOrder[(myIndex + 2) % 4]), // Opposite player
      left: getPlayerBySeat(seatOrder[(myIndex + 3) % 4]), // Previous player clockwise
    };

    return arrangement;
  };

  const currentPlayer = getCurrentPlayer();
  const playersArrangement = arrangePlayersForDisplay();

  if (!isInitialized) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-600">正在进入房间...</p>
        </div>
      </div>
    );
  }

  return (
    <DndProvider backend={HTML5Backend}>
      <div className="min-h-screen bg-gradient-to-br from-green-50 to-blue-50">
        {/* Header */}
        <div className="bg-white shadow-md p-4">
          <div className="max-w-6xl mx-auto flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <button
                onClick={handleBack}
                className="flex items-center space-x-2 text-gray-600 hover:text-gray-800"
              >
                <ArrowLeft size={20} />
                <span>返回大厅</span>
              </button>
              
              <div className="text-lg font-semibold text-gray-800">
                房间 {roomId}
              </div>
            </div>

            <div className="flex items-center space-x-4">
              {/* Connection status */}
              <div className="flex items-center space-x-2">
                {isConnected ? (
                  <>
                    <Wifi size={16} className="text-green-500" />
                    <span className="text-green-600 text-sm">已连接</span>
                  </>
                ) : (
                  <>
                    <WifiOff size={16} className="text-red-500" />
                    <span className="text-red-600 text-sm">未连接</span>
                  </>
                )}
              </div>

              {/* Player count */}
              <div className="flex items-center space-x-2">
                <Users size={16} className="text-gray-500" />
                <span className="text-gray-600 text-sm">
                  {players.length}/4 人
                </span>
              </div>

              {/* Game status */}
              <div className={`px-3 py-1 rounded-full text-sm font-medium ${
                status === 'waiting' ? 'bg-yellow-100 text-yellow-800' :
                status === 'playing' ? 'bg-green-100 text-green-800' :
                'bg-gray-100 text-gray-800'
              }`}>
                {status === 'waiting' ? '等待中' : 
                 status === 'playing' ? '游戏中' : '已结束'}
              </div>
            </div>
          </div>
        </div>

        {/* Connection error */}
        {showConnectionError && (
          <div className="bg-red-50 border border-red-200 p-4">
            <div className="max-w-6xl mx-auto flex items-center space-x-2">
              <AlertCircle size={20} className="text-red-500" />
              <span className="text-red-700">连接服务器失败，请刷新页面重试</span>
            </div>
          </div>
        )}

        {/* Game area */}
        <div className="max-w-6xl mx-auto p-4">
          <div className="grid grid-cols-12 grid-rows-8 gap-4 h-[calc(100vh-200px)]">
            {/* Top player */}
            <div className="col-span-4 col-start-5 row-span-1 row-start-1">
              {playersArrangement.top && (
                <PlayerInfo
                  player={playersArrangement.top}
                  isCurrentPlayer={currentPlayer?.seat === playersArrangement.top.seat}
                  position="top"
                />
              )}
            </div>

            {/* Left player */}
            <div className="col-span-3 col-start-1 row-span-2 row-start-3">
              {playersArrangement.left && (
                <PlayerInfo
                  player={playersArrangement.left}
                  isCurrentPlayer={currentPlayer?.seat === playersArrangement.left.seat}
                  position="left"
                />
              )}
            </div>

            {/* Right player */}
            <div className="col-span-3 col-start-10 row-span-2 row-start-3">
              {playersArrangement.right && (
                <PlayerInfo
                  player={playersArrangement.right}
                  isCurrentPlayer={currentPlayer?.seat === playersArrangement.right.seat}
                  position="right"
                />
              )}
            </div>

            {/* Table (center) */}
            <div className="col-span-6 col-start-4 row-span-3 row-start-2">
              <Table />
            </div>

            {/* Trump and game info */}
            <div className="col-span-2 col-start-1 row-span-1 row-start-6">
              {currentDeal && (
                <div className="bg-white rounded-lg p-3 shadow-md">
                  <div className="text-sm font-medium text-gray-800 mb-1">
                    主牌信息
                  </div>
                  <div className="text-lg font-bold text-red-600">
                    {currentDeal.trump || '待定'}
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    第 {currentDeal.dealId ? (currentDeal.dealId.split('-')[1] || '1') : '1'} 局
                  </div>
                </div>
              )}
            </div>

            {/* Game phase info */}
            <div className="col-span-2 col-start-11 row-span-1 row-start-6">
              {currentDeal && (
                <div className="bg-white rounded-lg p-3 shadow-md">
                  <div className="text-sm font-medium text-gray-800 mb-1">
                    游戏阶段
                  </div>
                  <div className="text-sm text-blue-600">
                    {currentDeal.phase === 'in_progress' ? '进行中' : currentDeal.phase}
                  </div>
                </div>
              )}
            </div>

            {/* My hand */}
            <div className="col-span-12 row-span-1 row-start-7">
              <Hand cards={myHand} />
            </div>

            {/* My player info */}
            <div className="col-span-4 col-start-5 row-span-1 row-start-8">
              {playersArrangement.bottom && (
                <PlayerInfo
                  player={playersArrangement.bottom}
                  isCurrentPlayer={currentPlayer?.seat === playersArrangement.bottom.seat}
                  isMySeat={true}
                  position="bottom"
                />
              )}
            </div>
          </div>
        </div>

        {/* Control bar */}
        <div className="fixed bottom-0 left-0 right-0">
          <ControlBar />
        </div>
      </div>
    </DndProvider>
  );
};

export default Room;