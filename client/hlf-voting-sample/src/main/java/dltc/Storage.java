package dltc;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.io.Serializable;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.Key;
import java.security.KeyFactory;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.SecureRandom;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.PKCS8EncodedKeySpec;
import java.security.spec.X509EncodedKeySpec;

public class Storage {
    private static String usersDir = "users";
    private static String keysDir = "keys";
    private static String workDir = System.getProperty("user.dir");

    private static String getUsersDir() {
        return usersDir;
    }

    private static String getUserDir(String msp, String userName) {
        return getWorkDir() + File.separator + getUsersDir() + File.separator + msp + File.separator + userName;
    }

    private static String getUserKeysDir(String userName) {
        return getWorkDir() + File.separator + getUsersDir() + File.separator + userName + File.separator
                + getKeysDir();
    }

    public static AppUser load(FileInputStream fileInputStream) throws IOException, ClassNotFoundException {
        ObjectInputStream decoder = new ObjectInputStream(fileInputStream);
        return (AppUser) decoder.readObject();
    }

    public static void save(AppUser appUser) throws IOException {
        File dir = new File(getWorkDir());
        if (!dir.exists()) {
            boolean created = dir.mkdirs();
            if (!created) {
                throw new IOException("cannot create directory");
            }
        }

        File file = new File(getUserDir(appUser.getMspId(), appUser.getName()));
        File usersDir = new File(getUsersDir());
        if (!usersDir.exists()) {
            usersDir.mkdirs();
        }
        if (!file.exists()) {
            file.mkdirs();
        }
        file = new File(file.getAbsoluteFile() + "/" + appUser.getName());
        if (!file.exists()) {
            boolean created = file.createNewFile();
            if (!created) {
                throw new IOException("File cannot be created");
            }
        }

        save(appUser, file);
    }

    public static AppUser load(String mspId, String userName) throws IOException, ClassNotFoundException {
        File file = new File(getUserDir(mspId, userName) + File.separator + userName);
        FileInputStream inputStream = new FileInputStream(file);
        return load(inputStream);
    }

    public static boolean exist(String mspId, String userName) {
        File file = new File(getUserDir(mspId, userName) + File.separator + userName);
        return file.exists();
    }

    public static KeyPair generateKeyPair(int bits) throws Exception {
        KeyPairGenerator generator = KeyPairGenerator.getInstance("RSA");
        generator.initialize(bits, new SecureRandom());
        return generator.generateKeyPair();
    }

    public static RSAPublicKey getPubKey(String userName) throws Exception {
        String filePath = getUserKeysDir(userName);
        File dir = new File(filePath);
        if (!dir.exists()) {
            boolean created = dir.mkdirs();
            System.out.print(created);
        }
        String fullName = filePath + "/" + userName;
        File file = new File(fullName + ".pub");
        if (!file.isFile()) {
            file.createNewFile();
            KeyPair pair = generateKeyPair(512);
            SaveKeys(pair, fullName);
        }

        return loadPubKey(file.getAbsolutePath());
    }

    public static RSAPrivateKey getPrivKey(String userName) throws Exception {
        String filesPath = getUserKeysDir(userName) + "/" + userName;
        File file = new File(filesPath + ".key");
        if (!file.isFile()) {
            KeyPair pair = generateKeyPair(1024);
            SaveKeys(pair, filesPath);
        }

        return loadPrivKey(file.getAbsolutePath());
    }

    public static void SaveKeys(KeyPair pair, String outFile) throws IOException {
        RSAPublicKey pubKey = (RSAPublicKey) pair.getPublic();
        RSAPrivateKey privKey = (RSAPrivateKey) pair.getPrivate();

        FileOutputStream out = new FileOutputStream(outFile + ".key");
        out.write(privKey.getEncoded());
        out.close();

        out = new FileOutputStream(outFile + ".pub");
        out.write(pubKey.getEncoded());
        out.close();
    }

    public static void saveSignedKey(String userName, BigInteger signedKey) throws IOException {
        String filesPath = getUserKeysDir(userName) + "/" + userName + ".signed";

        File file = new File(filesPath);
        if (!file.isFile()) {
            boolean created = file.createNewFile();
            if (!created) {
                throw new IOException("cannot create file");
            }
        }

        save(signedKey, file);
    }

    public static void save(Serializable obj, File file) throws IOException {
        FileOutputStream outputStream = new FileOutputStream(file);
        ObjectOutputStream objectOut = new ObjectOutputStream(outputStream);
        objectOut.writeObject(obj);
        objectOut.close();
    }

    public static RSAPublicKey loadPubKey(String filePath)
            throws IOException, NoSuchAlgorithmException, InvalidKeySpecException {
        Path path = Paths.get(filePath);
        byte[] bytes = Files.readAllBytes(path);

        X509EncodedKeySpec ks = new X509EncodedKeySpec(bytes);
        KeyFactory kf = KeyFactory.getInstance("RSA");
        PublicKey pub = kf.generatePublic(ks);

        return (RSAPublicKey) pub;
    }

    public static RSAPrivateKey loadPrivKey(String filePath)
            throws IOException, NoSuchAlgorithmException, InvalidKeySpecException {
        Path path = Paths.get(filePath);
        byte[] bytes = Files.readAllBytes(path);

        PKCS8EncodedKeySpec ks = new PKCS8EncodedKeySpec(bytes);
        KeyFactory kf = KeyFactory.getInstance("RSA");
        PrivateKey pvt = kf.generatePrivate(ks);

        return (RSAPrivateKey) pvt;
    }

    public static void clearKeys(String userName) {
        File dir = new File(getUserKeysDir(userName));
        for (File file : dir.listFiles()) {
            file.delete();
        }
    }

    public static void setUsersDir(String usersDir) {
        Storage.usersDir = usersDir;

    }

    public static String getKeysDir() {
        return keysDir;
    }

    public static void setKeysDir(String keysDir) {
        Storage.keysDir = keysDir;
    }

    public static String getWorkDir() {
        return workDir;
    }

    public static void setWorkDir(String workDir) {
        Storage.workDir = workDir;
    }
}
