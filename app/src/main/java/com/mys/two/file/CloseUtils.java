package com.mys.two.file;

import android.util.Log;

import java.io.Closeable;

public class CloseUtils {

    public static void closeSilently(Closeable closeable) {
        if (closeable != null) {
            try {
                closeable.close();
            } catch (Exception e) {
                Log.e("-----1", "e=" + e);
            }
        }
    }

    public static void closeReader(java.io.Reader reader) {
        if (reader != null) {
            try {
                reader.close();
            } catch (Exception e) {
                Log.e("-----1", "e=" + e);
            }
        }
    }

}
