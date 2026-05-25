import { WebSocket } from 'k6/websockets';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { check } from 'k6';
import { sleep } from 'k6';

export const options = {
    vus: 10,
    iterations: 10,
    insecureSkipTLSVerify: true,
};

/**
 * This test simulates a bunch of users trying to login without 
 * registering first. The expected result is that all the login attempts will fail.
 */
export default function () {
    let randomUser = randomString(10)
    let randomPassword = randomString(10)

    const ws = WebSocket("wss://localhost:8080/ws")
    ws.addEventListener("open", () => {
        ws.addEventListener("message", (event) => {
            let reply = JSON.parse(event.data)
            
            if (reply.type == "login_failed" || reply.type == "login_success") {
                check(reply, {
                    "is login failed": (r) => r.type == "login_failed",
                })
                ws.close()
                return
            }

        })

        login(ws, randomUser, randomPassword)

    });
}

function register(ws, randomUser, randomPassword) {
        sleep((Math.random() * 5) + 4)
        ws.send(JSON.stringify({
            type: "register_user",
            user: {
                name: randomUser,
                password: randomPassword,
            },
        }))
}

function login(ws, randomUser, randomPassword) {
    sleep((Math.random() * 5) + 4)
    ws.send(JSON.stringify({
        type: "login_user",
        user: {
            name: randomUser,
            password: randomPassword,
        },
    }))
}