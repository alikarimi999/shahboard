import { currentGame } from './gameState.js';
import { connectWebSocket } from './ws.js';
import { initializeBoard } from './board.js';
import { config } from './config.js';
import {
    handleGameChatMsg, handleMoveApproved, handlePlayerJoined, handlePlayerLeft,
    handleGameEnded, handleError, handleViewGame, handleViewersList
} from './eventHandlers.js';
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
    connectWebSocket(`${config.baseUrl}/wsgateway/ws`).then(connection => {
        currentGame.ws = connection;

        currentGame.ws.registerMessageHandler("msg_approved", handleGameChatMsg);
        currentGame.ws.registerMessageHandler("move_approved", handleMoveApproved);
        currentGame.ws.registerMessageHandler("game_ended", handleGameEnded);
        currentGame.ws.registerMessageHandler("player_joined", handlePlayerJoined);
        currentGame.ws.registerMessageHandler("player_left", handlePlayerLeft);
        currentGame.ws.registerMessageHandler("view_game", handleViewGame);
        currentGame.ws.registerMessageHandler("viewers_list", handleViewersList);
        currentGame.ws.registerMessageHandler("err", handleError);

        const game = initializeBoard(false);
        const data = {
            game_id: currentGame.gameId,
            timestamp: Date.now()
        };

    }).catch(error => {
        console.error('Connection failed:', error);
    });
}


window.addEventListener("beforeunload", () => {
    currentGame.ws.unregisterMessageHandler("msg_approved");
    currentGame.ws.unregisterMessageHandler("move_approved");
    currentGame.ws.unregisterMessageHandler("game_ended");
    currentGame.ws.unregisterMessageHandler("player_connection_updated");
    currentGame.ws.unregisterMessageHandler("view_game");
    currentGame.ws.unregisterMessageHandler("viewers_list");
    currentGame.ws.unregisterMessageHandler("err");
});