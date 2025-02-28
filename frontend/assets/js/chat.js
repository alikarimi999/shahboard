import { currentGame } from './gameState.js';

export function sendGameChatMsg(textMsg) {
    console.log("currentGame: ", currentGame.player)
    const data = {
        sender_id: currentGame.player.id,
        game_id: currentGame.gameId,
        content: textMsg,
        timestamp: Date.now()
    }

    const jsonData = JSON.stringify(data);
    const base64Data = btoa(jsonData);

    const msg = {
        type: "msg_send",
        data: base64Data
    };

    currentGame.ws.sendMessage(msg);
}