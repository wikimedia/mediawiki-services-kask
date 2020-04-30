const { assert, utils } = require('../index');
const fs = require('fs');
const os = require('os');
const fsp = fs.promises;

const testRootDir = `${os.tmpdir()}/${utils.uniq()}`;
const testConfigsDir = `${testRootDir}/configs`;
const testConfigFiles = [
    [`${testConfigsDir}/quibble.json`, `{ "file": "${testConfigsDir}/quibble.json" }`],
    [`${testConfigsDir}/example.json`, `{ "file": "${testConfigsDir}/example.json" }`],
    [`${testRootDir}/.api-testing.config.json`, `{ "file": "${testRootDir}/.api-testing.config.json" }`]
];

// Setup our test configs in the temp directory
const createTestConfigs = async () => {
    await fsp.mkdir(testRootDir);
    await fsp.mkdir(testConfigsDir);
    const fileWritePromises = testConfigFiles.map(
        (fileInfo) => fsp.writeFile(fileInfo[0], fileInfo[1])
    );
    await Promise.all(fileWritePromises);
};

// Setup our test configs in the temp directory
const deleteTestConfigs = async () => {
    // NOTE: rmdir does not support recursion in node 11 and earlier.
    const filesInConfigDir = await fsp.readdir(testConfigsDir, { withFileTypes: true });
    await Promise.all(filesInConfigDir.map((dirent) => fsp.unlink(`${testConfigsDir}/${dirent.name}`)));

    const filesInRootDir = await fsp.readdir(testRootDir, { withFileTypes: true });
    await Promise.all(filesInRootDir.map(
        (dirent) => dirent.isDirectory() ?
            fsp.rmdir(`${testRootDir}/${dirent.name}`) :
            fsp.unlink(`${testRootDir}/${dirent.name}`)
    ));

    // await fsp.rmdir(testConfigsDir);
    await fsp.rmdir(testRootDir);
};

describe('Configuration', () => {
    let envVar;
    const getConfig = require('../lib/config');

    before(async () => {
        // Save the env var for other tests
        envVar = process.env.API_TESTING_CONFIG_FILE;
        delete process.env.API_TESTING_CONFIG_FILE;
        await createTestConfigs();
    });

    after(async () => {
        await deleteTestConfigs();

        if (envVar) {
            process.env.API_TESTING_CONFIG_FILE = envVar;
        }
    });

    describe(`Using ${testRootDir} as the configuration root folder`, () => {
        it('Use .api-testing.config.json file if API_TESTING_CONFIG_FILE does not exist', () => {
            delete process.env.API_TESTING_CONFIG_FILE;
            assert.deepEqual(getConfig(testRootDir), { file: `${testRootDir}/.api-testing.config.json` });
        });

        it('Select full path config set in API_TESTING_CONFIG_FILE env variable over local config', () => {
            process.env.API_TESTING_CONFIG_FILE = `${testConfigsDir}/quibble.json`;
            assert.deepEqual(getConfig(testRootDir), { file: `${testConfigsDir}/quibble.json` });
            delete process.env.API_TESTING_CONFIG_FILE;
        });

        it('Throw exception if config file set in API_TESTING_CONFIG_FILE does not exist', () => {
            process.env.API_TESTING_CONFIG_FILE = 'idonotexist.json';
            assert.throws(() => getConfig(testRootDir), Error, /API_TESTING_CONFIG_FILE was set but neither/);
            delete process.env.API_TESTING_CONFIG_FILE;
        });

        describe('Renaming required root folder config ".api-testing.config.json"', () => {
            it('Throws exception if ".api-testing.config.json" doesnt exist and API_TESTING_CONFIG_FILE is not set', () => {
                delete process.env.API_TESTING_CONFIG_FILE;
                fs.rename(`${testRootDir}/.api-testing.config.json`, `${testRootDir}/wrong.json`, (err) => {
                    assert.throws(() => getConfig(testRootDir), Error, /Missing local config!/);
                });
            });
        });
    });

    describe('Using REST_BASE_URL for configuration', () => {
        it('should return a json when REST_BASE_URL is set', () => {
            process.env.REST_BASE_URL = 'http://localhost:8081/';

            assert.deepEqual(getConfig(), { base_uri: process.env.REST_BASE_URL });
            delete process.env.REST_BASE_URL;
        });
    });
});
