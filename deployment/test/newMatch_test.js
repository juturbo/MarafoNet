import { WebSocket } from 'k6/websockets';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { check } from 'k6';
import { sleep } from 'k6';

export const options = {
    vus: 4,
    iterations: 4,
    insecureSkipTLSVerify: true,
};

export default function () {
    let randomUser = randomString(10)
    let randomPassword = randomString(10)

    let userRegistered = false
    let userLoggedIn = false

    const ws = WebSocket("wss://localhost:8080/ws")
    ws.addEventListener("open", () => {
        ws.addEventListener("message", (event) => {
            let reply = JSON.parse(event.data)
            if (reply.type == "register_success" || reply.type == "register_failure") {
                check(reply, {
                    "is user registered": (r) => r.type == "register_success",
                })
                if (reply.type == "register_success") {
                    console.log("Registration successful for user:", randomUser)
                    login(ws, randomUser, randomPassword)
                    console.log("Logging in user:", randomUser)
                    userLoggedIn = true
                    return
                }
            }
            else if (reply.type == "login_success" || reply.type == "login_failure") {
                check(reply, {
                    "is user logged in": (r) => r.type == "login_success",
                })
                console.log("Login response received for user:", randomUser)
                firstJoin(ws)
                return
            }
            else if (reply.type == "game_update") {
                check(reply, {
                    "game state received": (r) => r.type == "game_update",
                })
                console.log("Game state received for user:", randomUser)
                ws.close()
            }

        })

        if (!userRegistered) {
            console.log("Registering user:", randomUser)
            register(ws, randomUser, randomPassword)
            userRegistered = true
        };

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

function firstJoin(ws) {
    ws.send(JSON.stringify({
        type: "first_join",
    }))
}