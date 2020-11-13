package dltc;

import org.hyperledger.fabric.sdk.BlockEvent.TransactionEvent;
import org.hyperledger.fabric.sdk.BlockListener;
import org.hyperledger.fabric.sdk.ChaincodeID;
import org.hyperledger.fabric.sdk.Channel;
import org.hyperledger.fabric.sdk.Enrollment;
import org.hyperledger.fabric.sdk.EventHub;
import org.hyperledger.fabric.sdk.HFClient;
import org.hyperledger.fabric.sdk.Orderer;
import org.hyperledger.fabric.sdk.Peer;
import org.hyperledger.fabric.sdk.ProposalResponse;
import org.hyperledger.fabric.sdk.QueryByChaincodeRequest;
import org.hyperledger.fabric.sdk.TransactionProposalRequest;
import org.hyperledger.fabric.sdk.exception.InvalidArgumentException;
import org.hyperledger.fabric.sdk.exception.ProposalException;
import org.hyperledger.fabric.sdk.exception.TransactionException;
import org.hyperledger.fabric.sdk.security.CryptoSuite;
import org.hyperledger.fabric_ca.sdk.Attribute;
import org.hyperledger.fabric_ca.sdk.HFCAClient;
import org.hyperledger.fabric_ca.sdk.RegistrationRequest;
import org.json.JSONObject;

import java.io.IOException;
import java.math.BigInteger;
import java.security.KeyFactory;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;
import java.security.spec.X509EncodedKeySpec;
import java.util.Arrays;
import java.util.Base64;
import java.util.Collection;
import java.util.Iterator;
import java.util.Properties;
import java.util.Random;
import java.util.Vector;

public class FabricClient {
    private static final String channelName = "mychannel";
    private static final String chaincodeName = "votingcc";
    private static final String mspIdemix = "idemixMSPID1";

    private HFCAClient caClient;
    private HFClient client = null;
    private HFClient clientIdemix = null;

    public FabricClient(String url) {
        this.caClient = getHfCaClient("http://127.0.0.1:7054", null);
    }

    public String createVoting(AppUser appUser, JSONObject voting) throws Exception {
        return invokeBlockChain(appUser, false, "createVoting", voting.toString());
    }

    public String getVoting(AppUser appUser, String votingId) throws Exception {
        return queryBlockChain(appUser, false, "getVoting", votingId);
    }

    public String getResults(AppUser appUser, String votingId) throws Exception {
        return queryBlockChain(appUser, false, "getResults", votingId);
    }

    public void registerUserInVotingIdemix(AppUser appUser, JSONObject voting) throws Exception {

        BigInteger signedKey = getBlindSignature(appUser, voting);
        Storage.saveSignedKey(appUser.getName(), signedKey);

        invokeRegistration(appUser, voting.getString("id"), signedKey, Storage.getPubKey(appUser.getName()));
    }

    public String voteIdemix(AppUser appUser, JSONObject ballot) throws Exception {
        RSAPublicKey myPubKey = Storage.getPubKey(appUser.getName());
        RSAPrivateKey myPrivKey = Storage.getPrivKey(appUser.getName());

        byte[] myHash = getHashOfPubKey(myPubKey);
        BigInteger hashB = new BigInteger(1, myHash);
        BigInteger encrHash = hashB.modPow(myPrivKey.getPrivateExponent(), myPrivKey.getModulus());

        String encodedPubKey = Base64.getEncoder().encodeToString(myPubKey.getEncoded());
        ballot
            .put("key", encodedPubKey)
            .put("signedKey", encrHash.toString(10));

        return invokeBlockChain(appUser, true, "voteIdemix", ballot.toString());
    }

    private BigInteger getBlindSignature(AppUser appUser, JSONObject voting) throws Exception {
        RSAPublicKey myPubKey = Storage.getPubKey(appUser.getName());
        KeyFactory kf = KeyFactory.getInstance("RSA");
        X509EncodedKeySpec keySpecX509 = new X509EncodedKeySpec(Base64.getDecoder().decode(voting.getString("pubKey")));
        RSAPublicKey votingPubKey = (RSAPublicKey) kf.generatePublic(keySpecX509);

        BigInteger R = getRandom(8);
        BigInteger RInv = R.modInverse(votingPubKey.getModulus());
        String data = getDataForBlindSign(votingPubKey, myPubKey, R);

        JSONObject userRegData = new JSONObject()
                .put("userId", appUser.getName())
                .put("votingId", voting.getString("id"))
                .put("data", data);

        BigInteger signedData = new BigInteger(
                invokeBlockChain(appUser, false, "getBlindSign", userRegData.toString()));
        BigInteger signedKey = signedData.multiply(RInv);

        // check
        BigInteger myHashForCheck = signedKey.modPow(votingPubKey.getPublicExponent(), votingPubKey.getModulus());
        byte[] myHash = getHashOfPubKey(myPubKey);
        BigInteger hashB = new BigInteger(1, myHash);

        if (hashB.compareTo(myHashForCheck) != 0) {
            throw new Exception("Hashs are not the same. blind signature is invalid");
        }

        return signedKey;
    }

    private String invokeRegistration(AppUser appUser, String votingId, BigInteger signedKey, RSAPublicKey publicKey)
            throws Exception {

        String encodedPubKey = Base64.getEncoder().encodeToString(publicKey.getEncoded());
        JSONObject userTallingData = new JSONObject().put("votingId", votingId).put("key", encodedPubKey)
                .put("signedKey", signedKey.toString(10));

        return invokeBlockChain(appUser, true, "registerUserInVotingIdemix", userTallingData.toString());
    }

    private String getDataForBlindSign(RSAPublicKey pubKey, RSAPublicKey myPubKey, BigInteger R)
            throws IOException, NoSuchAlgorithmException {
        byte[] hash = getHashOfPubKey(myPubKey);
        BigInteger REncr = R.modPow(pubKey.getPublicExponent(), pubKey.getModulus());
        BigInteger hashedKey = new BigInteger(1, hash);
        BigInteger multNum = REncr.multiply(hashedKey);

        return multNum.toString(10);
    }

    private byte[] getHashOfPubKey(RSAPublicKey pubKey) throws IOException, NoSuchAlgorithmException {
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        return digest.digest(pubKey.getEncoded());
    }

    private BigInteger getRandom(int numBytes) {
        byte[] r = new byte[numBytes];
        new Random().nextBytes(r);
        return new BigInteger(1, r);
    }

    public HFCAClient getHfCaClient(String caUrl, Properties caClientProperties) {
        try {
            CryptoSuite cryptoSuite = CryptoSuite.Factory.getCryptoSuite();
            HFCAClient caClient = HFCAClient.createNewInstance(caUrl, caClientProperties);
            caClient.setCryptoSuite(cryptoSuite);
            return caClient;
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }

    public AppUser enrollUser(String userId, String org, String msp, String secret) throws
            Exception {
        Enrollment enrollment = caClient.enroll(userId, secret);
        AppUser appUser = new AppUser(userId, org, msp, enrollment);

        return appUser;
    }

    public String registerUser(AppUser registrar, String userId, String org, Boolean
            isAdmin, String secret)
            throws Exception {
        RegistrationRequest rr = new RegistrationRequest(userId, org);
        rr.setSecret(secret);

        return registerUser(registrar, rr, isAdmin);
    }

    private String registerUser(AppUser registrar, RegistrationRequest rr, boolean isAdmin) throws
            Exception {
        if (isAdmin) {
            Attribute attr = new Attribute("role", "2");
            rr.addAttribute(attr);
        }

        return caClient.register(rr, registrar);
    }

    private HFClient getHFClient(AppUser appUser, boolean isIdemix) throws Exception {
        if (isIdemix) {
            if (clientIdemix != null) {
                if (clientIdemix.getUserContext().getName().equals(appUser.getName())
                        && clientIdemix.getUserContext().getAffiliation().equals(appUser.getAffiliation())) {
                    return clientIdemix;
                }
            }
            clientIdemix = getClient(appUser, isIdemix);
            return clientIdemix;
        }

        if (client != null) {
            if (client.getUserContext().getEnrollment().equals(appUser.getEnrollment())) {
                return client;
            }
        }
        client = getClient(appUser, isIdemix);
        return client;
    }

    public HFClient getClient(AppUser appUser) throws Exception {
        HFClient client = getHfClient();
        client.setUserContext(appUser);
        getChannel(client);

        return client;
    }

    private Channel getChannel(HFClient client) throws InvalidArgumentException, TransactionException {
        Peer peer = client.newPeer("peer0.org1.sample.com", "grpc://127.0.0.1:57051");
        EventHub eventHub = client.newEventHub("eventhub01", "grpc://127.0.0.1:57053");
        Orderer orderer = client.newOrderer("orderer1.sample.com", "grpc://127.0.0.1:57050");
        Channel channel = client.newChannel(channelName);
        channel.addPeer(peer);
        channel.addEventHub(eventHub);
        channel.addOrderer(orderer);
        channel.initialize();

        return channel;
    }

    private HFClient getHfClient() throws Exception {
        CryptoSuite cryptoSuite = CryptoSuite.Factory.getCryptoSuite();
        HFClient client = HFClient.createNewInstance();
        client.setCryptoSuite(cryptoSuite);
        return client;
    }

    public HFClient getClient(AppUser appUser, boolean isIdemix) throws Exception {
        if (isIdemix) {
            AppUser newAppUser = new AppUser(appUser.getName(), appUser.getAffiliation(), appUser.getMspId(),
                    caClient.idemixEnroll(appUser.getEnrollment(), mspIdemix));
            return getClient(newAppUser);
        }

        return getClient(appUser);
    }

    private String invokeBlockChain(AppUser appUser, boolean isIdemix, String function, String... args)
            throws Exception, ProposalException, InvalidArgumentException {
        HFClient client = getHFClient(appUser, isIdemix);
        Channel channel = client.getChannel(channelName);

        TransactionProposalRequest tpr = client.newTransactionProposalRequest();
        ChaincodeID chaincodeID = ChaincodeID.newBuilder().setName(chaincodeName).build();
        tpr.setChaincodeID(chaincodeID);
        tpr.setFcn(function);
        tpr.setArgs(args);

        Collection<ProposalResponse> resps = channel.sendTransactionProposal(tpr);
        byte[] response = handlePorposalResponses(resps);

        Vector<String> txnIds = new Vector<>();
        String txnId = resps.iterator().next().getTransactionID();
        String blockEventListenerHandle = setBlockEventListener(channel, txnIds);

        channel.sendTransaction(resps);

        boolean eventDone = false;
        eventDone = waitForBlockEvent(150, channel, txnIds, blockEventListenerHandle, txnId);

        if (!eventDone) {
            return null;
        }

        return response == null ? null : new String(response);
    }

    private String queryBlockChain(AppUser appUser, boolean isIdemix, String function, String... args)
            throws Exception {
        HFClient client = getHFClient(appUser, isIdemix);
        Channel channel = client.getChannel(channelName);

        QueryByChaincodeRequest qpr = client.newQueryProposalRequest();
        ChaincodeID chaincodeID = ChaincodeID.newBuilder().setName(chaincodeName).build();
        qpr.setChaincodeID(chaincodeID);
        qpr.setFcn(function);
        qpr.setArgs(args);

        Collection<ProposalResponse> presps = channel.queryByChaincode(qpr);
        byte[] response = handlePorposalResponses(presps);

        return response == null ? null : new String(response);
    }

    private byte[] handlePorposalResponses(Collection<ProposalResponse> presps) throws Exception {
        byte[] response = null;

        for (ProposalResponse pres : presps) {
            if (pres.getStatus() != ProposalResponse.Status.SUCCESS) {
                throw new Exception(pres.getMessage());
            }

            if (response == null) {
                response = pres.getChaincodeActionResponsePayload();
                continue;
            }

            if (!Arrays.equals(response, pres.getChaincodeActionResponsePayload())) {
                throw new Exception("Proposal responses are not the same");
            }
        }

        return response;
    }

    private String setBlockEventListener(Channel channel, Vector<String> completedTxns)
            throws InvalidArgumentException {

        BlockListener blockListener = blockEvent -> {
            System.out.println("block data count = " + blockEvent.getBlock().getData().getDataCount());
            Iterator<TransactionEvent> iterator = blockEvent.getTransactionEvents().iterator();
            while (iterator.hasNext()) {
                TransactionEvent next = iterator.next();
                System.out.println("txn id = " + next.getTransactionID());
                completedTxns.add(next.getTransactionID());
            }
        };
        String eventListenerHandle = channel.registerBlockListener(blockListener);
        return eventListenerHandle;
    }

    private boolean waitForBlockEvent(Integer timeout, Channel channel, Vector<String> chaincodeEvents,
            String chaincodeEventListenerHandle, String txnId) throws InvalidArgumentException {
        boolean eventDone = false;
        if (chaincodeEventListenerHandle != null && txnId != null) {
            for (int i = 0; i < timeout; i++) {
                if (chaincodeEvents.contains(txnId)) {
                    eventDone = true;
                    break;
                } else {
                    try {
                        Thread.sleep(100);
                    } catch (InterruptedException e) {
                        e.printStackTrace();
                    }
                }
            }
        }

        channel.unregisterBlockListener(chaincodeEventListenerHandle);
        return eventDone;
    }
}