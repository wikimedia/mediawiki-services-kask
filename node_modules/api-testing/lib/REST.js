const supertest = require('supertest');
const config = require('./config')();

class RESTClient {
    /**
     * Constructs a new agent for making HTTP requests to the MediaWiki
     * REST API. The agent acts like a browser session and has its own
     * cookie jar.
     */

    constructor(endpoint = 'rest.php/v1') {
        this.req = supertest.agent(config.base_uri + endpoint);
    }

    /**
     * Constructs an HTTP request to the REST API and returns the
     * corresponding supertest Test object, which behaves like a
     * superagent Request. It can be used like a Promise that resolves
     * to a Response.
     *
     * The request has not been sent when this method returns,
     * and can still be modified like a superagent request.
     * Call end() or then(), or use await to send the request.
     *
     * @param {string} endpoint
     * @param {string} method
     * @param {Object|string} params
     * @param {string} contentType
     * @return {Promise<*>}
     */
    async request(endpoint, method, params = {}, contentType = 'application/json') {
        let req;
        switch (method.toUpperCase()) {
            case 'GET':
                req = this.req.get(endpoint)
                    .query(params);
                break;
            case 'POST':
                req = this.req.post(endpoint)
                    .send(params)
                    .set('Content-Type', contentType);
                break;
            case 'PUT':
                req = this.req.put(endpoint)
                    .send(params)
                    .set('Content-Type', contentType);
                break;
            case 'DELETE':
                req = this.req.del(endpoint);
                break;
            default:
                throw new Error(`The following method is unsupported: ${method}`);
        }

        return req;
    }

    /**
     * Constructs a GET request and returns the
     * corresponding supertest Test object
     *
     * @param {string} endpoint
     * @param {Object|null}params
     * @return {Promise<*>}
     */
    async get(endpoint, params = {}) {
        return this.request(endpoint, 'GET', params);
    }

    /**
     * Constructs a POST request and returns the
     * corresponding supertest Test object
     *
     * @param {string} endpoint
     * @param {Object|string} params
     * @param {string} contentType
     * @return {Promise<*>}
     */
    async post(endpoint, params = {}, contentType = 'application/json') {
        return this.request(endpoint, 'POST', params, contentType);
    }

    /**
     * Constructs a PUT request and returns the
     * corresponding supertest Test object
     *
     * @param {string} endpoint
     * @param {Object|string} params
     * @param {string} contentType
     * @return {Promise<*>}
     */
    async put(endpoint, params = {}, contentType = 'application/json') {
        return this.request(endpoint, 'PUT', params, contentType);
    }

    /**
     * Constructs a DELETE request and returns the
     * corresponding supertest Test object
     *
     * @param {string} endpoint
     * @return {Promise<*>}
     */
    async del(endpoint) {
        return this.request(endpoint, 'DELETE');
    }
}

module.exports = RESTClient;
