import React from 'react';
import { useDrop } from 'react-dnd';
import Card from './Card';
import { Card as CardType, CardGroup } from '../types';
import { useGameState, useGameActions } from '../store';

interface TableProps {
  className?: string;
}

const Table: React.FC<TableProps> = ({ className = '' }) => {
  const { currentDeal } = useGameState();
  const { playCards } = useGameActions();

  const [{ isOver, canDrop }, drop] = useDrop({
    accept: 'card',
    drop: (item: { card: CardType }) => {
      // Card dropped on table, play it
      playCards([item.card]);
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  });

  const tablePlay = currentDeal?.tablePlay;
  const lastPlayer = currentDeal?.lastPlayer;

  const getPlayerDisplayName = (seat: string) => {
    const seatNames = {
      'east': '东',
      'south': '南',
      'west': '西',
      'north': '北'
    };
    return seatNames[seat as keyof typeof seatNames] || seat;
  };

  const getCardGroupDescription = (cardGroup: CardGroup) => {
    const cardCount = cardGroup.cards.length;
    
    if (cardCount === 1) {
      return '单张';
    } else if (cardCount === 2) {
      return '对子';
    } else if (cardCount === 3) {
      return '三条';
    } else if (cardCount >= 5) {
      return '顺子';
    } else {
      return `${cardCount}张`;
    }
  };

  return (
    <div
      ref={drop}
      className={`
        relative bg-green-100 rounded-lg border-2 border-dashed
        min-h-32 p-6 flex flex-col items-center justify-center
        ${isOver ? 'border-blue-500 bg-blue-50' : 'border-green-300'}
        ${canDrop ? 'shadow-lg' : ''}
        ${className}
      `}
    >
      {/* Drop hint */}
      {isOver && (
        <div className="absolute inset-0 bg-blue-200 bg-opacity-50 rounded-lg flex items-center justify-center">
          <div className="text-blue-800 font-bold text-lg">
            拖拽到此处出牌
          </div>
        </div>
      )}

      {/* Table content */}
      {tablePlay && tablePlay.cards.length > 0 ? (
        <div className="text-center">
          {/* Player info */}
          {lastPlayer && (
            <div className="mb-4 text-sm text-gray-600">
              <span className="font-medium">
                {getPlayerDisplayName(lastPlayer)} 出牌
              </span>
              <span className="ml-2 px-2 py-1 bg-gray-200 rounded">
                {getCardGroupDescription(tablePlay)}
              </span>
            </div>
          )}

          {/* Cards */}
          <div className="flex justify-center items-center space-x-1 flex-wrap">
            {tablePlay.cards.map((card, index) => (
              <div
                key={`${card.suit}-${card.rank}-${index}`}
                className="relative"
                style={{
                  marginLeft: index > 0 ? '-0.3rem' : 0,
                  zIndex: index,
                }}
              >
                <Card
                  card={card}
                  isPlayable={false}
                  isDraggable={false}
                  size="medium"
                />
              </div>
            ))}
          </div>

          {/* Card group info */}
          <div className="mt-4 text-sm text-gray-600">
            <div>牌型：{getCardGroupDescription(tablePlay)}</div>
            {tablePlay.value > 0 && (
              <div>点数：{tablePlay.value}</div>
            )}
          </div>
        </div>
      ) : (
        <div className="text-center text-gray-500">
          <div className="text-4xl mb-2">🎴</div>
          <div className="text-lg font-medium">出牌区</div>
          <div className="text-sm">等待玩家出牌</div>
        </div>
      )}

      {/* Game instructions */}
      <div className="absolute top-2 left-2 text-xs text-gray-500 max-w-xs">
        <div>• 点击选择手牌</div>
        <div>• 双击快速出单张</div>
        <div>• 拖拽到此处出牌</div>
      </div>
    </div>
  );
};

export default Table;