# scoutup
Simple Dev tool for running local Blockscout instances.

The primary goal of this tool is to provide a simple way to run a local explorer for development and testing purposes.

### Prerequisites
* docker
* supersim ([installation](https://supersim.pages.dev/getting-started/installation))
* anvil ([installation](https://book.getfoundry.sh/getting-started/installation))

### Build
```bash
$ git clone git@github.com:blockscout/scoutup.git
$ cd scoutup
$ go build
```

### When using Supersim
0. **Important:** For predeployed contracts to be indexed correctly, 
   ensure you have anvil (foundry) version which contains [e5ec47b](https://github.com/foundry-rs/foundry/commit/e5ec47b88208fdc48575359e0a5c44f85570ef63) commit.
   Currently, latest stable release does not contain the commit.  You can set it up by running either:
    ```
    foundryup --commit e5ec47b
    ```
   or (the better way) just install latest nightly release
   ```
   foundryup --install nightly
   ```

1. Run `supersim`:
   ```
   supersim --interop.autorelay true
   ```
2. Run local explorers for all `supersim` chains:
   ```
   ./scoutup --supersim
   ```
3. The command above will fetch the configuration of the supersim network via the supersim admin RPC, set up the necessary Blockscout instances accordingly, and run the local Blockscout instances using docker.
4. When the application starts, it displays a configuration log showing details for each chain in the terminal. The “Local” chain corresponds to the L1 environment, while “OPChainA” (and similarly OPChainB, OPChainC, etc.) corresponds to Optimism L2 environments. If you want to open the web interface for a particular chain, use the “Frontend” address listed under that chain. If you need direct access to the chain’s API endpoints, use the corresponding “Backend” address.    


### When using Anvil
1. Run local `anvil` node:
```
anvil --host 0.0.0.0
```
2. Run local explorer for anvil using `scoutup`:
```
./scoutup
```
3. The command above will run the local Blockscout instance with the default config using docker.


### Cleanup
`scoutup` attempts to stop and remove all running containers and delete all temporary files when stopping. However, depending on the termination process, some dangling containers and temporary files may remain. In such cases, it is recommended to run the following command to clean up:
```
./scoutup clean
```

### Troubleshooting
1. **Q:** When running with `l2.count > 2`, sending interoperable tokens results in the following error on the recipient node:
      
   `failed to fetch interop start: Insufficient funds for gas * price + value`
   
   **A:** There is an issue with older versions of anvil.
      Make sure that you have installed at least the "[nightly (2025-02-17)](https://github.com/foundry-rs/foundry/releases/tag/nightly-51b75c83045a963a48f9cd8d765e0f2e775acb3e)"