const { use } = require('chai');
const utils = require('./utils');

module.exports = use(function (_chai, _utils) {
    const assert = _chai.assert;

    /**
     * Compares two titles, applying some normalization
     * @param {string} act
     * @param {string} exp
     * @param {string} msg
     */
    assert.sameTitle = (act, exp, msg) => {
        new _chai.Assertion(utils.dbkey(act), msg, assert.deepEqual, true).to.eql(utils.dbkey(exp));
    };
});
