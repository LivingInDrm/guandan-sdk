import React, { useMemo } from 'react';
import { useDrop } from 'react-dnd';
import Card from './Card';
import { Card as CardType } from '../types';
import { usePlayerState, useGameActions } from '../store';

interface HandProps {
  cards: CardType[];
  className?: string;
}

const Hand: React.FC<HandProps> = ({ cards, className = '' }) => {
  const { selectedCards, isMyTurn, canPlay } = usePlayerState();
  const { selectCard, deselectCard, playCards } = useGameActions();

  const [{ isOver }, drop] = useDrop({
    accept: 'card',
    drop: (item: { card: CardType }) => {
      // Card dropped back to hand, deselect it
      deselectCard(item.card);
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
    }),
  });

  const sortedCards = useMemo(() => {
    return [...cards].sort((a, b) => {
      // Sort by suit first, then by rank
      const suitOrder = { 'Spades': 0, 'Hearts': 1, 'Clubs': 2, 'Diamonds': 3, 'Joker': 4 };
      const rankOrder = {
        '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9, '10': 10,
        'J': 11, 'Q': 12, 'K': 13, 'A': 14, '小王': 15, '大王': 16
      };
      
      const suitA = suitOrder[a.suit as keyof typeof suitOrder] || 0;
      const suitB = suitOrder[b.suit as keyof typeof suitOrder] || 0;
      
      if (suitA !== suitB) {
        return suitA - suitB;
      }
      
      const rankA = rankOrder[a.rank as keyof typeof rankOrder] || 0;
      const rankB = rankOrder[b.rank as keyof typeof rankOrder] || 0;
      
      return rankA - rankB;
    });
  }, [cards]);

  const isCardSelected = (card: CardType) => {
    return selectedCards.some(selected => 
      selected.suit === card.suit && selected.rank === card.rank
    );
  };

  const handleCardClick = (card: CardType) => {
    if (!canPlay) return;
    
    if (isCardSelected(card)) {
      deselectCard(card);
    } else {
      selectCard(card);
    }
  };

  const handleCardDoubleClick = (card: CardType) => {
    if (!canPlay) return;
    
    // Double click to play single card
    playCards([card]);
  };

  return (
    <div
      ref={drop}
      data-testid="hand"
      className={`
        relative min-h-20 p-4 rounded-lg border-2 border-dashed
        ${isOver ? 'border-blue-500 bg-blue-50' : 'border-gray-300'}
        ${className}
      `}
    >
      <div className="flex justify-center items-end space-x-1 flex-wrap">
        {sortedCards.length === 0 ? (
          <div className="text-gray-500 text-center py-8">
            <div className="text-4xl mb-2">🎴</div>
            <div>手牌已出完</div>
          </div>
        ) : (
          sortedCards.map((card, index) => (
            <div
              key={`${card.suit}-${card.rank}-${index}`}
              className="relative"
              style={{
                marginLeft: index > 0 ? '-0.5rem' : 0,
                zIndex: index,
              }}
            >
              <Card
                card={card}
                isSelected={isCardSelected(card)}
                isPlayable={canPlay}
                isDraggable={canPlay}
                size="medium"
                onClick={handleCardClick}
                onDoubleClick={handleCardDoubleClick}
              />
            </div>
          ))
        )}
      </div>
      
      {/* Selected cards info */}
      {selectedCards.length > 0 && (
        <div className="absolute top-2 right-2 bg-blue-500 text-white px-2 py-1 rounded text-sm">
          已选择 {selectedCards.length} 张
        </div>
      )}
      
      {/* Turn indicator */}
      {isMyTurn && (
        <div className="absolute top-2 left-2 bg-green-500 text-white px-2 py-1 rounded text-sm animate-pulse">
          轮到你了
        </div>
      )}
    </div>
  );
};

export default Hand;