import React, { useState } from 'react';

import './NameEntryScreen.css';

export default function NameEntryScreen({ ws, onAuthSuccess, onSwitchToRegister }) {
    const [name, setName] = useState('');
    const[password, setPassword] = useState('');

    const handleSubmit = (event) => {
        event.preventDefault();

        if (ws && name.trim() && password.trim()) {
            const payload = {
                type: 'login_user',
                playerName: name,
                password: password,
                payload: null,
            };
            ws.send(JSON.stringify(payload));
        }
    };

    ws.onmessage = (event) => {
        const response = JSON.parse(event.data);
        if (response.type === 'login_failed') {
            //setError(response.message);
            //setLoading(false);
        } else if (response.type === 'login_success') {
            onAuthSuccess();
        }
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
            </form>
            <button onClick={onSwitchToRegister} type="button">
                Don't have an account? Register here
            </button>
        </div>
    );
}