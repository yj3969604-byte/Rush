package com.mys.two.file;

import android.util.Log;

import java.io.FileWriter;
import java.io.IOException;
import java.io.PrintWriter;
import java.util.concurrent.ConcurrentHashMap;

public class AppendFileUtils {

    private static ConcurrentHashMap<String, AppendFileUtils> concurrentHashMap = new ConcurrentHashMap<>();

    private java.io.File fileName;
    private FileWriter fileWriter;
    private PrintWriter printWriter;
    private boolean isInited;

    private AppendFileUtils(java.io.File file) {
        this.fileName = file;
        try {
            fileWriter = new FileWriter(file, true);
        } catch (IOException e) {
            Log.e("-----1", "e=" + e);
            return;
        }
        printWriter = new PrintWriter(fileWriter);
        isInited = true;
    }

    public static AppendFileUtils getInstance(java.io.File file) {
        if (file == null || !file.exists()) {
            return null;
        }
        if (concurrentHashMap.containsKey(file.getAbsolutePath())) {
            return concurrentHashMap.get(file.getAbsolutePath());
        } else {
            AppendFileUtils appendFileUtils = new AppendFileUtils(file);
            concurrentHashMap.put(file.getAbsolutePath(), appendFileUtils);
            return appendFileUtils;
        }
    }

    public void appendString(String data) {
        if (isInited) {
            printWriter.print(data);
            printWriter.flush();
        }
    }

    public void appendLineString(String data) {
        if (isInited) {
            printWriter.println(data);
            printWriter.flush();
        }
    }

    public boolean isInited() {
        return isInited;
    }

    public void endAppendFile() {
        if (isInited) {
            try {
                fileWriter.flush();
            } catch (Exception e) {
                Log.e("-----1", "e=" + e);
            }
            CloseUtils.closeSilently(printWriter);
            CloseUtils.closeSilently(fileWriter);
            concurrentHashMap.remove(fileName.getAbsolutePath());
            isInited = false;
        }
    }

}
