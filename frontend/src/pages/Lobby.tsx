import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Users, Play, RefreshCw } from 'lucide-react';

interface Room {
  roomId: string;
  playerCount: number;
  maxPlayers: number;
  isEmpty: boolean;
}

const Lobby: React.FC = () => {
  const navigate = useNavigate();
  const [roomName, setRoomName] = useState('');
  const [joinRoomId, setJoinRoomId] = useState('');
  const [selectedSeat, setSelectedSeat] = useState<number>(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [rooms, setRooms] = useState<Room[]>([]);
  const [refreshing, setRefreshing] = useState(false);

  const seatNames = ['东 (East)', '南 (South)', '西 (West)', '北 (North)'];

  const handleCreateRoom = async () => {
    if (!roomName.trim()) {
      setError('请输入房间名称');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/room', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          roomName: roomName.trim(),
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || '创建房间失败');
      }

      const data = await response.json();
      
      // Navigate to room with the created room ID
      navigate(`/room/${data.roomId}?seat=${selectedSeat}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建房间失败');
    } finally {
      setLoading(false);
    }
  };

  const handleJoinRoom = async (roomId?: string) => {
    const targetRoomId = roomId || joinRoomId;
    
    if (!targetRoomId.trim()) {
      setError('请输入房间ID');
      return;
    }

    setLoading(true);
    setError('');

    try {
      // First, get room info to check available seats
      const roomInfoResponse = await fetch(`/api/room/${targetRoomId}`);
      if (!roomInfoResponse.ok) {
        throw new Error('房间不存在或无法访问');
      }
      
      const roomInfo = await roomInfoResponse.json();
      const occupiedSeats = roomInfo.players.map((player: any) => player.seat);
      
      // Check if selected seat is available
      if (occupiedSeats.includes(selectedSeat)) {
        // Find first available seat
        let availableSeat = -1;
        for (let seat = 0; seat < 4; seat++) {
          if (!occupiedSeats.includes(seat)) {
            availableSeat = seat;
            break;
          }
        }
        
        if (availableSeat === -1) {
          throw new Error('房间已满，无法加入');
        }
        
        // Use the available seat instead
        const seatUsed = availableSeat;
        
        const response = await fetch(`/api/room/${targetRoomId}/join`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            seat: seatUsed,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || '加入房间失败');
        }

        // Navigate to room with the actual seat used
        navigate(`/room/${targetRoomId}?seat=${seatUsed}`);
      } else {
        // Selected seat is available, use it
        const response = await fetch(`/api/room/${targetRoomId}/join`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            seat: selectedSeat,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || '加入房间失败');
        }

        // Navigate to room
        navigate(`/room/${targetRoomId}?seat=${selectedSeat}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加入房间失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRefreshRooms = async () => {
    setRefreshing(true);
    
    try {
      const response = await fetch('/api/rooms');
      if (response.ok) {
        const data = await response.json();
        setRooms(data);
      }
    } catch (err) {
      console.error('Failed to refresh rooms:', err);
    } finally {
      setRefreshing(false);
    }
  };

  React.useEffect(() => {
    handleRefreshRooms();
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-blue-50 p-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-2">掼蛋游戏</h1>
          <p className="text-gray-600">选择座位，创建或加入房间开始游戏</p>
        </div>

        {/* Error message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-600">{error}</p>
          </div>
        )}

        {/* Seat selection */}
        <div className="mb-8 bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-800">选择座位</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {seatNames.map((name, index) => (
              <button
                key={index}
                onClick={() => setSelectedSeat(index)}
                className={`p-4 rounded-lg border-2 transition-all ${
                  selectedSeat === index
                    ? 'border-blue-500 bg-blue-50 text-blue-700'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <div className="text-center">
                  <div className="text-lg font-medium">{name}</div>
                  <div className="text-sm text-gray-500">座位 {index}</div>
                </div>
              </button>
            ))}
          </div>
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          {/* Create Room */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4 text-gray-800 flex items-center">
              <Plus className="mr-2" size={20} />
              创建房间
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  房间名称
                </label>
                <input
                  type="text"
                  value={roomName}
                  onChange={(e) => setRoomName(e.target.value)}
                  placeholder="输入房间名称"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  disabled={loading}
                />
              </div>
              <button
                onClick={handleCreateRoom}
                disabled={loading}
                className="w-full bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {loading ? '创建中...' : '创建房间'}
              </button>
            </div>
          </div>

          {/* Join Room */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4 text-gray-800 flex items-center">
              <Users className="mr-2" size={20} />
              加入房间
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  房间ID
                </label>
                <input
                  type="text"
                  value={joinRoomId}
                  onChange={(e) => setJoinRoomId(e.target.value)}
                  placeholder="输入房间ID"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                  disabled={loading}
                />
              </div>
              <button
                onClick={() => handleJoinRoom()}
                disabled={loading}
                className="w-full bg-green-500 text-white py-2 px-4 rounded-md hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {loading ? '加入中...' : '加入房间'}
              </button>
            </div>
          </div>
        </div>

        {/* Room List */}
        <div className="mt-8 bg-white rounded-lg shadow-md p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold text-gray-800">房间列表</h2>
            <button
              onClick={handleRefreshRooms}
              disabled={refreshing}
              className="flex items-center px-4 py-2 text-gray-600 hover:text-gray-800 disabled:opacity-50 transition-colors"
            >
              <RefreshCw className={`mr-2 ${refreshing ? 'animate-spin' : ''}`} size={16} />
              刷新
            </button>
          </div>

          {rooms.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <Users size={48} className="mx-auto mb-4 opacity-50" />
              <p>暂无房间</p>
            </div>
          ) : (
            <div className="space-y-3">
              {rooms.map((room) => (
                <div
                  key={room.roomId}
                  className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:bg-gray-50"
                >
                  <div className="flex items-center">
                    <div className="flex-1">
                      <div className="font-medium text-gray-800">
                        房间 {room.roomId}
                      </div>
                      <div className="text-sm text-gray-500">
                        {room.playerCount}/{room.maxPlayers} 人
                      </div>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-3">
                    <div className={`px-2 py-1 rounded-full text-xs font-medium ${
                      room.isEmpty
                        ? 'bg-gray-100 text-gray-600'
                        : room.playerCount >= room.maxPlayers
                        ? 'bg-red-100 text-red-600'
                        : 'bg-green-100 text-green-600'
                    }`}>
                      {room.isEmpty ? '空闲' : room.playerCount >= room.maxPlayers ? '已满' : '可加入'}
                    </div>
                    
                    <button
                      onClick={() => handleJoinRoom(room.roomId)}
                      disabled={loading || room.playerCount >= room.maxPlayers}
                      className="flex items-center px-3 py-1 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >
                      <Play size={14} className="mr-1" />
                      加入
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Instructions */}
        <div className="mt-8 bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-800">游戏说明</h2>
          <div className="text-gray-600 space-y-2">
            <p>• 掼蛋是一个4人对战的扑克游戏</p>
            <p>• 选择座位：东南西北四个位置，东西为一队，南北为一队</p>
            <p>• 创建房间后，其他玩家可以通过房间ID加入</p>
            <p>• 所有4位玩家进入后，游戏自动开始</p>
            <p>• 游戏目标：通过配合，率先出完手牌获胜</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Lobby;