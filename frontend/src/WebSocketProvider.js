// WebSocketProvider.js
import React, { createContext, useEffect, useRef, useState } from 'react';

export const WebSocketContext = createContext(null);

export const WebSocketProvider = ({ children }) => {
  const ws = useRef(null);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    // Connect to your K8s load balancer
    ws.current = new WebSocket('wss://localhost:5000/ws');

    ws.current.onopen = () => setIsConnected(true);
    
    ws.current.onclose = () => {
      setIsConnected(false);
      // FAULT TOLERANCE: Implement your auto-reconnect logic here
      console.log('Connection lost. Reconnecting to a healthy Pod...');
    };

    return () => {
      ws.current?.close();
    };
  }, []);

  return (
    <WebSocketContext.Provider value={{ ws: ws.current, isConnected }}>
      {children}
    </WebSocketContext.Provider>
  );
};