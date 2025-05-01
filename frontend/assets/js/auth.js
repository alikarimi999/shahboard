
import { user } from './user.js';
import { config } from './config.js';

// Function to dynamically load Google Sign-In SDK
function loadGoogleSDK() {
    return new Promise((resolve, reject) => {
        if (document.getElementById("google-sdk")) {
            resolve(); // SDK is already loaded
            return;
        }

        const script = document.createElement("script");
        script.src = "https://accounts.google.com/gsi/client";
        script.id = "google-sdk";
        script.async = true;
        script.defer = true;
        script.onload = resolve;
        script.onerror = reject;
        document.head.appendChild(script);
    });
}

// Function to load login UI if the user is not logged in
function loadLoginUI() {
    if (user.jwt_token == null) {
        fetch("./assets/components/login.html")
            .then(response => response.text())
            .then(html => {
                document.body.insertAdjacentHTML("beforeend", html); // Append login UI

                // Ensure SDK & UI are loaded before initializing Google Sign-In
                loadGoogleSDK().then(() => {
                    initializeGoogleSignIn();
                    checkUserStatus();
                }).catch(error => console.error("Error loading Google SDK:", error));
            })
            .catch(error => console.error("Error loading login UI:", error));
    }
}

// Function to initialize Google Sign-In after loading the UI
function initializeGoogleSignIn() {
    if (window.google) {
        google.accounts.id.initialize({
            client_id: "103572145818-otri5g8tq5uu1lv2il163tjti4na2v74.apps.googleusercontent.com",
            callback: handleCredentialResponse,
        });

        google.accounts.id.renderButton(
            document.getElementById("signin-button"),
            { theme: "outline", size: "large" }
        );

        google.accounts.id.prompt(); // Show One Tap sign-in if enabled
    } else {
        console.error("Google Sign-In SDK not loaded yet.");
    }
}
function handleCredentialResponse(response) {
    sendTokenToBackend(response.credential)
        .then(userData => {
            if (userData) {
                localStorage.setItem("user", JSON.stringify(userData));
                user.update();
                window.location.reload();
            } else {
                console.error("Failed to get user data from the backend.");
            }
        })
        .catch(error => {
            console.error("Error handling credential response:", error);
        });
}

async function sendTokenToBackend(token) {
    try {
        const response = await fetch(`${config.baseUrl}/auth/oauth/google`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ token: token }),
        });

        if (!response.ok) {
            throw new Error(`Server returned ${response.status}: ${await response.text()}`);
        }

        const data = await response.json();
        return data;
    } catch (error) {
        console.error("Error sending token to backend:", error);
        return null; // Or throw error to propagate it
    }
}


async function continueAsGuest() {
    try {
        const response = await fetch(`${config.baseUrl}/auth/guest`);

        if (!response.ok) {
            throw new Error(`Guest login failed: ${await response.text()}`);
        }

        const userData = await response.json();
        userData.picture = `https://api.dicebear.com/9.x/bottts/svg?seed=${userData.id}`;
        userData.is_guest = true;
        localStorage.setItem("user", JSON.stringify(userData));
        user.update();
        window.location.reload();
    } catch (error) {
        console.error("Error during guest login:", error);
        alert("Failed to continue as guest. Please try again.");
    }
}

// Function to decode JWT token
function parseJwt(token) {
    let base64Url = token.split('.')[1];
    let base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    let jsonPayload = decodeURIComponent(atob(base64).split('').map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    return JSON.parse(jsonPayload);
}


// Function to log out
export function logout() {
    user.clean();
    localStorage.removeItem("user"); // Clear stored data
    const popupMessage = document.getElementById("popup-message");
    const overlay = document.getElementById("overlay");

    if (popupMessage) popupMessage.style.display = "block";
    if (overlay) overlay.style.display = "block";
}

// Function to check if the user is logged in and update UI
function checkUserStatus() {
    if (user.id == null) {
        // Show pop-up and overlay if they exist
        const popupMessage = document.getElementById("popup-message");
        const overlay = document.getElementById("overlay");

        if (popupMessage) popupMessage.style.display = "block";
        if (overlay) overlay.style.display = "block";
    }
}

// Load login UI when the DOM is fully loaded
document.addEventListener("DOMContentLoaded", function () {
    loadLoginUI();
});

window.continueAsGuest = continueAsGuest;
