import React from 'react';
import ListGroup from 'react-bootstrap/ListGroup';
import './LobbyScreen.css';

export default function LobbyScreen() {

    return (
        <div className="lobby-screen">
            <h1>Lobby:</h1>
            <ListGroup>
                <ListGroup.Item>Cras justo odio</ListGroup.Item>
                <ListGroup.Item>Dapibus ac facilisis in</ListGroup.Item>
                <ListGroup.Item>Morbi leo risus</ListGroup.Item>
                <ListGroup.Item>Porta ac consectetur ac</ListGroup.Item>
                <ListGroup.Item>Vestibulum at eros</ListGroup.Item>
            </ListGroup>
        </div>
    );
}