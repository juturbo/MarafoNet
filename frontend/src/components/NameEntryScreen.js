import React from 'react';

import './NameEntryScreen.css';

export default function NameEntryScreen() {

    return (
        <div className="name-entry-screen">
            <h1>Name Entry Screen</h1>
            <form>
                <input type="text" placeholder="Enter your name" />
                <button type="submit">Join Game</button>
            </form>
        </div>
    );
}