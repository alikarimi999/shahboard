import { formatDate } from "./utils.js";

export function showUserProfile(id, profile, element, connected, disconnected_at) {
    const gameEndDuration = 120;

    if (element && element.timerInterval) {
        clearInterval(element.timerInterval);
        element.timerInterval = null;
    }

    if (!element || !(element instanceof HTMLElement)) {
        console.error('Invalid element passed to showUserProfile:', element);
        return;
    }

    if (connected) {
        element.innerHTML = `<a href="/profile.html?userId=${id}" class="profile-link" target="_blank" rel="noopener noreferrer">✅ ${profile.name} (${profile.score})</a>`;
    } else {
        const validDisconnectedAt = (typeof disconnected_at === 'number' && !isNaN(disconnected_at))
            ? disconnected_at
            : Math.floor(Date.now() / 1000);

        if (validDisconnectedAt !== disconnected_at) {
            console.warn(`Invalid disconnected_at (${disconnected_at}), using current time: ${validDisconnectedAt}`);
        }

        const gameEndTime = validDisconnectedAt + gameEndDuration;

        // Use a container for the timer with an icon and the time
        element.innerHTML = `
            <span class="timer-container" title="Opponent must reconnect by this time">
            <span class="timer-label">END IN</span>
            <span class="timer-icon">⏳</span>
                <span class="timer"></span>
            </span>
            <a href="/profile.html?userId=${id}" class="profile-link" target="_blank" rel="noopener noreferrer"> ${profile.name} (${profile.score})</a>
        `;
        const timerSpan = element.querySelector('.timer');
        const timerContainer = element.querySelector('.timer-container');

        if (!timerSpan || !timerContainer) {
            console.error('Timer elements not found in element:', element);
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
                timerContainer.classList.add('timer-urgent');
            }
        };

        updateTimer();
        if (gameEndTime - Math.floor(Date.now() / 1000) > 0) {
            element.timerInterval = setInterval(updateTimer, 1000);
        }
    }

    element.onmouseenter = () => {
        showProfileSummary(id, profile, element);
    };
}

export function showProfileSummary(id, profile, element) {
    hideProfileSummary(); // Remove any existing tooltip


    const tooltip = document.createElement('a');
    tooltip.className = 'custom-tooltip';
    tooltip.href = `/profile.html?userId=${id}`;
    tooltip.target = '_blank';
    tooltip.rel = 'noopener noreferrer';
    tooltip.style.textDecoration = 'none';

    // Tooltip content
    tooltip.innerHTML = `
        <img src="${profile.avatar_url || 'default-avatar.jpg'}" alt="Profile Picture">
        <div class="tooltip-content">
            <strong>${profile.email || 'Unknown'}</strong>
            <p>UID: <span>${id}</span>
            <p>Score: ${profile.score || 'N/A'}</p>
            <p>Level: ${profile.level || 'N/A'}</p>
            <p>Joined: ${formatDate(profile.created_at)}</p>
        </div>
    `;

    // Append tooltip to the body
    document.body.appendChild(tooltip);

    // Position tooltip near opponent name
    const rect = element.getBoundingClientRect();
    tooltip.style.left = `${rect.left + window.scrollX + 20}px`;
    tooltip.style.top = `${rect.top + window.scrollY - 10}px`;

    // Delay hiding when mouse leaves element
    element.onmouseleave = () => {
        setTimeout(() => {
            if (!tooltip.matches(':hover')) {
                hideProfileSummary();
            }
        }, 200); // Short delay to allow mouse movement
    };

    // Prevent tooltip from closing when hovered
    tooltip.onmouseenter = () => {
        tooltip.dataset.locked = "true"; // Prevent hiding
    };

    tooltip.onmouseleave = () => {
        tooltip.dataset.locked = "false"; // Allow hiding
        hideProfileSummary();
    };
}

export function hideProfileSummary() {
    const tooltip = document.querySelector('.custom-tooltip');
    if (tooltip && tooltip.dataset.locked !== "true") {
        tooltip.remove();
    }
}