import React from 'react';
import { Play, SkipForward, RotateCcw, Settings, Info } from 'lucide-react';
import { usePlayerState, useGameActions, useErrorState } from '../store';

interface ControlBarProps {
  className?: string;
}

const ControlBar: React.FC<ControlBarProps> = ({ className = '' }) => {
  const { selectedCards, isMyTurn, canPlay } = usePlayerState();
  const { playCards, pass, deselectCard } = useGameActions();
  const { showValidationError, errorMessage, clearError } = useErrorState();

  const handlePlayCards = () => {
    if (selectedCards.length === 0) {
      return;
    }
    
    playCards(selectedCards);
  };

  const handlePass = () => {
    pass();
  };

  const handleClearSelection = () => {
    selectedCards.forEach(card => deselectCard(card));
  };

  const isPlayDisabled = !canPlay || !isMyTurn || selectedCards.length === 0;
  const isPassDisabled = !canPlay || !isMyTurn;

  return (
    <div className={`bg-white border-t border-gray-200 p-4 ${className}`}>
      {/* Error message */}
      {showValidationError && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="text-red-600">⚠️</div>
            <span className="text-red-700">{errorMessage}</span>
          </div>
          <button
            onClick={clearError}
            className="text-red-600 hover:text-red-800"
          >
            ✕
          </button>
        </div>
      )}

      {/* Main controls */}
      <div className="flex items-center justify-between">
        {/* Left side - Card info */}
        <div className="flex items-center space-x-4">
          <div className="text-sm text-gray-600">
            {selectedCards.length > 0 ? (
              <span>已选择 {selectedCards.length} 张牌</span>
            ) : (
              <span>请选择要出的牌</span>
            )}
          </div>
          
          {selectedCards.length > 0 && (
            <button
              onClick={handleClearSelection}
              className="flex items-center space-x-1 px-3 py-1 text-gray-600 hover:text-gray-800 border border-gray-300 rounded"
            >
              <RotateCcw size={16} />
              <span>清空选择</span>
            </button>
          )}
        </div>

        {/* Center - Main action buttons */}
        <div className="flex items-center space-x-3">
          <button
            onClick={handlePlayCards}
            disabled={isPlayDisabled}
            className={`
              flex items-center space-x-2 px-6 py-2 rounded-lg font-medium transition-all
              ${isPlayDisabled
                ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                : 'bg-blue-500 text-white hover:bg-blue-600 shadow-md hover:shadow-lg'
              }
            `}
          >
            <Play size={18} />
            <span>出牌</span>
          </button>

          <button
            onClick={handlePass}
            disabled={isPassDisabled}
            className={`
              flex items-center space-x-2 px-6 py-2 rounded-lg font-medium transition-all
              ${isPassDisabled
                ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                : 'bg-orange-500 text-white hover:bg-orange-600 shadow-md hover:shadow-lg'
              }
            `}
          >
            <SkipForward size={18} />
            <span>过牌</span>
          </button>
        </div>

        {/* Right side - Settings and info */}
        <div className="flex items-center space-x-2">
          <button className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded">
            <Settings size={18} />
          </button>
          
          <button className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded">
            <Info size={18} />
          </button>
        </div>
      </div>

      {/* Status indicator */}
      <div className="mt-3 text-center">
        {!isMyTurn && (
          <div className="text-sm text-gray-500">
            等待其他玩家出牌...
          </div>
        )}
        
        {isMyTurn && (
          <div className="text-sm text-green-600 font-medium">
            轮到你出牌了！
          </div>
        )}
      </div>

      {/* Quick actions */}
      <div className="mt-3 flex justify-center space-x-4">
        <button
          className="text-xs text-gray-500 hover:text-gray-700 underline"
          onClick={() => {
            // TODO: Show game rules
          }}
        >
          游戏规则
        </button>
        
        <button
          className="text-xs text-gray-500 hover:text-gray-700 underline"
          onClick={() => {
            // TODO: Show shortcuts
          }}
        >
          快捷键
        </button>
        
        <button
          className="text-xs text-gray-500 hover:text-gray-700 underline"
          onClick={() => {
            // TODO: Request hint
          }}
        >
          提示
        </button>
      </div>
    </div>
  );
};

export default ControlBar;