import React, { useState, useEffect } from 'react';

import './NameEntryScreen.css';

export default function NameEntryScreen({ ws, onAuthSuccess, onSwitchToRegister }) {
    const [name, setName] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');

    useEffect(() => {
        if (!ws) return;

        const handleMessage = (event) => {
            try {
                const response = JSON.parse(event.data);
                console.log('📨 LogInScreen received:', response);
                
                if (response.type === 'login_failed') {
                    console.error('❌ Login failed:', response.message);
                    setError(response.message || 'Login failed');
                } else if (response.type === 'login_success') {
                    console.log('✅ Login successful for:', name);
                    setError('');
                    onAuthSuccess(name);
                }
            } catch (err) {
                console.error('Error parsing login response:', err);
            }
        };

        ws.addEventListener('message', handleMessage);
        return () => ws.removeEventListener('message', handleMessage);
    }, [ws, name, onAuthSuccess]);

    const handleSubmit = (event) => {
        event.preventDefault();
        setError('');

        if (!ws) {
            setError('WebSocket not connected');
            return;
        }

        if (!name.trim() || !password.trim()) {
            setError('Please enter both name and password');
            return;
        }

        const payload = {
            type: 'login_user',
            user: {
                Name: name,      // Capital N
                Password: password,  // Capital P
            },
            payload: null,
        };
        
        console.log('📤 Sending login payload:', payload);
        ws.send(JSON.stringify(payload));
    };

    return (
        <div className="name-entry-screen">
            <h1>Name Entry Screen</h1>
            <form onSubmit={handleSubmit}>
                <input
                    type="text"
                    placeholder="Enter your name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                />
                <input
                    type="password"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />
                <button type="submit">Log in and Join Game</button>
                {error && <p style={{ color: 'red' }}>{error}</p>}
            </form>
            <button onClick={onSwitchToRegister} type="button">
                Don't have an account? Register here
            </button>
        </div>
    );
}