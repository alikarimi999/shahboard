import { formatDate } from "./utils.js";

export function showUserProfile(id, profile, element, connected) {
    if (connected) {
        element.innerHTML = `<a href="/profile.html?userId=${id}" class="profile-link" target="_blank" rel="noopener noreferrer">✅ ${profile.name} (${profile.score})</a>`;
    } else {
        element.innerHTML = `<a href="/profile.html?userId=${id}" class="profile-link" target="_blank" rel="noopener noreferrer">❌ ${profile.name} (${profile.score})</a>`;
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