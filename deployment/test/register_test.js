import { WebSocket } from 'k6/websockets';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { check } from 'k6';
import { sleep } from 'k6';

export const options = {
	iterations: 300,
	vus: 300,
	insecureSkipTLSVerify: true,
};

const payload = {
  type: "register_user",
  user: {
	name: randomString(10),
	password: randomString(10),
  },
}

export default function () {
	const ws = WebSocket("wss://localhost:8080/ws")
	ws.addEventListener("open", () => {
		ws.addEventListener("message", (event) => {
			let reply = JSON.parse(event.data)
			check(reply, {
				"is user registered": (r) => r.type == "register_success"
			})
			ws.close()
		})
		sleep((Math.random() * 5) + 4)
		ws.send(JSON.stringify(payload))
	});
}