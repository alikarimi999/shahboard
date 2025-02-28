import { currentGame } from './gameState.js';
import { user } from './user.js'

export async function findMatch() {
    const response = await fetch("http://localhost:8082/user/match", {
        method: "GET",
        headers: {
            "Authorization": `Bearer ${user.jwt_token}`,
            "Content-Type": "application/json"
        }
    });

    if (!response.ok) {
        throw new Error(`Failed to find match: ${response.statusText}`);
    }

    const matchData = await response.json();
    const matchJson = JSON.stringify(matchData);
    const encoder = new TextEncoder();
    const matchBinary = encoder.encode(matchJson); // Convert to Uint8Array
    const matchBase64 = btoa(String.fromCharCode(...matchBinary)); // Convert binary to Base64

    currentGame.matchId = matchData.id;

    const message = {
        id: matchData.id,
        type: "find_match",
        timestamp: Date.now(),
        data: matchBase64
    };

    console.log("message: ", message);

    // Send message over WebSocket
    currentGame.ws.sendMessage(message);
}

