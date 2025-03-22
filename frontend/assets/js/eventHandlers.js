import { currentGame } from './gameState.js';
import { updateBoardPosition, ColorWhite, ColorBlack } from './board.js';
import { user } from './user.js'
import { showErrorMessage } from './error.js';

const GameOutcome = Object.freeze({
    NoOutcome: "*",
    WhiteWon: "1-0",
    BlackWon: "0-1",
    Draw: "1/2-1/2",
});

export function handleGameCreated(base64Data) {
    const gameData = JSON.parse(atob(base64Data));

    if (!gameData || !gameData.game_id || currentGame.matchId !== gameData.match_id) return;

    const { player1, player2 } = gameData;
    const userId = user.id;

    if (player1.id === userId) {
        currentGame.color = player1.color === 1 ? 'w' : 'b';
        currentGame.opponent.id = player2.id;
        currentGame.player.id = player1.id;
    } else if (player2.id === userId) {
        currentGame.color = player2.color === 1 ? 'w' : 'b';
        currentGame.opponent.id = player1.id;
        currentGame.player.id = player2.id;
    }

    currentGame.gameId = gameData.game_id;
    currentGame.game.reset();
    updateBoardPosition();
    currentGame.board.orientation(currentGame.color === 'w' ? 'white' : 'black');

    document.dispatchEvent(new Event("game_created"));
}

export function handleGameChatCreated(base64Data) {
    const chatData = JSON.parse(atob(base64Data));
    if (chatData.game_id === currentGame.gameId) {
        document.dispatchEvent(new CustomEvent("chat_created"));
    }
}

export function handleMoveApproved(base64Data) {
    const moveData = JSON.parse(atob(base64Data));

    // This is for validating and applying opponent's move in player mode.
    if (currentGame.isPlayer && moveData.game_id === currentGame.gameId &&
        moveData.player_id === currentGame.opponent.id && moveData.index - 1 === currentGame.game.history().length) {
        currentGame.game.move(moveData.move);
        updateBoardPosition();
        return;
    }

    // This is for validating and applying player's move in player mode.
    // this is for situations that user playing with multiple sessions.
    if (currentGame.isPlayer && moveData.game_id === currentGame.gameId &&
        moveData.player_id === currentGame.player.id && moveData.index - 1 === currentGame.game.history().length) {
        currentGame.game.move(moveData.move);
        updateBoardPosition();
        return;
    }

    // This is for validating and applying players moves in view mode
    if (moveData.game_id === currentGame.gameId && (moveData.player_id === currentGame.player.id ||
        moveData.player_id === currentGame.opponent.id) && moveData.index - 1 === currentGame.game.history().length) {
        currentGame.game.move(moveData.move);
        updateBoardPosition();
    }
}

export function handleGameChatMsg(base64Data) {
    const chatData = JSON.parse(atob(base64Data));
    if (currentGame.isPlayer && chatData.game_id === currentGame.gameId && chatData.sender_id === currentGame.opponent.id) {
        document.dispatchEvent(new CustomEvent("opponent_msg", { detail: { content: chatData.content } }));
    } else if (chatData.game_id === currentGame.gameId && (chatData.sender_id === currentGame.player.id || chatData.sender_id === currentGame.opponent.id)) {
        document.dispatchEvent(new CustomEvent("chat_msg", {
            detail: {
                sender: chatData.sender_id,
                content: chatData.content
            }
        }));
    }
}

export function handlePlayerConnectionUpdate(base64Data) {
    const playerData = JSON.parse(atob(base64Data));
    if (currentGame.isPlayer && playerData.player_id === currentGame.opponent.id) {
        if (playerData.connected) {
            document.dispatchEvent(new Event("opponent_connected"));
        } else {
            document.dispatchEvent(new Event("opponent_disconnected"));
        }
    } else {
        // Dispatch player_disconnected event with player id 
        if (playerData.connected) {
            document.dispatchEvent(new CustomEvent("player_connected", { detail: { id: playerData.player_id } }));

        } else {
            document.dispatchEvent(new CustomEvent("player_disconnected", { detail: { id: playerData.player_id } }));
        }
    }
}

export function handleGameEnded(base64Data) {

    const gameData = JSON.parse(atob(base64Data));

    if (gameData.game_id !== currentGame.gameId) return;

    const eventMap = {
        [GameOutcome.WhiteWon]: currentGame.color === ColorWhite ? "win" : "lose",
        [GameOutcome.BlackWon]: currentGame.color === ColorBlack ? "win" : "lose",
        [GameOutcome.Draw]: "draw",
    };

    const result = eventMap[gameData.outcome];

    if (!result) {
        console.error("Unknown game outcome:", gameData.outcome);
        return;
    }

    var desc = "";
    if (result == "win" && gameData.desc == "player_left") {
        desc = "Opponent left the game!";
    }

    // Dispatch a custom event with the game result
    const event = new CustomEvent("game_ended", {
        detail: { gameId: gameData.game_id, result, desc },
    });
    document.dispatchEvent(event);
}

export function handleResumeGame(base64Data) {
    const gameData = JSON.parse(atob(base64Data));
    if (gameData.game_id !== currentGame.gameId) return;
    const event = new CustomEvent("pgn_received", {
        detail: {
            gameId: gameData.game_id,
            pgn: gameData.pgn
        }
    })

    document.dispatchEvent(event);
}

export function handleViewGame(base64Data) {
    const gameData = JSON.parse(atob(base64Data));
    if (gameData.game_id !== currentGame.gameId) return;
    const event = new CustomEvent("pgn_received", {
        detail: {
            gameId: gameData.game_id,
            pgn: gameData.pgn
        }
    })
    document.dispatchEvent(event);
}

export function handleViewersList(base64Data) {
    const gameData = JSON.parse(atob(base64Data));
    if (gameData.game_id !== currentGame.gameId) return;
    const event = new CustomEvent("viewers_list", {
        detail: {
            gameId: gameData.game_id,
            list: gameData.list
        }
    })

    document.dispatchEvent(event);
}


export function handleError(base64Data) {
    try {
        showErrorMessage(atob(base64Data));
    } catch (e) {
        console.error("Error decoding error message:", e);
    }

}
