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
function PlayerPosition({ playerName, position, isFirstPlayer, isCurrentPlayer }) {
    return (
        <div className={`player-position player-${position}`}>
            <div className="player-name">
                {playerName}
                {isFirstPlayer && <span className="first-badge">Di mano</span>}
                {isCurrentPlayer && <span className="first-badge">Tocca a te</span>}
            </div>
            <div className="player-cards-back">
                <img src="/assets/cards/back.png" alt="Card back" className="back-card" />
                <img src="/assets/cards/back.png" alt="Card back" className="back-card" />
                <img src="/assets/cards/back.png" alt="Card back" className="back-card" />
            </div>
        </div>
    );
}

// Main TableScreen component
export default function TableScreen({ gameUpdate, currentPlayerName, onPlayAgain }) {
    const [gameState, setGameState] = useState(gameUpdate);
    const [sortEnabled, setSortEnabled] = useState(false);
    const [showLastTrick, setShowLastTrick] = useState(true);
    const { ws, error, clearError } = useContext(WebSocketContext);
    
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
    
    // Handle exit button - close WebSocket and exit to login screen
    const handleExit = () => {
        if (ws) {
            ws.close();
        }
        window.location.reload();
    };
    
    // Handle play again button - send first_join to go back to lobby
    const handlePlayAgain = () => {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            console.error('WebSocket not connected');
            return;
        }
        
        const message = {
            type: 'play_again'
        };
        
        ws.send(JSON.stringify(message));
        console.log('Sent play_again message to return to lobby');
        
        // Navigate back to lobby screen
        if (onPlayAgain) {
            onPlayAgain();
        }
    };
    
    useEffect(() => {
        if (gameUpdate) {
            setGameState(gameUpdate);
        }
    }, [gameUpdate]);
    
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
            {/* Row 1: Total Points, Top Player, Empty */}
            <div className="total-points-info">
                <div className="total-points-title">Total Points</div>
                <div className="total-points-score">
                    {gameState.TotalPoints && gameState.TotalPoints[currentPlayer.TeamId] !== undefined ? gameState.TotalPoints[currentPlayer.TeamId] : 0}
                    <span className="score-separator">-</span>
                    {gameState.TotalPoints && gameState.TotalPoints[1 - currentPlayer.TeamId] !== undefined ? gameState.TotalPoints[1 - currentPlayer.TeamId] : 0}
                </div>
            </div>
            
            {/* Top Player - Teammate */}
            <div className="position top">
                {top ? (
                    <PlayerPosition playerName={top.Name} position="top" isFirstPlayer={gameState.FirstPlayer === top.Name} isCurrentPlayer={gameState.CurrentPlayer === top.Name} />
                ) : (
                    <div className="player-position">No teammate</div>
                )}
            </div>
            
            {/* Last Trick - Toggle Button and Display */}
            <div className="position-empty last-trick-slot">
                <div className="last-trick-header">
                    <button 
                        className="last-trick-button"
                        onClick={() => setShowLastTrick(!showLastTrick)}
                        title="Toggle last trick details"
                        aria-expanded={showLastTrick}
                    >
                        {showLastTrick ? '▼ Nascondi presa' : '▶ Mostra presa'}
                    </button>
                </div>
                {showLastTrick && gameState.LastTrick && gameState.LastTrick.length > 0 && (
                    <div className="last-trick-container">
                        <div className="last-trick-title">Ultima Presa</div>
                        <div className="last-trick-cards">
                            {gameState.LastTrick.map((playedCard, index) => (
                                <div key={index} className="table-card-slot last-trick-card-slot">
                                    <div className="table-card-player">{playedCard.PlayerName}</div>
                                    <Card card={playedCard.Card} />
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </div>
            
            {/* Row 2: Left Player, Table Center, Right Player */}
            {/* Left Player - Opponent */}
            <div className="position left">
                {left ? (
                    <PlayerPosition playerName={left.Name} position="left" isFirstPlayer={gameState.FirstPlayer === left.Name} isCurrentPlayer={gameState.CurrentPlayer === left.Name} />
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
            </div>
            
            {/* Right Player - Opponent */}
            <div className="position right">
                {right ? (
                    <PlayerPosition playerName={right.Name} position="right" isFirstPlayer={gameState.FirstPlayer === right.Name} isCurrentPlayer={gameState.CurrentPlayer === right.Name} />
                ) : (
                    <div className="player-position">No opponent</div>
                )}
            </div>
            
            {/* Bottom Player - Current Player with visible cards */}
            <div className="position bottom">
                <div className="current-player">
                    <div className="player-name">
                        {bottom.Name} (Team {bottom.TeamId})
                        {gameState.FirstPlayer === bottom.Name && <span className="first-badge">Di mano</span>}
                        {gameState.CurrentPlayer === bottom.Name && <span className="first-badge">Tocca a te</span>}
                    </div>
                    <div className="sort-button-container">
                        <button 
                            className={`sort-button ${sortEnabled ? 'active' : ''}`}
                            onClick={() => setSortEnabled(!sortEnabled)}
                            title="Sort cards by suit and rank"
                        >
                            Ordina
                        </button>
                    </div>
                    <div className="player-hand">
                        {gameState.Hand && gameState.Hand.length > 0 ? (
                            (sortEnabled ? [...gameState.Hand].sort((a, b) => a.Suit - b.Suit || b.Rank - a.Rank) : gameState.Hand).map((card, index) => (
                                <Card key={index} card={card} onClick={handleCardClick} />
                            ))
                        ) : (
                            <div className="empty-hand">No cards</div>
                        )}
                    </div>
                </div>
            </div>
            
            {/* Game Result - if game is over */}
            {gameState.WinnerTeam !== null && gameState.WinnerTeam !== undefined && (
                <div className="game-result">
                    <div className="result-text">Team {gameState.WinnerTeam} Wins!</div>
                    <div className="result-players">
                        {gameState.WinnerPlayers && gameState.WinnerPlayers.join(', ')}
                    </div>
                    <div className="result-buttons">
                        <button className="result-button play-again-btn" onClick={handlePlayAgain}>
                            Play Again
                        </button>
                        <button className="result-button exit-btn" onClick={handleExit}>
                            Exit to Login
                        </button>
                    </div>
                </div>
            )}

            {/* Trump Selector - shows for First player */}
            <TrumpSelector 
                isFirstPlayer={gameState.FirstPlayer === bottom.Name && (!gameState.TrumpSuit || gameState.TrumpSuit === 'None')}
                gameID={gameState.gameID}
            />
        </div>
    );
}