This guide outlines how to deploy the explorer using a local qrysm-gzond testnet. Utilized postgres, redis and little_bigtable as data storage

# Install docker
If you never worked with Docker, [this short video](https://www.youtube.com/watch?v=rOTqprHv1YE) gives an overview to understand roughly what we will do with it.

Now, let us install it:
```
sudo apt update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
```

# Install kurtosis-cli
Kurtosis is a software which will launch the different parts of a test network and the explorer, all running locally, using Docker. You will not have to deal with it (nor with Docker), because automating the launch of interdependent modules with Docker and configuring them is the point of Kurtosis. [This short video](https://www.loom.com/share/4256e2b84e5840d3a0a941a80037aebe) gives an overview if it is your first time.

Now, let us install it:
```
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt install kurtosis-cli
```

# Install golang
You will find the last version of Go [on this page](https://go.dev/doc/install). The commands that you will type to install it will look like this:

```
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
```
Add the golang binaries to the path by adding the following lines to your _~/.profile_ file and then logout & login again.
```
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin
```
The second line is not mentionned in the installation instructions of Go's website but will be necessary for our system.
Before continuing, restarting your computer now might save you from unexplained errors during the next steps.

# Clone the explorer repository
```
cd ~
git clone https://github.com/theQRL/qrl-beaconchain-explorer.git
cd qrl-beaconchain-explorer
```

# Build the explorer binaries
```
sudo apt install build-essential
make all
```

# Start postgres, redis, little_bigtable & the qrl test network
```
cd ~/qrl-beaconchain-explorer/local-deployment/
kurtosis clean -a && kurtosis run --enclave my-testnet . --args-file network_params.yaml
```
Later in your developer life (after having started Kurtosis and stopped it a few times), if you encounter an error at this step, you might need to clean up bugged cache files from previous runs that Kurtosis or Docker left behind.
The `./stop` script [in this repository](https://github.com/thib-wien/scripts-localnetworkandexplorer) gathers cleaning commands which worked for their author (it might save you hours of browsing Stack Overflow and GitHub's issues).

# Generate the explorer config file for the deployed testnet
```
cd ~/qrl-beaconchain-explorer/local-deployment/
bash provision-explorer-config.sh
```
This will generate a _config.yml_ to be used by the explorer and then create the bigtable & postgres schema.

# Start the explorer modules
```
cd ~/qrl-beaconchain-explorer/local-deployment/
docker compose up -d
```
You can start / stop the exporter submodules using `docker compose`

# Convenience-script run.sh
Above, we have started / stopped the local chain + the explorer manually. The `run.sh` script in this directory can be used to start and stop everything automatically. Just run `./run.sh start` to start the whole system, wait a bit and browse http://localhost:8080 to see it in action. You can run `./run.sh sql` to explore the sql-database. Everything can be stopped and cleaned up with `./run.sh stop`.

# Exit validators
Exiting individual validators can be done using the provided `exit_validator.sh` script. Requires [https://github.com/wealdtech/ethdo](ethdo) to be available on the path.
```
bash exit_validators.sh -i validator_index -m "memonic" -b "http://bn_api_host:bn_api_port"
```