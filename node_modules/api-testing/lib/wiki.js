const supertest = require('supertest');
const crypto = require('crypto');
const querystring = require('querystring');
const { assert } = require('./assert');
const config = require('./config')();

/**
 * Runs some pending jobs.
 *
 * @param {int} n The number of jobs to run.
 * @return {Promise<number>} Zero if there are no more jobs to be run,
 * and a number grater than zero if there are more jobs ready to be run.
 * That number may or may not represent the number of jobs remaining
 * in the queue.
 */
const runJobs = async (n = 1) => {
    if (!config.secret_key) {
        throw Error('Missing secret_key configuration value. ' +
            'Set secret_key to the value of $wgSecretKey from LocalSettings.php');
    }

    const sig = (params) => {
        const data = {};
        const keys = Object.keys(params).sort();

        for (var k of keys) {
            data[k] = params[k];
        }

        const s = querystring.stringify(data);
        const hmac = crypto.createHmac('sha1', config.secret_key).update(s);
        return hmac.digest('hex');
    };

    const params = {
        title: 'Special:RunJobs',
        maxjobs: n,
        maxtime: Math.max(n * 10, 60),
        async: '', // false
        stats: '1', // true
        tasks: '', // what does this mean?
        sigexpiry: Math.ceil(Date.now() / 1000 + 60 * 60) // one hour
    };

    params.signature = sig(params);

    const response = await supertest.agent(config.base_uri).post('index.php').type('form').send(params);

    assert.isDefined(response.body.reached, `Text: ${response.text}`);

    if (response.body.reached === 'none-ready') {
        // The backend reports that no more jobs are ready.
        return 0;
    } else {
        // If response.body.jobs is empty, we may be hitting an infinite
        // loop here. That should not happen.
        assert.isNotEmpty(response.body.jobs);

        // There is no reliable way to get the current size of the job queue.
        // Just return some number to indicate that there is more work to be done.
        return 100;
    }
};

/**
 * Returns a promise that will resolve when all jobs in the wiki's job queue
 * have been run.
 *
 * @return {Promise<void>}
 */
const runAllJobs = async () => {
    const log = () => {}; // TODO: allow optional logging

    while (true) {
        log('Running jobs...');
        const jobsRemaining = await runJobs(10);

        if (jobsRemaining) {
            log(`Still ${jobsRemaining} in the queue.`);
        } else {
            break;
        }
    }
};

const getSecretKey = () => {
    return config.secret_key || null;
};

module.exports = {
    runAllJobs,
    getSecretKey
};
