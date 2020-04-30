'use strict';

module.exports = (baseDir = process.cwd()) => {
    const fs = require('fs');

    const configsDir = `${baseDir}/configs`;
    const configFileEnv = process.env.API_TESTING_CONFIG_FILE;
    const baseURLEnv = process.env.REST_BASE_URL;

    let requireFile = configFileEnv;

    if (baseURLEnv) {
        return {
            base_uri: baseURLEnv
        };
    } else if (requireFile) {
        if (!fs.existsSync(requireFile)) {
            // was it just the filename without the default config dir?
            requireFile = `${configsDir}/${configFileEnv}`;
            if (!fs.existsSync(requireFile)) {
                throw Error(`API_TESTING_CONFIG_FILE was set but neither '${configFileEnv}' nor '${requireFile}' exist.`);
            }
        }
    } else {
        // If .api-testing.config.json doesnt exist in root folder, throw helpful error
        const localConfigFile = '.api-testing.config.json';
        requireFile = `${baseDir}/${localConfigFile}`;
        if (!fs.existsSync(requireFile)) {
            throw Error(`Missing local config! Please create a ${localConfigFile} config`);
        }
    }

    return require(requireFile);
};
