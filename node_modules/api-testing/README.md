# API-Testing

API-Testing is a library for end-to-end integration tests for MediaWiki's
[Action API](https://www.mediawiki.org/wiki/API:Main_page) and [REST API](https://www.mediawiki.org/wiki/API:REST_API).
You can run tests locally by installing the NPM package and configuring it to access a test wiki or a service. The
library is implemented in JavaScript for node.js, using the [supertest](https://www.npmjs.com/package/supertest)
HTTP testing library, the [Chai](https://www.npmjs.com/package/chai) assertion
library, and the [Mocha](https://www.npmjs.com/package/mocha) testing framework.

## Documentation

See the [wiki page](https://www.mediawiki.org/wiki/MediaWiki_API_integration_tests) for
information about setting up a testing environment, running tests, and writing tests.

## Contributing

To open a patch request, see the guide to using [Gerrit](https://www.mediawiki.org/wiki/Gerrit) for Wikimedia projects.

To review open tasks or file a bug report, visit [Phabricator](https://phabricator.wikimedia.org/maniphest/?project=PHID-PROJ-tvow7mknofgk3l2okyyc&statuses=open()&group=none&order=newest#R).

## License

This project is licensed under the terms of the
[GNU General Public License, version 2 or later](LICENSE).
