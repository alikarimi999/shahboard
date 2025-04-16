import { showUserProfile } from "./profile-summary.js";

export function showGameResult(winner, loser, result, desc) {
    document.getElementById('resultModal').style.display = 'block';

    const winnerAvatar = document.getElementById('winner-avatar');
    const winnerName = document.getElementById('winner-name');
    const loserAvatar = document.getElementById('loser-avatar');
    const loserName = document.getElementById('loser-name');
    const resultDescription = document.getElementById('result-description');


    winnerAvatar.src = winner.avatar_url;
    winnerAvatar.alt = `${winner.name}'s Avatar`;
    winnerName.textContent = winner.name;
    winnerAvatar.style.backgroundImage = `url(${winner.avatar_url})`;
    if (result == "draw") {
        winnerAvatar.className = "result-avatar draw";
    } else {
        winnerAvatar.className = "result-avatar winner";
    }

    loserAvatar.src = loser.avatar_url;
    loserAvatar.alt = `${loser.name}'s Avatar`;
    loserName.textContent = loser.name;
    loserAvatar.style.backgroundImage = `url(${loser.avatar_url})`;
    if (result == "draw") {
        loserAvatar.className = "result-avatar draw";
    } else {
        loserAvatar.className = "result-avatar loser";
    }

    showUserProfile(winner.id, winner, winnerName, true);
    showUserProfile(loser.id, loser, loserName, true);

    if (result == "draw") {
        resultDescription.textContent = "It's a draw!";
    } else {
        resultDescription.textContent = desc
    }
}

async function loadGameResultModalHTML() {
    const response = await fetch('../../result.html');
    const html = await response.text();
    const container = document.createElement('div');
    container.innerHTML = html;
    document.body.appendChild(container);
}

document.addEventListener('DOMContentLoaded', loadGameResultModalHTML);
