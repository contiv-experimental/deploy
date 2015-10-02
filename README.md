
### Contiv Deploy

Deploy utilizes libcompose to launch applications but with ability to apply the policies needed for 
networking and storage.

#### How to try it out

1. Checkout the tree: 
git clone https://github.com/contiv/deploy.git

2. Compile and run unit tests to ensure you have correct environment
cd deploy
make 

3. Launch containers
```
$ cd example
$ deploy -file docker-compose.yml --labels="io.contiv.env:prod"
```

4. Use docker-compose to ps/stop/scale containers
```
$ docker-compose ps
$ docker-compose stop
$ docker-compose restart
```
