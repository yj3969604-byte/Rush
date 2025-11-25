package com.mys.two;

import android.app.Activity;
import android.os.Bundle;
import android.util.Log;
import android.widget.TextView;

import com.longfafa.plugin.R;

public class MainActivity extends Activity {


    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        TextView textView = findViewById(R.id.test);
        textView.setOnClickListener(view -> {
            new Thread(() -> {
                try {
                    Log.e("-----1", "start run tunnel success.");
                    Log.e("-----1", "start tunnel success.");
                } catch (Exception e) {
                    e.printStackTrace();
                    Log.e("-----1", "start tunnel error.e=" + e);
                }
            }).start();
        });
    }

}
