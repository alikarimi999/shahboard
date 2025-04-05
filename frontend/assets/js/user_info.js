import { user } from "./user.js";
import { config } from "./config.js";

export async function getUserProfile(userId) {
    if (!userId) {
        console.error("User ID is required.");
        return null;
    }

    try {
        const response = await fetch(`${config.baseUrl}/profile/users/${userId}`,
            {
                method: 'GET',
                headers: {
                    "Authorization": `Bearer ${user.jwt_token}`,
                    'Content-Type': 'application/json',
                },
            }
        );
        if (!response.ok) {
            throw new Error('Failed to fetch profile');
        }

        const data = await response.json();
        return data; // Return profile data for reuse
    } catch (error) {
        console.error('Error fetching user profile:', error);
        return null;
    }
}

export async function getUserRating(userId) {
    if (!userId) {
        console.error("User ID is required.");
        return null;
    }

    try {
        const response = await fetch(`${config.baseUrl}/profile/rating/${userId}`,
            {
                method: 'GET',
                headers: {
                    "Authorization": `Bearer ${user.jwt_token}`,
                    'Content-Type': 'application/json',
                },
            }
        );
        if (!response.ok) {
            throw new Error('Failed to fetch profile');
        }

        const data = await response.json();
        return data; // Return profile data for reuse
    } catch (error) {
        console.error('Error fetching rating profile:', error);
        return null;
    }
}