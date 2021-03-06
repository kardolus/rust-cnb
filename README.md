# Rust Cloud Native Buildpack
To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

You will need the [pack cli](https://github.com/buildpack/pack) in order to use this buildpack. You can either get the cli yourself or 
you can use the `install_tools` script from the scripts folder of this repo. 

For more info on Cloud Native Buildpacks have a look at https://buildpacks.io/.

## Configuration
You can use a buildpack.yml file to configure rust and rustup versions for your rust-app. Here's an example buildpack.yml
```
---
rustup:
  version: 1.16.0
rust:
  version: nightly

# Version can be: nightly-2016-06-03, nightly, 1.32.0, stable etc
```

## Quick test
- Pull the stacks
```
docker pull cfbuildpacks/cflinuxfs3-cnb-experimental:build
docker pull cfbuildpacks/cflinuxfs3-cnb-experimental:run
```
- Pull the cflinuxfs3 builder: `docker pull kardolus/fs3builder`
- Package the buildpack: `./scripts/package.sh`
- Get the pack-cli: `./scripts/install_tools.sh`
- Create an OCI image:
```
.bin/pack build rustapp --builder kardolus/fs3builder --buildpack </path/to/packaged/buildpack> -p integration/testdata/simple_app/ --no-pull
```
- Run the image: `docker run -it rustapp`
