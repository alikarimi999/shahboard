import { currentGame } from './gameState.js';
import { connectWebSocket } from './ws.js';
import { config } from './config.js';
import { initializeBoard } from './board.js';
import {
    handleGameCreated, handleGameChatCreated, handleGameChatMsg, handleMoveApproved,
    handlePlayerJoined, handlePlayerLeft, handleGameEnded, handleError, handleResumeGame, handleViewersList
} from './eventHandlers.js';

connectWebSocket(`${config.baseUrl}/wsgateway/ws`).then(connection => {
    currentGame.ws = connection;

    currentGame.ws.registerMessageHandler("game_created", handleGameCreated);
    currentGame.ws.registerMessageHandler("chat_created", handleGameChatCreated);
    currentGame.ws.registerMessageHandler("msg_approved", handleGameChatMsg);
    currentGame.ws.registerMessageHandler("move_approved", handleMoveApproved);
    currentGame.ws.registerMessageHandler("game_ended", handleGameEnded);
    currentGame.ws.registerMessageHandler("player_joined", handlePlayerJoined);
    currentGame.ws.registerMessageHandler("player_left", handlePlayerLeft);
    currentGame.ws.registerMessageHandler("resume_game", handleResumeGame);
    currentGame.ws.registerMessageHandler("viewers_list", handleViewersList);
    currentGame.ws.registerMessageHandler("err", handleError);

    initializeBoard(true);

}).catch(error => {
    console.error('Connection failed:', error);
});


window.addEventListener("beforeunload", () => {
    currentGame.ws.unregisterMessageHandler("game_created");
    currentGame.ws.unregisterMessageHandler("chat_created");
    currentGame.ws.unregisterMessageHandler("move_approved");
    currentGame.ws.unregisterMessageHandler("game_ended");
    currentGame.ws.unregisterMessageHandler("player_connection_updated");
    currentGame.ws.unregisterMessageHandler("resume_game");
    currentGame.ws.unregisterMessageHandler("viewers_list");
    currentGame.ws.unregisterMessageHandler("err");
});