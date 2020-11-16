package dltc;

import org.hyperledger.fabric.sdk.NetworkConfig;
import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;

public class Main {
    private static String admin;
    private static String secretA;
    private static String user;
    private static NetworkConfig.OrgInfo orgInfo;

    public static void main(String[] args) {
        try {
            NetworkConfig networkConfig = NetworkConfig.fromYamlFile(new File("src/main/resources/network-config.yaml"));
            Main.orgInfo = networkConfig.getClientOrganization();
            NetworkConfig.UserInfo userInfo = orgInfo.getCertificateAuthorities().get(0).getRegistrars().iterator().next();
            Main.admin = userInfo.getName();
            Main.secretA = userInfo.getEnrollSecret();
            Main.user = orgInfo.getName() + "User";
            FabricClient fabricClient = new FabricClient(networkConfig);

            AppUser appUser = getAppUser(fabricClient);

            if (orgInfo.getName().equals("org1")) {
                fillLedger(fabricClient, appUser);
            } else {
                fabricClient.readBlocks(appUser);
            }


        } catch (Throwable ex) {
            ex.printStackTrace();
        }
    }

    private static void fillLedger(FabricClient fabricClient, AppUser appUser) throws Exception {
        String votingId = fabricClient.createVoting(appUser, getVoting());
        System.out.println("votingId: " + votingId);

        String voting = fabricClient.getVoting(appUser, votingId);
        System.out.println("voting: " + voting);

        fabricClient.registerUserInVotingIdemix(appUser, new JSONObject(voting));
        fabricClient.voteIdemix(appUser, getBallot(votingId));

        String results = fabricClient.getResults(appUser, votingId);
        System.out.println("results: " + results);
    }

    private static AppUser getAppUser(FabricClient fabricClient) throws Exception {
        AppUser appUser = null;
        if (!Storage.exist(admin)) {
            appUser = fabricClient.enrollUser(admin, orgInfo.getName(), orgInfo.getMspId(), secretA);
            Storage.save(appUser);
        }

        if (Storage.exist(user)) {
            return Storage.load(user);
        }

        AppUser registrar = Storage.load(admin);
        String secret = user;
        try {
            fabricClient.registerUser(registrar, user, orgInfo.getName(), false, secret);
            appUser = fabricClient.enrollUser(user, orgInfo.getName(), orgInfo.getMspId(), secret);
        } catch (Throwable ex) {
            appUser = fabricClient.enrollUser(user, orgInfo.getName(), orgInfo.getMspId(), secret);
        }

        Storage.save(appUser);

        return appUser;
    }

    private static JSONObject getVoting() throws Exception{
        return new JSONObject()
            .put("title", "test voting")
            .put("id", "testVOTEID7")
            .put("participants", new JSONArray()
                    .put("user1")
                    .put("user2")
                    .put(user))
            .put("questions", new JSONArray()
                    .put(new JSONObject()
                        .put("title", "question1")
                        .put("answers", new JSONArray()
                                .put(new JSONObject()
                                    .put("value","answers1")
                                    .put("value","answers2"))))
                    .put(new JSONObject()
                        .put("title", "question2")
                        .put("answers", new JSONArray()
                                .put(new JSONObject()
                                    .put("value","answers1")
                                    .put("value","answers2")))));
    }

    private static JSONObject getBallot(String votingId) throws Exception{
        return new JSONObject()
            .put("votingId", votingId)
            .put("questions", new JSONArray()
                    .put(new JSONObject()
                        .put("title", "question1")
                        .put("answers", new JSONArray()
                                .put(new JSONObject()
                                    .put("value","answers1"))))
                    .put(new JSONObject()
                        .put("title", "question2")
                        .put("answers", new JSONArray()
                                .put(new JSONObject()
                                    .put("value","answers2")))));
    }
}
