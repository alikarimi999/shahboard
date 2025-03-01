import { currentGame } from './gameState.js';
import { connectWebSocket } from './ws.js';
import { initializeBoard } from './board.js';
import { handleGameChatMsg, handleMoveApproved, handlePlayerConnectionUpdate, handleGameEnded, handleError } from './eventHandlers.js';
import { showErrorMessage } from './error.js';

function getGameIdFromURL() {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('game_id');
}

// Get game ID from URL
const gameId = getGameIdFromURL();

if (!gameId) {
    // If no game ID, show an error and redirect to home
    showErrorMessage("Invalid game ID");
    window.location.href = '/'; // Redirect to the home page (adjust the URL as needed)
} else {
    currentGame.gameId = gameId;
    connectWebSocket('http://localhost:8083/ws').then(connection => {
        currentGame.ws = connection;

        currentGame.ws.registerMessageHandler("msg_approved", handleGameChatMsg);
        currentGame.ws.registerMessageHandler("move_approved", handleMoveApproved);
        currentGame.ws.registerMessageHandler("game_ended", handleGameEnded);
        currentGame.ws.registerMessageHandler("player_connection_updated", handlePlayerConnectionUpdate);
        currentGame.ws.registerMessageHandler("err", handleError);

        const game = initializeBoard(false);
        const data = {
            game_id: currentGame.gameId,
            timestamp: Date.now()
        };

        const jsonData = JSON.stringify(data);
        const base64Data = btoa(jsonData);

        // Send view game message
        currentGame.ws.sendMessage({
            type: "view",
            data: base64Data
        });

    }).catch(error => {
        console.error('Connection failed:', error);
    });
}


window.addEventListener("beforeunload", () => {
    currentGame.ws.unregisterMessageHandler("move_approved");
    currentGame.ws.unregisterMessageHandler("game_ended");
    currentGame.ws.unregisterMessageHandler("player_connection_updated");
    currentGame.ws.unregisterMessageHandler("err");
});