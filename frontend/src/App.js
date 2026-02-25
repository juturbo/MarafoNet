// App.js
import React, { useState, useEffect, useContext } from 'react';
import { WebSocketContext } from './WebSocketProvider';

// Import your views
import NameEntryScreen from './components/NameEntryScreen';
import LobbyScreen from './components/LobbyScreen';
import TableScreen from './components/TableScreen';

function App() {
  const { ws, isConnected } = useContext(WebSocketContext);
  
  // Define your state here
  const [phase, setPhase] = useState('NAME_ENTRY'); // 'NAME_ENTRY' | 'LOBBY' | 'PLAYING'
  const [gameState, setGameState] = useState(null);
  const [theme, setTheme] = useState('table_green');

  useEffect(() => {
    const greenFeltBg = 'url("/images/green-felt.jpg")'; 
    const woodTableBg = 'url("/images/wood-table.jpg")';

    if (theme === 'table_wood') {
      document.body.style.backgroundImage = woodTableBg;
      document.body.style.color = '#ecf0f1'; // Light text for dark wood
    } else {
      document.body.style.backgroundImage = greenFeltBg;
      document.body.style.color = '#ffffff'; // White text for green felt
    }

    // Essential CSS to make the image look good on any screen size
    document.body.style.backgroundSize = 'cover';       // Stretches/shrinks to cover the whole screen
    document.body.style.backgroundPosition = 'center';  // Centers the image
    document.body.style.backgroundAttachment = 'fixed'; // Prevents the image from moving if you scroll
    document.body.style.backgroundRepeat = 'no-repeat'; // Stops the image from tiling
    
    // Optional: Add a smooth fade effect when switching
    document.body.style.transition = 'background-image 0.5s ease-in-out';
  }, [theme]);

  useEffect(() => {
    if (!ws) return;

    // Listen for messages from the Go/Etcd backend
    ws.onmessage = (event) => {
      const payload = JSON.parse(event.data);

      // CHANGE OF STATE LOGIC
      switch (payload.type) {
        case 'LOBBY_JOINED':
          setPhase('LOBBY');
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
        return <LobbyScreen ws={ws} />;
      case 'PLAYING':
        return <TableScreen ws={ws} gameState={gameState} />;
      default:
        return <div>Unknown Phase</div>;
    }
  };

  return (
    <div style={{ position: 'relative', minHeight: '100vh', padding: '20px' }}>
      
      {/* 3. The Persistent Button */}
      {/* Because this is outside the switch statement, it NEVER unmounts */}
      <button 
        onClick={() => setTheme(theme === 'table_green' ? 'table_wood' : 'table_green')}
        style={{ position: 'absolute', top: '10px', right: '10px', zIndex: 1000 }}
      >
        Toggle {theme === 'table_green' ? 'Wood Table' : 'Green Felt'} Theme
      </button>

      {/* 4. The dynamic screen content */}
      {renderScreen()}
      
    </div>
  );
}

export default App;