import React from 'react';
import { User, Crown, Wifi, WifiOff } from 'lucide-react';
import { Player, SeatID } from '../types';

interface PlayerInfoProps {
  player: Player;
  isCurrentPlayer?: boolean;
  isMySeat?: boolean;
  position?: 'top' | 'bottom' | 'left' | 'right';
  className?: string;
}

const PlayerInfo: React.FC<PlayerInfoProps> = ({
  player,
  isCurrentPlayer = false,
  isMySeat = false,
  position = 'bottom',
  className = '',
}) => {
  const getSeatDisplayName = (seat: SeatID) => {
    const seatNames = {
      'east': '东',
      'south': '南',
      'west': '西',
      'north': '北'
    };
    return seatNames[seat] || seat;
  };

  const getSeatColor = (seat: SeatID) => {
    const colors = {
      'east': 'bg-red-100 text-red-800',
      'south': 'bg-blue-100 text-blue-800',
      'west': 'bg-green-100 text-green-800',
      'north': 'bg-yellow-100 text-yellow-800'
    };
    return colors[seat] || 'bg-gray-100 text-gray-800';
  };

  const getPositionClasses = (position: string) => {
    switch (position) {
      case 'top':
        return 'flex-col items-center text-center';
      case 'bottom':
        return 'flex-col items-center text-center';
      case 'left':
        return 'flex-row items-center text-left';
      case 'right':
        return 'flex-row-reverse items-center text-right';
      default:
        return 'flex-col items-center text-center';
    }
  };

  const getCardCountPosition = (position: string) => {
    switch (position) {
      case 'top':
        return 'mt-2';
      case 'bottom':
        return 'mb-2';
      case 'left':
        return 'ml-2';
      case 'right':
        return 'mr-2';
      default:
        return 'mt-2';
    }
  };

  return (
    <div
      className={`
        ${getPositionClasses(position)}
        p-3 rounded-lg border-2 transition-all
        ${isCurrentPlayer ? 'border-blue-500 bg-blue-50 shadow-lg' : 'border-gray-200 bg-white'}
        ${isMySeat ? 'ring-2 ring-purple-500' : ''}
        ${className}
      `}
    >
      {/* Player Avatar and Info */}
      <div className="flex items-center space-x-2">
        {/* Avatar */}
        <div className={`
          relative w-12 h-12 rounded-full flex items-center justify-center
          ${getSeatColor(player.seat)}
          ${isCurrentPlayer ? 'ring-2 ring-blue-500' : ''}
        `}>
          <User size={20} />
          
          {/* Connection status */}
          <div className="absolute -top-1 -right-1">
            {player.connected ? (
              <Wifi size={12} className="text-green-500" />
            ) : (
              <WifiOff size={12} className="text-red-500" />
            )}
          </div>
        </div>

        {/* Player details */}
        <div className="flex-1">
          <div className="flex items-center space-x-1">
            <span className="font-medium text-gray-800">
              {player.name}
            </span>
            {isMySeat && (
              <Crown size={14} className="text-purple-500" />
            )}
          </div>
          
          <div className="text-xs text-gray-500">
            {getSeatDisplayName(player.seat)}
          </div>
        </div>
      </div>

      {/* Hand count */}
      <div className={`${getCardCountPosition(position)} flex items-center space-x-2`}>
        <div className={`
          px-2 py-1 rounded text-sm font-medium
          ${player.handCount === 0 
            ? 'bg-green-100 text-green-800' 
            : 'bg-gray-100 text-gray-800'
          }
        `}>
          {player.handCount === 0 ? '已完成' : `${player.handCount} 张`}
        </div>

        {/* Level indicator */}
        <div className="px-2 py-1 bg-yellow-100 text-yellow-800 rounded text-sm">
          {player.level}级
        </div>
      </div>

      {/* Status indicators */}
      <div className="flex items-center space-x-1 mt-1">
        {isCurrentPlayer && (
          <div className="px-2 py-1 bg-blue-500 text-white rounded text-xs animate-pulse">
            轮到此人
          </div>
        )}
        
        {!player.connected && (
          <div className="px-2 py-1 bg-red-500 text-white rounded text-xs">
            离线
          </div>
        )}
        
        {player.handCount === 0 && (
          <div className="px-2 py-1 bg-green-500 text-white rounded text-xs">
            获胜
          </div>
        )}
      </div>
    </div>
  );
};

export default PlayerInfo;