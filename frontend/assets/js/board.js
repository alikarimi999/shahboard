import { currentGame } from './gameState.js';
import { user } from './user.js';
import { getUserLiveGameId, parsePGN } from './game_utils.js';
import { showErrorMessage } from './error.js';
import { getUserProfile } from './user_info.js';

const whiteSquareYellow = '#DAA520';
const blackSquareYellow = '#AA7600';

export const ColorWhite = 'w';
export const ColorBlack = 'b';

let selectedSquare = null;

export function initializeBoard(isPlayer) {
    currentGame.game = new Chess(); // Create a new game instance

    if (isPlayer) {
        currentGame.isPlayer = true;
    }

    var config = {
        draggable: currentGame.isPlayer,
        position: 'start',
        onDragStart: onDragStart,
        onDrop: onDrop,
        onMouseoutSquare: onMouseoutSquare,
        onMouseoverSquare: onMouseoverSquare,
        onSnapEnd: onSnapEnd,
        pieceTheme: '../../assets/img/pieces/{piece}.svg',
    };
    currentGame.board = Chessboard('board', config);

    if (user.loggedIn) {
        let gameIdPromise = isPlayer ? getUserLiveGameId(user.id) : Promise.resolve(currentGame.gameId);
        return gameIdPromise.then(gameId => {
            if (!gameId) return;
            currentGame.gameId = gameId;

            if (isPlayer) {
                currentGame.player.id = user.id;
            }

            const loadingSpinner = document.getElementById("loading-spinner");
            loadingSpinner.classList.add("active");

            // Prepare WebSocket message
            const data = { game_id: gameId, timestamp: Date.now() };
            const jsonData = JSON.stringify(data);
            const base64Data = btoa(jsonData);
            const messageType = isPlayer ? "resume_game" : "view_game";

            currentGame.ws.sendMessage({
                type: messageType,
                data: base64Data
            });

            // Wait for "pgn_received" event with a 10-second timeout
            return new Promise((resolve) => {
                const timeout = setTimeout(() => {
                    showErrorMessage("Game data could not be loaded.");
                    loadingSpinner.classList.remove("active");
                }, 20000);

                function handlePgnReceived(event) {
                    clearTimeout(timeout);
                    document.removeEventListener("pgn_received", handlePgnReceived);

                    const responseData = event.detail;

                    if (!responseData || !responseData.pgn) {
                        showErrorMessage("Game data could not be loaded.");
                        return;
                    }

                    const pgn = parsePGN(responseData.pgn);

                    if (isPlayer) {
                        currentGame.player.id = user.id;

                        if (pgn.parsed.w === user.id) {
                            currentGame.color = "w";
                            currentGame.opponent.id = pgn.parsed.b;
                        } else {
                            currentGame.color = "b";
                            currentGame.opponent.id = pgn.parsed.w;
                        }
                    } else {
                        currentGame.player.id = pgn.parsed.w;
                        currentGame.opponent.id = pgn.parsed.b;
                        currentGame.color = "w";  // Viewer always sees the game as if they are white
                    }


                    // Fetch player and opponent profiles
                    Promise.all([
                        getUserProfile(currentGame.player.id).then(profile => {
                            currentGame.player.profile = profile;
                            currentGame.player.email = profile.email;
                            currentGame.player.name = profile.name;
                            currentGame.player.avatar_url = profile.avatar_url
                        }),
                        getUserProfile(currentGame.opponent.id).then(profile => {
                            currentGame.opponent.profile = profile;
                            currentGame.opponent.email = profile.email;
                            currentGame.opponent.name = profile.name;
                            currentGame.opponent.avatar_url = profile.avatar_url
                        })
                    ]).then(() => {
                        // Reset game and set board orientation
                        currentGame.game.reset();
                        currentGame.board.orientation(currentGame.color === 'w' ? 'white' : 'black');
                        currentGame.game.load_pgn(pgn.raw);
                        updateBoardPosition();
                        loadingSpinner.classList.remove("active");

                        document.dispatchEvent(new Event("game_created"));
                        document.dispatchEvent(new Event("pgn_applied"));
                        resolve();
                    }).catch(() => {
                        showErrorMessage("Failed to load player profiles.");
                        resolve();
                    });
                }

                // Attach the event listener
                document.addEventListener("pgn_received", handlePgnReceived, { once: true });
            });
        });
    }
}

function highlightCurrentSquare(square) {
    const $square = $('#board .square-' + square);
    const background = $square.hasClass('black-3c85d') ? blackSquareYellow : whiteSquareYellow;
    $square.css('background', background);
}

function removeDotSquares() {
    $('#board .square-55d63').css('background', '').find('.dot').remove();
}

function dotSquare(square, isThreatened) {
    const $square = $('#board .square-' + square);
    if (isThreatened) {
        $square.css('background', 'radial-gradient(circle, rgba(139, 0, 0, 0.6) 100%, rgba(139, 0, 0, 0))');
    } else {
        const dotStyle = `
            position: absolute;
            width: 30%;
            height: 30%;
            background-color: rgba(0, 0, 0, 0.3);
            border-radius: 50%;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
        `;
        $square.append($('<div class="dot"></div>').attr('style', dotStyle));
    }
}

export function updateBoardPosition() {
    currentGame.board.position(currentGame.game.fen());
}

function onDragStart(source, piece) {
    // Prevent picking up pieces if the game is over
    if (currentGame.game.game_over()) return false;

    if (currentGame.gameId) {
        // Prevent dragging opponent's pieces
        if ((currentGame.color === ColorWhite && piece.search(/^b/) !== -1) ||
            (currentGame.color === ColorBlack && piece.search(/^w/) !== -1)) {
            return false;
        }
    }

    if ((currentGame.color === ColorWhite && piece.search(/^b/) !== -1) ||
        (currentGame.color === ColorBlack && piece.search(/^w/) !== -1)) {
        return false;
    }

}

function onDrop(source, target) {
    removeDotSquares();

    if (currentGame.gameId && !currentGame.player.connected) {
        showDisconnectedMessage(false);
        return 'snapback';
    }
    if (currentGame.gameId && !currentGame.opponent.connected) {
        showDisconnectedMessage(true);
        return 'snapback';
    }

    const move = currentGame.game.move({ from: source, to: target, promotion: 'q' });

    if (!move) return 'snapback';

    if (currentGame.gameId) sendMove(move.san);
    // moveSound.play();

    updateBoardPosition();
}

function onMouseoverSquare(square, piece) {
    var moves = currentGame.game.moves({
        square: square,
        verbose: true
    });

    if (moves.length === 0) return;

    // Highlight valid move squares with a dot
    if (!selectedSquare) {
        highlightCurrentSquare(square);
        moves.forEach(move => {
            // Check if the move is a capture by looking for 'x' in the SAN
            if (move.san.includes('x')) {
                dotSquare(move.to, true); // Highlight as a threatened (capture) square
            } else {
                dotSquare(move.to, false); // Highlight as a normal valid move
            }
        });
    }
}

function onMouseoutSquare(square, piece) {
    if (!selectedSquare) {
        removeDotSquares();
    }
}

function onSnapEnd() {
    currentGame.board.position(currentGame.game.fen())
}


function sendMove(move) {
    const data = {
        game_id: currentGame.gameId,
        player_id: user.id,
        move: move,
        index: currentGame.game.history().length,
        timestamp: Date.now()
    };

    const jsonData = JSON.stringify(data);
    const base64Data = btoa(jsonData);

    const msg = {
        type: "player_moved",
        data: base64Data
    };

    currentGame.ws.sendMessage(msg);
}

function handleSquareSelection(square) {
    if (currentGame.gameId) {

        if (!currentGame.isPlayer) {
            showNotAllowedMessage();
            selectedSquare = null;
            removeDotSquares();
            return;
        }

        var piece = currentGame.game.get(square);
        if (piece && piece.color !== currentGame.color) {
            showNotAllowedMessage();
            return;
        }
    }


    selectedSquare = square;
    highlightCurrentSquare(square);
    // Get valid moves for the selected square
    var moves = currentGame.game.moves({
        square: square,
        verbose: true
    });

    // If no valid moves, reset selection
    if (moves.length === 0) {
        selectedSquare = null;
        return;
    }

    // Highlight valid move squares
    moves.forEach(move => {
        if (move.san.includes('x')) {
            dotSquare(move.to, true); // Highlight as a threatened (capture) square
        } else {
            dotSquare(move.to, false); // Highlight as a normal valid move
        }
    });
}


function attemptMove(destinationSquare) {
    // If no square is selected, ignore the click
    if (!selectedSquare) return;

    if (selectedSquare === destinationSquare) {
        // Deselect the square if it's clicked again
        selectedSquare = null;
        removeDotSquares();
        return;
    }

    if (!currentGame.isPlayer) {
        showNotAllowedMessage();
        selectedSquare = null;
        removeDotSquares();
        return;
    }

    if (!currentGame.player.connected) {
        showDisconnectedMessage(false);
        selectedSquare = null;
        removeDotSquares();
        return;
    }

    if (!currentGame.opponent.connected) {
        showDisconnectedMessage(true);
        selectedSquare = null;
        removeDotSquares();
        return;
    }

    const piece = currentGame.game.get(destinationSquare)
    if (piece && piece.color == currentGame.game.turn()) {
        selectedSquare = destinationSquare; // Change selection to the new piece
        removeDotSquares(); // Remove previous highlights
        highlightCurrentSquare(destinationSquare); // Highlight new selection
        var moves = currentGame.game.moves({
            square: destinationSquare,
            verbose: true
        });

        // If no valid moves, reset selection
        if (moves.length === 0) {
            selectedSquare = null;
            return;
        }

        moves.forEach(move => {
            if (move.san.includes('x')) {
                dotSquare(move.to, true); // Highlight as a threatened (capture) square
            } else {
                dotSquare(move.to, false); // Highlight as a normal valid move
            }
        });

        return;
    }

    // Try to make a move
    var move = currentGame.game.move({
        from: selectedSquare,
        to: destinationSquare,
        promotion: 'q' // Always promote to a queen for simplicity
    });

    // If the move is illegal, keep the selection
    if (move === null) {
        selectedSquare = null;
        return;
    }

    if (currentGame.gameId) {
        sendMove(move.san);
    }

    // emit a move event with its index and san
    const moveIndex = currentGame.game.history().length;
    const moveEvent = new CustomEvent('move', { detail: { index: moveIndex, san: move.san } });
    document.dispatchEvent(moveEvent);

    // moveSound.play();

    // Update the board position
    currentGame.board.position(currentGame.game.fen());

    // Clear the selection
    selectedSquare = null;

    // Remove all highlights and dots
    removeDotSquares();
}

// Add click-based event handlers using event delegation
$('#board').on('click', '.square-55d63, .piece-417db', function (event) {
    var square = $(this).closest('.square-55d63').attr('data-square');

    if (!square) return; // Avoid undefined values

    if (!selectedSquare) {
        handleSquareSelection(square);
    } else {
        attemptMove(square);
    }
});

$(window).resize(function () {
    currentGame.board.resize();
});

function showDisconnectedMessage(opponent) {
    if ($('#disconnected').length) return; // Avoid duplicate messages

    let txtMsg = "You are disconnected!";
    if (opponent) {
        txtMsg = "Opponent disconnected!";
    }

    showErrorMessage(txtMsg);
}

function showNotAllowedMessage() {
    showErrorMessage("Not allowed!");
}
