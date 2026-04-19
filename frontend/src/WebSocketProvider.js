// WebSocketProvider.js
import React, { createContext, useEffect, useRef, useState } from 'react';

export const WebSocketContext = createContext(null);

export const WebSocketProvider = ({ children }) => {
  const ws = useRef(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState(null);

  const address = 'wss://localhost:8080/ws';
  useEffect(() => {
    // Connect to your K8s load balancer
    ws.current = new WebSocket(address);

    ws.current.onopen = () => {
      setIsConnected(true);
      setError(null);
      console.log('Connected to WebSocket server');
    };
    
    ws.current.onclose = () => {
      setIsConnected(false);
      // FAULT TOLERANCE: Implement your auto-reconnect logic here
      console.log('Connection lost. Reconnecting to a healthy Pod...');
    };

    ws.current.addEventListener('message', (event) => {
      try {
        const message = JSON.parse(event.data);
        console.log('WebSocket message received:', message);
        if (message.type === 'error' && message.message) {
          console.log('Setting error to:', message.message);
          setError(message.message);
        }
      } catch (e) {
        console.error('Error parsing WebSocket message:', e);
      }
    });

    //return () => {
    //  ws.current?.close();TODO
    //};
  }, []);

  const clearError = () => setError(null);

  return (
    <WebSocketContext.Provider value={{ ws: ws.current, isConnected, error, clearError }}>
      {children}
    </WebSocketContext.Provider>
  );
};