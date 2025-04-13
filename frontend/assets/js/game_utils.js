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

