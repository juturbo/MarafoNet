import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import { WebSocketProvider } from './WebSocketProvider';
import { MockWebSocketProvider } from './MockWebSocketProvider';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <MockWebSocketProvider>
      <App />
    </MockWebSocketProvider>
  </React.StrictMode>
);