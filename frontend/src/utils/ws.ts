import { WSMessage, SnapshotMessage, EventMessage, CONNECTION_STATUS } from '../types';

export interface WSClientOptions {
  url: string;
  reconnectAttempts?: number;
  reconnectDelay?: number;
  heartbeatInterval?: number;
  onOpen?: () => void;
  onMessage?: (message: WSMessage) => void;
  onClose?: (event: CloseEvent) => void;
  onError?: (error: Event) => void;
  onReconnect?: (attempt: number) => void;
}

export class WSClient {
  private ws: WebSocket | null = null;
  private url: string;
  private options: Required<WSClientOptions>;
  private reconnectAttempts = 0;
  private maxReconnectAttempts: number;
  private reconnectDelay: number;
  private heartbeatInterval: number;
  private heartbeatTimer: NodeJS.Timeout | null = null;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private isReconnecting = false;
  private isClosed = false;

  constructor(options: WSClientOptions) {
    this.url = options.url;
    this.maxReconnectAttempts = options.reconnectAttempts ?? 5;
    this.reconnectDelay = options.reconnectDelay ?? 3000;
    this.heartbeatInterval = options.heartbeatInterval ?? 30000;
    
    this.options = {
      url: options.url,
      reconnectAttempts: this.maxReconnectAttempts,
      reconnectDelay: this.reconnectDelay,
      heartbeatInterval: this.heartbeatInterval,
      onOpen: options.onOpen ?? (() => {}),
      onMessage: options.onMessage ?? (() => {}),
      onClose: options.onClose ?? (() => {}),
      onError: options.onError ?? (() => {}),
      onReconnect: options.onReconnect ?? (() => {}),
    };
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      this.isClosed = false;
      
      try {
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
          this.reconnectAttempts = 0;
          this.isReconnecting = false;
          this.startHeartbeat();
          this.options.onOpen();
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event);
        };

        this.ws.onclose = (event) => {
          this.stopHeartbeat();
          this.options.onClose(event);
          
          if (!this.isClosed && !this.isReconnecting) {
            this.attemptReconnect();
          }
        };

        this.ws.onerror = (error) => {
          this.options.onError(error);
          reject(error);
        };

      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect(): void {
    this.isClosed = true;
    this.isReconnecting = false;
    this.stopHeartbeat();
    this.clearReconnectTimer();
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  send(message: WSMessage): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        reject(new Error('WebSocket is not connected'));
        return;
      }

      try {
        this.ws.send(JSON.stringify(message));
        resolve();
      } catch (error) {
        reject(error);
      }
    });
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  getConnectionStatus(): keyof typeof CONNECTION_STATUS {
    if (!this.ws) return 'DISCONNECTED';
    
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return 'CONNECTING';
      case WebSocket.OPEN:
        return 'CONNECTED';
      case WebSocket.CLOSING:
      case WebSocket.CLOSED:
        return 'DISCONNECTED';
      default:
        return 'ERROR';
    }
  }

  private handleMessage(event: MessageEvent): void {
    try {
      console.log('=== RAW WebSocket Message ===');
      console.log('Raw data:', event.data);
      const message: WSMessage = JSON.parse(event.data);
      console.log('Parsed message:', JSON.stringify(message, null, 2));
      console.log('=== END RAW Message ===');
      
      // Handle ping/pong
      if (message.t === 'ping') {
        this.send({ t: 'pong' }).catch(console.error);
        return;
      }
      
      if (message.t === 'pong') {
        return;
      }
      
      this.options.onMessage(message);
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    
    this.heartbeatTimer = setInterval(() => {
      if (this.isConnected()) {
        this.send({ t: 'ping' }).catch(() => {
          // Heartbeat failed, connection might be broken
          this.disconnect();
        });
      }
    }, this.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private attemptReconnect(): void {
    if (this.isReconnecting || this.isClosed || 
        this.reconnectAttempts >= this.maxReconnectAttempts) {
      return;
    }

    this.isReconnecting = true;
    this.reconnectAttempts++;
    
    this.options.onReconnect(this.reconnectAttempts);
    
    this.reconnectTimer = setTimeout(() => {
      this.connect()
        .then(() => {
          this.isReconnecting = false;
        })
        .catch(() => {
          this.isReconnecting = false;
          this.attemptReconnect();
        });
    }, this.reconnectDelay);
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }
}

// Factory functions for specific message types
export const createPlayCardsMessage = (cards: string[]): WSMessage => ({
  t: 'PlayCards',
  data: { cards }
});

export const createPassMessage = (): WSMessage => ({
  t: 'Pass'
});

// Message type guards
export const isSnapshotMessage = (message: WSMessage): message is SnapshotMessage => {
  return message.t === 'Snapshot';
};

export const isEventMessage = (message: WSMessage): message is EventMessage => {
  return message.t === 'Event';
};

export const isErrorMessage = (message: WSMessage): message is WSMessage & { error: string } => {
  return message.t === 'Error' && 'error' in message;
};

// Utility functions
export const parseCardString = (cardStr: string): { suit: string; rank: string } => {
  if (cardStr === '小王') return { suit: 'Joker', rank: '小王' };
  if (cardStr === '大王') return { suit: 'Joker', rank: '大王' };
  
  const suitMap: Record<string, string> = {
    '♥': 'Hearts',
    '♦': 'Diamonds',
    '♣': 'Clubs',
    '♠': 'Spades'
  };
  
  const suit = suitMap[cardStr[0]] || 'Unknown';
  const rank = cardStr.slice(1);
  
  return { suit, rank };
};

export const formatCardString = (card: { suit: string; rank: string }): string => {
  if (card.suit === 'Joker') return card.rank;
  
  const suitMap: Record<string, string> = {
    'Hearts': '♥',
    'Diamonds': '♦',
    'Clubs': '♣',
    'Spades': '♠'
  };
  
  return suitMap[card.suit] + card.rank;
};

export default WSClient;