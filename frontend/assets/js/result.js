import { showUserProfile } from "./profile-summary.js";

export function showGameResult(winnerId, loserId, winnerProfile, loserProfile, result, desc) {
    document.getElementById('resultModal').style.display = 'block';

    const winnerAvatar = document.getElementById('winner-avatar');
    const winnerName = document.getElementById('winner-name');
    const loserAvatar = document.getElementById('loser-avatar');
    const loserName = document.getElementById('loser-name');
    const resultDescription = document.getElementById('result-description');


    winnerAvatar.src = winnerProfile.avatar_url;
    winnerAvatar.alt = `${winnerProfile.name}'s Avatar`;
    winnerName.textContent = winnerProfile.name;
    winnerAvatar.style.backgroundImage = `url(${winnerProfile.avatar_url})`;
    if (result == "draw") {
        winnerAvatar.className = "result-avatar draw";
    } else {
        winnerAvatar.className = "result-avatar winner";
    }

    loserAvatar.src = loserProfile.avatar_url;
    loserAvatar.alt = `${loserProfile.name}'s Avatar`;
    loserName.textContent = loserProfile.name;
    loserAvatar.style.backgroundImage = `url(${loserProfile.avatar_url})`;
    if (result == "draw") {
        loserAvatar.className = "result-avatar draw";
    } else {
        loserAvatar.className = "result-avatar loser";
    }

    showUserProfile(winnerId, winnerProfile, winnerName, true);
    showUserProfile(loserId, loserProfile, loserName, true);

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
