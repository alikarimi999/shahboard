import { user } from './user.js';
import { logout } from './auth.js';

export function connectWebSocket(url) {
    if (user.jwt_token === null || user.jwt_token === "") {
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
    let resolveConnection;
    let isFirstPong = true; // Add flag to track first pong after connection

    const connectionPromise = new Promise((resolve) => {
        resolveConnection = resolve;
    });

    function initializeSocket() {
        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            console.log("Connected to WebSocket");
            pingInterval = setInterval(() => {
                if (socket.readyState === WebSocket.OPEN && pongReceived) {
                    socket.send(new Uint8Array([0x0]));
                    pongReceived = false;
                }
            }, 1000);
        };

        socket.onmessage = (event) => {
            if (event.data instanceof Blob) {
                const reader = new FileReader();
                reader.onload = function () {
                    const byteArray = new Uint8Array(reader.result);
                    handleBinaryMessage(byteArray);
                };
                reader.readAsArrayBuffer(event.data);
            } else if (typeof event.data === "string") {
                if (!event.data || event.data.trim() === "") {
                    console.warn("Received empty or whitespace-only message:", event.data);
                    return;
                }
                try {
                    const parsedMessage = JSON.parse(event.data);
                    handleTextMessage(parsedMessage);
                } catch (e) {
                    console.error("Failed to parse WebSocket message as JSON:", event.data, e);
                }
            } else {
                console.warn("Unknown message type:", event.data);
            }
        };

        socket.onclose = (event) => {
            clearInterval(pingInterval);
            if (event.code === 1006) {
                alert("Connection closed: " + event.reason);
            } else {
                console.warn("Unexpected disconnection:", event);
                alert("Connection closed: " + event.reason);
            }
            window.location.reload();
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
            if (isFirstPong) {
                // Dispatch connected event on first pong
                // document.dispatchEvent(new Event("websocket_connected"));
                // document.dispatchEvent(new Event("opponent_connected"));
                isFirstPong = false;
            }

            resolveConnection({
                sendMessage,
                registerMessageHandler,
                unregisterMessageHandler,
                close: () => {
                    socket.close();
                    clearInterval(pingInterval);
                    clearInterval(healthCheckInterval);
                }
            });
        } else {
            console.warn("Unknown binary message received:", byteArray);
        }
    }

    function checkServerHealth() {
        if (user.jwt_token === null || user.jwt_token === "") {
            return;
        }

        if (user.jwt_token != null && Date.now() - lastReceivedTime > 10000) {
            console.log("Server unresponsive: ", Date.now() - lastReceivedTime);
            // WebSocket disconnected event 
            document.dispatchEvent(new Event("websocket_disconnected"));
            socket.close();
            clearInterval(pingInterval);
            lastReceivedTime = Date.now();
            pongReceived = true;
            initializeSocket();
        }
    }

    healthCheckInterval = setInterval(checkServerHealth, 5000);

    return connectionPromise;
}