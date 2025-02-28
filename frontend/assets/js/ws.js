import { user } from './user.js';

export function connectWebSocket(url) {
    if (user.jwt_token === "") {
        console.error("User is not authenticated");
        return null;
    }

    const wsUrl = `${url}?token=${user.jwt_token}`;
    let socket;
    let lastReceivedTime = Date.now();
    let pingInterval;
    let pongReceived = true;
    let healthCheckInterval;
    const messageHandlers = {};

    function initializeSocket() {
        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            console.log("Connected to WebSocket");

            pingInterval = setInterval(() => {
                if (socket.readyState === WebSocket.OPEN && pongReceived) {
                    const pingMessage = new Uint8Array([0x0]);
                    socket.send(pingMessage);
                    pongReceived = false;
                }
            }, 1000);
        };

        socket.onmessage = (event) => {
            if (event.data instanceof Blob) {
                const reader = new FileReader();
                reader.onload = function () {
                    const arrayBuffer = reader.result;
                    const byteArray = new Uint8Array(arrayBuffer);
                    handleBinaryMessage(byteArray);
                };
                reader.readAsArrayBuffer(event.data);
            } else if (typeof event.data === "string") {
                let jsonData = JSON.parse(event.data);
                handleTextMessage(jsonData);
            } else {
                console.warn("Unknown message type:", event.data);
            }
        };

        socket.onclose = () => {
            console.log("Disconnected from WebSocket");
            clearInterval(pingInterval);
        };
    }

    initializeSocket();

    function sendMessage(msg) {
        if (socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(msg));
        }
    }

    function registerMessageHandler(type, handler) {
        messageHandlers[type] = handler;
    }

    function unregisterMessageHandler(type) {
        delete messageHandlers[type];
    }

    function handleTextMessage(message) {
        const { type, data } = message;
        if (messageHandlers[type]) {
            messageHandlers[type](data);
        } else {
            console.warn("Unhandled message type:", type);
        }
    }

    function handleBinaryMessage(byteArray) {
        const messageType = byteArray[0];
        if (messageType === 0x1) {
            lastReceivedTime = Date.now();
            pongReceived = true;
        } else {
            console.warn("Unknown binary message received:", byteArray);
        }
    }

    function checkServerHealth() {
        if (Date.now() - lastReceivedTime > 10000) {
            console.log("Server unresponsive, reconnecting...");
            socket.close();
            clearInterval(pingInterval);
            lastReceivedTime = Date.now();
            pongReceived = true;
            initializeSocket();
        }
    }

    healthCheckInterval = setInterval(checkServerHealth, 5000);

    return {
        sendMessage,
        registerMessageHandler,
        unregisterMessageHandler,
        close: () => {
            socket.close();
            clearInterval(pingInterval);
            clearInterval(healthCheckInterval);
        }
    };
}