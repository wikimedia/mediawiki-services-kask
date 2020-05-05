'use strict';

const assert  = require('assert');
const request = require('supertest');


const TEST_URL = process.env.TEST_URL;


function randomString(len) {
   var result = '';
   var chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
   var charsLen = chars.length;
   for (let i = 0; i < len; i++) {
      result += chars.charAt(Math.floor(Math.random() * charsLen));
   }
   return result;
}


function newValue(len) {
    return `KASK:V:TEST_${randomString(len)}`;
}


describe('CRUD', function() {
    const key = `KASK:K:TEST_${randomString(64)}`;
    const val = newValue(128);
    const agent = request(TEST_URL);

    it('creates', function(done) {
        // GET should 404 (key has not yet been written)
        agent
            .get(`/${key}`)
            .end((err, {status}) => {
                assert.equal(status, 404, 'test value exists prior to create');
                // POST should return 201
                agent
                    .post(`/${key}`)
                    .set('Content-Type', 'application/octet-stream')
                    .send(val)
                    .expect(201, done);
        });
    });
    it('reads', function(done) {
        agent
            .get(`/${key}`)
            .expect('Content-Type', 'application/octet-stream')
            .expect(response => assert.equal(response.body, val))
            .expect(200, done);
    });
    it('updates', function(done) {
        const v = newValue(128);

        // POST an updated value
        agent
            .post(`/${key}`)
            .set('Content-Type', 'application/octet-stream')
            .send(v)
            .end((err, {status}) => {
                assert.equal(status, 201);
                // GET updated value and validate
                agent
                    .get(`/${key}`)
                    .expect('Content-Type', 'application/octet-stream')
                    .expect((response) => {
                        assert.equal(response.body.toString(), v, 'unexpected value after update');
                    })
                    .expect(200, done);
            });
    });
    it('deletes', function(done) {
        // DELETE value
        agent
            .delete(`/${key}`)
            .end((err, {status}) => {
                assert.equal(status, 204);
                // GET value (should now 404)
                agent
                    .get(`/${key}`)
                    .expect(404, done);
            });
    });
});
