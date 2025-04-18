import { showProfileSummary } from './profile-summary.js';
import { getUserProfile, getUserRating } from './user_info.js';
import { showErrorMessage } from './error.js';
import { user } from './user.js';
import { formatDate } from './utils.js';
import { config } from './config.js';

const opponentProfiles = new Map();

let currentPage = 1;
const pageSize = 10;

function getUserIdFromUrl() {
    const params = new URLSearchParams(window.location.search);
    return params.get('userId'); // Get the 'id' parameter from URL
}

const userId = getUserIdFromUrl();
if (userId) {
    fetchUserProfile(userId);
    fetchGameHistory(userId, currentPage);
} else {
    showErrorMessage('User ID not found in URL.');
}

async function fetchUserProfile(userId) {
    try {
        const [userProfile, userRating] = await Promise.all([
            getUserProfile(userId),
            getUserRating(userId)
        ]);

        if (!userProfile) throw new Error('User profile not found');

        document.getElementById('profile-header').innerHTML = `
        <img src="${userProfile.avatar_url}" alt="${userProfile.name}" class="profile-avatar">
        <div id="profile-info" class="profile-info">

            <h2>${userProfile.email}</h2>
            <div class="profile-grid">
                <div>
                    <p><strong>UID:</strong> ${userId || 'N/A'}</p>
                    <p><strong>Current Score:</strong> ${userRating?.current_score ?? 'N/A'}</p>
                    <p><strong>Best Score:</strong> ${userRating?.best_score ?? 'N/A'}</p>
                    <p><strong>Level:</strong> ${userProfile.level.charAt(0).toUpperCase() + userProfile.level.slice(1) || 'N/A'}</p>
                </div>
                <div>
                    <p><strong>Games Played:</strong> ${userRating?.games_played ?? 'N/A'}</p>
                    <p><strong>Win:</strong> ${userRating?.games_won ?? 'N/A'}</p>
                    <p><strong>Loss:</strong> ${userRating?.games_lost ?? 'N/A'}</p>
                    <p><strong>Draw:</strong> ${userRating?.games_draw ?? 'N/A'}</p>
                </div>
            </div>
                <div class="profile-joined">
                    <p><strong>Joined:</strong> ${formatDate(userProfile.created_at) || 'N/A'}</p>
                 </div>
            </div>
        `;
    } catch (error) {
        console.error('Error fetching user profile:', error);
    }
}


// Fetch game history from backend
async function fetchGameHistory(userId, page) {
    try {
        if (page < 1) return;

        const response = await fetch(`${config.baseUrl}/profile/rating/history/${userId}?page=${page}&limit=${pageSize}`,
            {
                method: 'GET',
                headers: {
                    "Authorization": `Bearer ${user.jwt_token}`,
                    'Content-Type': 'application/json',
                },
            }
        );

        if (!response.ok) {
            throw new Error('Network response was not ok');
        }

        const data = await response.json();
        processGameHistory(data);

        currentPage = data.current_page;
        document.getElementById('page-info').textContent = `Page ${currentPage}`;
        document.getElementById('prev-btn').disabled = currentPage === 1;
        document.getElementById('next-btn').disabled = currentPage === data.total_pages;
    } catch (error) {
        console.error('Error fetching game history:', error);
    }
}

async function processGameHistory(data) {
    const gameHistoryTable = document.getElementById('game-history-table').getElementsByTagName('tbody')[0];
    gameHistoryTable.innerHTML = '';

    if (data.list && data.list.length > 0) {
        for (const game of data.list) {
            const row = gameHistoryTable.insertRow();

            const opponentCell = row.insertCell(0);
            opponentCell.textContent = `Loading...`;
            opponentCell.style.cursor = 'pointer';

            row.insertCell(1).textContent = game.game_id;
            row.insertCell(2).textContent = formatDate(game.timestamp, true);

            const resultCell = row.insertCell(3);
            resultCell.textContent = game.result.charAt(0).toUpperCase() + game.result.slice(1);
            resultCell.classList.add(game.result.toLowerCase());

            const eloChangeCell = row.insertCell(4);
            eloChangeCell.textContent = `${game.change > 0 ? '+' : ''}${game.change}`;
            eloChangeCell.classList.add(game.change > 0 ? 'positive' : 'negative');

            // Check if opponent profile is already cached
            if (opponentProfiles.has(game.opponent_id)) {
                // Use the cached profile
                const profile = opponentProfiles.get(game.opponent_id);
                opponentCell.innerHTML = `<a href="/profile.html?userId=${game.opponent_id}" class="profile-link" target="_blank" rel="noopener noreferrer">${profile.name}</a>`;

                const opponentLink = opponentCell.querySelector('.profile-link');
                opponentLink.addEventListener('mouseenter', () => showProfileSummary(game.opponent_id, profile, opponentCell));
            } else {
                getUserProfile(game.opponent_id).then(profile => {
                    opponentProfiles.set(game.opponent_id, profile);
                    opponentCell.innerHTML = `<a href="/profile.html?userId=${game.opponent_id}" class="profile-link" target="_blank" rel="noopener noreferrer">${profile.name}</a>`;

                    const opponentLink = opponentCell.querySelector('.profile-link');
                    opponentLink.addEventListener('mouseenter', () => showProfileSummary(game.opponent_id, profile, opponentCell));
                }).catch(() => {
                    opponentCell.textContent = 'Unknown';
                });
            }
        }
    } else {
        gameHistoryTable.innerHTML = '<tr><td colspan="5">No game history available.</td></tr>';
    }
}


function getCurrentPage() {
    return currentPage;
}

document.addEventListener("DOMContentLoaded", () => {
    const fetchHistoryBtn = document.getElementById("next-btn");

    if (fetchHistoryBtn) {
        fetchHistoryBtn.addEventListener("click", () => {
            fetchGameHistory(userId, getCurrentPage() + 1);
        });
    }
});

document.addEventListener("DOMContentLoaded", () => {
    const fetchHistoryBtn = document.getElementById("prev-btn");

    if (fetchHistoryBtn) {
        fetchHistoryBtn.addEventListener("click", () => {
            fetchGameHistory(userId, getCurrentPage() - 1);
        });
    }
});