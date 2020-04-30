
const { action, assert, REST, utils } = require('api-testing');


function newValue(len) {
    return `KASK:V:TEST_${utils.uniq(len)}`;
}


describe('Basic CRUD', () => {
    const client = new REST('');
    const key = `KASK:K:TEST_${utils.uniq(64)}`;
    const val = newValue(128);

    it('creates', async () => {
        return client.get(`/${key}`)
            .then(({status}) => {
                assert.strictEqual(status, 404, `${key} should not (yet) exist`);
            })
            .then(() => {
                client.post(`/${key}`, val, 'application/octet-stream')
                    .then(({status}) => {
                        assert.strictEqual(status, 201, `bad status code creating ${key}`);
                    });
            });
    });
    it('reads', async () => {
        const {status, headers, body} = await client.get(`/${key}`);
        assert.strictEqual(status, 200);
        assert.strictEqual(headers['content-type'], 'application/octet-stream');
        assert.strictEqual(body.toString(), val);
    });
    it('updates', async () => {
        const v = newValue(128);

        return client.post(`/${key}`, v, 'application/octet-stream')
            .then(({status}) => {
                assert.strictEqual(status, 201);
            })
            .then(() => {
                client.get(`/${key}`)
                    .then(({status, headers, body}) => {
                        assert.strictEqual(status, 200);
                        assert.strictEqual(headers['content-type'], 'application/octet-stream');
                        assert.strictEqual(body.toString(), v);
                    });
            });
    });
    it('deletes', async () => {
        return client.del(`/${key}`)
            .then(({status}) => {
                assert.strictEqual(status, 204);
            })
            .then(() => {
                client.get(`/${key}`)
                    .then(({status}) => {
                        assert.strictEqual(status, 404);
                    });
            });
    });
});
