export const user = {
    id: null,
    email: null,
    name: null,
    picture: null,
    jwt_token: null,

    update: function () {
        const storedUser = localStorage.getItem("user");
        if (storedUser) {
            const parsed = JSON.parse(storedUser);
            this.id = parsed.id || null;
            this.email = parsed.email || null;
            this.name = parsed.name || null;
            this.picture = parsed.picture || null;
            this.jwt_token = parsed.jwt_token || null;
        }
    },

    clean: function () {
        this.id = null;
        this.email = null;
        this.name = null;
        this.picture = null;
        this.jwt_token = null;
    },

    loggedIn: function () {
        return this.id !== null;
    }
};

const initializeUser = () => {
    user.update();
};

initializeUser();