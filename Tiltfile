#!/usr/bin/env python

# get some color in earthly outputs
os.environ['FORCE_COLOR'] = '1'

config.define_bool("no-volumes")
cfg = config.parse()

load('ext://namespace', 'namespace_create')
namespace_create('orderer')
namespace_create('org1')
namespace_create('org2')

clk_k8s = 'clk -a --force-color k8s -c ' + k8s_context() + ' '
kc_secret = 'kubectl create secret --dry-run=client -o yaml '

dk_run = 'docker run --rm -u $(id -u):$(id -g) -v $PWD/config:/config hyperledger/fabric-tools:2.3 '
if not os.path.exists('./config/generated/crypto-config'):
    local(dk_run + ' cryptogen generate --config=/config/crypto-config.yaml --output=/config/generated/crypto-config')
if not os.path.exists('./config/generated/genesis.block'):
    local(dk_run + 'env FABRIC_CFG_PATH=/config configtxgen -profile TwoOrgsOrdererGenesis -channelID system-channel -outputBlock /config/generated/genesis.block')
if not os.path.exists('./config/generated/star.tx'):
    local(dk_run + 'env FABRIC_CFG_PATH=/config configtxgen -profile TwoOrgsChannel -outputCreateChannelTx /config/generated/star.tx -channelID star')

for org in ['org1', 'org2']:
    local(dk_run + 'env FABRIC_CFG_PATH=/config configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate /config/generated/anchor-star-' + org + '.tx -channelID star -asOrg ' + org)

#### orderers ####

k8s_yaml(local(kc_secret + '-n orderer generic hlf--genesis --from-file=./config/generated/genesis.block', quiet=True))
k8s_yaml(local(kc_secret + '-n orderer generic hlf--ord-admincert --from-file=cert.pem=./config/generated/crypto-config/ordererOrganizations/orderer/users/Admin@orderer/msp/signcerts/Admin@orderer-cert.pem', quiet=True))
for orderer in ['orderer1', 'orderer2', 'orderer3']:
    # create secrets
    k8s_yaml(local(kc_secret + '-n orderer generic hlf--' + orderer + '-idcert --from-file=./config/generated/crypto-config/ordererOrganizations/orderer/orderers/' + orderer + '.orderer/msp/signcerts/' + orderer + '.orderer-cert.pem', quiet=True))
    k8s_yaml(local(kc_secret + '-n orderer generic hlf--' + orderer + '-idkey --from-file=./config/generated/crypto-config/ordererOrganizations/orderer/orderers/' + orderer + '.orderer/msp/keystore/priv_sk', quiet=True))
    k8s_yaml(local(kc_secret + '-n orderer generic hlf--' + orderer + '-cacert --from-file=./config/generated/crypto-config/ordererOrganizations/orderer/orderers/' + orderer + '.orderer/msp/cacerts/ca.orderer-cert.pem', quiet=True))
    k8s_yaml(local(kc_secret + '-n orderer tls hlf--' + orderer + '-tls --key=./config/generated/crypto-config/ordererOrganizations/orderer/orderers/' + orderer + '.orderer/tls/server.key --cert=./config/generated/crypto-config/ordererOrganizations/orderer/orderers/' + orderer + '.orderer/tls/server.crt', quiet=True))
    k8s_yaml(local(kc_secret + '-n orderer generic hlf--' + orderer + '-tlsrootcert --from-file=cacert.pem=./config/generated/crypto-config/ordererOrganizations/orderer/orderers/' + orderer + '.orderer/tls/ca.crt', quiet=True))

    k8s_yaml(
        helm(
            'helm/hlf-ord',
            namespace='orderer',
            name=orderer,
            values=['helm/hlf-ord/values-' + orderer + '.yaml'],
        )
    )
    k8s_resource(orderer + '-hlf-ord', labels=['orderer'])
    if config.tilt_subcommand == 'up':
        local(clk_k8s + 'add-domain ' + orderer + '.orderer.localhost')
    if config.tilt_subcommand == 'down' and not cfg.get("no-volumes"):
        local('kubectl --context ' + k8s_context() + ' -n orderer delete pvc --selector=app.kubernetes.io/instance=' + orderer + ' --wait=false')


#### peers ####

custom_build(
    "eniblock/hlf-helper",
    'earthly ./helper+docker --ref=$EXPECTED_REF',
    ["./helper"],
)
custom_build(
    "eniblock/hlf-ccid",
    'earthly ./ccid+docker --ref=$EXPECTED_REF',
    ["./ccid"],
)

for org in ['org1', 'org2']:
    k8s_yaml(local(kc_secret + '-n ' + org + ' generic starchannel --from-file=./config/generated/star.tx --from-file=./config/generated/anchor-star-' + org + '.tx', quiet=True))
    k8s_yaml(local(kc_secret + '-n ' + org + ' generic hlf--ord-tlsrootcert --from-file=cacert.pem=./config/generated/crypto-config/ordererOrganizations/orderer/orderers/orderer1.orderer/tls/ca.crt', quiet=True))
    for peer in ['peer1', 'peer2']:
        k8s_yaml(local(kc_secret + '-n ' + org + ' generic hlf--' + peer + '-idcert --from-file=./config/generated/crypto-config/peerOrganizations/' + org + '/peers/' + peer + '.' + org + '/msp/signcerts/' + peer + '.' + org + '-cert.pem', quiet=True))
        k8s_yaml(local(kc_secret + '-n ' + org + ' generic hlf--' + peer + '-idkey --from-file=./config/generated/crypto-config/peerOrganizations/' + org + '/peers/' + peer + '.' + org + '/msp/keystore/priv_sk', quiet=True))
        k8s_yaml(local(kc_secret + '-n ' + org + ' generic hlf--' + peer + '-cacert --from-file=./config/generated/crypto-config/peerOrganizations/' + org + '/peers/' + peer + '.' + org + '/msp/cacerts/ca.' + org + '-cert.pem', quiet=True))

        k8s_yaml(local(kc_secret + '-n ' + org + ' tls hlf--' + peer + '-tls --key=./config/generated/crypto-config/peerOrganizations/' + org + '/peers/' + peer + '.' + org + '/tls/server.key --cert=./config/generated/crypto-config/peerOrganizations/' + org + '/peers/' + peer + '.' + org + '/tls/server.crt', quiet=True))
        k8s_yaml(local(kc_secret + '-n ' + org + ' generic hlf--' + peer + '-tlsrootcert --from-file=cacert.pem=./config/generated/crypto-config/peerOrganizations/' + org + '/peers/' + peer + '.' + org + '/tls/ca.crt', quiet=True))

        k8s_yaml(local(kc_secret + '-n ' + org + ' tls hlf--' + peer + '-tls-client --key=./config/generated/crypto-config/peerOrganizations/' + org + '/users/Admin@' + org + '/tls/client.key --cert=./config/generated/crypto-config/peerOrganizations/' + org + '/users/Admin@' + org + '/tls/client.crt', quiet=True))
        k8s_yaml(local(kc_secret + '-n ' + org + ' generic hlf--' + peer + '-client-tlsrootcert --from-file=./config/generated/crypto-config/peerOrganizations/' + org + '/users/Admin@' + org + '/tls/ca.crt', quiet=True))

        k8s_yaml(local(kc_secret + '-n ' + org + ' generic hlf--' + peer + '-admincert --from-file=cert.pem=./config/generated/crypto-config/peerOrganizations/' + org + '/users/Admin@' + org + '/msp/signcerts/Admin@' + org + '-cert.pem', quiet=True))
        k8s_yaml(local(kc_secret + '-n ' + org + ' generic hlf--' + peer + '-adminkey --from-file=cert.pem=./config/generated/crypto-config/peerOrganizations/' + org + '/users/Admin@' + org + '/msp/keystore/priv_sk', quiet=True))

        k8s_yaml(
            helm(
                'helm/hlf-peer',
                namespace=org,
                name=peer,
                values=['helm/hlf-peer/values-' + org + '-' + peer + '.yaml'],
            )
        )
        k8s_resource(peer + '-hlf-peer:statefulset:' + org, labels=[org])
        k8s_resource(peer + '-hlf-peer-couchdb:statefulset:' + org, labels=[org])
        k8s_resource(peer + '-hlf-peer-fabcar:deployment:' + org, labels=[org])
        k8s_resource(peer + '-hlf-peer-jc-star:job:' + org, labels=[org])
        k8s_resource(peer + '-hlf-peer-rc-fabcar:job:' + org, labels=[org])
        if config.tilt_subcommand == 'up':
            local(clk_k8s + 'add-domain ' + peer + '.' + org + '.localhost')
        if config.tilt_subcommand == 'down' and not cfg.get("no-volumes"):
            local('kubectl --context ' + k8s_context() + ' -n ' + org + ' delete pvc --selector=app.kubernetes.io/instance=' + peer + ' --wait=false')

custom_build(
    "eniblock/hlf-fabcar",
    'earthly ./fabcar+docker --ref=$EXPECTED_REF',
    ["./fabcar"],
)

#### lint ####

local_resource('ord lint', 'earthly ./helm/hlf-ord+lint', 'hlf-ord/', allow_parallel=True)
local_resource('peer lint', 'earthly ./helm/hlf-peer+lint', 'hlf-peer/', allow_parallel=True)
