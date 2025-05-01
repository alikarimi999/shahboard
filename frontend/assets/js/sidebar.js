import { user } from "./user.js";
import { logout } from "./auth.js";
import { copyToClipboard } from "./copy.js";

document.addEventListener("DOMContentLoaded", () => {
    fetch("../../sidebar.html")
        .then(response => response.text())
        .then(html => {
            document.getElementById("sidebar-container").innerHTML = html;
            setupSidebar(); // Initialize sidebar functionality
            updateUserProfile(); // Update user profile after sidebar loads
        })
        .catch(error => console.error("Error loading sidebar:", error));
});

function setupSidebar() {
    const burgerBtn = document.getElementById("burger-btn");
    const closeBtn = document.getElementById("close-btn");
    const sidebar = document.getElementById("sidebar");

    if (burgerBtn && sidebar) {
        burgerBtn.addEventListener("mouseover", () => {
            sidebar.classList.toggle("open");
            burgerBtn.classList.add("hidden");
        });
    }

    if (sidebar) {
        sidebar.addEventListener("mouseleave", () => {
            sidebar.classList.remove("open");
            burgerBtn.classList.remove("hidden");
        });
    }

    if (closeBtn) {
        closeBtn.addEventListener("click", () => {
            sidebar.classList.remove("open");
            burgerBtn.classList.remove("hidden");
        });
    }

    // Logout functionality
    const logoutBtn = document.getElementById("logout-btn");
    if (logoutBtn) {
        logoutBtn.addEventListener("click", () => {
            logout();
            window.location.reload();
        });
    }
}

function updateUserProfile() {
    if (user.id) {
        document.getElementById("user-avatar").src = user.avatar_url;
        document.getElementById("user-name").innerText = user.name;
        document.getElementById("logout-btn").classList.remove("hidden");
    }

    if (user.email) {
        document.getElementById("user-email").classList.remove("hidden");
        document.getElementById("email-text").innerText = maskText(user.email);
        document.getElementById("email-text").dataset.full = user.email; // Store full email
    }

    if (user.id) {
        document.getElementById("user-uid").classList.remove("hidden");
        document.getElementById("uid-text").innerText = maskText(user.id);
        document.getElementById("uid-text").dataset.full = user.id; // Store full UID

        const profileLink = document.getElementById("profile-menu-link");
        profileLink.href = `/profile.html?userId=${user.id}`;
    }

    // Add copy functionality
    document.getElementById("user-email").addEventListener("click", () => copyToClipboard(user.email));
    document.getElementById("user-uid").addEventListener("click", () => copyToClipboard(user.id));
}


function maskText(text) {
    if (!text || text.length < 7) return text; // Avoid masking short text
    return text.slice(0, 3) + "***" + text.slice(-3);
}