// JavaScript to handle profile summary on hover
document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('.username').forEach(username => {
        username.addEventListener('mouseenter', (e) => {
            console.log('Profile summary script loaded>>>>> ');
            const usernameElement = e.target;
            const profileSummary = document.createElement('div');
            profileSummary.classList.add('profile-summary');

            // Get data attributes
            const email = usernameElement.getAttribute('data-email');
            const uid = usernameElement.getAttribute('data-uid');
            const level = usernameElement.getAttribute('data-level');
            const ranking = usernameElement.getAttribute('data-ranking');

            // Set the content of the profile summary
            profileSummary.innerHTML = `
                <p><strong>Email:</strong> ${email}</p>
                <p><strong>UID:</strong> ${uid}</p>
                <p><strong>Level:</strong> ${level}</p>
                <p><strong>Ranking:</strong> ${ranking}</p>
            `;

            // Append to the parent of the username (where we want the summary to appear)
            usernameElement.appendChild(profileSummary);
        });

        username.addEventListener('mouseleave', (e) => {
            const profileSummary = e.target.querySelector('.profile-summary');
            if (profileSummary) {
                profileSummary.remove();  // Remove the profile summary when mouse leaves
            }
        });
    });
});
