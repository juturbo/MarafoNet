import { WebSocket } from 'k6/websockets';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { check } from 'k6';
import { sleep } from 'k6';

export const options = {
    insecureSkipTLSVerify: true,
    scenarios: {
        play_cards: {
            executor: 'per-vu-iterations',
            vus: 10,
            iterations: 1,
            exec: 'playCards'
        },
        choose_trump: {
            executor: 'per-vu-iterations',
            vus: 10,
            iterations: 1,
            exec: 'chooseTrump'
        }
    },
};

/**
 * This test simulates a bunch of users trying to login without 
 * registering first. The expected result is that all the login attempts will fail.
 */
export function playCards() {

    const ws = WebSocket("wss://localhost:8080/ws")
    ws.addEventListener("open", () => {
        ws.addEventListener("message", (event) => {
            let reply = JSON.parse(event.data)
            
            check(reply, {
                "invalid requests": (r) => r.type == "invalid_request",
            })
            ws.close()
            return

        })

        playCard(ws, {rank: 1, suit: 1})

    });
}

export function chooseTrump() {

    const ws = WebSocket("wss://localhost:8080/ws")
    ws.addEventListener("open", () => {
        ws.addEventListener("message", (event) => {
            let reply = JSON.parse(event.data)
            
            check(reply, {
                "invalid requests": (r) => r.type == "invalid_request",
            })
            ws.close()
            return

        })

        sendTrump(ws)

    });
}

function sendTrump(ws) {
    ws.send(JSON.stringify({
        type: "set_trump",
        payload: {
            gameId: null,
            suit: 1,
        },
    }))
}

function playCard(ws, card) {
    ws.send(JSON.stringify({
        type: "play_card",
        payload: {
            rank: card.Rank,
            suit: card.Suit,
        },
    }))
}
