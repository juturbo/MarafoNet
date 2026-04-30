import React, { useContext } from 'react';
import { WebSocketContext } from '../WebSocketProvider';
import './TrumpSelector.css';

export default function TrumpSelector({ isFirstPlayer, gameID }) {
    const { ws } = useContext(WebSocketContext);

    React.useEffect(() => {
        console.log('TrumpSelector render - isFirstPlayer:', isFirstPlayer);
    }, [isFirstPlayer]);

    const suits = [
        { number: 1, name: 'Bastoni', label: '♣' },
        { number: 2, name: 'Coppe', label: '♥' },
        { number: 3, name: 'Denare', label: '♦' },
        { number: 4, name: 'Spade', label: '♠' }
    ];

    const handleSuitSelect = (suitNumber) => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            const message = {
                type: 'set_trump',
                payload: {
                    gameId: gameID,
                    suit: Number(suitNumber)
                }
            };
            console.log('Sending trump selection message:', message);
            ws.send(JSON.stringify(message));
            console.log('Trump selected:', suitNumber, 'GameID:', gameID);
        }
    };

    if (!isFirstPlayer) {
        return null;
    }

    return (
        <div className="trump-selector-overlay">
            <div className="trump-selector-modal">
                <div className="trump-selector-title">Choose Trump Suit</div>
                <div className="trump-selector-buttons">
                    {suits.map(suit => (
                        <button
                            key={suit.number}
                            className="trump-button"
                            onClick={() => handleSuitSelect(suit.number)}
                            title={suit.name}
                        >
                            <span className="suit-symbol">{suit.label}</span>
                            <span className="suit-name">{suit.name}</span>
                        </button>
                    ))}
                </div>
            </div>
        </div>
    );
}
