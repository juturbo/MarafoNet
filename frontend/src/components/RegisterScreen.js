import React, { useState } from 'react';

import './NameEntryScreen.css';

export default function RegisterScreen({ ws, onRegisterSuccess }) {
    const [name, setName] = useState('');
    const[password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');

    const handleSubmit = (event) => {
        event.preventDefault();

        if (password !== confirmPassword) {
            alert('Passwords do not match');
            return;
        }

        if (ws && name.trim() && password.trim()) {
            const payload = {
                type: 'register_user',
                user: {
                    Name: name,
                    Password: password,
                },
                payload: null,
            };
            ws.send(JSON.stringify(payload));
            console.log('Sent:', payload);
        }
    };

    ws.onmessage = (event) => {
        const response = JSON.parse(event.data);
        console.log('Received:', response);
        if (response.type === 'register_failed') {
            //setError(response.message);
            //setLoading(false);
        } else if (response.type === 'register_success') {
            onRegisterSuccess();
        }
    };

    return (
        <div className="name-entry-screen">
            <h1>Register</h1>
            <form onSubmit={handleSubmit}>
                <input
                    type="text"
                    placeholder="Enter your username"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                />
                <input
                    type="password"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />
                <input
                    type="password"
                    placeholder="Confirm your password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                />
                <button type="submit">Register</button>
            </form>
        </div>
    );
}