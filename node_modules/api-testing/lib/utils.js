module.exports = {
    /**
     * Returns a unique string of random alphanumeric characters.
     *
     * @param {int} n the desired number of characters
     * @return {string}
     */
    uniq(n = 10) {
        const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let result = '';

        for (let i = 0; i < n; i++) {
            result += characters.charAt(Math.floor(Math.random() * characters.length));
        }
        return result;
    },

    /**
     * Returns a unique title for use in tests.
     *
     * @param {string} prefix
     * @return {string}
     */
    title(prefix = '') {
        return prefix + this.uniq();
    },

    /**
     * Returns a promise that will resolve in no less than the given number of milliseconds.
     * @param {int} ms wait time in milliseconds
     * @return {Promise<void>}
     */
    sleep(ms = 1000) {
        return new Promise((resolve) => setTimeout(resolve, ms));
    },

    /**
     * Converts a title string to DB key form by replacing any spaces with underscores.
     * @param {string} title
     * @return {string}
     */
    dbkey(title) {
        return title.replace(/ /g, '_');
    }
};
