// App.js
import React, { useState, useEffect, useContext } from 'react';
import { WebSocketContext } from './WebSocketProvider';

// Import your views
import NameEntryScreen from './components/NameEntryScreen';
import LobbyScreen from './components/LobbyScreen';
import TableScreen from './components/TableScreen';
import BackgroundButton from './components/BackgroundButton';

function App() {
  const { ws, isConnected } = useContext(WebSocketContext);
  
  // Define your state here
  const [phase, setPhase] = useState('NAME_ENTRY'); // 'NAME_ENTRY' | 'LOBBY' | 'PLAYING'
  const [gameState, setGameState] = useState(null);
  const [lobbyInfo, setLobbyInfo] = useState(null);

  useEffect(() => {
    if (!ws) return;

    // Listen for messages from the Go/Etcd backend
    ws.onmessage = (event) => {
      const payload = JSON.parse(event.data);

      // CHANGE OF STATE LOGIC
      switch (payload.type) {
        case 'LOBBY_JOINED':
          setPhase('LOBBY');
          setLobbyInfo(payload.data);
          break;
        case 'GAME_STARTED':
          setPhase('PLAYING');
          setGameState(payload.data);
          break;
        case 'STATE_UPDATE':
          // Update cards on the table, scores, whose turn it is
          setGameState(payload.data);
          break;
        case 'SESSION_RECOVERED':
          // Triggered after a K8s pod failure and successful reconnect
          setPhase(payload.data.currentPhase);
          setGameState(payload.data.gameState);
          break;
        default:
          console.warn('Unknown message type:', payload.type);
      }
    };
  }, [ws]);

  if (!isConnected) {
    return <div>Connecting to MarafoNet Cluster...</div>;
  }

// Helper function to keep the return statement clean
  const renderScreen = () => {
    switch (phase) {
      case 'NAME_ENTRY':
        return <NameEntryScreen ws={ws} />;
      case 'LOBBY':
        return <LobbyScreen ws={ws} lobbyState={lobbyInfo} />;
      case 'PLAYING':
        return <TableScreen ws={ws} gameState={gameState} />;
      default:
        return <div>Unknown Phase</div>;
    }
  };

  return (
    <div style={{ position: 'relative', minHeight: '100vh', padding: '20px' }}>
      {renderScreen()}
    </div>
  );
}

export default App;