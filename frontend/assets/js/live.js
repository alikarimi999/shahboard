import { getUserProfile } from "./user_info.js";
import { showProfileSummary } from "./profile-summary.js";
import { user } from "./user.js";
import { config } from "./config.js";

const profileCache = new Map();


async function fetchLiveGamesData() {
    try {
        const response = await fetch(`${config.baseUrl}/game/live/data`,
            {
                method: 'GET',
                headers: {
                    "Authorization": `Bearer ${user.jwt_token}`,
                    'Content-Type': 'application/json',
                },
            }
        ); // Replace with actual API endpoint
        const data = await response.json();

        if (data.total > 0) {
            document.querySelector('.page-header h2').textContent = `Live Games: ${data.total}`;
        }

        if (Array.isArray(data.list)) {
            displayGames(data.list);
        } else {
            console.error('Unexpected data format:', data);
        }
    } catch (error) {
        console.error('Error fetching live games:', error);
    }
}

async function displayGames(games) {
    const container = document.getElementById('games-container');
    container.innerHTML = '';

    for (const game of games) {
        const [player1, player2] = await Promise.all([
            fetchUserProfile(game.player1.id),
            fetchUserProfile(game.player2.id)
        ]);

        const player1Name = player1?.name || "Unknown Player 1";
        const player2Name = player2?.name || "Unknown Player 2";

        const gameElement = document.createElement('div');
        gameElement.className = 'game-card'; // Use game-card class defined in CSS
        gameElement.innerHTML = `
            <h3 class="game-title">
                <a href="/profile.html?userId=${game.player1.id}" 
                   class="profile-link" 
                   target="_blank" 
                   rel="noopener noreferrer">
                    ${player1Name}
                </a> 
                vs 
                <a href="/profile.html?userId=${game.player2.id}" 
                   class="profile-link" 
                   target="_blank" 
                   rel="noopener noreferrer">
                    ${player2Name}
                </a>
            </h3>
            <p class="game-info">Game ID: ${game.game_id} | Viewers: ${game.viewers_number}</p>
            <a href="/view.html?game_id=${game.game_id}" class="watch-button" target="_blank" 
                   rel="noopener noreferrer">Watch</a>
        `;

        container.appendChild(gameElement);

        // Add hover event listeners for showing profile summaries
        const player1Link = gameElement.querySelector(`a[href="/profile.html?userId=${game.player1.id}"]`);
        const player2Link = gameElement.querySelector(`a[href="/profile.html?userId=${game.player2.id}"]`);

        player1Link.addEventListener('mouseenter', () => showProfileSummary(game.player1.id, player1, player1Link));
        player2Link.addEventListener('mouseenter', () => showProfileSummary(game.player2.id, player2, player2Link));
    }
}

async function fetchUserProfile(userId) {
    if (profileCache.has(userId)) {
        return profileCache.get(userId);
    }

    const profile = await getUserProfile(userId);
    if (profile) {
        profileCache.set(userId, profile); // Store in cache
    }
    return profile;
}

fetchLiveGamesData();
setInterval(fetchLiveGamesData, 30000); // Refresh data every 30 seconds