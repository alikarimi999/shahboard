import { user } from './user.js'
import { config } from './config.js'

export async function getUserLiveGameId() {
    try {
        const apiUrl = `${config.baseUrl}/game/live/user/${user.id}`;
        const response = await fetch(apiUrl, {
            method: 'GET',
            headers: {
                "Authorization": `Bearer ${user.jwt_token}`,
                'Content-Type': 'application/json',
            },
        });
        const data = await response.json();
        return data.game_id;
    } catch (error) {
        console.error('Error fetching user live game ID:', error.message);
        return null;
    }
}

// export async function getLivePgnByUser(userId) {
//     try {
//         const apiUrl = `http://localhost:8081/games/live?user_id=${userId}`;
//         return await fetchPgn(apiUrl);
//     } catch (error) {
//         console.error('Error fetching PGN by user:', error.message);
//         return null;
//     }
// }

// export async function getLivePgnByGame(gameId) {
//     try {
//         const apiUrl = `http://localhost:8081/games/live?game_id=${gameId}`;
//         return await fetchPgn(apiUrl);
//     } catch (error) {
//         console.error('Error fetching PGN by game:', error.message);
//         return null;
//     }
// }

// async function fetchPgn(apiUrl) {
//     try {
//         const response = await fetch(apiUrl, {
//             method: 'GET',
//             headers: {
//                 'Content-Type': 'application/json',
//                 // Add any required authentication headers here
//                 "Authorization": `Bearer ${user.jwt_token}`,
//             }
//         });

//         if (!response.ok) {
//             throw new Error(`HTTP error! status: ${response.status}`);
//         }

//         const data = await response.json();

//         const gameObject = {
//             id: data.id,
//             pgn: {
//                 raw: data.pgn,
//                 parsed: {}
//             }
//         };

//         const pgnLines = data.pgn.split('\n');
//         pgnLines.forEach(line => {
//             const match = line.match(/\[(\w+)\s+"([^"]+)"\]/);
//             if (match) {
//                 const [, key, value] = match;
//                 gameObject.pgn.parsed[key.toLowerCase()] = value;
//             }
//         });

//         const movesSection = data.pgn.split('\n\n')[1]?.trim();
//         if (movesSection) {
//             gameObject.pgn.moves = movesSection;
//         }
//         console.log('Game Object:', gameObject);
//         return gameObject;
//     } catch (error) {
//         console.error('Error fetching PGN:', error.message);
//         return null;
//     }
// }


export function parsePGN(pgn) {
    const gamePgn = {
        raw: pgn,
        parsed: {}
    };

    const pgnLines = pgn.split('\n');
    pgnLines.forEach(line => {
        const match = line.match(/\[(\w+)\s+"([^"]+)"\]/);
        if (match) {
            const [, key, value] = match;
            gamePgn.parsed[key.toLowerCase()] = value;
        }
    });

    const movesSection = pgn.split('\n\n')[1]?.trim();
    if (movesSection) {
        gamePgn.moves = movesSection;
    }

    return gamePgn;
}

