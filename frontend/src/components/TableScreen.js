import React, { useState, useEffect, useContext } from 'react';

import './TableScreen.css';
import TrumpSelector from './TrumpSelector';
import DangerAlert from './DangerAlert';
import { WebSocketContext } from '../WebSocketProvider';

// Card component to display individual cards
function Card({ card, onClick }) {
    if (!card) return null;
    
    // Map numeric suit values to folder names
    // Clubs/Bastoni = 1, Cups/Coppe = 2, Coins/Denare = 3, Swords/Spade = 4
    const suitFolders = {
        1: 'bastoni',
        2: 'coppe',
        3: 'denare',
        4: 'spade',
    };
    
    // Map numeric rank values to card numbers (1-10)
    // Ace = 1, Two = 2, ..., King = 10
    const rankValues = {
        1: '1',   // Ace
        2: '2',   // Two
        3: '3',   // Three
        4: '4',   // Four
        5: '5',   // Five
        6: '6',   // Six
        7: '7',   // Seven
        8: '8',   // Jack
        9: '9',   // Knight
        10: '10', // King
    };
    
    const suitFolder = suitFolders[card.Suit] || 'bastoni';
    const rankValue = rankValues[card.Rank] || '1';
    const imagePath = `/assets/cards/${suitFolder}/${rankValue}.png`;
    
    return (
        <div className="card" onClick={() => onClick && onClick(card.Rank, card.Suit)}>
            <img 
                src={imagePath} 
                alt={`Card ${card.Rank} of ${card.Suit}`}
                className="card-image"
                onError={(e) => {
                    console.error(`Failed to load card image: ${imagePath}`);
                    e.target.style.display = 'none';
                }}
            />
        </div>
    );
}

// Player position indicator (for other players)
function PlayerPosition({ playerName, position }) {
    return (
        <div className={`player-position player-${position}`}>
            <div className="player-name">{playerName}</div>
            <div className="player-cards-back">
                <img src="/assets/cards/back.png" alt="Card back" className="back-card" />
                <img src="/assets/cards/back.png" alt="Card back" className="back-card" />
                <img src="/assets/cards/back.png" alt="Card back" className="back-card" />
            </div>
        </div>
    );
}

// Main TableScreen component
export default function TableScreen({ matchUpdate, currentPlayerName }) {
    const [gameState, setGameState] = useState(matchUpdate);
    const [trumpSelected, setTrumpSelected] = useState(false);
    const { ws, error, clearError } = useContext(WebSocketContext);
    console.log('TableScreen received error from context:', error);
    
    // Map suit numbers to suit names
    const getSuitName = (suitNumber) => {
        const suitNames = {
            1: 'Bastoni',
            2: 'Coppe',
            3: 'Denare',
            4: 'Spade',
        };
        return suitNames[suitNumber] || 'None';
    };
    
    // Handle card click - send play_card message
    const handleCardClick = (rank, suit) => {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            console.error('WebSocket not connected');
            return;
        }
        
        const message = {
            type: 'play_card',
            payload: {
                rank: rank,
                suit: suit
            }
        };
        
        ws.send(JSON.stringify(message));
        console.log('Sent play_card message:', message);
    };
    
    useEffect(() => {
        if (matchUpdate) {
            setGameState(matchUpdate);
        }
    }, [matchUpdate]);
    
    if (!gameState || !gameState.Players || gameState.Players.length === 0) {
        return <div className="table-screen">Waiting for game data...</div>;
    }
    
    // Find current player index
    const currentPlayerIndex = gameState.Players.findIndex(p => p.Name === currentPlayerName);
    if (currentPlayerIndex === -1) {
        return <div className="table-screen">Player not found in game</div>;
    }
    
    const currentPlayer = gameState.Players[currentPlayerIndex];
    const currentTeamId = currentPlayer.TeamId;
    
    // Get player positions around the table based on teams
    const bottom = currentPlayer;
    
    // Find teammate (same TeamId, different player)
    const top = gameState.Players.find(p => p.TeamId === currentTeamId && p.Name !== currentPlayerName);
    
    // Find opponents (different TeamId)
    const opponents = gameState.Players.filter(p => p.TeamId !== currentTeamId);
    const left = opponents[0] || null;
    const right = opponents[1] || null;
    
    // Get table cards
    const tableCards = gameState.Table || [];
    
    return (
        <div className="table-screen">
            <DangerAlert message={error} onClose={clearError} />
            {/* Row 1: Empty, Top Player, Empty */}
            <div className="position-empty"></div>
            
            {/* Top Player - Teammate */}
            <div className="position top">
                {top ? (
                    <PlayerPosition playerName={top.Name} position="top" />
                ) : (
                    <div className="player-position">No teammate</div>
                )}
            </div>
            
            <div className="position-empty"></div>
            
            {/* Row 2: Left Player, Table Center, Right Player */}
            {/* Left Player - Opponent */}
            <div className="position left">
                {left ? (
                    <PlayerPosition playerName={left.Name} position="left" />
                ) : (
                    <div className="player-position">No opponent</div>
                )}
            </div>
            
            {/* Center - Table Cards */}
            <div className="table-center">
                <div className="trump-info">
                    Briscola: <span className="trump-suit">{getSuitName(gameState.TrumpSuit)}</span>
                </div>
                <div className="table-cards">
                    {tableCards.length > 0 ? (
                        tableCards.map((playedCard, index) => (
                            <div key={index} className="table-card-slot">
                                <div className="table-card-player">{playedCard.PlayerName}</div>
                                <Card card={playedCard.Card} />
                            </div>
                        ))
                    ) : (
                        <div className="empty-table">Table is empty</div>
                    )}
                </div>
                <div className="match-info">
                    <div>Team 0: {gameState.MatchPoints[0]} pts</div>
                    <div>Team 1: {gameState.MatchPoints[1]} pts</div>
                </div>
            </div>
            
            {/* Right Player - Opponent */}
            <div className="position right">
                {right ? (
                    <PlayerPosition playerName={right.Name} position="right" />
                ) : (
                    <div className="player-position">No opponent</div>
                )}
            </div>
            
            {/* Bottom Player - Current Player with visible cards */}
            <div className="position bottom">
                <div className="current-player">
                    <div className="player-name">
                        {bottom.Name} (Team {bottom.TeamId})
                        {gameState.FirstPlayer === bottom.Name && <span className="first-badge">FIRST</span>}
                    </div>
                    <div className="player-hand">
                        {bottom.Hand && bottom.Hand.length > 0 ? (
                            bottom.Hand.map((card, index) => (
                                <Card key={index} card={card} onClick={handleCardClick} />
                            ))
                        ) : (
                            <div className="empty-hand">No cards</div>
                        )}
                    </div>
                </div>
            </div>
            
            {/* Match Result - if game is over */}
            {gameState.WinnerTeam !== null && gameState.WinnerTeam !== undefined && (
                <div className="match-result">
                    <div className="result-text">Team {gameState.WinnerTeam} Wins!</div>
                    <div className="result-players">
                        {gameState.WinnerPlayers && gameState.WinnerPlayers.join(', ')}
                    </div>
                </div>
            )}

            {/* Trump Selector - shows for First player */}
            <TrumpSelector 
                isFirstPlayer={gameState.FirstPlayer === bottom.Name && (!gameState.TrumpSuit || gameState.TrumpSuit === 'None')}
                matchID={gameState.MatchID}
                onTrumpSelected={() => setTrumpSelected(true)}
            />
        </div>
    );
}