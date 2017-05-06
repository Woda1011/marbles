
##### Shutdown
docker-compose -f four-peer-ca.yaml down

##### Start
docker-compose -f four-peer-ca.yaml up -d

##### CLI for VP01
docker exec -it vp0_1 bash

##### Docker Status
docker ps

##### Clean up Docker 
docker rm -f $(docker ps -aq)
docker rmi $(docker images -q)


##### Git Clone
git clone https://github.com/Woda1011/marbles.git


##### Chaincode deployment hash
app.js
Zeile 197

##### four peer credentials
"enrollId": "bob", "enrollSecret": "NOE63pEQbL25"


######
ssh blockchain@blockchain-12b7p377.cloudapp.net