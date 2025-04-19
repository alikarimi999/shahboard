import { getUserProfile } from "./user_info.js";
import { showProfileSummary } from "./profile-summary.js";
import { user } from "./user.js";
import { config } from "./config.js";

const profileCache = new Map();

async function fetchLiveGamesData() {
    const spinner = document.getElementById('refresh-spinner');
    spinner.classList.remove('hidden');

    try {
        const response = await fetch(`${config.baseUrl}/game/live/data`, {
            method: 'GET',
            headers: {
                "Authorization": `Bearer ${user.jwt_token}`,
                'Content-Type': 'application/json',
            },
        });
        const data = await response.json();

        const header = document.querySelector('.page-header h2');
        header.textContent = `Live Games: ${data.total || 0}`;

        if (Array.isArray(data.list)) {
            displayGames(data.list);
        } else {
            console.error('Unexpected data format:', data);
        }

        const lastUpdated = document.getElementById('last-updated');
        lastUpdated.textContent = `Last Updated: ${new Date().toLocaleTimeString()}`;
    } catch (error) {
        console.error('Error fetching live games:', error);
    } finally {
        spinner.classList.add('hidden');
    }
}

async function displayGames(games) {
    const container = document.getElementById('games-container');
    const sortOption = document.getElementById('sort-games').value;

    // Identify the top three games by viewer count
    const topViewerGames = [...games].sort((a, b) => b.viewers_number - a.viewers_number);
    const topThreeGameIds = topViewerGames.slice(0, 3).map(game => game.game_id);

    // Apply user's sort preference for display
    let sortedGames = [...games];
    if (sortOption === 'viewers-desc') {
        sortedGames.sort((a, b) => b.viewers_number - a.viewers_number);
    } else if (sortOption === 'viewers-asc') {
        sortedGames.sort((a, b) => a.viewers_number - b.viewers_number);
    } else if (sortOption === 'time-desc') {
        sortedGames.sort((a, b) => b.started_at.localeCompare(a.started_at));
    }

    container.innerHTML = '';

    for (const game of sortedGames) {
        const [player1, player2] = await Promise.all([
            fetchUserProfile(game.player1.id),
            fetchUserProfile(game.player2.id)
        ]);

        const player1Name = player1?.name || "Unknown Player 1";
        const player2Name = player2?.name || "Unknown Player 2";

        let player1DisconnectedAt = null;
        let player2DisconnectedAt = null;
        if (Array.isArray(game.players_disconnection)) {
            const player1Disconnection = game.players_disconnection.find(d => d.player_id === game.player1.id);
            const player2Disconnection = game.players_disconnection.find(d => d.player_id === game.player2.id);
            player1DisconnectedAt = player1Disconnection?.disconnected_at || null;
            player2DisconnectedAt = player2Disconnection?.disconnected_at || null;
        }

        const gameElement = document.createElement('div');
        gameElement.className = 'game-card';

        const player1Container = document.createElement('span');
        player1Container.className = 'player-container';
        player1Container.innerHTML = `
            <a href="/profile.html?userId=${game.player1.id}" 
               class="profile-link" 
               target="_blank" 
               rel="noopener noreferrer">
                ${player1Name}
            </a>
        `;
        if (player1DisconnectedAt) {
            showPlayerTimer(player1Container, player1DisconnectedAt);
        }

        const player2Container = document.createElement('span');
        player2Container.className = 'player-container';
        player2Container.innerHTML = `
            <a href="/profile.html?userId=${game.player2.id}" 
               class="profile-link" 
               target="_blank" 
               rel="noopener noreferrer">
                ${player2Name}
            </a>
        `;
        if (player2DisconnectedAt) {
            showPlayerTimer(player2Container, player2DisconnectedAt);
        }

        // Apply "Popular" badge to the top three games by viewers
        const isPopular = topThreeGameIds.includes(game.game_id);

        // Convert start time to readable format
        const startTime = game.started_at
            ? new Date(game.started_at * 1000).toLocaleTimeString()
            : 'Unknown';

        gameElement.innerHTML = `
            ${isPopular ? '<span class="popular-badge">🔥 Popular</span>' : ''}
            <h3 class="game-title">
                ${player1Container.outerHTML} vs ${player2Container.outerHTML}
            </h3>
            <div class="game-info">
                <div class="info-item">
                    <span class="info-label">Game:</span>
                    <span class="info-value">${game.game_id}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Viewers:</span>
                    <span class="info-value">${game.viewers_number}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Started:</span>
                    <span class="info-value">${startTime}</span>
                </div>
            </div>
            <a href="/view.html?game_id=${game.game_id}" class="watch-button" target="_blank" 
               rel="noopener noreferrer">Watch</a>
        `;

        container.appendChild(gameElement);

        const player1Link = gameElement.querySelector(`a[href="/profile.html?userId=${game.player1.id}"]`);
        const player2Link = gameElement.querySelector(`a[href="/profile.html?userId=${game.player2.id}"]`);
        player1Link.addEventListener('mouseenter', () => showProfileSummary(game.player1.id, player1, player1Link));
        player2Link.addEventListener('mouseenter', () => showProfileSummary(game.player2.id, player2, player2Link));
    }
}

function showPlayerTimer(element, disconnected_at) {
    const gameEndDuration = 120;

    const validDisconnectedAt = (typeof disconnected_at === 'number' && !isNaN(disconnected_at))
        ? disconnected_at
        : Math.floor(Date.now() / 1000);

    const gameEndTime = validDisconnectedAt + gameEndDuration;

    const timerContainer = document.createElement('span');
    timerContainer.className = 'timer-container';
    timerContainer.title = 'Player must reconnect by this time';
    timerContainer.innerHTML = `
        <span class="timer-disconnected">❌</span>
        <span class="timer"></span>
    `;
    element.insertBefore(timerContainer, element.firstChild);

    const timerSpan = timerContainer.querySelector('.timer');
    if (!timerSpan) {
        console.error('Timer span not found in element:', element);
        return;
    }

    const updateTimer = () => {
        const nowSeconds = Math.floor(Date.now() / 1000);
        const timeLeftSeconds = gameEndTime - nowSeconds;

        if (timeLeftSeconds <= 0) {
            timerSpan.textContent = '00:00';
            if (element.timerInterval) {
                clearInterval(element.timerInterval);
                element.timerInterval = null;
            }
        } else {
            const minutes = Math.floor(timeLeftSeconds / 60);
            const seconds = timeLeftSeconds % 60;
            timerSpan.textContent = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
            if (timeLeftSeconds <= 10) {
                timerContainer.classList.add('timer-critical');
                timerContainer.classList.remove('timer-urgent');
            } else if (timeLeftSeconds <= 30) {
                timerContainer.classList.add('timer-urgent');
                timerContainer.classList.remove('timer-critical');
            } else {
                timerContainer.classList.remove('timer-urgent', 'timer-critical');
            }
        }
    };

    updateTimer();
    if (gameEndTime - Math.floor(Date.now() / 1000) > 0) {
        element.timerInterval = setInterval(updateTimer, 1000);
    }
}

async function fetchUserProfile(userId) {
    if (profileCache.has(userId)) {
        return profileCache.get(userId);
    }
    const profile = await getUserProfile(userId);
    if (profile) {
        profileCache.set(userId, profile);
    }
    return profile;
}

fetchLiveGamesData();
setInterval(fetchLiveGamesData, 30000);

document.getElementById('sort-games').addEventListener('change', fetchLiveGamesData);