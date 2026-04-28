import React, { useEffect } from 'react';
import './LobbyScreen.css';

export default function LobbyScreen({ ws }) {
    useEffect(() => {
        if (!ws) {
            console.log('LobbyScreen: WebSocket not connected');
            return;
        }
        const payload = {
            type: 'first_join'
        }

        try {
            ws.send(JSON.stringify(payload));
        } catch (error) {
            console.error('LobbyScreen: Error sending message:', error);
        }
    }, [ws]);

    return (
        <div className="lobby-screen">
            <h1>Lobby:</h1>
            <p>Wait while we find players.</p>
        </div>
    );
}