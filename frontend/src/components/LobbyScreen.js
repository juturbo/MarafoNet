import React from 'react';
import ListGroup from 'react-bootstrap/ListGroup';
import './LobbyScreen.css';

export default function LobbyScreen({ ws }) {

    const payload = {
        type: 'first_join'
    }

    ws.send(JSON.stringify(payload));
    console.log('Sent:', payload);

    ws.onmessage = (event) => {
        console.log(event.data)
    };

    return (
        <div className="lobby-screen">
            <h1>Lobby:</h1>
            <p>Wait while we find players.</p>
        </div>
    );
}