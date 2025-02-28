import { user } from "./user.js";
import { logout } from "./auth.js";

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
    console.log("Updating user profile...   ", user);
    if (user.picture && user.name) {
        document.getElementById("user-avatar").src = user.picture;
        document.getElementById("user-name").innerText = user.name;
        document.getElementById("logout-btn").classList.remove("hidden");
    }
}
