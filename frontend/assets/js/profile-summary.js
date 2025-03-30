export function showProfileSummary(id, profile, element) {
    hideProfileSummary(); // Remove any existing tooltip

    const lastActive = profile.last_active_at
        ? new Date(profile.last_active_at * 1000).toLocaleString('en-US', {
            year: 'numeric',
            month: 'short',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false, // 24-hour format
        })
        : 'Unknown';

    // Create tooltip container
    const tooltip = document.createElement('div');
    tooltip.className = 'custom-tooltip';

    // Tooltip content
    tooltip.innerHTML = `
        <img src="${profile.avatar_url || 'default-avatar.jpg'}" alt="Profile Picture">
        <div class="tooltip-content">
            <strong>${profile.email || 'Unknown'}</strong>
            <p>UID: <span>${id}</span>
            <p>Score: ${profile.score || 'N/A'}</p>
            <p>Level: ${profile.level || 'N/A'}</p>
            <p>Last Active: ${lastActive}</p>
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