# Installing `triton-kubernetes` CLI
Each [release on github](https://github.com/joyent/triton-kubernetes/releases) has associated binaries built which can be used to easily install `triton-kubernetes` CLI.

## Install on Linux
There are three packages available for Linux. An RPM, DEB and a standalone binary.

### Linux install using RPM package
Download the `triton-kubernetes` [rpm package](https://github.com/joyent/triton-kubernetes/releases).

From the same directory as where rpm package was downloaded, run the following command:
```bash
rpm -i triton-kubernetes_v0.9.0_linux-amd64.rpm
```
<sub>Replace `triton-kubernetes_v0.9.0_linux-amd64.rpm` with the package name that was downloaded.</sub>

### Linux install using DEB package
Download the `triton-kubernetes` [deb package](https://github.com/joyent/triton-kubernetes/releases).

From the same directory as where deb package was downloaded, run the following command:
```bash
dpkg -i triton-kubernetes_v0.9.0_linux-amd64.deb
```
<sub>Replace `triton-kubernetes_v0.9.0_linux-amd64.deb` with the package name that was downloaded.</sub>

### Linux install using standalone binary
Triton Multi-Cloud Kubernetes CLI has a standalone Linux binary available.
Download the `triton-kubernetes` [Linux binary](https://github.com/joyent/triton-kubernetes/releases).
Move the binary to `/usr/local/bin/` or somewhere in your `$PATH`.

## Install on OSX
Triton Multi-Cloud Kubernetes CLI has a standalone OSX binary available.
Download the `triton-kubernetes` [OSX binary](https://github.com/joyent/triton-kubernetes/releases).
Move the binary to `/usr/local/bin/` or somewhere in your `$PATH`.