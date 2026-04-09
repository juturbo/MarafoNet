import React, { useEffect } from 'react';
import ListGroup from 'react-bootstrap/ListGroup';
import './LobbyScreen.css';

export default function LobbyScreen({ ws }) {
    useEffect(() => {
        if (!ws) {
            console.log('⚠️ LobbyScreen: WebSocket not connected');
            return;
        }
        
        console.log('📤 LobbyScreen: Sending first_join...');
        const payload = {
            type: 'first_join'
        }

        try {
            ws.send(JSON.stringify(payload));
            console.log('✅ LobbyScreen: Sent payload:', payload);
            console.log('🔗 WebSocket state:', ws.readyState, '(0=CONNECTING, 1=OPEN, 2=CLOSING, 3=CLOSED)');
        } catch (error) {
            console.error('❌ LobbyScreen: Error sending message:', error);
        }
    }, [ws]);

    return (
        <div className="lobby-screen">
            <h1>Lobby:</h1>
            <p>Wait while we find players.</p>
        </div>
    );
}