<p align="center">
  <a href="https://github.com/xylevy/BTCSteroids/">
    <img src="images/logo.png" alt="Logo" width=72 height=72>
  </a>

  <h3 align="center">BTCSteroids</h3>

  <p align="center">
    Bitcoin address balance checker on steroids.

  </p>
  
 [![Issues][issues-shield]][issues-url]
 [![MIT License][license-shield]][license-url]
 
  



## Table of contents

- [Quick start](#quick-start)
- [What's included](#whats-included)
- [Use Cases](#use-cases)
- [Thanks](#thanks)
- [Copyright and license](#copyright-and-license)


## Quick start

Building and running the Go executable.

- Make sure Go is installed, GOROOT & GOPATH path and all that good stuff is set.

- Use the ```go get``` command to install required packages. Syntax : ```go get package-import-path```

- From the project directory or pointing to ```main.go``` run ```go build```. Syntax ```go build -o /path/BTCSteroids/main_go.exe /path/BTCSteroids/main.go```

- Check the specified output path for the built binary

- To run the built file, just run it from the command console, in this case ```main_go.exe```  


## What's included



```text
BTCSteroids/
├─── main.go
├─── README.md
+--- electrum/
|    ├── misc.go
|	 ├── network.go
|    └── server.go
|
+--- operations/  
|    ├── string_in_slice.go
|    └── unique_in_slice.go
|
+--- services/
|	 └── address_to_scripthash.go
| 
|
└─── steroids/
	 ├── bitcoinchain_com.go
	 ├── blockchain_info.go
	 ├── blockcypher_com.go
	 ├── blockonomics_co.go
	 ├── bloom_flayer.go
	 ├── electrum_x.go
	 ├── hammer.go
	 ├── sample_addresses.go
	 ├── smartbit_com.go
	 └── workers.go
```
This project includes some workers that use APIs, go-electrum to check the given addresses against nodes and a local checker using bloom filters. The Blockonomics worker  needs an API Key. Add it/change in the file ```blockonomics_co.go```.
A thing to note is that you may need to update node addresses in ```electrum_x.go```


## Use cases
This is a project I picked up while learning Go Programming and was inspired by https://github.com/ashishbhate/hammer. Golang is fast and I have an interest in blockchain technologies.
- This program can be used to watch a list of addresses, check ```sample_addresses.go``` 
- You can a worker that checks against local DB/TSV/CSV file of addresses with balance using bloom filters. Check ```bloom_flayer.go```. To improve efficacy, we can further check the results with one of the other workers to weed out false-positives.
** As you  can tell by now, this worker is more geared more towards "bitcoin-cracking" (I mean, it is pointless and inefficient to use this to check balances or watch bitcoin wallets) ** 
- Both the above use cases can be chained with an address generator using worlists,random generation or whatever means, modifying the program so as to use stdin or reading the addresses from a file.

NOTE: The local implementation used to serve the database of addresses with balances using bloom filters has been removed from this version*/

### Attribution

<img src="https://media.flaticon.com/dist/min/img/logo/flaticon_negative.svg" width=25% height=25%>

Icons made by Freepik from [Flaticon](https://www.flaticon.com)

### Thanks
- https://github.com/ashishbhate/hammer
- https://github.com/Ismaestro/markdown-template
- https://github.com/checksum0/go-electrum
- https://gobyexample.com/waitgroups
- https://api.smartbit.com.au/v1/blockchain/address -> NOTE: *defunct*
- https://blockexplorer.com/api-ref
- https://www.blockchain.com/api/q
- https://www.blockcypher.com/dev/bitcoin/#rate-limits-and-tokens
- https://bitcoinchain.com/api#api_blockchain
- https://1209k.com/bitcoin-eye/ele.php -> NOTE: *for getting electrum node addresses*




## Copyright and license

Code and documentation copyright 2011-2022 the authors. Code released under the [MIT License](https://github.com/xylevy/BTCSteroids/blob/master/LICENSE).

[issues-shield]: https://img.shields.io/github/issues/xylevy/BTCSteroids.svg?style=for-the-badge
[issues-url]: https://github.com/xylevy/BTCSteroids/issues
[license-shield]: https://img.shields.io/github/license/xylevy/BTCSteroids.svg?style=for-the-badge
[license-url]: https://github.com/xylevy/BTCSteroids/blob/master/LICENSE
