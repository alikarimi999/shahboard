import { currentGame } from './gameState.js';

const viewerCount = document.getElementById("viewerNumber");
const viewerList = document.getElementById("viewerList");
const viewerContainer = document.getElementById("viewerCount");

viewerContainer.addEventListener("mouseenter", () => {
    viewerList.style.display = "block";
});

viewerContainer.addEventListener("mouseleave", () => {
    viewerList.style.display = "none";
});

document.addEventListener("viewers_list", (event) => {
    const { gameId, list } = event.detail;
    if (gameId === currentGame.gameId) {
        let viewers = fetchViewersInfo(list)
        showViewers(viewers)
    }
})

function showViewers(list) {
    viewerCount.textContent = list.length;
    viewerList.innerHTML = "";

    list.forEach(viewer => {
        const viewerItem = document.createElement("div");
        viewerItem.classList.add("viewer");
        viewerItem.innerHTML = `
            <img src="${viewer.avatar}" alt="${viewer.id}">
            <span>${viewer.id}</span>
        `;
        viewerList.appendChild(viewerItem);
    });
}

// set a random profile picture for the user and return an array of objects with the user's id and the profile picture
function fetchViewersInfo(ids) {
    return ids.map(id => ({
        id,
        avatar: `https://i.pravatar.cc/30?img=${Math.floor(Math.random() * 70) + 1}`
    }));
}