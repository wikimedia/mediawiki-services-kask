const { assert } = require('./assert');
const Client = require('./actionapi');
const utils = require('./utils');
const config = require('./config')();
const wiki = require('./wiki');

const singletons = {};

module.exports = {

    async root() {
        if (singletons.root) {
            return singletons.root;
        }

        const root = new Client();
        await root.login(config.root_user.name,
            config.root_user.password);

        await root.loadTokens(['createaccount', 'userrights', 'csrf']);

        const rightsToken = await root.token('userrights');
        assert.notEqual(rightsToken, '+\\');

        singletons.root = root;
        return root;
    },

    async user(name, groups = [], tokens = ['csrf']) {
        if (singletons[name]) {
            return singletons[name];
        }

        // TODO: Use a fixed user name for Alice. Works only on a blank wiki.
        let uname = utils.title(name);
        const passwd = utils.uniq();
        const root = await this.root();
        const client = new Client();

        const account = await client.createAccount({ username: uname, password: passwd });
        uname = account.username;

        if (groups.length) {
            // HACK: This reduces the chance of race conditions due to
            // replication lag. For the proper solution, see T230211.
            await wiki.runAllJobs();

            const groupResult = await root.addGroups(uname, groups);
            assert.sameMembers(groupResult.added, groups);
        }

        await client.account(uname, passwd);

        if (tokens.length) {
            await client.loadTokens(tokens);
        }

        singletons[name] = client;
        return client;
    },

    async alice() {
        return this.user('Alice');
    },

    async bob() {
        return this.user('Bob');
    },

    async mindy() {
        return this.user('Mindy', ['sysop'], ['userrights', 'csrf']);
    },

    async robby() {
        return this.user('Robby', ['bot']);
    },

    getAnon() {
        return new Client();
    }
};
