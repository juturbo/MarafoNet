// App.js
import React, { useState, useEffect, useContext } from 'react';
import { WebSocketContext } from './WebSocketProvider';

// Import your views
import NameEntryScreen from './components/LogInScreen';
import LobbyScreen from './components/LobbyScreen';
import TableScreen from './components/TableScreen';
import RegisterScreen from './components/RegisterScreen';
import BackgroundButton from './components/BackgroundButton';

function App() {
  const { ws, isConnected } = useContext(WebSocketContext);
  
  // Define your state here
  // PHASES: 'LOG_IN' | 'REGISTER' | 'LOBBY' | 'PLAYING'
  const [phase, setPhase] = useState('LOG_IN');
  const [gameState, setGameState] = useState(null);
  const [lobbyInfo, setLobbyInfo] = useState(null);
  const [currentPlayerName, setCurrentPlayerName] = useState(null);

  useEffect(() => {
    if (!ws) {
      console.log('WebSocket is not connected yet');
      return;
    }

    console.log('App.js: Setting up WebSocket message handler');

    // Listen for game update messages
    const handleGameMessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        console.log('📨 App.js received:', payload);

        if(payload.type === "match_update") {
          console.log('🎮 Phase changing to PLAYING');
          setPhase('PLAYING');
          setGameState(payload.match);
        }
      } catch (err) {
        console.error('Error parsing message in App.js:', err);
      }
    };

    const handleError = (error) => {
      console.error('❌ WebSocket error:', error);
    };

    const handleOpen = () => {
      console.log('✅ WebSocket opened');
    };

    const handleClose = () => {
      console.log('❌ WebSocket closed');
    };

    ws.addEventListener('message', handleGameMessage);
    ws.addEventListener('error', handleError);
    ws.addEventListener('open', handleOpen);
    ws.addEventListener('close', handleClose);

    return () => {
      ws.removeEventListener('message', handleGameMessage);
      ws.removeEventListener('error', handleError);
      ws.removeEventListener('open', handleOpen);
      ws.removeEventListener('close', handleClose);
    };
  }, [ws]);

  if (!isConnected) {
    return <div>Connecting to MarafoNet Cluster...</div>;
  }

// Helper function to keep the return statement clean
  const renderScreen = () => {
    switch (phase) {
      case 'LOG_IN':
        return <NameEntryScreen ws={ws} onAuthSuccess={onAuthSuccess} onSwitchToRegister={onSwitchToRegister} />;
      case 'REGISTER':
        return <RegisterScreen ws={ws} onRegisterSuccess={onRegisterSuccess} />;
      case 'LOBBY':
        return <LobbyScreen ws={ws} lobbyState={lobbyInfo} />;
      case 'PLAYING':
        return <TableScreen matchUpdate={gameState} currentPlayerName={currentPlayerName} />;
      default:
        return <div>Unknown Phase</div>;
    }
  };

  const onAuthSuccess = (playerName) => {
    setCurrentPlayerName(playerName);
    setPhase('LOBBY');
  };

  const onRegisterSuccess = () => {
    setPhase('LOG_IN');
  };

  const onSwitchToRegister = () => {
    setPhase('REGISTER');
  };

  return (
    <div style={{ position: 'relative', minHeight: '100vh', padding: '20px' }}>
      {renderScreen()}
    </div>
  );
}

export default App;