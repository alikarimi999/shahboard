import { currentGame } from "./gameState.js";
import { getUserProfile } from "./user_info.js";

const viewerCount = document.getElementById("viewerNumber");
const viewerList = document.getElementById("viewerList");
const viewerContainer = document.getElementById("viewerCount");

// Cache to store viewer profiles and avoid redundant requests
const viewerProfiles = new Map();

// Show viewer list on hover
viewerContainer.addEventListener("mouseenter", () => {
    viewerList.style.display = "block";
});

viewerContainer.addEventListener("mouseleave", (event) => {
    // Check if the mouse is leaving the entire viewer container
    if (!viewerContainer.contains(event.relatedTarget)) {
        viewerList.style.display = "none";
    }
});

// Prevent hiding when hovering over viewerList itself
viewerList.addEventListener("mouseenter", () => {
    viewerList.style.display = "block";
});

viewerList.addEventListener("mouseleave", (event) => {
    if (!viewerContainer.contains(event.relatedTarget)) {
        viewerList.style.display = "none";
    }
});

// Handle viewer updates
document.addEventListener("viewers_list", async (event) => {
    const { gameId, list } = event.detail;
    if (gameId === currentGame.gameId) {
        const viewers = await fetchViewersInfo(list);
        showViewers(viewers);
    }
});

// Display viewer avatars and usernames
function showViewers(viewers) {
    viewerCount.textContent = viewers.length;
    viewerList.innerHTML = "";

    viewers.forEach((viewer) => {
        const viewerItem = document.createElement("div");
        viewerItem.classList.add("viewer");
        viewerItem.innerHTML = `
            <img src="${viewer.avatar_url}" alt="${viewer.name}">
            <span>${viewer.name}</span>
        `;

        viewerList.appendChild(viewerItem);
    });
}

// Fetch viewer profiles (check cache first)
async function fetchViewersInfo(ids) {
    const viewers = [];

    for (const id of ids) {
        if (viewerProfiles.has(id)) {
            viewers.push(viewerProfiles.get(id));
        } else {
            const profile = await getUserProfile(id);
            if (profile) {
                viewerProfiles.set(id, profile);
                viewers.push(profile);
            }
        }
    }

    return viewers;
}
