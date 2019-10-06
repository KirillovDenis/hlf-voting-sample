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
        return getWorkDir() + File.separator + usersDir;
    }

    private static String getUserDir(String userName) {
        return getWorkDir() + File.separator + getUsersDir() + File.separator + userName;
    }

    private static String getUserKeysDir(String userName) {
        return getWorkDir() + File.separator + getUsersDir() + File.separator + userName + File.separator + getKeysDir();
    }

    static void serialize(AppUser appUser, String userName) throws IOException {

        String filePath = getUserDir(appUser.getName());
        File dir = new File(filePath);
        if (!dir.exists()) {
            dir.mkdirs();
        }

        try (ObjectOutputStream oos = new ObjectOutputStream(
                Files.newOutputStream(Paths.get(filePath + "/" + userName + ".jso")))) {
            oos.writeObject(appUser);
        }
    }

    static AppUser tryDeserialize(String name, String enrName) throws Exception {
        if (Files.exists(Paths.get(getWorkDir() + "/" + getUsersDir() + "/" + name + "/" + enrName + ".jso"))) {
            return deserialize(name, enrName);
        }
        return null;
    }

    static AppUser deserialize(String name, String enrName) throws Exception {
        try (ObjectInputStream decoder = new ObjectInputStream(Files.newInputStream(Paths.get(getUserDir(name) + "/" + enrName + ".jso")))) {
            return (AppUser) decoder.readObject();
        }
    }

    public static void save(AppUser appUser, FileOutputStream outputStream) throws IOException {
        ObjectOutputStream objectOut = new ObjectOutputStream(outputStream);
        objectOut.writeObject(appUser);
        objectOut.close();
    }

    public static AppUser load(FileInputStream fileInputStream) throws IOException, ClassNotFoundException {
        ObjectInputStream decoder = new ObjectInputStream(fileInputStream);
        return (AppUser) decoder.readObject();
    }

    public static Object load(File file) throws IOException, ClassNotFoundException {
        FileInputStream fileInputStream = new FileInputStream(file);
        ObjectInputStream decoder = new ObjectInputStream(fileInputStream);
        Object obj = decoder.readObject();
        decoder.close();
        return obj;
    }

    public static Object loadObj(FileInputStream inputStream) throws IOException, ClassNotFoundException {
        ObjectInputStream decoder = new ObjectInputStream(inputStream);
        return decoder.readObject();
    }

    public static void save(AppUser appUser) throws IOException {
        File dir = new File(getWorkDir());
        if (!dir.exists()) {
            boolean created = dir.mkdirs();
            if (!created) {
                throw new IOException("cannot create directory");
            }
        }

        File file = new File(getUserDir(appUser.getName()));
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

    public static AppUser load(String userName) throws IOException, ClassNotFoundException {
        File file = new File(getUserDir(userName) + "/" + userName);
        FileInputStream inputStream = new FileInputStream(file);
        return load(inputStream);
    }

    public static boolean exist(String userName) {
        File file = new File(getUserDir(userName) + "/" + userName);
        return  file.exists();
    }

    public static KeyPair generateKeyPair(int bits) throws Exception {
        KeyPairGenerator generator = KeyPairGenerator.getInstance("RSA");
        generator.initialize(bits, new SecureRandom());
        return generator.generateKeyPair();
    }

    public static RSAPublicKey getPubKey(String userName)
            throws Exception, IOException, NoSuchAlgorithmException, InvalidKeySpecException {

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

    public static RSAPrivateKey getPrivKey(String userName)
            throws Exception, IOException, NoSuchAlgorithmException, InvalidKeySpecException {

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

    public static void SaveKey(Key key, FileOutputStream outputStream) throws IOException {
        outputStream.write(key.getEncoded());
        outputStream.close();
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

    public static BigInteger loadSignedKey(String userName) throws IOException, ClassNotFoundException {
        String filesPath = getUserKeysDir(userName) + "/" + userName + ".signed";
        File file = new File(filesPath);

        return (BigInteger) load(file);
    }

    public static void save(Serializable obj, File file) throws IOException {
        FileOutputStream outputStream = new FileOutputStream(file);
        ObjectOutputStream objectOut = new ObjectOutputStream(outputStream);
        objectOut.writeObject(obj);
        objectOut.close();
    }

    public static void save(Serializable obj, FileOutputStream fileOutputStream) throws IOException {
        ObjectOutputStream objectOut = new ObjectOutputStream(fileOutputStream);
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

//    public static RSAPublicKey loadPubKey(FileInputStream inputStream)
//            throws IOException, NoSuchAlgorithmException, InvalidKeySpecException {
//
//        byte[] bytes = ByteStreams.toByteArray(inputStream);
//
//        X509EncodedKeySpec ks = new X509EncodedKeySpec(bytes);
//        KeyFactory kf = KeyFactory.getInstance("RSA");
//        PublicKey pub = kf.generatePublic(ks);
//
//        return (RSAPublicKey) pub;
//    }

    public static RSAPrivateKey loadPrivKey(String filePath)
            throws IOException, NoSuchAlgorithmException, InvalidKeySpecException {
        Path path = Paths.get(filePath);
        byte[] bytes = Files.readAllBytes(path);

        PKCS8EncodedKeySpec ks = new PKCS8EncodedKeySpec(bytes);
        KeyFactory kf = KeyFactory.getInstance("RSA");
        PrivateKey pvt = kf.generatePrivate(ks);

        return (RSAPrivateKey) pvt;
    }

//    public static RSAPrivateKey loadPrivKey(FileInputStream inputStream)
//            throws IOException, NoSuchAlgorithmException, InvalidKeySpecException {
//
//        byte[] bytes = ByteStreams.toByteArray(inputStream);
//
//        PKCS8EncodedKeySpec ks = new PKCS8EncodedKeySpec(bytes);
//        KeyFactory kf = KeyFactory.getInstance("RSA");
//        PrivateKey pvt = kf.generatePrivate(ks);
//
//        return (RSAPrivateKey) pvt;
//    }

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
