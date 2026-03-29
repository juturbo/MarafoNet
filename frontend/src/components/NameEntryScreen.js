import React, { useState } from 'react';

import './NameEntryScreen.css';

export default function NameEntryScreen({ ws }) {
    const [name, setName] = useState('');

    const handleSubmit = (event) => {
        event.preventDefault();

        if (ws && name.trim()) {
            const payload = {
                type: 'first_join',
                playerName: name,
                payload: null,
                uuid: ""
            };
            ws.send(JSON.stringify(payload));
            console.log('Sent:', payload);
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
                <button type="submit">Join Game</button>
            </form>
        </div>
    );
}