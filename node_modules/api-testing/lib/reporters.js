'use strict';


const Mocha = require('mocha');
const {EVENT_RUN_END, EVENT_TEST_FAIL} = Mocha.Runner.constants;


class NrpeReporter {
    constructor(runner) {
        const stats = runner.stats;
        const errMsgs = [];

        runner
            .on(EVENT_TEST_FAIL, (test, err) => {
                errMsgs.push(`${test.title}: ${err.message}`);
            })
            .once(EVENT_RUN_END, () => {
                const {passes, failures} = stats;
                const ratio = `${passes}/${(passes + failures)}`;

                if (stats.failures > 0) {
                    console.log(errMsgs.join('; '), `(${ratio} tests passed)`);
                } else {
                    console.log(`All good (${ratio} tests passed) `);
                }
            });
    }
}


module.exports = NrpeReporter;
