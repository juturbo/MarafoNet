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

    const ws = WebSocket("wss://localhost:8080/ws")
    ws.addEventListener("open", () => {
        ws.addEventListener("message", (event) => {
            let reply = JSON.parse(event.data)
            if (reply.type == "register_success" || reply.type == "register_failure") {
                check(reply, {
                    "is user registered": (r) => r.type == "register_success",
                })
            }
            else if (reply.type == "login_success" || reply.type == "login_failure") {
                check(reply, {
                    "is user logged in": (r) => r.type == "login_success",
                })
            }
        })

        registerAndJoin(ws, randomUser, randomPassword)

        
    });
}

function registerAndJoin(ws, randomUser, randomPassword) {
        sleep((Math.random() * 5) + 4)
        ws.send(JSON.stringify({
            type: "register_user",
            user: {
                name: randomUser,
                password: randomPassword,
            },
        }))
        sleep((Math.random() * 5) + 4)
        ws.send(JSON.stringify({
            type: "first_join",
            user: {
                name: randomUser,
                password: randomPassword,
            }
        }))
        sleep((Math.random() * 5) + 4)
}