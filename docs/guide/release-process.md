## The Release Process

The release process involves two steps: publishing binaries to Github releases (using Travis CI) and adding the new release to the homebrew registry.

To start building and releasing, Travis CI must be configured for the triton-kubernetes repo. To let Travis CI upload Github releases, an encrypted Github API Key must be provided in the .travis.yml file.
1. Create a Github API Key and save the API Key locally (https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line)
2. Install the travis cli https://github.com/travis-ci/travis.rb#installation
3. Run travis encrypt <your-github-api-key-here>
4. Enter the encrypted text in the .travis.yml file at deploy.api-key.secure


### Creating a Release

To create a release, you just need to create and push a git tag with the proper name (e.g. v0.0.1). Once the tag is set, Travis CI will begin the release process.
1. Using the git cli, checkout the commit you would like to release
2. Run `git tag v0.0.0`, where `v0.0.0` is the desired version number
3. Run `git push origin v0.0.0` 
4. TravisCI will begin building the binaries using the commit that was tagged.
4. After Travis CI is done, check the Github releases page to verify that the binaries have been uploaded.

### Creating a Release Manually

To build the binaries locally, you will need to install the following on your machine:
* rpmbuild (For OS X, you can run `brew install rpm`)
* [fpm](https://github.com/jordansissel/fpm)

Then run `make build VERSION=0.0.0` where 0.0.0 is replaced with the desired version number.

### Updating the homebrew registry

The github repository at: https://github.com/Homebrew/homebrew-core serves as the homebrew registry. To submit a pull request to that repo:
1. Fork the homebrew-core repository
2. In the Formula folder, add a brew formula for triton-kubernetes (triton-kubernetes.rb). Requires:
    * SHA256 checksum of the .tar.gz file from the Github release (use shasum command)
    * new version number
    * See example at https://github.com/cg50x/homebrew-test/blob/master/Formula/triton-kubernetes.rb
3. Submit a pull request to the homebrew-core repository. Once it is merged, you can verify that it works by running `brew update` and then `brew install triton-kubernetes`.