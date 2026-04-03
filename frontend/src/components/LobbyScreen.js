import React from 'react';
import ListGroup from 'react-bootstrap/ListGroup';
import './LobbyScreen.css';

export default function LobbyScreen({ ws }) {

    const payload = {
        type: 'fist_join'
    }

    ws.send(JSON.stringify(payload));
    console.log('Sent:', payload);

    return (
        <div className="lobby-screen">
            <h1>Lobby:</h1>
            <p>Wait while we find players.</p>
        </div>
    );
}