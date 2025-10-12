#include "shared/plugin.h"
#include "include/cef_app.h"
#include <windows.h>
#include <thread>
#include <chrono>
#include <iostream>

int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE, LPSTR, int) {
    int error_code;
    PluginInstanceHandle h1 = PluginCreateInstance(
            "profileA",
            "MyUaAgent/1.0",
            "http://127.0.0.1:7891",
            "",
            "",
            &error_code
    );
    if (error_code != 0) {
        std::cerr << "Failed to create instance h1" << std::endl;
        return 1;
    }
    PluginLoadURL(h1, "file:///D:/test.html");
    // Create second browser instance
    PluginInstanceHandle h2 = PluginCreateInstance(
            "profileB",
            "MyUaAgent/2.0",
            "http://127.0.0.1:7891",
            "",
            "",
            &error_code
    );
    if (error_code != 0) {
        std::cerr << "Failed to create instance h2" << std::endl;
        PluginDestroyInstance(h1);
        return 1;
    }
    PluginLoadURL(h2, "file:///D:/test.html"); // Local HTML test

    std::thread([&h1, &h2]() {
        std::this_thread::sleep_for(std::chrono::seconds(10));
        PluginEvalJS(h1, "alert('User-Agent: ' + navigator.userAgent);");
        PluginEvalJS(h2, "alert('User-Agent: ' + navigator.userAgent);");
    }).detach();
    std::thread([&h1, &h2]() {
        std::this_thread::sleep_for(std::chrono::seconds(20));
        PluginLoadURL(h1, "https://httpbin.org/headers");
        PluginLoadURL(h2, "https://httpbin.org/headers");
    }).detach();
    std::thread([]() {
        std::this_thread::sleep_for(std::chrono::seconds(120));
        CefQuitMessageLoop();
    }).detach();
    CefRunMessageLoop();
    PluginDestroyInstance(h1);
    PluginDestroyInstance(h2);
    CefShutdown();
    return 0;
}

