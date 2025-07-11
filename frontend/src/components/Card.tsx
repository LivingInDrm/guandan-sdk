import React from 'react';
import { useDrag } from 'react-dnd';
import { Card as CardType } from '../types';

interface CardProps {
  card: CardType;
  isSelected?: boolean;
  isPlayable?: boolean;
  isDraggable?: boolean;
  size?: 'small' | 'medium' | 'large';
  onClick?: (card: CardType) => void;
  onDoubleClick?: (card: CardType) => void;
}

const Card: React.FC<CardProps> = ({
  card,
  isSelected = false,
  isPlayable = true,
  isDraggable = true,
  size = 'medium',
  onClick,
  onDoubleClick,
}) => {
  const [{ isDragging }, drag] = useDrag({
    type: 'card',
    item: { card },
    canDrag: isDraggable && isPlayable,
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  });

  const getSuitColor = (suit: string) => {
    switch (suit) {
      case 'Hearts':
      case 'Diamonds':
        return 'text-red-500';
      case 'Clubs':
      case 'Spades':
        return 'text-black';
      case 'Joker':
        return 'text-purple-600';
      default:
        return 'text-gray-800';
    }
  };

  const getSuitSymbol = (suit: string) => {
    switch (suit) {
      case 'Hearts':
        return '♥';
      case 'Diamonds':
        return '♦';
      case 'Clubs':
        return '♣';
      case 'Spades':
        return '♠';
      case 'Joker':
        return '🃏';
      default:
        return '?';
    }
  };

  const getSizeClasses = (size: string) => {
    switch (size) {
      case 'small':
        return 'w-8 h-12 text-xs';
      case 'medium':
        return 'w-12 h-16 text-sm';
      case 'large':
        return 'w-16 h-22 text-base';
      default:
        return 'w-12 h-16 text-sm';
    }
  };

  const handleClick = () => {
    if (onClick && isPlayable) {
      onClick(card);
    }
  };

  const handleDoubleClick = () => {
    if (onDoubleClick && isPlayable) {
      onDoubleClick(card);
    }
  };

  return (
    <div
      ref={isDraggable ? drag : undefined}
      data-testid="card"
      className={`
        ${getSizeClasses(size)}
        bg-white border-2 rounded-lg shadow-md cursor-pointer
        flex flex-col items-center justify-between p-1
        transition-all duration-200
        ${isSelected ? 'border-blue-500 bg-blue-50 transform -translate-y-2' : 'border-gray-300'}
        ${isDragging ? 'opacity-50' : ''}
        ${isPlayable ? 'hover:shadow-lg hover:scale-105' : 'opacity-50 cursor-not-allowed'}
        ${!isPlayable ? 'grayscale' : ''}
      `}
      onClick={handleClick}
      onDoubleClick={handleDoubleClick}
    >
      {/* Top left corner */}
      <div className={`self-start ${getSuitColor(card.suit)} font-bold`}>
        <div className="text-center">
          <div className="leading-none">{card.rank}</div>
          <div className="leading-none">{getSuitSymbol(card.suit)}</div>
        </div>
      </div>

      {/* Center symbol */}
      <div className={`${getSuitColor(card.suit)} text-lg font-bold`}>
        {card.suit === 'Joker' ? (
          <div className="text-center">
            <div className="text-2xl">🃏</div>
            <div className="text-xs">{card.rank}</div>
          </div>
        ) : (
          getSuitSymbol(card.suit)
        )}
      </div>

      {/* Bottom right corner (rotated) */}
      <div className={`self-end ${getSuitColor(card.suit)} font-bold transform rotate-180`}>
        <div className="text-center">
          <div className="leading-none">{card.rank}</div>
          <div className="leading-none">{getSuitSymbol(card.suit)}</div>
        </div>
      </div>
    </div>
  );
};

export default Card;