package dltc;

import org.json.JSONArray;
import org.json.JSONObject;

public class Main {
    private static final String url = "127.0.0.1";
    private static final String admin = "admin";
    private static final String secretA = "adminpw";
    private static final String user = "user1";
    private static final String org = "org1";
    private static final String msp = "Org1MSP";

    public static void main(String[] args) {
        try {
            FabricClient fabricClient = new FabricClient(url);

            AppUser appUser = getAppUser(fabricClient);

            String votingId = "d3ee6cc5-1e0f-49a2-a419-2a52b16d204a";// fabricClient.createVoting(appUser, getVoting());
            System.out.println("votingId: " + votingId);
            
            String voting = fabricClient.getVoting(appUser, votingId);
            System.out.println("voting: " + voting);

            fabricClient.registerUserInVotingIdemix(appUser, new JSONObject(voting));
            fabricClient.voteIdemix(appUser, getBallot(votingId));
            
            String results = fabricClient.getResults(appUser, votingId);
            System.out.println("results: " + results);

        } catch (Throwable ex) {
            ex.printStackTrace();
        }
    }

    private static AppUser getAppUser(FabricClient fabricClient) throws Exception {
        AppUser appUser = null;
        if (!Storage.exist(admin)) {
            appUser = fabricClient.enrollUser(admin, org, msp, secretA);
            Storage.save(appUser);
        }

        if (Storage.exist(user)) {
            return Storage.load(user);
        }

        AppUser registrar = Storage.load(admin);
        String secret = user;
        try {
            fabricClient.registerUser(registrar, user, org, false, secret);
            appUser = fabricClient.enrollUser(user, org, msp, secret);
        } catch (Throwable ex) {
            appUser = fabricClient.enrollUser(user, org, msp, secret);
        }

        Storage.save(appUser);

        return appUser;
    }

    private static JSONObject getVoting() throws Exception{
        return new JSONObject()
            .put("title", "test voting")
            .put("participants", new JSONArray()
                    .put("user1")
                    .put("user2")
                    .put("user3"))
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
