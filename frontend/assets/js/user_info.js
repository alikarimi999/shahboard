export async function getUserProfile(userId) {
    if (!userId) {
        console.error("User ID is required.");
        return null;
    }

    try {
        const response = await fetch(`http://localhost:8085/users/${userId}`);
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
        const response = await fetch(`http://localhost:8085/rating/${userId}`);
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