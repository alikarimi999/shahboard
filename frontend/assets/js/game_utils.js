export async function getLivePgn(gameId, userId) {
    try {
        let apiUrl;
        if (userId) {
            apiUrl = `http://localhost:8081/games/live?user_id=${userId}`;
        } else {
            apiUrl = `http://localhost:8081/games/live?game_id=${gameId}`;
        }

        const response = await fetch(apiUrl, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                // Add any required authentication headers here
                // 'Authorization': 'Bearer your-token-here'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        const gameObject = {
            id: data.id,
            pgn: {
                raw: data.pgn,
                parsed: {}
            }
        };

        const pgnLines = data.pgn.split('\n');
        pgnLines.forEach(line => {
            const match = line.match(/\[(\w+)\s+"([^"]+)"\]/);
            if (match) {
                const [, key, value] = match;
                gameObject.pgn.parsed[key.toLowerCase()] = value;
            }
        });

        const movesSection = data.pgn.split('\n\n')[1]?.trim();
        if (movesSection) {
            gameObject.pgn.moves = movesSection;
        }
        console.log('Game Object:', gameObject);
        return gameObject;
    } catch (error) {
        console.error('Error fetching game:', error.message);
        return null;
    }
}