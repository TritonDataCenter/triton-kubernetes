## Backup

`triton-kubernetes` is able to backup deployments into Manta, or any s3 compatible storage using minio.

Here is a demo of running backup against the [demo baremetal cluster](https://github.com/mesoform/triton-kubernetes/blob/master/docs/guide/bare-metal/cluster.md):
[![asciicast](https://asciinema.org/a/9O7U5UgUtaMZDnARV8KoDahsq.png)](https://asciinema.org/a/9O7U5UgUtaMZDnARV8KoDahsq)

`triton-kubernetes` cli can take a configuration file (yaml) with `--config` option to run in silent mode. To read about the yaml arguments, look at the [silent-install documentation](https://github.com/mesoform/triton-kubernetes/tree/master/docs/guide/silent-install-yaml.md).
