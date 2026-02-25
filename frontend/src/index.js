import React from 'react';
import ReactDOM from 'react-dom/client';
import 'bootstrap/dist/css/bootstrap.min.css';
import './index.css';
import App from './App';
import { WebSocketProvider } from './WebSocketProvider';
import BackgroundButton from './components/BackgroundButton';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <BackgroundButton />
    <WebSocketProvider>
      <App />
    </WebSocketProvider>
  </React.StrictMode>
);