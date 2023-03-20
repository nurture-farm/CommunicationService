# Prerequisite

### install librdkafka and pkg-config

brew install librdkafka
brew install pkg-config

### Set GOPRIVATE

You can run this command in terminal but it would be better if you set this in your
path variable in your bash profile.

export GOPRIVATE=code.nurture.farm

### Run temporal in your local

Temporal should run at 7233 port. If you are running at some other port. Please change the following 
config in communication_service.json

"temporal": {
    "worker": {
    "namespace": "default",
    "address": "localhost:7233"
    }
}

### Run in your local

chmod 777 run_local.sh
./run_local.sh

### Deployment in dev
cd /home/ubuntu/deployments/core/communication-service
kubectl apply -f deployment.yaml

#### Rollout restart
ssh -i ~/.ssh/afs-bastion.pem ubuntu@11.0.48.183
kubectl rollout restart deployment communication-service-deployment -n core

### Troubleshoot

You may need to set following variable:

LD_LIBRARY_PATH=/usr/local/opt/openssl/lib:"${LD_LIBRARY_PATH}"                    
CPATH=/usr/local/opt/openssl/include:"${CPATH}"                                    
PKG_CONFIG_PATH=/usr/local/opt/openssl/lib/pkgconfig:"${PKG_CONFIG_PATH}"          
export LD_LIBRARY_PATH CPATH PKG_CONFIG_PATH
CPPFLAGS=-I/usr/local/opt/openssl/include LDFLAGS=-L/usr/local/opt/openssl/lib 
PKG_CONFIG_PATH=/usr/local/opt/openssl/lib/pkgconfig







# CommunicationService
# CommunicationService
