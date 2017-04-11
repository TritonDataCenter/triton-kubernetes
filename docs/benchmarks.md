# Benchmarking with [simple-container-benchmarks](https://github.com/misterbisson/simple-container-benchmarks)

**System performance test of Triton KVM and AWS VM using [simple-container-benchmarks](https://github.com/misterbisson/simple-container-benchmarks):**  
This test was between a general purpose VM AWS and a comparable Triton KVM. Triton KVM package used for this test is k4-general-kvm-31.75G, and the AWS VM t2.2xlarge.

For all tests, ubuntu-16.04 image was used on both the Triton KVMs and AWS VMs.  

To get write performance, it pipes a gigabyte of zeros to a file on the filesystem:  
/disk request : 1073741824 bytes (1.1 GB)

To test CPU performance, it fetches random numbers and md5 hashes them:  
/cpu request : 268435456 bytes (268 MB)


For more detailed information on what the benchmarking does, click [here](https://github.com/misterbisson/simple-container-benchmarks#how-the-tests-work).  

## k4-general-kvm-31.75G vs t2.2xlarge

|                       	| Triton                    	| AWS                    	|
|-----------------------	|---------------------------	|------------------------	|
| /disk request average 	| 8.470108 s, 128.8 MB/s     	| 13.02394 s, 83.45 MB/s 	|
| /cpu request average  	| 16.85761 s, 15.96 MB/s     	| 21.63477 s, 12.4 MB/s 	|
| total mem             	| 32689108                   	| 32946296                	|
| used mem              	| 1593180                    	| 1595780                 	|
| free mem              	| 31095928                   	| 31350516                	|
| shared mem            	| 9016                      	| 8944                   	|
| buffers mem           	| 153116                    	| 153028                 	|
| cached mem            	| 951076                    	| 949996                 	|
| Architecture          	| x86_64                       	| x86_64                	|
| CPU op-mode(s)         	| 32-bit, 64-bit               	| 32-bit, 64-bit         	|
| Byte Order             	| Little Endian               	| Little Endian            	|
| CPU(s)                 	| 8                           	| 8                     	|
| On-line CPU(s) list   	| 0-7                         	| 0-7                      	|
| Thread(s) per core    	| 1                         	| 1                     	|
| Core(s) per socket    	| 1                         	| 8                      	|
| Socket(s)             	| 8                         	| 1                     	|
| NUMA node(s)          	| 1                            	| 1                     	|
| Vendor ID              	| GenuineIntel                 	| GenuineIntel          	|
| CPU family             	| 6                         	| 6                        	|
| Model                  	| 45                         	| 63                       	|
| Stepping               	| 7                          	| 2                       	|
| CPU MHz                	| 2600.128                    	| 2400.046              	|
| BogoMIPS               	| 5200.25                      	| 4800.09                  	|
| Hypervisor vendor      	| -                           	| Xen                    	|
| Virtualization type    	| -                         	| full                    	|
| L1d cache             	| 32K                       	| 32K                    	|
| L1i cache             	| 32K                          	| 32K                    	|
| L2 cache              	| 4096K                     	| 256K                   	|
| L3 cache              	| -                         	| 30720K                   	|
| NUMA node0 CPU(s)      	| 0-7                          	| 0-7                   	|
