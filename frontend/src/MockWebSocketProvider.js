// MockWebSocketProvider.js
import React, { createContext, useState, useEffect } from 'react';
import { WebSocketContext } from './WebSocketProvider';

export const MockWebSocketProvider = ({ children }) => {
  const [isConnected, setIsConnected] = useState(false);

  // Create a fake WebSocket object that just logs to the console
  const fakeWs = {
    send: (message) => console.log('Mock WebSocket sent:', message),
    close: () => console.log('Mock WebSocket closed'),
  };

  useEffect(() => {
    // Simulate the network connection delay
    setTimeout(() => {
      setIsConnected(true);
      
      // Optional: Simulate receiving a message from the Go backend after 2 seconds
      // to test your App.js state transitions automatically!
      setTimeout(() => {
        if (fakeWs.onmessage) {
          fakeWs.onmessage({
            data: JSON.stringify({
              type: 'GAME_STARTED',
              data: { /* Dummy Marafone table state here */ }
            })
          });
        }
      }, 2000);

    }, 500);
  }, []);
  // Notice we use the REAL WebSocketContext.Provider, but pass it FAKE values
  return (
    <WebSocketContext.Provider value={{ ws: fakeWs, isConnected }}>
      {children}
    </WebSocketContext.Provider>
  );
};
