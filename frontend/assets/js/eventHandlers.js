import { currentGame } from './gameState.js';
import { updateBoardPosition, ColorWhite, ColorBlack } from './board.js';
import { user } from './user.js'

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
    if (currentGame.isPlayer && moveData.game_id === currentGame.gameId && moveData.player_id === currentGame.opponent.id) {
        currentGame.game.move(moveData.move);
        updateBoardPosition();
    } else if (moveData.game_id === currentGame.gameId && (moveData.player_id === currentGame.player.id || moveData.player_id === currentGame.opponent.id)) {
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
            console.log("Opponent connected");
        } else {
            document.dispatchEvent(new Event("opponent_disconnected"));
            console.log("Opponent disconnected");
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
    console.log("Game Ended:", base64Data);

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

    // const eventType = eventMap[gameData.outcome];
    // if (eventType) {
    //     document.dispatchEvent(new Event(eventType));
    // }
}

// export function handleResumeGame(base64Data) {
//     console.log("Resume Game:", base64Data);
//     const gameData = JSON.parse(atob(base64Data));

//     if (!gameData.game_id) {
//         return
//     }

//     return getLivePgn(gameData.game_id).then(game => {
//         if (game) {
//             currentGame.gameId = gameData.game_id;
//             currentGame.player.id = user.id;
//             if (game.pgn.parsed.w === user.id) {
//                 currentGame.color = "w";
//                 currentGame.opponent.id = game.pgn.parsed.b;
//             } else {
//                 currentGame.color = "b";
//                 currentGame.opponent.id = game.pgn.parsed.w;
//             }

//             currentGame.game.reset();
//             currentGame.board.orientation(currentGame.color === 'w' ? 'white' : 'black');
//             currentGame.game.load_pgn(game.pgn.raw);
//             updateBoardPosition();

//             document.dispatchEvent(new Event("game_created"));
//         }
//     });

// }

export function handleError(base64Data) {
    console.error("Error:", atob(base64Data));
}
